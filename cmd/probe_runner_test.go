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

var _ = Describe("Running the Probes", func() {
	Describe("RunQueryProbe", func() {
		var (
			queryProbe *cmdfakes.FakeQueryProbe

			ctx    context.Context
			logger *lagertest.TestLogger

			timeout time.Duration

			expectedDurations []time.Duration
			someErr           error
		)

		BeforeEach(func() {
			queryProbe = new(cmdfakes.FakeQueryProbe)

			ctx = context.Background()
			logger = lagertest.NewTestLogger("run-query-probe")

			timeout = 5 * time.Second

			someErr = errors.New("some-error")

			expectedDurations = []time.Duration{1 * time.Second, 2 * time.Second}
			queryProbe.RunReturns(true, expectedDurations, nil)
		})

		It("calls the setup, run, and cleanup with a uuid", func() {
			correct, durations, err := RunQueryProbe(ctx, logger, queryProbe, timeout)

			Expect(queryProbe.SetupCallCount()).To(Equal(1))
			Expect(queryProbe.RunCallCount()).To(Equal(1))
			Expect(queryProbe.CleanupCallCount()).To(Equal(1))

			Expect(err).NotTo(HaveOccurred())
			Expect(correct).To(BeTrue())
			Expect(durations).To(Equal(expectedDurations))
		})

		Context("timeouts", func() {
			It("errors if it times out", func() {
				queryProbe.RunStub = func(ctx context.Context, logger lager.Logger, uniqueSuffix string) (bool, []time.Duration, error) {
					done := make(chan struct{}, 1)

					go func() {
						time.Sleep(10 * time.Millisecond)
						// Run completed successfully
						close(done)
					}()
					select {
					case <-ctx.Done():
						// Context was cancelled before computation succeeded
						// i.e. computation was cancelled
						return false, nil, ctx.Err()
					case <-done:
						// Long computation succeeded
						return true, nil, nil
					}
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
					done := make(chan struct{}, 1)

					go func() {
						time.Sleep(10 * time.Millisecond)
						// Run completed successfully
						close(done)
					}()
					select {
					case <-ctx.Done():
						// Context was cancelled before computation succeeded
						// i.e. computation was cancelled
						return false, nil, ctx.Err()
					case <-done:
						// Long computation succeeded
						return true, nil, nil
					}
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
			var (
				cleanupErr error
			)

			BeforeEach(func() {
				cleanupErr = errors.New("cleanup-error")

				queryProbe.SetupReturns(someErr)
				queryProbe.CleanupReturns(cleanupErr)
			})

			It("returns the setup error", func() {
				_, _, err := RunQueryProbe(ctx, logger, queryProbe, timeout)

				Expect(err).To(MatchError(someErr))
				Expect(err).NotTo(MatchError(cleanupErr))
			})
		})

		Context("when run and cleanup fail", func() {
			var (
				cleanupErr error
			)

			BeforeEach(func() {
				cleanupErr = errors.New("cleanup-error")

				queryProbe.RunReturns(false, nil, someErr)
				queryProbe.CleanupReturns(cleanupErr)
			})

			It("returns the run error", func() {
				_, _, err := RunQueryProbe(ctx, logger, queryProbe, timeout)

				Expect(err).To(MatchError(someErr))
				Expect(err).NotTo(MatchError(cleanupErr))
			})
		})
	})
})
