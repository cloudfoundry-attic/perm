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
	ProbeHistogramWindow      = 5 // Minutes
	ProbeHistogramRefreshTime = 1 * time.Minute
)

func RunProbe(ctx context.Context,
	logger lager.Logger,
	probe *monitor.Probe,
	statter *monitor.Statter,
	probeInterval, probeTimeout time.Duration,
) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for range time.NewTicker(ProbeHistogramRefreshTime).C {
			statter.Rotate()
		}
	}()

	go func() {
		defer wg.Done()

		for range time.NewTicker(probeInterval).C {
			correct, durations, err := cmd.RunProbe(ctx, logger, probe, probeTimeout)

			if err != nil {
				statter.SendFailedProbe(logger.Session("metrics"))
			} else if !correct {
				statter.SendIncorrectProbe(logger.Session("metrics"))
			} else {
				for _, d := range durations {
					statter.RecordProbeDuration(logger.Session("metrics"), d)
				}
				statter.SendCorrectProbe(logger.Session("metrics"))
			}
		}
	}()

	wg.Wait()
}
