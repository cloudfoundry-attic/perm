package statsdx

import (
	"time"

	"code.cloudfoundry.org/perm/pkg/logx"
	"github.com/cactus/go-statsd-client/statsd"
)

const (
	alwaysSample   = 1
	failureMessage = "failed-to-send-metric"
)

type Statter struct {
	statsdClient statsd.Statter
	logger       logx.Logger
}

func NewStatter(logger logx.Logger, statsdClient statsd.Statter) *Statter {
	return &Statter{
		statsdClient: statsdClient,
	}
}

func (s *Statter) Inc(metric string, value int64) {
	if err := s.statsdClient.Inc(metric, value, alwaysSample); err != nil {
		s.logger.Error(failureMessage, err, logx.Data{
			Key:   "metric",
			Value: metric,
		}, logx.Data{
			Key:   "value",
			Value: value,
		})
	}
}

func (s *Statter) Gauge(metric string, value int64) {
	if err := s.statsdClient.Gauge(metric, value, alwaysSample); err != nil {
		s.logger.Error(failureMessage, err, logx.Data{
			Key:   "metric",
			Value: metric,
		}, logx.Data{
			Key:   "value",
			Value: value,
		})
	}
}

func (s *Statter) TimingDuration(metric string, value time.Duration) {
	if err := s.statsdClient.TimingDuration(metric, value, alwaysSample); err != nil {
		s.logger.Error(failureMessage, err, logx.Data{
			Key:   "metric",
			Value: metric,
		}, logx.Data{
			Key:   "value",
			Value: value,
		})
	}
}
