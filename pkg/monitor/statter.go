package monitor

import (
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/monitor/stats"
	"github.com/cactus/go-statsd-client/statsd"
)

const (
	MetricFailure = 0.0
	MetricSuccess = 1.0

	AlwaysSendMetric = 1.0

	MetricProbeRunsSuccess = "perm.probe.runs.success"
	MetricProbeRunsCorrect = "perm.probe.runs.correct"

	MetricProbeTimingMax  = "perm.probe.responses.timing.max"  // gauge
	MetricProbeTimingP50  = "perm.probe.responses.timing.p50"  // gauge
	MetricProbeTimingP90  = "perm.probe.responses.timing.p90"  // gauge
	MetricProbeTimingP99  = "perm.probe.responses.timing.p99"  // gauge
	MetricProbeTimingP999 = "perm.probe.responses.timing.p999" // gauge
)

//go:generate counterfeiter . PermStatter

type PermStatter interface {
	statsd.Statter
	RecordProbeDuration(logger lager.Logger, d time.Duration)
	SendFailedProbe(logger lager.Logger)
	SendIncorrectProbe(logger lager.Logger)
	SendCorrectProbe(logger lager.Logger)
	SendStats(logger lager.Logger)
}

type Statter struct {
	statsd.Statter
	histogram *stats.Histogram
}

func NewStatter(statter statsd.Statter) *Statter {
	s := &Statter{
		histogram: stats.NewHistogram(stats.HistogramOptions{
			Name:        "perm.probe.responses.timing",
			MaxDuration: time.Second * 3,
			Buckets:     []float64{50, 90, 99, 99.9},
		}),
	}
	s.Statter = statter

	return s
}

func (s *Statter) RecordProbeDuration(logger lager.Logger, d time.Duration) {
	err := s.histogram.Observe(d)
	if err != nil {
		logger.Error(failedToRecordHistogramValue, err, lager.Data{
			"value": d,
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
	for k, v := range s.histogram.Collect() {
		s.Statter.Gauge(k, v, AlwaysSendMetric)
	}
}

func (s *Statter) sendGauge(logger lager.Logger, name string, value int64) {
	err := s.Statter.Gauge(name, value, AlwaysSendMetric)
	if err != nil {
		logger.Error(failedToSendMetric, err, lager.Data{
			"metric": name,
		})
	}
}

type histogramRecorder struct {
	recordHistogram func(map[string]int64) error
}

func (r *histogramRecorder) RecordHistogram(values map[string]int64) error {
	return r.recordHistogram(values)
}
