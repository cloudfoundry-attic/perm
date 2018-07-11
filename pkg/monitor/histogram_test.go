package monitor_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/perm/pkg/monitor"
)

var _ = Describe("ThreadSafeHistogram", func() {
	var (
		subject *ThreadSafeHistogram
	)

	BeforeEach(func() {
		subject = NewThreadSafeHistogram(5, 5)
	})

	Describe("#Max", func() {
		It("returns the highest recorded value", func() {
			Expect(subject.Max()).To(Equal(int64(0)))

			subject.RecordValue(10)
			subject.RecordValue(12345)
			subject.RecordValue(-30)
			subject.RecordValue(678)

			Expect(subject.Max()).To(Equal(int64(12345)))
		})
	})

	Describe("#ValueAtQuantile", func() {
		It("returns the value at the given quantile", func() {
			Expect(subject.ValueAtQuantile(50)).To(Equal(int64(0)))

			subject.RecordValue(1)
			subject.RecordValue(2)
			subject.RecordValue(3)

			Expect(subject.ValueAtQuantile(84)).To(Equal(int64(3)))
			Expect(subject.ValueAtQuantile(50)).To(Equal(int64(2)))
		})
		It("understands p100 as a max", func() {
			for j := int64(1); j <= 5; j++ {
				for i := int64(0); i <= 100; i++ {
					subject.RecordValue(i + j)
				}
			}
			maxValue := int64(105)
			Expect(subject.ValueAtQuantile(100)).To(Equal(maxValue))
			Expect(subject.Max()).To(Equal(maxValue))
		})
		It("reports quantiles and max from the same time window", func() {
			for j := int64(5); j > 0; j-- {
				subject.Rotate()
				for i := int64(100); i > 0; i-- {
					subject.RecordValue(i + j)
				}
			}
			maxValue := int64(105)
			Expect(subject.ValueAtQuantile(100)).To(Equal(maxValue))
			Expect(subject.Max()).To(Equal(maxValue))
		})
	})

	Describe("#Rotate", func() {
		It("resets the values once it's rotated out of the window size", func() {
			Expect(subject.ValueAtQuantile(50)).To(Equal(int64(0)))

			subject.RecordValue(1)
			subject.RecordValue(2)
			subject.RecordValue(3)

			Expect(subject.ValueAtQuantile(50)).To(Equal(int64(2)))

			subject.Rotate()
			Expect(subject.ValueAtQuantile(50)).To(Equal(int64(2)))
			subject.Rotate()
			Expect(subject.ValueAtQuantile(50)).To(Equal(int64(2)))
			subject.Rotate()
			Expect(subject.ValueAtQuantile(50)).To(Equal(int64(2)))
			subject.Rotate()
			Expect(subject.ValueAtQuantile(50)).To(Equal(int64(2)))
			subject.Rotate()
			Expect(subject.ValueAtQuantile(50)).To(Equal(int64(0)))
		})
	})
})
