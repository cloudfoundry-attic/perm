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

	Describe(".RunQueryProbe", func() {
		var (
			queryProbe *cmdfakes.FakeQueryProbe

			ctx    context.Context
			logger *lagertest.TestLogger

			timeout time.Duration

			expectedDurations []time.Duration
		)

		BeforeEach(func() {
			queryProbe = new(cmdfakes.FakeQueryProbe)

			ctx = context.Background()
			logger = lagertest.NewTestLogger("run-query-probe")

			timeout = 5 * time.Second

			expectedDurations = []time.Duration{1 * time.Second, 2 * time.Second}
			queryProbe.RunReturns(true, expectedDurations, nil)
		})

		It("calls the setup, run, and cleanup with a uuid", func() {
			correct, durations, err := RunQueryProbe(ctx, logger, queryProbe, timeout)

			Expect(queryProbe.SetupCallCount()).To(Equal(1))
			Expect(queryProbe.RunCallCount()).To(Equal(1))
			Expect(queryProbe.CleanupCallCount()).To(Equal(1))

			_, _, setupId := queryProbe.SetupArgsForCall(0)
			_, _, runId := queryProbe.RunArgsForCall(0)
			_, _, cleanupId := queryProbe.CleanupArgsForCall(0)

			Expect(err).NotTo(HaveOccurred())
			Expect(correct).To(BeTrue())
			Expect(durations).To(Equal(expectedDurations))

			Expect(setupId).To(Equal(runId))
			Expect(runId).To(Equal(cleanupId))
		})

		Context("timeouts", func() {
			It("errors if it times out", func() {
				queryProbe.RunStub = func(ctx context.Context, logger lager.Logger, uniqueSuffix string) (bool, []time.Duration, error) {
					err := cancellableComputation(ctx, 10*time.Millisecond)
					if err != nil {
						return false, nil, err
					}

					return true, nil, err
				}

				_, _, err := RunQueryProbe(ctx, logger, queryProbe, 10*time.Nanosecond)

				Expect(queryProbe.SetupCallCount()).To(Equal(1))
				Expect(queryProbe.RunCallCount()).To(Equal(1))
				Expect(queryProbe.CleanupCallCount()).To(Equal(1))

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(context.DeadlineExceeded))
			})

			It("succeeds if the timeout is not exceeded", func() {
				queryProbe.RunStub = func(ctx context.Context, logger lager.Logger, uniqueSuffix string) (bool, []time.Duration, error) {
					err := cancellableComputation(ctx, 10*time.Millisecond)
					if err != nil {
						return false, nil, err
					}

					return true, nil, err
				}

				_, _, err := RunQueryProbe(ctx, logger, queryProbe, 20*time.Millisecond)

				Expect(queryProbe.SetupCallCount()).To(Equal(1))
				Expect(queryProbe.RunCallCount()).To(Equal(1))
				Expect(queryProbe.CleanupCallCount()).To(Equal(1))

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the setup fails", func() {
			BeforeEach(func() {
				queryProbe.SetupReturns(someErr)
			})

			It("returns the error, calling cleanup, but not calling run", func() {
				_, _, err := RunQueryProbe(ctx, logger, queryProbe, timeout)

				Expect(queryProbe.SetupCallCount()).To(Equal(1))
				Expect(queryProbe.RunCallCount()).To(Equal(0))
				Expect(queryProbe.CleanupCallCount()).To(Equal(1))

				Expect(err).To(MatchError(someErr))
			})
		})

		Context("when the run fails", func() {
			BeforeEach(func() {
				queryProbe.RunReturns(false, nil, someErr)
			})

			It("still runs cleanup, but returns the error", func() {
				_, _, err := RunQueryProbe(ctx, logger, queryProbe, timeout)

				Expect(queryProbe.SetupCallCount()).To(Equal(1))
				Expect(queryProbe.RunCallCount()).To(Equal(1))
				Expect(queryProbe.CleanupCallCount()).To(Equal(1))

				Expect(err).To(MatchError(someErr))
			})
		})

		Context("when cleanup fails", func() {
			BeforeEach(func() {
				queryProbe.CleanupReturns(someErr)
			})

			It("returns the error", func() {
				_, _, err := RunQueryProbe(ctx, logger, queryProbe, timeout)

				Expect(queryProbe.SetupCallCount()).To(Equal(1))
				Expect(queryProbe.RunCallCount()).To(Equal(1))
				Expect(queryProbe.CleanupCallCount()).To(Equal(1))

				Expect(err).To(MatchError(someErr))
			})
		})

		Context("when setup and cleanup fail", func() {
			BeforeEach(func() {
				queryProbe.SetupReturns(someErr)
				queryProbe.CleanupReturns(someOtherErr)
			})

			It("returns the setup error", func() {
				_, _, err := RunQueryProbe(ctx, logger, queryProbe, timeout)

				Expect(err).To(MatchError(someErr))
				Expect(err).NotTo(MatchError(someOtherErr))
			})
		})

		Context("when run and cleanup fail", func() {
			BeforeEach(func() {
				queryProbe.RunReturns(false, nil, someErr)
				queryProbe.CleanupReturns(someOtherErr)
			})

			It("returns the run error", func() {
				_, _, err := RunQueryProbe(ctx, logger, queryProbe, timeout)

				Expect(err).To(MatchError(someErr))
				Expect(err).NotTo(MatchError(someOtherErr))
			})
		})
	})

	Describe(".RunAdminProbe", func() {
		var (
			ctx        context.Context
			logger     *lagertest.TestLogger
			adminProbe *cmdfakes.FakeAdminProbe
			timeout    time.Duration
		)

		BeforeEach(func() {
			ctx = context.Background()
			logger = lagertest.NewTestLogger("run-admin-probe")
			adminProbe = new(cmdfakes.FakeAdminProbe)
			timeout = 1 * time.Second
		})

		It("calls run and cleanup with a uuid", func() {
			err := RunAdminProbe(
				ctx,
				logger,
				adminProbe,
				timeout,
			)

			Expect(err).NotTo(HaveOccurred())

			Expect(adminProbe.RunCallCount()).To(Equal(1))
			Expect(adminProbe.CleanupCallCount()).To(Equal(1))

			_, _, runId := adminProbe.RunArgsForCall(0)
			_, _, cleanupId := adminProbe.CleanupArgsForCall(0)

			Expect(runId).To(Equal(cleanupId))
		})

		Describe("Timeouts", func() {
			Context("when the run times out", func() {
				It("runs the cleanup and returns an error", func() {
					adminProbe.RunStub = func(ctx context.Context, logger lager.Logger, uniqueSuffix string) error {
						return cancellableComputation(ctx, 20*time.Millisecond)
					}

					err := RunAdminProbe(
						ctx,
						logger,
						adminProbe,
						10*time.Millisecond,
					)

					Expect(err).To(MatchError(context.DeadlineExceeded))
					Expect(adminProbe.CleanupCallCount()).To(Equal(1))
				})
			})

			Context("when the run finishes in time", func() {
				It("finishes successfully", func() {
					adminProbe.RunStub = func(ctx context.Context, logger lager.Logger, uniqueSuffix string) error {
						return cancellableComputation(ctx, 5*time.Millisecond)
					}

					err := RunAdminProbe(
						ctx,
						logger,
						adminProbe,
						10*time.Millisecond,
					)

					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Describe("Correct cleanup every time", func() {
			Context("when the run fails", func() {
				It("runs the cleanup", func() {
					adminProbe.RunReturns(someErr)

					err := RunAdminProbe(
						ctx,
						logger,
						adminProbe,
						10*time.Millisecond,
					)

					Expect(err).To(MatchError(someErr))
					Expect(adminProbe.CleanupCallCount()).To(Equal(1))
				})
			})
		})

		Describe("when cleanup fails", func() {
			It("fails", func() {
				adminProbe.CleanupReturns(someErr)

				actualErr := RunAdminProbe(
					ctx,
					logger,
					adminProbe,
					10*time.Millisecond,
				)

				Expect(actualErr).To(MatchError(someErr))
			})
		})

		Describe("when run and cleanup fail", func() {
			It("returns the run error", func() {
				adminProbe.RunReturns(someErr)
				adminProbe.CleanupReturns(someOtherErr)

				actualErr := RunAdminProbe(
					ctx,
					logger,
					adminProbe,
					10*time.Millisecond,
				)

				Expect(actualErr).To(MatchError(someErr))
			})
		})
	})
})
