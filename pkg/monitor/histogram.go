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

const (
	ProbeHistogramWindow      = 5 // Minutes
	ProbeHistogramRefreshTime = 5 * time.Minute
	SigFigs                   = 5
)

func NewHistogramSet() *HistogramSet {
	h := map[string]*hdrhistogram.WindowedHistogram{}

	return &HistogramSet{
		rw:         &sync.RWMutex{},
		histograms: h,
	}
}

func (h *HistogramSet) Max(label string) int64 {
	h.rw.RLock()
	defer h.rw.RUnlock()

	_, ok := h.histograms[label]
	if !ok {
		return 0
	}

	return h.histograms[label].Merge().Max()
}

func (h *HistogramSet) RecordValue(label string, v int64) error {
	h.rw.Lock()
	defer h.rw.Unlock()

	_, ok := h.histograms[label]
	if !ok {
		h.addHistogram(label)
	}

	return h.histograms[label].Current.RecordValue(v)
}

func (h *HistogramSet) ValueAtQuantile(label string, q float64) int64 {
	h.rw.RLock()
	defer h.rw.RUnlock()

	_, ok := h.histograms[label]
	if !ok {
		return 0
	}

	return h.histograms[label].Merge().ValueAtQuantile(q)
}

func (h *HistogramSet) Rotate() {
	h.rw.Lock()
	defer h.rw.Unlock()

	for _, histogram := range h.histograms {
		histogram.Rotate()
	}
}

func (h *HistogramSet) addHistogram(label string) {
	h.histograms[label] = hdrhistogram.NewWindowed(ProbeHistogramWindow, 0, int64(time.Minute*10), SigFigs)
}
