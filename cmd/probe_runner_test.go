package cmd_test

import (
	"errors"

	. "code.cloudfoundry.org/perm/cmd"

	"context"

	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/cmd/cmdfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var cancellableComputation = func(ctx context.Context, timeout time.Duration) error {
	done := make(chan struct{}, 1)

	go func() {
		time.Sleep(timeout)
		// Run completed successfully
		close(done)
	}()
	select {
	case <-ctx.Done():
		// Context was cancelled before computation succeeded
		// i.e. computation was cancelled
		return ctx.Err()
	case <-done:
		// Long computation succeeded
		return nil
	}
}

var _ = Describe("Running the Probes", func() {
	var (
		someErr      error
		someOtherErr error
	)

	BeforeEach(func() {
		someErr = errors.New("some-error")
		someOtherErr = errors.New("some-other-error")
	})

	Describe(".RunProbe", func() {
		var (
			probe *cmdfakes.FakeProbe

			ctx    context.Context
			logger *lagertest.TestLogger

			timeout time.Duration

			expectedDurations []time.Duration
		)

		BeforeEach(func() {
			probe = new(cmdfakes.FakeProbe)

			ctx = context.Background()
			logger = lagertest.NewTestLogger("run-probe")

			timeout = 5 * time.Second

			expectedDurations = []time.Duration{1 * time.Second, 2 * time.Second}
			probe.RunReturns(true, expectedDurations, nil)
		})

		It("calls the setup, run, and cleanup with a uuid", func() {
			correct, durations, err := RunProbe(ctx, logger, probe, timeout)

			Expect(probe.SetupCallCount()).To(Equal(1))
			Expect(probe.RunCallCount()).To(Equal(1))
			Expect(probe.CleanupCallCount()).To(Equal(1))

			_, _, setupId := probe.SetupArgsForCall(0)
			_, _, runId := probe.RunArgsForCall(0)
			_, _, cleanupId := probe.CleanupArgsForCall(0)

			Expect(err).NotTo(HaveOccurred())
			Expect(correct).To(BeTrue())
			Expect(durations).To(Equal(expectedDurations))

			Expect(setupId).To(Equal(runId))
			Expect(runId).To(Equal(cleanupId))
		})

		Context("timeouts", func() {
			It("errors if it times out", func() {
				probe.RunStub = func(
					ctx context.Context,
					logger lager.Logger,
					uniqueSuffix string,
				) (bool, []time.Duration, error) {
					err := cancellableComputation(ctx, 10*time.Millisecond)
					if err != nil {
						return false, nil, err
					}

					return true, nil, err
				}

				_, _, err := RunProbe(ctx, logger, probe, 10*time.Nanosecond)

				Expect(probe.SetupCallCount()).To(Equal(1))
				Expect(probe.RunCallCount()).To(Equal(1))
				Expect(probe.CleanupCallCount()).To(Equal(1))

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(context.DeadlineExceeded))
			})

			It("succeeds if the timeout is not exceeded", func() {
				probe.RunStub = func(
					ctx context.Context,
					logger lager.Logger,
					uniqueSuffix string,
				) (bool, []time.Duration, error) {
					err := cancellableComputation(ctx, 10*time.Millisecond)
					if err != nil {
						return false, nil, err
					}

					return true, nil, err
				}

				_, _, err := RunProbe(ctx, logger, probe, 300*time.Millisecond)

				Expect(probe.SetupCallCount()).To(Equal(1))
				Expect(probe.RunCallCount()).To(Equal(1))
				Expect(probe.CleanupCallCount()).To(Equal(1))

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the setup fails", func() {
			BeforeEach(func() {
				probe.SetupReturns(someErr)
			})

			It("returns the error, calling cleanup, but not calling run", func() {
				_, _, err := RunProbe(ctx, logger, probe, timeout)

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
				_, _, err := RunProbe(ctx, logger, probe, timeout)

				Expect(probe.SetupCallCount()).To(Equal(1))
				Expect(probe.RunCallCount()).To(Equal(1))
				Expect(probe.CleanupCallCount()).To(Equal(1))

				Expect(err).To(MatchError(someErr))
			})
		})

		Context("when cleanup fails", func() {
			BeforeEach(func() {
				probe.CleanupReturns(someErr)
			})

			It("returns the error", func() {
				_, _, err := RunProbe(ctx, logger, probe, timeout)

				Expect(probe.SetupCallCount()).To(Equal(1))
				Expect(probe.RunCallCount()).To(Equal(1))
				Expect(probe.CleanupCallCount()).To(Equal(1))

				Expect(err).To(MatchError(someErr))
			})
		})

		Context("when setup and cleanup fail", func() {
			BeforeEach(func() {
				probe.SetupReturns(someErr)
				probe.CleanupReturns(someOtherErr)
			})

			It("returns the setup error", func() {
				_, _, err := RunProbe(ctx, logger, probe, timeout)

				Expect(err).To(MatchError(someErr))
				Expect(err).NotTo(MatchError(someOtherErr))
			})
		})

		Context("when run and cleanup fail", func() {
			BeforeEach(func() {
				probe.RunReturns(false, nil, someErr)
				probe.CleanupReturns(someOtherErr)
			})

			It("returns the run error", func() {
				_, _, err := RunProbe(ctx, logger, probe, timeout)

				Expect(err).To(MatchError(someErr))
				Expect(err).NotTo(MatchError(someOtherErr))
			})
		})
	})
})
