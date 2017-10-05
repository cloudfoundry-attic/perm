package main

import (
	"net"
	"os"

	"github.com/cactus/go-statsd-client/statsd"
	flags "github.com/jessevdk/go-flags"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"code.cloudfoundry.org/lager"

	"crypto/x509"
	"io/ioutil"

	"context"

	"strconv"

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
	Hostname      string         `long:"hostname" description:"Hostname used to resolve the address of Perm" required:"true"`
	Port          int            `long:"port" description:"Port used to connect to Perm" required:"true"`
	CACertificate []cmd.FileFlag `long:"ca-certificate" description:"File path of Perm's CA certificate"`
}

type statsDOptions struct {
	Hostname string `long:"hostname" description:"Hostname used to connect to StatsD server" required:"true"`
	Port     int    `long:"port" description:"Port used to connect to StatsD server" required:"true"`
}

const (
	prefix = "localhost"

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

	logger, _ := parserOpts.Logger.Logger("perm")

	logger.Debug("starting")
	defer logger.Debug("finishing")

	statsDAddr := net.JoinHostPort(parserOpts.StatsD.Hostname, strconv.Itoa(parserOpts.StatsD.Port))
	s, err := statsd.NewClient(statsDAddr, prefix)
	if err != nil {
		logger.Fatal(messages.FailedToConnectToStatsD, err, lager.Data{
			"addr": statsDAddr,
		})
		os.Exit(1)
	}
	defer s.Close()

	pool := x509.NewCertPool()

	for _, certPath := range parserOpts.Perm.CACertificate {
		caPem, err := ioutil.ReadFile(certPath.Path())
		if err != nil {
			logger.Fatal(messages.FailedToReadCertificate, err, lager.Data{
				"location": certPath,
			})
			os.Exit(1)
		}

		if ok := pool.AppendCertsFromPEM(caPem); !ok {
			logger.Fatal(messages.FailedToAppendCertToPool, err, lager.Data{
				"location": certPath,
			})
			os.Exit(1)
		}
	}

	addr := net.JoinHostPort(parserOpts.Perm.Hostname, strconv.Itoa(parserOpts.Perm.Port))
	creds := credentials.NewClientTLSFromCert(pool, parserOpts.Perm.Hostname)

	g, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		logger.Fatal(messages.FailedToGRPCDial, err, lager.Data{
			"addr": addr,
		})
		os.Exit(1)
	}
	defer g.Close()

	p := protos.NewRoleServiceClient(g)

	adminProbe := &monitor.AdminProbe{
		RoleServiceClient: p,
	}

	ctx := context.Background()
	adminProbeLogger := logger.Session("admin-probe")

	adminProbe.Cleanup(ctx, adminProbeLogger.Session("cleanup"))
	err = adminProbe.Run(ctx, adminProbeLogger.Session("run"))

	if e := s.Inc(MetricAdminProbeRunsTotal, 1, AlwaysSendMetric); e != nil {
		adminProbeLogger.Session("metrics").Error(messages.FailedToSendMetric, err, lager.Data{
			"metric": MetricAdminProbeRunsTotal,
		})
	}
	if err != nil {
		if e := s.Inc(MetricAdminProbeRunsFailed, 1, AlwaysSendMetric); e != nil {
			adminProbeLogger.Session("metrics").Error(messages.FailedToSendMetric, err, lager.Data{
				"metric": MetricAdminProbeRunsFailed,
			})
		}
	}

	os.Exit(0)
}
