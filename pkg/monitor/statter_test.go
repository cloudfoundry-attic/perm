package monitor_test

import (
	. "code.cloudfoundry.org/perm/pkg/monitor"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/pkg/monitor/monitorfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Statter", func() {
	var (
		histogram *ThreadSafeHistogram
		statsd    *monitorfakes.FakePermStatter

		logger *lagertest.TestLogger

		statter *Statter
	)

	BeforeEach(func() {
		histogram = NewThreadSafeHistogram(1, 3)
		statsd = new(monitorfakes.FakePermStatter)

		logger = lagertest.NewTestLogger("statter")

		statter = &Statter{
			statsd,
			histogram,
		}
	})

	Describe("SendFailedProbe", func() {
		It("sends a failure for the stat", func() {
			statter.SendFailedProbe(logger)

			Expect(statsd.GaugeCallCount()).To(Equal(1))

			metricName, value, rate := statsd.GaugeArgsForCall(0)
			Expect(metricName).To(Equal("perm.probe.runs.success"))
			Expect(value).To(Equal(int64(0)))
			Expect(rate).To(Equal(float32(1.0)))
		})
	})

	Describe("SendIncorrectProbe", func() {
		It("sends a failure and incorrect stat", func() {
			statter.SendIncorrectProbe(logger)

			Expect(statsd.GaugeCallCount()).To(Equal(2))

			metricName, value, rate := statsd.GaugeArgsForCall(0)
			Expect(metricName).To(Equal("perm.probe.runs.success"))
			Expect(value).To(Equal(int64(0)))
			Expect(rate).To(Equal(float32(1.0)))

			metricName, value, rate = statsd.GaugeArgsForCall(1)
			Expect(metricName).To(Equal("perm.probe.runs.correct"))
			Expect(value).To(Equal(int64(0)))
			Expect(rate).To(Equal(float32(1.0)))
		})
	})

	Describe("SendCorrectProbe", func() {
		It("sends successful and correct gauges, and 90, 99, 99.9th, and max quantile stats", func() {
			statter.RecordProbeDuration(logger, 1)

			statter.SendCorrectProbe(logger)

			Expect(statsd.GaugeCallCount()).To(Equal(6))

			metricName, value, rate := statsd.GaugeArgsForCall(0)
			Expect(metricName).To(Equal("perm.probe.runs.success"))
			Expect(value).To(Equal(int64(1)))
			Expect(rate).To(Equal(float32(1.0)))

			metricName, value, rate = statsd.GaugeArgsForCall(1)
			Expect(metricName).To(Equal("perm.probe.runs.correct"))
			Expect(value).To(Equal(int64(1)))
			Expect(rate).To(Equal(float32(1.0)))

			metricName, value, rate = statsd.GaugeArgsForCall(2)
			Expect(metricName).To(Equal("perm.probe.responses.timing.p90"))
			Expect(value).To(Equal(int64(1)))
			Expect(rate).To(Equal(float32(1.0)))

			metricName, value, rate = statsd.GaugeArgsForCall(3)
			Expect(metricName).To(Equal("perm.probe.responses.timing.p99"))
			Expect(value).To(Equal(int64(1)))
			Expect(rate).To(Equal(float32(1.0)))

			metricName, value, rate = statsd.GaugeArgsForCall(4)
			Expect(metricName).To(Equal("perm.probe.responses.timing.p999"))
			Expect(value).To(Equal(int64(1)))
			Expect(rate).To(Equal(float32(1.0)))

			metricName, value, rate = statsd.GaugeArgsForCall(5)
			Expect(metricName).To(Equal("perm.probe.responses.timing.max"))
			Expect(value).To(Equal(int64(1)))
			Expect(rate).To(Equal(float32(1.0)))
		})
	})
})
