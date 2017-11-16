package main

import (
	"sync"

	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
	"github.com/cactus/go-statsd-client/statsd"
	"github.com/codahale/hdrhistogram"
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

func recordHistogramDuration(logger lager.Logger, rw *sync.RWMutex, histogram *hdrhistogram.WindowedHistogram, d time.Duration) {
	rw.Lock()
	defer rw.Unlock()

	err := histogram.Current.RecordValue(int64(d))
	if err != nil {
		logger.Error(messages.FailedToRecordHistogramValue, err, lager.Data{
			"value": int64(d),
		})
	}
}

func sendHistogramQuantile(logger lager.Logger, statter statsd.Statter, rw *sync.RWMutex, histogram *hdrhistogram.WindowedHistogram, quantile float64, metric string) {
	rw.RLock()
	defer rw.RUnlock()

	v := histogram.Current.ValueAtQuantile(quantile)
	sendGauge(logger, statter, metric, v)
}

func sendHistogramMax(logger lager.Logger, statter statsd.Statter, rw *sync.RWMutex, histogram *hdrhistogram.WindowedHistogram, metric string) {
	rw.RLock()
	defer rw.RUnlock()

	v := histogram.Current.Max()
	sendGauge(logger, statter, metric, v)
}
