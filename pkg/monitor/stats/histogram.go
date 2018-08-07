package stats

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/codahale/hdrhistogram"
)

type HistogramOptions struct {
	Name        string
	Buckets     []float64
	MaxDuration time.Duration
}

type Histogram struct {
	name    string
	buckets []float64

	// this is the max number of values to store in a window
	// when this value is reached, all values in the oldest window will be discarded
	// ie. if there are 2 windows, at any time there will be between countBeforeRotation
	// to 2*countBeforeRotation number of values for histogram calculations
	countBeforeRotation int64

	histogram *hdrhistogram.WindowedHistogram
}

func NewHistogram(opts HistogramOptions) *Histogram {
	// e.g., you need >= 2 data points for 50, >= 4 for 25 or 75, >= 100 for 99, >= 1000 for 99.9, etc.
	// Doesn't currently work well if the number has a repeating decimal, e.g., 66.6...
	var minCountNeeded int64

	for _, b := range opts.Buckets {
		m := int64(100)
		for b != math.Trunc(b) {
			m *= 10
			b *= 10
		}

		count := m / gcd(int64(math.Trunc(b)), m)
		if count > minCountNeeded {
			minCountNeeded = count
		}
	}

	// Attempt to keep enough data points to smooth over noise,
	// e.g., so that the most granular non-max bucket is distinguishable from max
	minCountNeeded *= 3

	return &Histogram{
		name:    opts.Name,
		buckets: opts.Buckets,
		// 10% rounded up of the min number of data points needed
		countBeforeRotation: int64(math.Ceil(float64(minCountNeeded) / 10)),
		// set to 11 windows so that combined, the histogram windows will always contain the min number
		// of data points needed for the most granular percentile, and additionally up to 10% more data
		// points for rotating the histogram windows
		// more histogram windows means that the oldest data will be dropped more frequently, minimizing
		// the number of times the same max value will be sent, thus reducing skew in graphs of max values
		// NOTE: 11 windows become suboptimal when there are less than 10 datapoints needed
		histogram: hdrhistogram.NewWindowed(11, 0, durationToMilliseconds(opts.MaxDuration), 1),
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

// CountBeforeRotation is for testing
func (h *Histogram) CountBeforeRotation() int64 {
	return h.countBeforeRotation
}

func durationToMilliseconds(d time.Duration) int64 {
	// division between two int64 values will round down
	return int64(d / time.Millisecond)
}

func gcd(x, y int64) int64 {
	for y != 0 {
		x, y = y, x%y
	}

	return x
}
