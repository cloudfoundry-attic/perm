package main

import (
	"time"

	"code.cloudfoundry.org/perm/pkg/logx"
	"code.cloudfoundry.org/perm/pkg/monitor"
	"code.cloudfoundry.org/perm/pkg/monitor/recording"
	"code.cloudfoundry.org/perm/pkg/monitor/stats"
	"github.com/cactus/go-statsd-client/statsd"
)

func Probe(statsDClient statsd.Statter, histogram *stats.Histogram, probe *monitor.Probe, frequency time.Duration, logger logx.Logger) {
	probeLogger := logger.WithName("probe")
	probeLogger.Debug(starting)
	defer probeLogger.Debug(finished)

	ticker := time.NewTicker(frequency)

	for range ticker.C {
		err := probe.Run()
		if err == nil {
			probeLogger.Debug(success)
			sendGauge(probeCorrect, 1, statsDClient, probeLogger, err)
			sendGauge(probeSuccess, 1, statsDClient, probeLogger, err)

			for metric, value := range histogram.Collect() {
				sendGauge(metric, value, statsDClient, probeLogger, err)
			}

			continue
		}

		// Errors:
		// - HasAssignedPermissionError (record incorrect metric, record failure metric)
		// - HasUnassignedPermissionError (record incorrect metric, record failure metric)
		// - ExceededMaxLatencyError (record correct metric, record failure metric)
		// - error from API call
		//   - FailedToObserveDurationError (no duration recorded, record error count)
		//   - API error (timeout/network/etc., no duration recorded, record error count)
		switch err.(type) {
		case monitor.HasAssignedPermissionError:
			probeLogger.Error(failed, err)
			sendGauge(probeCorrect, 0, statsDClient, probeLogger, err)
			sendGauge(probeSuccess, 0, statsDClient, probeLogger, err)
		case monitor.HasUnassignedPermissionError:
			probeLogger.Error(failed, err)
			sendGauge(probeCorrect, 0, statsDClient, probeLogger, err)
			sendGauge(probeSuccess, 0, statsDClient, probeLogger, err)
		case monitor.ExceededMaxLatencyError:
			probeLogger.Error(failed, err)
			sendGauge(probeCorrect, 1, statsDClient, probeLogger, err)
			sendGauge(probeSuccess, 0, statsDClient, probeLogger, err)
		case recording.FailedToObserveDurationError:
			probeLogger.Error(failed, err)
			sendInc(probeFailedToObserve, statsDClient, probeLogger, err)
		default:
			probeLogger.Error(failed, err)
			sendInc(probeAPIErrored, statsDClient, probeLogger, err)
		}
	}
}

func sendGauge(metric string, value int64, statsDClient statsd.Statter, logger logx.Logger, probeErr error) {
	if err := statsDClient.Gauge(metric, value, alwaysSend); err != nil {
		logger.Error(failedToSendMetric, probeErr)
	}
}

func sendInc(metric string, statsDClient statsd.Statter, logger logx.Logger, probeErr error) {
	if err := statsDClient.Inc(metric, 1, alwaysSend); err != nil {
		logger.Error(failedToSendMetric, probeErr)
	}
}
