package monitor_test

import (
	"time"

	. "code.cloudfoundry.org/perm/pkg/monitor"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/pkg/monitor/monitorfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Statter", func() {
	var (
		statsd *monitorfakes.FakePermStatter

		logger *lagertest.TestLogger

		statter *Statter
	)

	BeforeEach(func() {
		statsd = new(monitorfakes.FakePermStatter)

		logger = lagertest.NewTestLogger("statter")

		statter = NewStatter(statsd)
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
		It("sends successful and correct gauges", func() {
			statter.RecordProbeDuration(logger, 1)

			statter.SendCorrectProbe(logger)

			Expect(statsd.GaugeCallCount()).To(Equal(2))

			metricName, value, rate := statsd.GaugeArgsForCall(0)
			Expect(metricName).To(Equal("perm.probe.runs.success"))
			Expect(value).To(Equal(int64(1)))
			Expect(rate).To(Equal(float32(1.0)))

			metricName, value, rate = statsd.GaugeArgsForCall(1)
			Expect(metricName).To(Equal("perm.probe.runs.correct"))
			Expect(value).To(Equal(int64(1)))
			Expect(rate).To(Equal(float32(1.0)))
		})
	})

	Describe("SendStats", func() {
		It("sends 50, 90, 99, 99.9th, and max quantile stats", func() {
			statter.RecordProbeDuration(logger, time.Second)

			statter.SendStats(logger)

			Expect(statsd.GaugeCallCount()).To(Equal(5))

			var metricNames []string

			for i := 0; i < 5; i++ {
				metricName, _, _ := statsd.GaugeArgsForCall(i)
				metricNames = append(metricNames, metricName)
			}

			Expect(metricNames).To(ContainElement("perm.probe.responses.timing.max"))
			Expect(metricNames).To(ContainElement("perm.probe.responses.timing.p50"))
			Expect(metricNames).To(ContainElement("perm.probe.responses.timing.p90"))
			Expect(metricNames).To(ContainElement("perm.probe.responses.timing.p99"))
			Expect(metricNames).To(ContainElement("perm.probe.responses.timing.p999"))
		})
	})
})
