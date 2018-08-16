package metrics

import "time"

type Statter interface {
	Inc(metric string, value int64, rate float32) error
	Gauge(metric string, value int64, rate float32) error
	TimingDuration(metric string, value time.Duration, rate float32) error
}
