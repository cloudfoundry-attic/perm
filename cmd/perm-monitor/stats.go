package main

import (
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
	"code.cloudfoundry.org/perm/monitor"
	"github.com/cactus/go-statsd-client/statsd"
)

const (
	MetricFailure = 0.0
	MetricSuccess = 1.0

	AlwaysSendMetric = 1.0
)

func sendGauge(logger lager.Logger, statter statsd.Statter, name string, value int64) {
	err := statter.Gauge(name, value, AlwaysSendMetric)
	if err != nil {
		logger.Error(messages.FailedToSendMetric, err, lager.Data{
			"metric": name,
		})
	}
}

func recordHistogramDuration(logger lager.Logger, histogram *monitor.Histogram, d time.Duration) {
	err := histogram.RecordValue(int64(d))
	if err != nil {
		logger.Error(messages.FailedToRecordHistogramValue, err, lager.Data{
			"value": int64(d),
		})
	}
}

func sendHistogramQuantile(logger lager.Logger, statter statsd.Statter, histogram *monitor.Histogram, quantile float64, metric string) {
	v := histogram.ValueAtQuantile(quantile)
	sendGauge(logger, statter, metric, v)
}

func sendHistogramMax(logger lager.Logger, statter statsd.Statter, histogram *monitor.Histogram, metric string) {
	v := histogram.Max()
	sendGauge(logger, statter, metric, v)
}
