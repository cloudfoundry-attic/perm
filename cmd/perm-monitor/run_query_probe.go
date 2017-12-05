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

	MetricQueryProbeRunsSuccess = "perm.probe.query.runs.success"
	MetricQueryProbeRunsCorrect = "perm.probe.query.runs.correct"

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
	var rw = &sync.RWMutex{}

	wg.Add(1)
	go rotateHistogramPeriodically(wg, rw, QueryProbeHistogramRefreshTime, histogram)

	err = probe.Setup(ctx, setupLogger)
	defer probe.Cleanup(ctx, cleanupLogger)

	ticker := time.NewTicker(QueryProbeTickDuration)

	for range ticker.C {
		func(logger lager.Logger) {
			cctx, cancel := context.WithTimeout(ctx, QueryProbeTimeout)
			defer cancel()

			correct, durations, err = probe.Run(cctx, runLogger)

			if err != nil {
				sendGauge(logger, statter, MetricQueryProbeRunsSuccess, 0.0)
			} else if !correct {
				sendGauge(logger, statter, MetricQueryProbeRunsSuccess, 0.0)
				sendGauge(logger, statter, MetricQueryProbeRunsCorrect, 0.0)
			} else {
				sendGauge(logger, statter, MetricQueryProbeRunsSuccess, 1.0)
				sendGauge(logger, statter, MetricQueryProbeRunsCorrect, 1.0)

				for _, d := range durations {
					recordHistogramDuration(logger, rw.RLocker(), histogram, d)
				}

				sendHistogramQuantile(logger, statter, rw.RLocker(), histogram, 90, MetricQueryProbeTimingP90)
				sendHistogramQuantile(logger, statter, rw.RLocker(), histogram, 99, MetricQueryProbeTimingP99)
				sendHistogramQuantile(logger, statter, rw.RLocker(), histogram, 99.9, MetricQueryProbeTimingP999)
				sendHistogramMax(logger, statter, rw.RLocker(), histogram, MetricQueryProbeTimingMax)
			}
		}(metricsLogger)
	}

	wg.Done()
}

func rotateHistogramPeriodically(wg *sync.WaitGroup, locker sync.Locker, d time.Duration, histogram *hdrhistogram.WindowedHistogram) {
	for range time.NewTicker(d).C {
		func() {
			locker.Lock()
			defer locker.Unlock()

			histogram.Rotate()
		}()
	}

	wg.Done()
}
