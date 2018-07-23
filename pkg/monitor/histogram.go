package monitor

import (
	"sync"

	"time"

	"github.com/codahale/hdrhistogram"
)

type HistogramSet struct {
	rw         *sync.RWMutex
	histograms map[string]*hdrhistogram.WindowedHistogram
}

func NewThreadSafeHistogram(windowSize int, sigfigs int) *HistogramSet {
	h := map[string]*hdrhistogram.WindowedHistogram{
		"total": hdrhistogram.NewWindowed(windowSize, 0, int64(time.Minute*10), sigfigs),
	}

	return &HistogramSet{
		rw:         &sync.RWMutex{},
		histograms: h,
	}
}

func (h *HistogramSet) Max(label string) int64 {
	h.rw.RLock()
	defer h.rw.RUnlock()

	return h.histograms[label].Merge().Max()
}

func (h *HistogramSet) RecordValue(label string, v int64) error {
	h.rw.Lock()
	defer h.rw.Unlock()

	return h.histograms[label].Current.RecordValue(v)
}

func (h *HistogramSet) ValueAtQuantile(label string, q float64) int64 {
	h.rw.RLock()
	defer h.rw.RUnlock()

	return h.histograms[label].Merge().ValueAtQuantile(q)
}

func (h *HistogramSet) Rotate() {
	h.rw.Lock()
	defer h.rw.Unlock()

	h.histograms["total"].Rotate()
}
