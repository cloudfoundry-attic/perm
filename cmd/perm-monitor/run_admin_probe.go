package main

import (
	"context"
	"sync"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/cmd"
	"code.cloudfoundry.org/perm/monitor"
)

func RunAdminProbe(
	ctx context.Context,
	logger lager.Logger,
	wg *sync.WaitGroup,
	probe *monitor.AdminProbe,
	statter *monitor.Statter,
	probeInterval, probeTimeout time.Duration,
) {
	defer wg.Done()

	var err error

	for range time.NewTicker(probeInterval).C {
		err = cmd.RunAdminProbe(ctx, logger, probe, probeTimeout)
		if err != nil {
			statter.SendFailedAdminProbe(logger.Session("metrics"))
		} else {
			statter.SendSuccessfulAdminProbe(logger.Session("metrics"))
		}
	}
}
