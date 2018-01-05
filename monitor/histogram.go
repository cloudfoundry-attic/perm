package monitor

import (
	"time"

	"sync"

	"github.com/codahale/hdrhistogram"
)

type Histogram struct {
	rw        *sync.RWMutex
	histogram *hdrhistogram.WindowedHistogram
}

func NewHistogram(windowSize int, minValue, maxValue time.Duration, sigfigs int) *Histogram {
	h := hdrhistogram.NewWindowed(windowSize, int64(minValue), int64(maxValue), sigfigs)

	return &Histogram{
		rw:        &sync.RWMutex{},
		histogram: h,
	}
}

func (h *Histogram) Max() int64 {
	h.rw.RLock()
	defer h.rw.RUnlock()

	return h.histogram.Current.Max()
}

func (h *Histogram) RecordValue(v int64) error {
	h.rw.Lock()
	defer h.rw.Unlock()

	return h.histogram.Current.RecordValue(v)
}

func (h *Histogram) ValueAtQuantile(q float64) int64 {
	h.rw.RLock()
	defer h.rw.RUnlock()

	return h.histogram.Merge().ValueAtQuantile(q)
}

func (h *Histogram) Rotate() {
	h.rw.Lock()
	defer h.rw.Unlock()

	h.histogram.Rotate()
}
