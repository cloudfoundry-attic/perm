package main

import (
	"context"
	"sync"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/cmd"
	"code.cloudfoundry.org/perm/pkg/monitor"
)

const (
	QueryProbeHistogramWindow      = 5 // Minutes
	QueryProbeHistogramRefreshTime = 1 * time.Minute
)

func RunQueryProbe(ctx context.Context,
	logger lager.Logger,
	probe *monitor.QueryProbe,
	statter *monitor.Statter,
	probeInterval, probeTimeout time.Duration,
) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for range time.NewTicker(QueryProbeHistogramRefreshTime).C {
			statter.Rotate()
		}
	}()

	go func() {
		defer wg.Done()

		for range time.NewTicker(probeInterval).C {
			correct, durations, err := cmd.RunQueryProbe(ctx, logger, probe, probeTimeout)

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

	wg.Wait()
}
