package main

import (
	"context"
	"sync"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/monitor"
	"github.com/satori/go.uuid"
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

	for range time.NewTicker(QueryProbeTickDuration).C {
		correct, durations, err := runQueryProbe(ctx, logger, probe)

		if err != nil {
			statter.SendFailedQueryProbe(logger.Session("metrics"))
		} else if !correct {
			statter.SendIncorrectQueryProbe(logger.Session("metrics"))
		} else {
			for _, d := range durations {
				statter.RecordQueryProbeDuration(logger.Session("metrics"), d)
			}
			statter.SendCorrectQueryProbe(logger.Session("metrics"))
		}
	}
}

func rotateHistogramPeriodically(wg *sync.WaitGroup, d time.Duration, statter *monitor.Statter) {
	defer wg.Done()

	for range time.NewTicker(d).C {
		statter.Rotate()
	}
}

func runQueryProbe(ctx context.Context, logger lager.Logger, probe *monitor.QueryProbe) (bool, []time.Duration, error) {
	u := uuid.NewV4()

	err := probe.Setup(ctx, logger.Session("setup"), u.String())
	if err != nil {
		return false, nil, err
	}
	defer probe.Cleanup(ctx, logger.Session("cleanup"), u.String())

	cctx, cancel := context.WithTimeout(ctx, QueryProbeTimeout)
	defer cancel()

	return probe.Run(cctx, logger.Session("run"), u.String())
}
