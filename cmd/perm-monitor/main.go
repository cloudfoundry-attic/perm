package main // import "code.cloudfoundry.org/perm/cmd/perm-monitor"

import (
	"net"
	"os"

	"github.com/cactus/go-statsd-client/statsd"
	flags "github.com/jessevdk/go-flags"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"code.cloudfoundry.org/lager"

	"crypto/x509"

	"context"

	"strconv"

	"sync"

	"code.cloudfoundry.org/perm/cmd"
	"code.cloudfoundry.org/perm/messages"
	"code.cloudfoundry.org/perm/monitor"
	"code.cloudfoundry.org/perm/protos"
)

type options struct {
	Perm permOptions `group:"Perm" namespace:"perm"`

	StatsD statsDOptions `group:"StatsD" namespace:"statsd"`

	Logger cmd.LagerFlag
}

type permOptions struct {
	Hostname      string                 `long:"hostname" description:"Hostname used to resolve the address of Perm" required:"true"`
	Port          int                    `long:"port" description:"Port used to connect to Perm" required:"true"`
	CACertificate []cmd.FileOrStringFlag `long:"ca-certificate" description:"File path of Perm's CA certificate"`
}

type statsDOptions struct {
	Hostname string `long:"hostname" description:"Hostname used to connect to StatsD server" required:"true"`
	Port     int    `long:"port" description:"Port used to connect to StatsD server" required:"true"`
}

func main() {
	parserOpts := &options{}
	parser := flags.NewParser(parserOpts, flags.Default)
	parser.NamespaceDelimiter = "-"

	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}

	logger, _ := parserOpts.Logger.Logger("perm-monitor")

	logger.Debug(messages.Starting)
	defer logger.Debug(messages.Finished)

	//////////////////////
	// Setup StatsD Client
	statsDAddr := net.JoinHostPort(parserOpts.StatsD.Hostname, strconv.Itoa(parserOpts.StatsD.Port))
	statsDClient, err := statsd.NewBufferedClient(statsDAddr, "", 0, 0)
	if err != nil {
		logger.Fatal(messages.FailedToConnectToStatsD, err, lager.Data{
			"addr": statsDAddr,
		})
		os.Exit(1)
	}
	defer statsDClient.Close()
	//////////////////////

	//////////////////////
	// Setup Perm GRPC Client
	//
	//// Setup TLS Credentials
	pool := x509.NewCertPool()

	for _, certPath := range parserOpts.Perm.CACertificate {
		caPem, e := certPath.Bytes(cmd.InjectableOS{}, cmd.InjectableIOReader{})
		if e != nil {
			logger.Fatal(messages.FailedToReadCertificate, e, lager.Data{
				"location": certPath,
			})
			os.Exit(1)
		}

		if ok := pool.AppendCertsFromPEM(caPem); !ok {
			logger.Fatal(messages.FailedToAppendCertToPool, e, lager.Data{
				"location": certPath,
			})
			os.Exit(1)
		}
	}

	addr := net.JoinHostPort(parserOpts.Perm.Hostname, strconv.Itoa(parserOpts.Perm.Port))
	creds := credentials.NewClientTLSFromCert(pool, parserOpts.Perm.Hostname)

	//// Setup GRPC connection
	g, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		logger.Fatal(messages.FailedToGRPCDial, err, lager.Data{
			"addr": addr,
		})
		os.Exit(1)
	}
	defer g.Close()

	roleServiceClient := protos.NewRoleServiceClient(g)
	permissionServiceClient := protos.NewPermissionServiceClient(g)
	//////////////////////

	ctx := context.Background()

	adminProbe := &monitor.AdminProbe{
		RoleServiceClient: roleServiceClient,
	}

	queryProbe := &monitor.QueryProbe{
		RoleServiceClient:       roleServiceClient,
		PermissionServiceClient: permissionServiceClient,
	}

	queryProbeHistogram := monitor.NewThreadSafeHistogram(
		QueryProbeHistogramWindow,
		QueryProbeMinResponseTime,
		QueryProbeMaxResponseTime,
		3,
	)
	statter := &monitor.Statter{
		StatsD:    statsDClient,
		Histogram: queryProbeHistogram,
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go RunAdminProbe(ctx, logger.Session("admin-probe"), &wg, adminProbe, statter)
	go RunQueryProbe(ctx, logger.Session("query-probe"), &wg, queryProbe, statter)

	wg.Wait()
	os.Exit(0)
}
