package monitor

import (
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/cactus/go-statsd-client/statsd"
)

const (
	MetricFailure = 0.0
	MetricSuccess = 1.0

	AlwaysSendMetric = 1.0

	MetricProbeRunsSuccess = "perm.probe.runs.success"
	MetricProbeRunsCorrect = "perm.probe.runs.correct"

	MetricProbeTimingMax  = "perm.probe.responses.timing.max"  // gauge
	MetricProbeTimingP90  = "perm.probe.responses.timing.p90"  // gauge
	MetricProbeTimingP99  = "perm.probe.responses.timing.p99"  // gauge
	MetricProbeTimingP999 = "perm.probe.responses.timing.p999" // gauge
)

//go:generate counterfeiter . PermStatter

type PermStatter interface {
	statsd.Statter
	Rotate()
	RecordProbeDuration(logger lager.Logger, d time.Duration)
	SendFailedProbe(logger lager.Logger)
	SendIncorrectProbe(logger lager.Logger)
	SendCorrectProbe(logger lager.Logger)
	SendStats(logger lager.Logger)
}

type Statter struct {
	statsd.Statter
	Histogram *ThreadSafeHistogram
}

func (s *Statter) Rotate() {
	s.Histogram.Rotate()
}

func (s *Statter) RecordProbeDuration(logger lager.Logger, d time.Duration) {
	err := s.Histogram.RecordValue(int64(d))
	if err != nil {
		logger.Error(failedToRecordHistogramValue, err, lager.Data{
			"value": int64(d),
		})
	}
}

func (s *Statter) SendFailedProbe(logger lager.Logger) {
	s.sendGauge(logger, MetricProbeRunsSuccess, MetricFailure)
}

func (s *Statter) SendIncorrectProbe(logger lager.Logger) {
	s.sendGauge(logger, MetricProbeRunsSuccess, MetricFailure)
	s.sendGauge(logger, MetricProbeRunsCorrect, MetricFailure)
}

func (s *Statter) SendCorrectProbe(logger lager.Logger) {
	s.sendGauge(logger, MetricProbeRunsSuccess, MetricSuccess)
	s.sendGauge(logger, MetricProbeRunsCorrect, MetricSuccess)
}

func (s *Statter) SendStats(logger lager.Logger) {
	s.sendHistogramQuantile(logger, 90, MetricProbeTimingP90)
	s.sendHistogramQuantile(logger, 99, MetricProbeTimingP99)
	s.sendHistogramQuantile(logger, 99.9, MetricProbeTimingP999)
	s.sendHistogramMax(logger, MetricProbeTimingMax)
}

func (s *Statter) sendGauge(logger lager.Logger, name string, value int64) {
	err := s.Gauge(name, value, AlwaysSendMetric)
	if err != nil {
		logger.Error(failedToSendMetric, err, lager.Data{
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
