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

	"time"

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

const (
	AlwaysSendMetric = 1.0

	MetricAdminProbeRunsTotal  = "perm.probe.admin.runs.total"
	MetricAdminProbeRunsFailed = "perm.probe.admin.runs.failed"
)

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

	p := protos.NewRoleServiceClient(g)
	//////////////////////

	ctx := context.Background()

	adminProbe := &monitor.AdminProbe{
		RoleServiceClient: p,
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go RunAdminProbe(ctx, logger.Session("admin-probe"), &wg, adminProbe, statsDClient)

	wg.Wait()
	os.Exit(0)
}

func RunAdminProbe(ctx context.Context, logger lager.Logger, wg *sync.WaitGroup, probe *monitor.AdminProbe, statter statsd.Statter) {
	var err error

	metricsLogger := logger.Session("metrics")
	cleanupLogger := logger.Session("cleanup")
	runLogger := logger.Session("run")

	ticker := time.NewTicker(30 * time.Second)

	for range ticker.C {
		func() {
			err = probe.Cleanup(ctx, cleanupLogger)
			if err != nil {
				return
			}

			cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			err = probe.Run(cctx, runLogger)

			if e := statter.Inc(MetricAdminProbeRunsTotal, 1, AlwaysSendMetric); e != nil {
				metricsLogger.Error(messages.FailedToSendMetric, err, lager.Data{
					"metric": MetricAdminProbeRunsTotal,
				})
			}
			if err != nil {
				if e := statter.Inc(MetricAdminProbeRunsFailed, 1, AlwaysSendMetric); e != nil {
					metricsLogger.Error(messages.FailedToSendMetric, err, lager.Data{
						"metric": MetricAdminProbeRunsFailed,
					})
				}
			}
		}()
	}

	wg.Done()
}
