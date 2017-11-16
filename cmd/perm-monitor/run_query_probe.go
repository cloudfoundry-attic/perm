package main

import (
	"context"
	"sync"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/monitor"
	"github.com/cactus/go-statsd-client/statsd"
	"github.com/codahale/hdrhistogram"
)

const (
	QueryProbeTickDuration = 1 * time.Second
	QueryProbeTimeout      = QueryProbeTickDuration

	QueryProbeMinResponseTime = 1 * time.Nanosecond
	QueryProbeMaxResponseTime = QueryProbeTimeout

	QueryProbeHistogramWindow      = 5 // Minutes
	QueryProbeHistogramRefreshTime = 1 * time.Minute

	MetricQueryProbeRunsTotal     = "perm.probe.query.runs.total"
	MetricQueryProbeRunsFailed    = "perm.probe.query.runs.failed"
	MetricQueryProbeRunsIncorrect = "perm.probe.query.runs.incorrect"

	MetricQueryProbeTimingMax  = "perm.probe.query.responses.timing.max"  // gauge
	MetricQueryProbeTimingP90  = "perm.probe.query.responses.timing.p90"  // gauge
	MetricQueryProbeTimingP99  = "perm.probe.query.responses.timing.p99"  // gauge
	MetricQueryProbeTimingP999 = "perm.probe.query.responses.timing.p999" // gauge
)

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

	histogram := hdrhistogram.NewWindowed(QueryProbeHistogramWindow, int64(QueryProbeMinResponseTime), int64(QueryProbeMaxResponseTime), 3)
	var rw sync.RWMutex

	wg.Add(1)
	go rotateHistogramPeriodically(wg, rw, QueryProbeHistogramRefreshTime, histogram)

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

func rotateHistogramPeriodically(wg *sync.WaitGroup, rw sync.RWMutex, d time.Duration, histogram *hdrhistogram.WindowedHistogram) {
	for range time.NewTicker(d).C {
		func() {
			rw.Lock()
			defer rw.Unlock()

			histogram.Rotate()
		}()

	}

	wg.Done()
}
