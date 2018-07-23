package cmd_test

import (
	"errors"

	. "code.cloudfoundry.org/perm/cmd"

	"context"

	"time"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/cmd/cmdfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/perm/pkg/monitor"
	"code.cloudfoundry.org/perm/pkg/monitor/monitorfakes"
)

var _ = Describe("Running the Probes", func() {
	var (
		someErr         error
		someOtherErr    error
		logger          *lagertest.TestLogger
		requestDuration time.Duration
		timeout         time.Duration
	)

	BeforeEach(func() {
		someErr = errors.New("some-error")
		someOtherErr = errors.New("some-other-error")
		logger = lagertest.NewTestLogger("run-probe")
		requestDuration = 100 * time.Millisecond
		timeout = 5 * time.Second
	})

	Describe(".GetProbeResults", func() {
		var (
			probe *cmdfakes.FakeProbe

			expectedSetupDurations   []monitor.LabeledDuration
			expectedRunDurations     []monitor.LabeledDuration
			expectedCleanupDurations []monitor.LabeledDuration
		)

		BeforeEach(func() {
			probe = new(cmdfakes.FakeProbe)

			expectedSetupDurations = []monitor.LabeledDuration{
				{Label: "CreateRole", Duration: 1 * time.Second},
				{Label: "AssignRole", Duration: 2 * time.Second},
			}
			expectedRunDurations = []monitor.LabeledDuration{
				{Label: "HasPermission", Duration: 3 * time.Second},
				{Label: "HasPermission", Duration: 4 * time.Second},
			}
			expectedCleanupDurations = []monitor.LabeledDuration{{Label: "DeleteRole", Duration: 5 * time.Second}}
			probe.RunReturns(true, expectedRunDurations, nil)
			probe.CleanupReturns(expectedCleanupDurations, nil)
			probe.SetupReturns(expectedSetupDurations, nil)
		})

		It("calls the setup, run, and cleanup with a uuid", func() {
			correct, durations, err := GetProbeResults(logger, probe, timeout)

			Expect(probe.SetupCallCount()).To(Equal(1))
			Expect(probe.RunCallCount()).To(Equal(1))
			Expect(probe.CleanupCallCount()).To(Equal(1))

			_, _, setupId := probe.SetupArgsForCall(0)
			_, _, runId := probe.RunArgsForCall(0)
			_, _, _, cleanupId := probe.CleanupArgsForCall(0)

			Expect(err).NotTo(HaveOccurred())
			Expect(correct).To(BeTrue())
			for _, expectedDuration := range expectedSetupDurations {
				Expect(durations).To(ContainElement(expectedDuration))
			}
			for _, expectedDuration := range expectedRunDurations {
				Expect(durations).To(ContainElement(expectedDuration))
			}
			for _, expectedDuration := range expectedCleanupDurations {
				Expect(durations).To(ContainElement(expectedDuration))
			}
			Expect(setupId).To(Equal(runId))
			Expect(runId).To(Equal(cleanupId))
		})

		Context("timeouts", func() {
			It("errors if it times out", func() {
				probe.RunReturns(false, nil, context.DeadlineExceeded)

				_, _, err := GetProbeResults(logger, probe, 10*time.Nanosecond)

				Expect(probe.SetupCallCount()).To(Equal(1))
				Expect(probe.RunCallCount()).To(Equal(1))
				Expect(probe.CleanupCallCount()).To(Equal(1))

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(context.DeadlineExceeded))
			})

			It("succeeds if the timeout is not exceeded", func() {
				probe.RunReturns(true, nil, nil)
				_, _, err := GetProbeResults(logger, probe, 300*time.Millisecond)

				Expect(probe.SetupCallCount()).To(Equal(1))
				Expect(probe.RunCallCount()).To(Equal(1))
				Expect(probe.CleanupCallCount()).To(Equal(1))

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the setup fails", func() {
			BeforeEach(func() {
				probe.SetupReturns(nil, someErr)
			})

			It("returns the error, calling cleanup, but not calling run", func() {
				_, _, err := GetProbeResults(logger, probe, timeout)

				Expect(probe.SetupCallCount()).To(Equal(1))
				Expect(probe.RunCallCount()).To(Equal(0))
				Expect(probe.CleanupCallCount()).To(Equal(1))

				Expect(err).To(MatchError(someErr))
			})
		})

		Context("when the run fails", func() {
			BeforeEach(func() {
				probe.RunReturns(false, nil, someErr)
			})

			It("still runs cleanup, but returns the error", func() {
				_, _, err := GetProbeResults(logger, probe, timeout)

				Expect(probe.SetupCallCount()).To(Equal(1))
				Expect(probe.RunCallCount()).To(Equal(1))
				Expect(probe.CleanupCallCount()).To(Equal(1))

				Expect(err).To(MatchError(someErr))
			})
		})

		Context("when cleanup fails", func() {
			BeforeEach(func() {
				probe.CleanupReturns([]monitor.LabeledDuration{}, someErr)
			})

			It("returns the error", func() {
				_, _, err := GetProbeResults(logger, probe, timeout)

				Expect(probe.SetupCallCount()).To(Equal(1))
				Expect(probe.RunCallCount()).To(Equal(1))
				Expect(probe.CleanupCallCount()).To(Equal(1))

				Expect(err).To(MatchError(someErr))
			})
		})

		Context("when setup and cleanup fail", func() {
			BeforeEach(func() {
				probe.SetupReturns(nil, someErr)
				probe.CleanupReturns([]monitor.LabeledDuration{}, someOtherErr)
			})

			It("returns the setup error", func() {
				_, _, err := GetProbeResults(logger, probe, timeout)

				Expect(err).To(MatchError(someErr))
				Expect(err).NotTo(MatchError(someOtherErr))
			})
		})

		Context("when run and cleanup fail", func() {
			BeforeEach(func() {
				probe.RunReturns(false, nil, someErr)
				probe.CleanupReturns([]monitor.LabeledDuration{}, someOtherErr)
			})

			It("returns the run error", func() {
				_, _, err := GetProbeResults(logger, probe, timeout)

				Expect(err).To(MatchError(someErr))
				Expect(err).NotTo(MatchError(someOtherErr))
			})
		})
	})

	Describe(".RecordProbeResults", func() {
		var probe *cmdfakes.FakeProbe
		var statter *monitorfakes.FakePermStatter

		BeforeEach(func() {
			probe = new(cmdfakes.FakeProbe)
			statter = new(monitorfakes.FakePermStatter)
		})

		It("reports failed probe when probe's setup fails", func() {
			probe.SetupReturns(nil, errors.New("error in setup"))
			RecordProbeResults(logger, probe, statter, requestDuration, timeout)
			Expect(statter.SendFailedProbeCallCount()).To(Equal(1))
		})

		It("reports failed probe when probe's cleanup fails", func() {
			probe.CleanupReturns([]monitor.LabeledDuration{}, errors.New("error in cleanup"))
			timeout = time.Second * 0
			RecordProbeResults(logger, probe, statter, requestDuration, timeout)
			Expect(statter.SendFailedProbeCallCount()).To(Equal(1))
		})

		It("reports incorrect probe when the probe wasn't correct", func() {
			probe.RunReturns(false, []monitor.LabeledDuration{}, nil)
			RecordProbeResults(logger, probe, statter, requestDuration, timeout)
			Expect(statter.SendIncorrectProbeCallCount()).To(Equal(1))
		})

		It("records probe durations and reports correct probe when durations are valid", func() {
			qd := monitor.LabeledDuration{Label: "qd", Duration: time.Millisecond * 30}
			durations := []monitor.LabeledDuration{qd, qd, qd}
			probe.RunReturns(true, durations, nil)
			RecordProbeResults(logger, probe, statter, requestDuration, timeout)
			Expect(statter.RecordProbeDurationCallCount()).To(Equal(3))
			Expect(statter.SendCorrectProbeCallCount()).To(Equal(1))
		})

		Context("when there is a duration that exceeds the query time window", func() {
			It("records the durations but also records a failure", func() {
				qd := monitor.LabeledDuration{Label: "qd", Duration: time.Millisecond * 130}
				durations := []monitor.LabeledDuration{qd, qd, qd}
				probe.RunReturns(true, durations, nil)
				RecordProbeResults(logger, probe, statter, requestDuration, timeout)
				Expect(statter.RecordProbeDurationCallCount()).To(Equal(3))
				Expect(statter.SendFailedProbeCallCount()).To(Equal(1))
				Expect(statter.SendCorrectProbeCallCount()).To(Equal(0))
			})
		})
	})
})
