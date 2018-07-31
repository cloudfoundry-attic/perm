package monitor

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/codahale/hdrhistogram"
)

type Histogram struct {
	name                string
	buckets             []float64
	countBeforeRotation int64
	histogram           *hdrhistogram.WindowedHistogram
}

type HistogramOptions struct {
	Name        string
	Buckets     []float64
	MinDuration time.Duration
	MaxDuration time.Duration
}

func NewHistogram(opts HistogramOptions) *Histogram {
	var countBeforeRotation int64
	// e.g., you need >= 2 data points for 50, >= 4 for 25 or 75, >= 100 for 99, >= 1000 for 99.9, etc.
	// Doesn't currently work well if the number has a repeating decimal, e.g., 66.6...
	for _, b := range opts.Buckets {
		m := int64(100)
		for b != math.Trunc(b) {
			m *= 10
			b *= 10
		}

		count := m / gcd(int64(math.Trunc(b)), m)
		if count > countBeforeRotation {
			countBeforeRotation = count
		}
	}

	return &Histogram{
		name:                opts.Name,
		buckets:             opts.Buckets,
		countBeforeRotation: countBeforeRotation,
		histogram:           hdrhistogram.NewWindowed(2, 0, durationToMilliseconds(opts.MaxDuration), 1),
	}
}

func (h *Histogram) Observe(duration time.Duration) error {
	if h.histogram.Current.TotalCount() >= h.countBeforeRotation {
		h.histogram.Rotate()
	}

	return h.histogram.Current.RecordValue(durationToMilliseconds(duration))
}

func (h *Histogram) Collect() map[string]int64 {
	histogram := h.histogram.Merge()
	values := make(map[string]int64)
	values[fmt.Sprintf("%s.max", h.name)] = histogram.Max()

	for _, b := range h.buckets {
		quantileLabel := strings.Replace(strconv.FormatFloat(b, 'f', -1, 64), ".", "", -1)
		values[fmt.Sprintf("%s.p%s", h.name, quantileLabel)] = histogram.ValueAtQuantile(b)
	}

	return values
}

func durationToMilliseconds(d time.Duration) int64 {
	return int64(d / time.Millisecond)
}

func gcd(x, y int64) int64 {
	for y != 0 {
		x, y = y, x%y
	}

	return x
}
