package monitor_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/perm/pkg/monitor"
)

var _ = Describe("Histogram", func() {
	var (
		histogramOptions HistogramOptions
	)

	BeforeEach(func() {
		histogramOptions = HistogramOptions{
			Name: "test.histogram",
		}
	})

	Describe("#Observe", func() {
		It("records the expected max", func() {
			histogramOptions.MaxDuration = time.Second
			subject := NewHistogram(histogramOptions)

			err := subject.Observe(time.Millisecond)
			Expect(err).NotTo(HaveOccurred())
			err = subject.Observe(time.Millisecond * 30)
			Expect(err).NotTo(HaveOccurred())
			err = subject.Observe(time.Millisecond * 55)
			Expect(err).NotTo(HaveOccurred())

			values := subject.Collect()
			Expect(values).To(HaveKeyWithValue("test.histogram.max", int64(55)))
		})

		It("fails if the value is larger than the MaxDuration", func() {
			histogramOptions.MaxDuration = time.Second
			subject := NewHistogram(histogramOptions)

			err := subject.Observe(time.Second)
			Expect(err).NotTo(HaveOccurred())

			err = subject.Observe(time.Hour)
			Expect(err).To(HaveOccurred())
		})

		It("fails if the value is negative", func() {
			histogramOptions.MaxDuration = time.Minute
			subject := NewHistogram(histogramOptions)

			err := subject.Observe(0)
			Expect(err).NotTo(HaveOccurred())

			err = subject.Observe(time.Second * -1)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#Collect", func() {
		It("returns labeled values for all buckets, including max", func() {
			histogramOptions.Buckets = []float64{50, 75, 99.9}
			subject := NewHistogram(histogramOptions)

			values := subject.Collect()
			Expect(values).To(HaveLen(len(histogramOptions.Buckets) + 1)) // 1 per bucket + max
			Expect(values).To(HaveKeyWithValue("test.histogram.max", int64(0)))
			Expect(values).To(HaveKeyWithValue("test.histogram.p50", int64(0)))
			Expect(values).To(HaveKeyWithValue("test.histogram.p75", int64(0)))
			Expect(values).To(HaveKeyWithValue("test.histogram.p999", int64(0)))
		})

		It("contains no values larger than the max", func() {
			histogramOptions.MaxDuration = time.Second
			subject := NewHistogram(histogramOptions)

			err := subject.Observe(time.Millisecond)
			Expect(err).NotTo(HaveOccurred())
			err = subject.Observe(time.Minute)
			Expect(err).To(HaveOccurred())

			values := subject.Collect()
			for _, v := range values {
				Expect(v).To(BeNumerically("<=", int64(time.Second/time.Millisecond)))
			}
		})

		It("contains the expected values for each bucket", func() {
			histogramOptions.MaxDuration = time.Millisecond * 5
			histogramOptions.Buckets = []float64{50, 85}
			subject := NewHistogram(histogramOptions)

			err := subject.Observe(time.Millisecond)
			Expect(err).NotTo(HaveOccurred())
			err = subject.Observe(time.Millisecond * 2)
			Expect(err).NotTo(HaveOccurred())
			err = subject.Observe(time.Millisecond * 3)
			Expect(err).NotTo(HaveOccurred())

			values := subject.Collect()
			Expect(values).To(HaveKeyWithValue("test.histogram.p50", int64(2)))
			Expect(values).To(HaveKeyWithValue("test.histogram.p85", int64(3)))
		})

		It("rotates values if enough data has been collected, based on the most granular bucket", func() {
			histogramOptions.MaxDuration = time.Millisecond * 5
			histogramOptions.Buckets = []float64{25, 50} // should rotate every 4 data points
			subject := NewHistogram(histogramOptions)

			countBeforeRotation := 4

			for i := 0; i < countBeforeRotation; i++ {
				err := subject.Observe(time.Millisecond * 1)
				Expect(err).NotTo(HaveOccurred())
			}

			// Should be:
			//   1 [1] 1 1
			Expect(subject.Collect()).To(HaveKeyWithValue("test.histogram.p50", int64(1)))

			for i := 0; i < countBeforeRotation; i++ {
				err := subject.Observe(time.Millisecond * 2)
				Expect(err).NotTo(HaveOccurred())
			}

			// Should be:
			//   1 1 1 [1] 2 2 2 2
			Expect(subject.Collect()).To(HaveKeyWithValue("test.histogram.p50", int64(1)))

			for i := 0; i < countBeforeRotation; i++ {
				err := subject.Observe(time.Millisecond * 3)
				Expect(err).NotTo(HaveOccurred())
			}

			// Should be:
			//   2 2 2 [2] 3 3 3 3
			// Without rotation, would be:
			//  1 1 1 1 2 [2] 2 2 3 3 3 3
			Expect(subject.Collect()).To(HaveKeyWithValue("test.histogram.p50", int64(2)))

			for i := 0; i < countBeforeRotation; i++ {
				err := subject.Observe(time.Millisecond * 4)
				Expect(err).NotTo(HaveOccurred())
			}

			// Should be:
			//   3 3 3 [3] 4 4 4 4
			// Without rotation, would be:
			//  1 1 1 1 2 2 2 [2] 3 3 3 3 4 4 4 4
			Expect(subject.Collect()).To(HaveKeyWithValue("test.histogram.p50", int64(3)))
		})
	})
})
