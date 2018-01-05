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

	MetricAdminProbeRunsSuccess = "perm.probe.admin.runs.success"

	MetricQueryProbeRunsSuccess = "perm.probe.query.runs.success"
	MetricQueryProbeRunsCorrect = "perm.probe.query.runs.correct"

	MetricQueryProbeTimingMax  = "perm.probe.query.responses.timing.max"  // gauge
	MetricQueryProbeTimingP90  = "perm.probe.query.responses.timing.p90"  // gauge
	MetricQueryProbeTimingP99  = "perm.probe.query.responses.timing.p99"  // gauge
	MetricQueryProbeTimingP999 = "perm.probe.query.responses.timing.p999" // gauge
)

type Statter struct {
	StatsD    statsd.Statter
	Histogram *monitor.Histogram
}

func (s *Statter) Rotate() {
	s.Histogram.Rotate()
}

func (s *Statter) RecordQueryProbeDuration(logger lager.Logger, d time.Duration) {
	err := s.Histogram.RecordValue(int64(d))
	if err != nil {
		logger.Error(messages.FailedToRecordHistogramValue, err, lager.Data{
			"value": int64(d),
		})
	}
}

func (s *Statter) SendFailedQueryProbe(logger lager.Logger) {
	s.sendGauge(logger, MetricQueryProbeRunsSuccess, MetricFailure)
}

func (s *Statter) SendIncorrectQueryProbe(logger lager.Logger) {
	s.sendGauge(logger, MetricQueryProbeRunsSuccess, MetricFailure)
	s.sendGauge(logger, MetricQueryProbeRunsCorrect, MetricFailure)
}

func (s *Statter) SendCorrectQueryProbe(logger lager.Logger) {
	s.sendGauge(logger, MetricQueryProbeRunsSuccess, MetricSuccess)
	s.sendGauge(logger, MetricQueryProbeRunsCorrect, MetricSuccess)
	s.sendHistogramQuantile(logger, 90, MetricQueryProbeTimingP90)
	s.sendHistogramQuantile(logger, 99, MetricQueryProbeTimingP99)
	s.sendHistogramQuantile(logger, 99.9, MetricQueryProbeTimingP999)
	s.sendHistogramMax(logger, MetricQueryProbeTimingMax)
}

func (s *Statter) SendFailedAdminProbe(logger lager.Logger) {
	s.sendGauge(logger, MetricAdminProbeRunsSuccess, MetricFailure)
}

func (s *Statter) SendSuccessfulAdminProbe(logger lager.Logger) {
	s.sendGauge(logger, MetricAdminProbeRunsSuccess, MetricSuccess)
}

func (s *Statter) sendGauge(logger lager.Logger, name string, value int64) {
	err := s.StatsD.Gauge(name, value, AlwaysSendMetric)
	if err != nil {
		logger.Error(messages.FailedToSendMetric, err, lager.Data{
			"metric": name,
		})
	}
}

func (s *Statter) sendHistogramQuantile(logger lager.Logger, quantile float64, metric string) {
	v := s.Histogram.ValueAtQuantile(quantile)
	s.sendGauge(logger, metric, v)
}

func (s *Statter) sendHistogramMax(logger lager.Logger, metric string) {
	v := s.Histogram.Max()
	s.sendGauge(logger, metric, v)
}
