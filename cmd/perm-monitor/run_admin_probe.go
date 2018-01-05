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
	AdminProbeTickDuration = 30 * time.Second
	AdminProbeTimeout      = 3 * time.Second

	MetricAdminProbeRunsSuccess = "perm.probe.admin.runs.success"
)

func RunAdminProbe(ctx context.Context, logger lager.Logger, wg *sync.WaitGroup, probe *monitor.AdminProbe, statter statsd.Statter) {
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
				sendGauge(metricsLogger, statter, MetricAdminProbeRunsSuccess, MetricFailure)
				return
			}

			cctx, cancel := context.WithTimeout(ctx, AdminProbeTimeout)
			defer cancel()

			err = probe.Run(cctx, runLogger)

			if err == nil {
				sendGauge(metricsLogger, statter, MetricAdminProbeRunsSuccess, MetricSuccess)
			} else {
				sendGauge(metricsLogger, statter, MetricAdminProbeRunsSuccess, MetricFailure)
			}
		}()
	}
}
