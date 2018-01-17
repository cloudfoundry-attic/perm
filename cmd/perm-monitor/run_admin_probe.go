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
	AdminProbeTickDuration = 30 * time.Second
	AdminProbeTimeout      = 3 * time.Second
)

func RunAdminProbe(
	ctx context.Context,
	logger lager.Logger,
	wg *sync.WaitGroup,
	probe *monitor.AdminProbe,
	statter *monitor.Statter,
) {
	defer wg.Done()

	var err error

	for range time.NewTicker(AdminProbeTickDuration).C {
		err = cmd.RunAdminProbe(ctx, logger, probe, AdminProbeTimeout)
		if err != nil {
			statter.SendFailedAdminProbe(logger.Session("metrics"))
		} else {
			statter.SendSuccessfulAdminProbe(logger.Session("metrics"))
		}
	}
}
