package main

import (
	"context"
	"sync"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/monitor"
	"github.com/cactus/go-statsd-client/statsd"
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
	defer wg.Done()

	histogram := monitor.NewHistogram(QueryProbeHistogramWindow, QueryProbeMinResponseTime, QueryProbeMaxResponseTime, 3)

	var innerWG sync.WaitGroup
	innerWG.Add(2)
	go rotateHistogramPeriodically(&innerWG, QueryProbeHistogramRefreshTime, histogram)
	go runProbePeriodically(ctx, logger, &innerWG, probe, statter, histogram)

	innerWG.Wait()
}

func runProbePeriodically(ctx context.Context, logger lager.Logger, wg *sync.WaitGroup, probe *monitor.QueryProbe, statter statsd.Statter, histogram *monitor.Histogram) {
	defer wg.Done()

	var (
		correct   bool
		err       error
		durations []time.Duration
	)

	metricsLogger := logger.Session("metrics")
	setupLogger := logger.Session("setup")
	cleanupLogger := logger.Session("cleanup")
	runLogger := logger.Session("run")

	err = probe.Setup(ctx, setupLogger)
	defer probe.Cleanup(ctx, cleanupLogger)

	ticker := time.NewTicker(QueryProbeTickDuration)
	for range ticker.C {
		func(logger lager.Logger) {
			cctx, cancel := context.WithTimeout(ctx, QueryProbeTimeout)
			defer cancel()

			correct, durations, err = probe.Run(cctx, runLogger)

			if err != nil {
				sendGauge(logger, statter, MetricQueryProbeRunsSuccess, MetricFailure)
			} else if !correct {
				sendGauge(logger, statter, MetricQueryProbeRunsSuccess, MetricFailure)
				sendGauge(logger, statter, MetricQueryProbeRunsCorrect, MetricFailure)
			} else {
				sendGauge(logger, statter, MetricQueryProbeRunsSuccess, MetricSuccess)
				sendGauge(logger, statter, MetricQueryProbeRunsCorrect, MetricSuccess)

				for _, d := range durations {
					recordHistogramDuration(logger, histogram, d)
				}

				sendHistogramQuantile(logger, statter, histogram, 90, MetricQueryProbeTimingP90)
				sendHistogramQuantile(logger, statter, histogram, 99, MetricQueryProbeTimingP99)
				sendHistogramQuantile(logger, statter, histogram, 99.9, MetricQueryProbeTimingP999)
				sendHistogramMax(logger, statter, histogram, MetricQueryProbeTimingMax)
			}
		}(metricsLogger)
	}
}

func rotateHistogramPeriodically(wg *sync.WaitGroup, d time.Duration, histogram *monitor.Histogram) {
	defer wg.Done()

	for range time.NewTicker(d).C {
		histogram.Rotate()
	}
}
