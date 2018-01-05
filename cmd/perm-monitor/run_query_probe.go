package main

import (
	"context"
	"sync"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/monitor"
)

const (
	QueryProbeTickDuration = 1 * time.Second
	QueryProbeTimeout      = QueryProbeTickDuration

	QueryProbeMinResponseTime = 1 * time.Nanosecond
	QueryProbeMaxResponseTime = QueryProbeTimeout

	QueryProbeHistogramWindow      = 5 // Minutes
	QueryProbeHistogramRefreshTime = 1 * time.Minute
)

func RunQueryProbe(ctx context.Context, logger lager.Logger, wg *sync.WaitGroup, probe *monitor.QueryProbe, statter *monitor.Statter) {
	defer wg.Done()

	var innerWG sync.WaitGroup
	innerWG.Add(2)
	go rotateHistogramPeriodically(&innerWG, QueryProbeHistogramRefreshTime, statter)
	go runProbePeriodically(ctx, logger, &innerWG, probe, statter)

	innerWG.Wait()
}

func runProbePeriodically(ctx context.Context, logger lager.Logger, wg *sync.WaitGroup, probe *monitor.QueryProbe, statter *monitor.Statter) {
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
				statter.SendFailedQueryProbe(metricsLogger)
			} else if !correct {
				statter.SendIncorrectQueryProbe(logger)
			} else {
				for _, d := range durations {
					statter.RecordQueryProbeDuration(logger, d)
				}
				statter.SendCorrectQueryProbe(logger)
			}
		}(metricsLogger)
	}
}

func rotateHistogramPeriodically(wg *sync.WaitGroup, d time.Duration, statter *monitor.Statter) {
	defer wg.Done()

	for range time.NewTicker(d).C {
		statter.Rotate()
	}
}
