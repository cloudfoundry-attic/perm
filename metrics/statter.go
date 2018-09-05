package metrics

import "time"

type Statter interface {
	Inc(metric string, value int64)
	Gauge(metric string, value int64)
	TimingDuration(metric string, value time.Duration)
}
