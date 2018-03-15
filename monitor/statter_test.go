package monitor_test

import (
	. "code.cloudfoundry.org/perm/monitor"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/monitor/monitorfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Statter", func() {
	var (
		histogram *ThreadSafeHistogram
		statsd    *monitorfakes.FakeStatter

		logger *lagertest.TestLogger

		statter *Statter
	)

	BeforeEach(func() {
		histogram = NewThreadSafeHistogram(1, 3)
		statsd = new(monitorfakes.FakeStatter)

		logger = lagertest.NewTestLogger("statter")

		statter = &Statter{
			StatsD:    statsd,
			Histogram: histogram,
		}
	})

	Describe("SendFailedQueryProbe", func() {
		It("sends a failure for the query stat", func() {
			statter.SendFailedQueryProbe(logger)

			Expect(statsd.GaugeCallCount()).To(Equal(1))

			metricName, value, rate := statsd.GaugeArgsForCall(0)
			Expect(metricName).To(Equal("perm.probe.query.runs.success"))
			Expect(value).To(Equal(int64(0)))
			Expect(rate).To(Equal(float32(1.0)))
		})
	})

	Describe("SendIncorrectQueryProbe", func() {
		It("sends a failure and incorrect query stat", func() {
			statter.SendIncorrectQueryProbe(logger)

			Expect(statsd.GaugeCallCount()).To(Equal(2))

			metricName, value, rate := statsd.GaugeArgsForCall(0)
			Expect(metricName).To(Equal("perm.probe.query.runs.success"))
			Expect(value).To(Equal(int64(0)))
			Expect(rate).To(Equal(float32(1.0)))

			metricName, value, rate = statsd.GaugeArgsForCall(1)
			Expect(metricName).To(Equal("perm.probe.query.runs.correct"))
			Expect(value).To(Equal(int64(0)))
			Expect(rate).To(Equal(float32(1.0)))
		})
	})

	Describe("SendCorrectQueryProbe", func() {
		It("sends successful and correct gauges, and 90, 99, 99.9th, and max quantile query stats", func() {
			statter.RecordQueryProbeDuration(logger, 1)

			statter.SendCorrectQueryProbe(logger)

			Expect(statsd.GaugeCallCount()).To(Equal(6))

			metricName, value, rate := statsd.GaugeArgsForCall(0)
			Expect(metricName).To(Equal("perm.probe.query.runs.success"))
			Expect(value).To(Equal(int64(1)))
			Expect(rate).To(Equal(float32(1.0)))

			metricName, value, rate = statsd.GaugeArgsForCall(1)
			Expect(metricName).To(Equal("perm.probe.query.runs.correct"))
			Expect(value).To(Equal(int64(1)))
			Expect(rate).To(Equal(float32(1.0)))

			metricName, value, rate = statsd.GaugeArgsForCall(2)
			Expect(metricName).To(Equal("perm.probe.query.responses.timing.p90"))
			Expect(value).To(Equal(int64(1)))
			Expect(rate).To(Equal(float32(1.0)))

			metricName, value, rate = statsd.GaugeArgsForCall(3)
			Expect(metricName).To(Equal("perm.probe.query.responses.timing.p99"))
			Expect(value).To(Equal(int64(1)))
			Expect(rate).To(Equal(float32(1.0)))

			metricName, value, rate = statsd.GaugeArgsForCall(4)
			Expect(metricName).To(Equal("perm.probe.query.responses.timing.p999"))
			Expect(value).To(Equal(int64(1)))
			Expect(rate).To(Equal(float32(1.0)))

			metricName, value, rate = statsd.GaugeArgsForCall(5)
			Expect(metricName).To(Equal("perm.probe.query.responses.timing.max"))
			Expect(value).To(Equal(int64(1)))
			Expect(rate).To(Equal(float32(1.0)))
		})
	})

	Describe("SendFailedAdminProbe", func() {
		It("sends a failure for the admin stat", func() {
			statter.SendFailedAdminProbe(logger)

			Expect(statsd.GaugeCallCount()).To(Equal(1))

			metricName, value, rate := statsd.GaugeArgsForCall(0)
			Expect(metricName).To(Equal("perm.probe.admin.runs.success"))
			Expect(value).To(Equal(int64(0)))
			Expect(rate).To(Equal(float32(1.0)))
		})
	})

	Describe("SendSuccessfulAdminProbe", func() {
		It("sends a success for the admin stat", func() {
			statter.SendSuccessfulAdminProbe(logger)

			Expect(statsd.GaugeCallCount()).To(Equal(1))

			metricName, value, rate := statsd.GaugeArgsForCall(0)
			Expect(metricName).To(Equal("perm.probe.admin.runs.success"))
			Expect(value).To(Equal(int64(1)))
			Expect(rate).To(Equal(float32(1.0)))
		})
	})
})
