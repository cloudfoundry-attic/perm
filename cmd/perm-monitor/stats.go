package main

import (
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
	"github.com/cactus/go-statsd-client/statsd"
)

const (
	AlwaysSendMetric = 1.0
)

func incrementStat(logger lager.Logger, statter statsd.Statter, name string) {
	err := statter.Inc(name, 1, AlwaysSendMetric)
	if err != nil {
		logger.Error(messages.FailedToSendMetric, err, lager.Data{
			"metric": name,
		})
	}
}

func sendGauge(logger lager.Logger, statter statsd.Statter, name string, value int64) {
	err := statter.Gauge(name, value, AlwaysSendMetric)
	if err != nil {
		logger.Error(messages.FailedToSendMetric, err, lager.Data{
			"metric": name,
		})
	}
}
