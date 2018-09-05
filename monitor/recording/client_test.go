package recording_test

import (
	"context"
	"errors"
	"time"

	"code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/perm"
	. "code.cloudfoundry.org/perm/monitor/recording"
	"code.cloudfoundry.org/perm/monitor/recording/recordingfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
)

var _ = Describe("Client", func() {
	var (
		now time.Time

		fakeClient   *recordingfakes.FakeClient
		fakeRecorder *recordingfakes.FakeRecorder
		fakeClock    *fakeclock.FakeClock

		subject *RecordingClient

		roleName    string
		actor       perm.Actor
		permissions []perm.Permission
		action      string
		resource    string

		ctx          context.Context
		testDuration time.Duration
	)

	BeforeEach(func() {
		now = time.Now()

		fakeClient = new(recordingfakes.FakeClient)
		fakeRecorder = new(recordingfakes.FakeRecorder)
		fakeClock = fakeclock.NewFakeClock(now)

		subject = NewClient(fakeClient, fakeRecorder, WithClock(fakeClock))

		roleName = uuid.NewV4().String()
		actor = perm.Actor{
			ID:        uuid.NewV4().String(),
			Namespace: uuid.NewV4().String(),
		}
		permissions = []perm.Permission{
			{
				Action:          uuid.NewV4().String(),
				ResourcePattern: uuid.NewV4().String(),
			},
		}
		action = uuid.NewV4().String()
		resource = uuid.NewV4().String()

		ctx = context.Background()
		testDuration = time.Millisecond * 10
	})

	Describe("#AssignRole", func() {
		Context("when no errors are encountered from AssignRole", func() {
			BeforeEach(func() {
				fakeClient.AssignRoleStub = func(context.Context, string, perm.Actor) error {
					fakeClock.Increment(testDuration)
					return nil
				}
			})

			It("records the duration of the call", func() {
				duration, err := subject.AssignRole(ctx, roleName, actor)
				Expect(err).NotTo(HaveOccurred())
				Expect(duration).To(Equal(testDuration))

				Expect(fakeRecorder.ObserveCallCount()).To(Equal(1))
				Expect(fakeRecorder.ObserveArgsForCall(0)).To(Equal(testDuration))
			})

			Context("when an error is encountered recording the duration", func() {
				It("returns the error and the duration", func() {
					observeErr := errors.New("test err")
					fakeRecorder.ObserveReturns(observeErr)

					duration, err := subject.AssignRole(ctx, roleName, actor)
					Expect(err).To(MatchError(FailedToObserveDurationError{Err: observeErr}))
					Expect(duration).To(Equal(testDuration))
				})
			})
		})

		Context("when an error is encountered from AssignRole", func() {
			It("returns the error and does not record the duration of the call", func() {
				returnedErr := errors.New("AssignRole error")
				fakeClient.AssignRoleReturns(returnedErr)

				duration, err := subject.AssignRole(ctx, roleName, actor)
				Expect(err).To(MatchError(returnedErr))
				Expect(duration).To(BeZero())

				Expect(fakeRecorder.ObserveCallCount()).To(Equal(0))
			})
		})
	})

	Describe("#CreateRole", func() {
		Context("when no errors are encountered from CreateRole", func() {
			BeforeEach(func() {
				fakeClient.CreateRoleStub = func(context.Context, string, ...perm.Permission) (perm.Role, error) {
					fakeClock.Increment(testDuration)
					return perm.Role{}, nil
				}
			})

			It("records the duration of the call", func() {
				_, duration, err := subject.CreateRole(ctx, roleName, permissions...)
				Expect(err).NotTo(HaveOccurred())
				Expect(duration).To(Equal(testDuration))

				Expect(fakeRecorder.ObserveCallCount()).To(Equal(1))
				Expect(fakeRecorder.ObserveArgsForCall(0)).To(Equal(testDuration))
			})

			Context("when an error is encountered recording the duration", func() {
				It("returns the error and the duration", func() {
					observeErr := errors.New("test err")
					fakeRecorder.ObserveReturns(observeErr)

					_, duration, err := subject.CreateRole(ctx, roleName, permissions...)
					Expect(err).To(MatchError(FailedToObserveDurationError{Err: observeErr}))
					Expect(duration).To(Equal(testDuration))
				})
			})
		})

		Context("when an error is encountered from CreateRole", func() {
			It("returns the error and does not record the duration of the call", func() {
				returnedErr := errors.New("CreateRole error")
				fakeClient.CreateRoleReturns(perm.Role{}, returnedErr)

				_, duration, err := subject.CreateRole(ctx, roleName, permissions...)
				Expect(err).To(MatchError(returnedErr))
				Expect(duration).To(BeZero())

				Expect(fakeRecorder.ObserveCallCount()).To(Equal(0))
			})
		})
	})

	Describe("#DeleteRole", func() {
		Context("when no errors are encountered from DeleteRole", func() {
			BeforeEach(func() {
				fakeClient.DeleteRoleStub = func(context.Context, string) error {
					fakeClock.Increment(testDuration)
					return nil
				}
			})

			It("records the duration of the call", func() {
				duration, err := subject.DeleteRole(ctx, roleName)
				Expect(err).NotTo(HaveOccurred())
				Expect(duration).To(Equal(testDuration))

				Expect(fakeRecorder.ObserveCallCount()).To(Equal(1))
				Expect(fakeRecorder.ObserveArgsForCall(0)).To(Equal(testDuration))
			})

			Context("when an error is encountered recording the duration", func() {
				It("returns the error and the duration", func() {
					observeErr := errors.New("test err")
					fakeRecorder.ObserveReturns(observeErr)

					duration, err := subject.DeleteRole(ctx, roleName)
					Expect(err).To(MatchError(FailedToObserveDurationError{Err: observeErr}))
					Expect(duration).To(Equal(testDuration))
				})
			})
		})

		Context("when an error is encountered from DeleteRole", func() {
			It("returns the error and does not record the duration of the call", func() {
				returnedErr := errors.New("DeleteRole error")
				fakeClient.DeleteRoleReturns(returnedErr)

				duration, err := subject.DeleteRole(ctx, roleName)
				Expect(err).To(MatchError(returnedErr))
				Expect(duration).To(BeZero())

				Expect(fakeRecorder.ObserveCallCount()).To(Equal(0))
			})
		})
	})

	Describe("#HasPermission", func() {
		Context("when no errors are encountered from HasPermission", func() {
			BeforeEach(func() {
				fakeClient.HasPermissionStub = func(context.Context, perm.Actor, string, string) (bool, error) {
					fakeClock.Increment(testDuration)
					return false, nil
				}
			})

			It("records the duration of the call", func() {
				_, duration, err := subject.HasPermission(ctx, actor, action, resource)
				Expect(err).ToNot(HaveOccurred())
				Expect(duration).To(Equal(testDuration))

				Expect(fakeRecorder.ObserveCallCount()).To(Equal(1))
				Expect(fakeRecorder.ObserveArgsForCall(0)).To(Equal(testDuration))
			})

			Context("when an error is encountered recording the duration", func() {
				It("returns the error and the duration", func() {
					observeErr := errors.New("test err")
					fakeRecorder.ObserveReturns(observeErr)

					_, duration, err := subject.HasPermission(ctx, actor, action, resource)
					Expect(err).To(MatchError(FailedToObserveDurationError{Err: observeErr}))
					Expect(duration).To(Equal(testDuration))
				})
			})
		})

		Context("when an error is encountered from HasPermission", func() {
			It("returns the error and does not record the duration of the call", func() {
				returnedErr := errors.New("HasPermission error")
				fakeClient.HasPermissionReturns(false, returnedErr)

				_, duration, err := subject.HasPermission(ctx, actor, action, resource)
				Expect(err).To(MatchError(returnedErr))
				Expect(duration).To(BeZero())

				Expect(fakeRecorder.ObserveCallCount()).To(Equal(0))
			})
		})
	})

	Describe("#UnassignRole", func() {
		Context("when no errors are encountered from UnassignRole", func() {
			BeforeEach(func() {
				fakeClient.UnassignRoleStub = func(context.Context, string, perm.Actor) error {
					fakeClock.Increment(testDuration)
					return nil
				}
			})

			It("records the duration of the call", func() {
				duration, err := subject.UnassignRole(ctx, roleName, actor)
				Expect(err).NotTo(HaveOccurred())
				Expect(duration).To(Equal(testDuration))

				Expect(fakeRecorder.ObserveCallCount()).To(Equal(1))
				Expect(fakeRecorder.ObserveArgsForCall(0)).To(Equal(testDuration))
			})

			Context("when an error is encountered recording the duration", func() {
				It("returns the error and the duration", func() {
					observeErr := errors.New("test err")
					fakeRecorder.ObserveReturns(observeErr)

					duration, err := subject.UnassignRole(ctx, roleName, actor)
					Expect(err).To(MatchError(FailedToObserveDurationError{Err: observeErr}))
					Expect(duration).To(Equal(testDuration))
				})
			})
		})

		Context("when an error is encountered from UnassignRole", func() {
			It("returns the error and does not record the duration of the call", func() {
				returnedErr := errors.New("UnassignRole error")
				fakeClient.UnassignRoleReturns(returnedErr)

				duration, err := subject.UnassignRole(ctx, roleName, actor)
				Expect(err).To(MatchError(returnedErr))
				Expect(duration).To(BeZero())

				Expect(fakeRecorder.ObserveCallCount()).To(Equal(0))
			})
		})
	})
})
