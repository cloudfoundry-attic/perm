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
	AcceptableQueryWindow     = 100 * time.Millisecond
)

func RunProbeWithFrequency(ctx context.Context,
	logger lager.Logger,
	probe *monitor.Probe,
	statter monitor.PermStatter,
	probeFrequency, probeTimeout time.Duration,
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

		for range time.NewTicker(probeFrequency).C {
			cmd.RecordProbeResults(ctx, logger, probe, probeTimeout, statter, AcceptableQueryWindow)
		}
	}()

	wg.Wait()
}
