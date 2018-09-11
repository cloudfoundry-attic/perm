package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"code.cloudfoundry.org/perm"
	cmdflags "code.cloudfoundry.org/perm/cmd/flags"
	"code.cloudfoundry.org/perm/cmd/internal/ioutilx"
	"code.cloudfoundry.org/perm/logx"
	"code.cloudfoundry.org/perm/metrics"
	"code.cloudfoundry.org/perm/monitor"
	"code.cloudfoundry.org/perm/monitor/recording"
	"github.com/cactus/go-statsd-client/statsd"
	flags "github.com/jessevdk/go-flags"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc/resolver"
)

type options struct {
	Perm permOptions `group:"Perm" namespace:"perm"`

	StatsD statsDOptions `group:"StatsD" namespace:"statsd"`

	Probe probeOptions `group:"Probe" namespace:"probe"`

	Logger cmdflags.LagerFlag
}

type permOptions struct {
	Host         string                 `long:"host" description:"Hostname used to resolve the address of Perm" default:"localhost"`
	Port         int                    `long:"port" description:"Port used to connect to Perm" default:"6283"`
	TLSCA        []ioutilx.FileOrString `long:"tls-ca" description:"File path of Perm's CA certificate (and the OAuth2 server's CA if --require-auth)"`
	RequireAuth  bool                   `long:"require-auth" description:"Enable the monitor to talk to perm using oauth"`
	TokenURL     string                 `long:"token-url" description:"URL to the OAuth2 server's token endpoint (only required if '--require-auth' is provided)"`
	ClientID     string                 `long:"client-id" description:"OAuth2 Client ID used to fetch token (only required if '--require-auth' is provided)"`
	ClientSecret string                 `long:"client-secret" description:"OAuth2 Client Secret used to fetch token (only required if '--require-auth' is provided)"`
}

type statsDOptions struct {
	Host string `long:"host" description:"Hostname used to connect to StatsD server" default:"localhost"`
	Port int    `long:"port" description:"Port used to connect to StatsD server" default:"8125"`
}

type probeOptions struct {
	Frequency      time.Duration `long:"frequency" description:"The amount of time between probe runs" default:"5s"`
	Timeout        time.Duration `long:"timeout" description:"The amount of time for each API call to complete; if exceeded, the probe will error its current run" default:"1s"`
	CleanupTimeout time.Duration `long:"cleanup-timeout" description:"If a probe run errors, this is the max allowed time for cleanup" default:"10s"`
	MaxLatency     time.Duration `long:"max-latency" description:"If any API call in the current probe run exceeds this value, a latency KPI failure will be recorded" default:"100ms"`
}

func main() {
	resolver.SetDefaultScheme("dns")
	parserOpts := &options{}
	parser := flags.NewParser(parserOpts, flags.Default)
	parser.NamespaceDelimiter = "-"

	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}

	logger := parserOpts.Logger.Logger("perm-monitor")

	//////////////////////
	// Setup StatsD Client
	//////////////////////
	statsDAddr := net.JoinHostPort(parserOpts.StatsD.Host, strconv.Itoa(parserOpts.StatsD.Port))
	statsDClient, err := statsd.NewBufferedClient(statsDAddr, "", 0, 0)
	if err != nil {
		logger.Error(failedToConnectToStatsD, err, logx.Data{
			Key:   "addr",
			Value: statsDAddr,
		})
		os.Exit(1)
	}
	defer statsDClient.Close()

	//////////////////////
	// Setup Perm GRPC Client
	//////////////////////
	pool := x509.NewCertPool()

	for _, certPath := range parserOpts.Perm.TLSCA {
		caPem, e := certPath.Bytes(ioutilx.InjectableOS{}, ioutilx.InjectableIOReader{})
		certLogger := logger.WithData(logx.Data{Key: "location", Value: certPath})
		if e != nil {
			certLogger.Error(failedToReadCertificate, e)
			os.Exit(1)
		}

		if ok := pool.AppendCertsFromPEM(caPem); !ok {
			certLogger.Error(failedToAppendCertToPool, e)
			os.Exit(1)
		}
	}

	addr := net.JoinHostPort(parserOpts.Perm.Host, strconv.Itoa(parserOpts.Perm.Port))
	opts := []perm.DialOption{perm.WithTLSConfig(&tls.Config{RootCAs: pool})}

	if parserOpts.Perm.RequireAuth {
		tsConfig := clientcredentials.Config{
			ClientID:     parserOpts.Perm.ClientID,
			ClientSecret: parserOpts.Perm.ClientSecret,
			TokenURL:     parserOpts.Perm.TokenURL,
		}

		oauth2Client := http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: pool,
				},
			},
		}
		ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &oauth2Client)

		tokenSource := tsConfig.TokenSource(ctx)

		opts = append(opts, perm.WithTokenSource(tokenSource))
	}

	client, err := perm.Dial(addr, opts...)
	if err != nil {
		logger.Error(failedToCreatePermClient, err)
		os.Exit(1)
	}
	defer client.Close()

	//////////////////////
	// Setup Probe
	//////////////////////
	histogram := metrics.NewHistogram(metrics.HistogramOptions{
		Name:        "perm.probe.responses.timing",
		Buckets:     []float64{50, 90, 95, 99, 99.9},
		MaxDuration: parserOpts.Probe.Timeout,
	})

	crHistogram := metrics.NewHistogram(metrics.HistogramOptions{
		Name:        "perm.probe.createrole.responses.timing",
		Buckets:     []float64{50, 90, 95, 99, 99.9},
		MaxDuration: parserOpts.Probe.Timeout,
	})

	histograms := map[string]recording.Recorder{
		"all":        histogram,
		"CreateRole": crHistogram,
	}

	recordingClient := recording.NewClient(client, histograms)

	probeOptions := []monitor.Option{
		monitor.WithTimeout(parserOpts.Probe.Timeout),
		monitor.WithCleanupTimeout(parserOpts.Probe.CleanupTimeout),
		monitor.WithMaxLatency(parserOpts.Probe.MaxLatency),
	}
	probe := monitor.NewProbe(recordingClient, statsDClient, logger, probeOptions...)

	logger.Info(starting)
	defer logger.Debug(finished)

	for range time.NewTicker(parserOpts.Probe.Frequency).C {
		probe.Run()
		for metric, duration := range histogram.Collect() {
			statsDClient.Gauge(metric, duration, 1)
		}
		for metric, duration := range crHistogram.Collect() {
			statsDClient.Gauge(metric, duration, 1)
		}
	}
}
