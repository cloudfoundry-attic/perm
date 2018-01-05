package main

import (
	"context"
	"sync"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/monitor"
)

const (
	AdminProbeTickDuration = 30 * time.Second
	AdminProbeTimeout      = 3 * time.Second
)

func RunAdminProbe(ctx context.Context, logger lager.Logger, wg *sync.WaitGroup, probe *monitor.AdminProbe, statter *Statter) {
	defer wg.Done()

	var err error

	metricsLogger := logger.Session("metrics")
	cleanupLogger := logger.Session("cleanup")
	runLogger := logger.Session("run")

	ticker := time.NewTicker(AdminProbeTickDuration)

	for range ticker.C {
		func() {
			err = probe.Cleanup(ctx, cleanupLogger)
			if err != nil {
				statter.SendFailedAdminProbe(metricsLogger)
				return
			}

			cctx, cancel := context.WithTimeout(ctx, AdminProbeTimeout)
			defer cancel()

			err = probe.Run(cctx, runLogger)

			if err == nil {
				statter.SendFailedAdminProbe(metricsLogger)
			} else {
				statter.SendSuccessfulAdminProbe(metricsLogger)
			}
		}()
	}
}
