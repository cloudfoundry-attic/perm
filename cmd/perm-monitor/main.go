package main // import "code.cloudfoundry.org/perm/cmd/perm-monitor"

import (
	"net"
	"os"

	"github.com/cactus/go-statsd-client/statsd"
	"github.com/codahale/hdrhistogram"
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

	AdminProbeTickDuration = 30 * time.Second
	AdminProbeTimeout      = 3 * time.Second

	QueryProbeTickDuration = 1 * time.Second
	QueryProbeTimeout      = 1 * time.Second

	MetricAdminProbeRunsTotal  = "perm.probe.admin.runs.total"
	MetricAdminProbeRunsFailed = "perm.probe.admin.runs.failed"

	MetricQueryProbeRunsTotal     = "perm.probe.query.runs.total"
	MetricQueryProbeRunsFailed    = "perm.probe.query.runs.failed"
	MetricQueryProbeRunsIncorrect = "perm.probe.query.runs.incorrect"

	MetricQueryProbeTimingMax  = "perm.probe.query.responses.timing.max"  // gauge
	MetricQueryProbeTimingP90  = "perm.probe.query.responses.timing.p90"  // gauge
	MetricQueryProbeTimingP99  = "perm.probe.query.responses.timing.p99"  // gauge
	MetricQueryProbeTimingP999 = "perm.probe.query.responses.timing.p999" // gauge
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

	var wg sync.WaitGroup
	wg.Add(2)

	go RunAdminProbe(ctx, logger.Session("admin-probe"), &wg, adminProbe, statsDClient)
	go RunQueryProbe(ctx, logger.Session("query-probe"), &wg, queryProbe, statsDClient)

	wg.Wait()
	os.Exit(0)
}

func RunAdminProbe(ctx context.Context, logger lager.Logger, wg *sync.WaitGroup, probe *monitor.AdminProbe, statter statsd.Statter) {
	var err error

	metricsLogger := logger.Session("metrics")
	cleanupLogger := logger.Session("cleanup")
	runLogger := logger.Session("run")

	ticker := time.NewTicker(AdminProbeTickDuration)

	for range ticker.C {
		func() {
			err = probe.Cleanup(ctx, cleanupLogger)
			if err != nil {
				return
			}

			cctx, cancel := context.WithTimeout(ctx, AdminProbeTimeout)
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

func RunQueryProbe(ctx context.Context, logger lager.Logger, wg *sync.WaitGroup, probe *monitor.QueryProbe, statter statsd.Statter) {
	var (
		correct   bool
		err       error
		durations []time.Duration
	)

	metricsLogger := logger.Session("metrics")
	setupLogger := logger.Session("setup")
	cleanupLogger := logger.Session("cleanup")
	runLogger := logger.Session("run")

	minResponseTime := 1 * time.Nanosecond
	maxResponseTime := 1 * time.Second
	histogram := hdrhistogram.NewWindowed(5, int64(minResponseTime), int64(maxResponseTime), 3)
	var rw sync.RWMutex

	wg.Add(1)
	go func() {
		for range time.NewTicker(1 * time.Minute).C {
			func() {
				rw.Lock()
				defer rw.Unlock()

				histogram.Rotate()
			}()

		}

		wg.Done()
	}()

	err = probe.Setup(ctx, setupLogger)
	defer probe.Cleanup(ctx, cleanupLogger)

	ticker := time.NewTicker(QueryProbeTickDuration)

	for range ticker.C {
		func() {
			cctx, cancel := context.WithTimeout(ctx, QueryProbeTimeout)
			defer cancel()

			correct, durations, err = probe.Run(cctx, runLogger)

			incrementStat(metricsLogger, statter, MetricQueryProbeRunsTotal)

			if err != nil {
				incrementStat(metricsLogger, statter, MetricQueryProbeRunsFailed)
			} else {
				if !correct {
					incrementStat(logger, statter, MetricQueryProbeRunsIncorrect)
				}

				for _, d := range durations {
					histogram.Current.RecordValue(int64(d))
				}

				rw.RLock()
				defer rw.RUnlock()

				p90 := histogram.Current.ValueAtQuantile(90)
				p99 := histogram.Current.ValueAtQuantile(99)
				p999 := histogram.Current.ValueAtQuantile(99.9)
				max := histogram.Current.Max()

				sendGauge(logger, statter, MetricQueryProbeTimingP90, p90)
				sendGauge(logger, statter, MetricQueryProbeTimingP99, p99)
				sendGauge(logger, statter, MetricQueryProbeTimingP999, p999)
				sendGauge(logger, statter, MetricQueryProbeTimingMax, max)
			}
		}()
	}

	wg.Done()
}

func incrementStat(logger lager.Logger, statter statsd.Statter, name string) {
	err := statter.Inc(name, 1, AlwaysSendMetric)
	if err != nil {
		logger.Error(messages.FailedToSendMetric, err, lager.Data{
			"metric": name,
		})
	}
}

func sendGauge(logger lager.Logger, statter statsd.Statter, name string, value int64) {
	err := statter.Gauge(name, value, AlwaysSendMetric)
	if err != nil {
		logger.Error(messages.FailedToSendMetric, err, lager.Data{
			"metric": name,
		})
	}
}
