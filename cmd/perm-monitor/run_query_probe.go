package main

import (
	"context"
	"sync"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/cmd"
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

func RunQueryProbe(ctx context.Context,
	logger lager.Logger,
	wg *sync.WaitGroup,
	probe *monitor.QueryProbe,
	statter *monitor.Statter,
) {
	defer wg.Done()

	var innerWG sync.WaitGroup
	innerWG.Add(2)
	go func() {
		defer innerWG.Done()

		for range time.NewTicker(QueryProbeHistogramRefreshTime).C {
			statter.Rotate()
		}
	}()

	go func() {
		defer innerWG.Done()

		for range time.NewTicker(QueryProbeTickDuration).C {
			correct, durations, err := cmd.RunQueryProbe(ctx, logger, probe, QueryProbeTimeout)

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
	}()

	innerWG.Wait()
}
