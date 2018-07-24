package monitor_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/perm/pkg/monitor"
)

var _ = Describe("ThreadSafeHistogram", func() {
	var (
		subject *HistogramSet
	)

	BeforeEach(func() {
		subject = NewHistogramSet()
	})

	Describe("#Max", func() {
		BeforeEach(func() {
			Expect(subject.Max("some-label")).To(Equal(int64(0)))

			subject.RecordValue("some-label", 10)
			subject.RecordValue("some-label", 12345)
			subject.RecordValue("some-label", -30)
			subject.RecordValue("some-label", 678)

			subject.RecordValue("some-other-label", 67890)
		})

		It("returns the highest recorded value", func() {
			Expect(subject.Max("some-label")).To(Equal(int64(12345)))
			Expect(subject.Max("some-other-label")).To(Equal(int64(67890)))
		})

		It("returns the highest recorded overall value", func() {
			Expect(subject.Max("overall")).To(Equal(int64(67890)))
		})
	})

	Describe("#ValueAtQuantile", func() {
		It("returns the value at the given quantile", func() {
			Expect(subject.ValueAtQuantile("some-label", 50)).To(Equal(int64(0)))

			subject.RecordValue("some-label", 1)
			subject.RecordValue("some-label", 2)
			subject.RecordValue("some-label", 3)

			Expect(subject.ValueAtQuantile("some-label", 84)).To(Equal(int64(3)))
			Expect(subject.ValueAtQuantile("some-label", 50)).To(Equal(int64(2)))
		})

		It("understands p100 as a max", func() {
			for j := int64(1); j <= 5; j++ {
				for i := int64(0); i <= 100; i++ {
					subject.RecordValue("some-label", i+j)
				}
			}
			maxValue := int64(105)
			Expect(subject.ValueAtQuantile("some-label", 100)).To(Equal(maxValue))
			Expect(subject.Max("some-label")).To(Equal(maxValue))
		})
		It("reports quantiles and max from the same time window", func() {
			for j := int64(5); j > 0; j-- {
				subject.Rotate()
				for i := int64(100); i > 0; i-- {
					subject.RecordValue("some-label", i+j)
				}
			}
			maxValue := int64(105)
			Expect(subject.ValueAtQuantile("some-label", 100)).To(Equal(maxValue))
			Expect(subject.Max("some-label")).To(Equal(maxValue))
		})

		It("records the overall quantiles if separate values are recorded", func() {
			Expect(subject.ValueAtQuantile("some-label", 50)).To(Equal(int64(0)))

			subject.RecordValue("some-label", 1)
			subject.RecordValue("some-label", 2)
			subject.RecordValue("some-label", 3)
			subject.RecordValue("some-other-label", 4)
			subject.RecordValue("yet-another-label", 5)

			Expect(subject.ValueAtQuantile("some-label", 99)).To(Equal(int64(3)))
			Expect(subject.ValueAtQuantile("overall", 99)).To(Equal(int64(5)))

			Expect(subject.ValueAtQuantile("some-label", 50)).To(Equal(int64(2)))
			Expect(subject.ValueAtQuantile("overall", 50)).To(Equal(int64(3)))
		})
	})

	Describe("#Rotate", func() {
		It("resets the values once it's rotated out of the window size", func() {
			Expect(subject.ValueAtQuantile("some-label", 50)).To(Equal(int64(0)))

			subject.RecordValue("some-label", 1)
			subject.RecordValue("some-label", 2)
			subject.RecordValue("some-label", 3)

			Expect(subject.ValueAtQuantile("some-label", 50)).To(Equal(int64(2)))

			subject.Rotate()
			Expect(subject.ValueAtQuantile("some-label", 50)).To(Equal(int64(2)))
			subject.Rotate()
			Expect(subject.ValueAtQuantile("some-label", 50)).To(Equal(int64(2)))
			subject.Rotate()
			Expect(subject.ValueAtQuantile("some-label", 50)).To(Equal(int64(2)))
			subject.Rotate()
			Expect(subject.ValueAtQuantile("some-label", 50)).To(Equal(int64(2)))
			subject.Rotate()
			Expect(subject.ValueAtQuantile("some-label", 50)).To(Equal(int64(0)))
		})

		It("rotates all values including the overall histogram's once they're out of the window size", func() {
			Expect(subject.ValueAtQuantile("some-label", 50)).To(Equal(int64(0)))

			subject.RecordValue("some-label", 1)
			subject.RecordValue("some-label", 2)
			subject.RecordValue("some-label", 3)
			subject.RecordValue("some-other-label", 4)
			subject.RecordValue("yet-another-label", 5)

			Expect(subject.ValueAtQuantile("overall", 50)).To(Equal(int64(3)))

			subject.Rotate()
			Expect(subject.ValueAtQuantile("overall", 50)).To(Equal(int64(3)))
			subject.Rotate()
			Expect(subject.ValueAtQuantile("overall", 50)).To(Equal(int64(3)))
			subject.Rotate()
			Expect(subject.ValueAtQuantile("overall", 50)).To(Equal(int64(3)))
			subject.Rotate()
			Expect(subject.ValueAtQuantile("overall", 50)).To(Equal(int64(3)))
			subject.Rotate()
			Expect(subject.ValueAtQuantile("overall", 50)).To(Equal(int64(0)))
		})
	})
})
