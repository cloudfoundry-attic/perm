package recording_test

import (
	"context"
	"errors"
	"time"

	"code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/perm/pkg/monitor/monitorfakes"
	. "code.cloudfoundry.org/perm/pkg/monitor/recording"
	"code.cloudfoundry.org/perm/pkg/monitor/recording/recordingfakes"
	"code.cloudfoundry.org/perm/pkg/perm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
)

var _ = Describe("Client", func() {
	var (
		now time.Time

		fakeClient   *monitorfakes.FakeClient
		fakeRecorder *recordingfakes.FakeDurationRecorder
		fakeClock    *fakeclock.FakeClock

		subject *Client

		roleName    string
		actor       perm.Actor
		permissions []perm.Permission
		action      string
		resource    string

		ctx context.Context
	)

	BeforeEach(func() {
		now = time.Now()

		fakeClient = new(monitorfakes.FakeClient)
		fakeRecorder = new(recordingfakes.FakeDurationRecorder)
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
	})

	Describe("#AssignRole", func() {
		Context("when no errors are encountered", func() {
			BeforeEach(func() {
				fakeClient.AssignRoleStub = func(context.Context, string, perm.Actor) error {
					fakeClock.Increment(time.Second * 5)
					return nil
				}
			})

			It("should record the duration of the call", func() {
				err := subject.AssignRole(ctx, roleName, actor)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeRecorder.ObserveCallCount()).To(Equal(1))

				duration := fakeRecorder.ObserveArgsForCall(0)
				Expect(duration).To(Equal(time.Second * 5))
			})

			It("returns an error if recording fails", func() {
				observeErr := errors.New("test err")
				fakeRecorder.ObserveStub = func(time.Duration) error {
					return observeErr
				}

				err := subject.AssignRole(ctx, roleName, actor)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(FailedToObserveDurationError{Err: observeErr}))
			})
		})

		Context("when an error is encountered", func() {
			It("should return the error and not record the duration of the call", func() {
				returnedErr := errors.New("AssignRole error")
				fakeClient.AssignRoleReturns(returnedErr)

				err := subject.AssignRole(ctx, roleName, actor)
				Expect(err).To(MatchError(returnedErr))

				Expect(fakeRecorder.ObserveCallCount()).To(Equal(0))
			})
		})
	})

	Describe("#CreateRole", func() {
		Context("when no errors are encountered", func() {
			It("should record the duration of the call", func() {
				fakeClient.CreateRoleStub = func(context.Context, string, ...perm.Permission) (perm.Role, error) {
					fakeClock.Increment(time.Second * 5)
					return perm.Role{}, nil
				}

				_, err := subject.CreateRole(ctx, roleName, permissions...)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeRecorder.ObserveCallCount()).To(Equal(1))

				duration := fakeRecorder.ObserveArgsForCall(0)
				Expect(duration).To(Equal(duration))
			})

			It("returns an error if recording fails", func() {
				observeErr := errors.New("test err")
				fakeRecorder.ObserveStub = func(time.Duration) error {
					return observeErr
				}

				_, err := subject.CreateRole(ctx, roleName, permissions...)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(FailedToObserveDurationError{Err: observeErr}))
			})
		})

		Context("when an error is encountered", func() {
			It("should return the error and not record the duration of the call", func() {
				returnedErr := errors.New("CreateRole error")
				fakeClient.CreateRoleReturns(perm.Role{}, returnedErr)

				_, err := subject.CreateRole(ctx, roleName, permissions...)
				Expect(err).To(MatchError(returnedErr))

				Expect(fakeRecorder.ObserveCallCount()).To(Equal(0))
			})
		})
	})

	Describe("#DeleteRole", func() {
		Context("when no errors are encountered", func() {
			It("should record the duration of the call", func() {
				fakeClient.DeleteRoleStub = func(context.Context, string) error {
					fakeClock.Increment(time.Second * 5)
					return nil
				}

				err := subject.DeleteRole(ctx, roleName)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeRecorder.ObserveCallCount()).To(Equal(1))

				duration := fakeRecorder.ObserveArgsForCall(0)
				Expect(duration).To(Equal(duration))
			})

			It("returns an error if recording fails", func() {
				observeErr := errors.New("test err")
				fakeRecorder.ObserveStub = func(time.Duration) error {
					return observeErr
				}

				err := subject.DeleteRole(ctx, roleName)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(FailedToObserveDurationError{Err: observeErr}))
			})
		})

		Context("when an error is encountered", func() {
			It("should return the error and not record the duration of the call", func() {
				returnedErr := errors.New("DeleteRole error")
				fakeClient.DeleteRoleReturns(returnedErr)

				err := subject.DeleteRole(ctx, roleName)
				Expect(err).To(MatchError(returnedErr))

				Expect(fakeRecorder.ObserveCallCount()).To(Equal(0))
			})
		})
	})

	Describe("#UnassignRole", func() {
		Context("when no errors are encountered", func() {
			It("should record the duration of the call", func() {
				fakeClient.UnassignRoleStub = func(context.Context, string, perm.Actor) error {
					fakeClock.Increment(time.Second * 5)
					return nil
				}

				err := subject.UnassignRole(ctx, roleName, actor)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeRecorder.ObserveCallCount()).To(Equal(1))

				duration := fakeRecorder.ObserveArgsForCall(0)
				Expect(duration).To(Equal(duration))
			})

			It("returns an error if recording fails", func() {
				observeErr := errors.New("test err")
				fakeRecorder.ObserveStub = func(time.Duration) error {
					return observeErr
				}

				err := subject.UnassignRole(ctx, roleName, actor)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(FailedToObserveDurationError{Err: observeErr}))
			})
		})

		Context("when an error is encountered", func() {
			It("should return the error and not record the duration of the call", func() {
				returnedErr := errors.New("UnassignRole error")
				fakeClient.UnassignRoleReturns(returnedErr)

				err := subject.UnassignRole(ctx, roleName, actor)
				Expect(err).To(MatchError(returnedErr))

				Expect(fakeRecorder.ObserveCallCount()).To(Equal(0))
			})
		})
	})

	Describe("#HasPermission", func() {
		Context("when no errors are encountered", func() {
			It("should record the duration of the call", func() {
				fakeClient.HasPermissionStub = func(context.Context, perm.Actor, string, string) (bool, error) {
					fakeClock.Increment(time.Second * 5)
					return false, nil
				}

				_, _ = subject.HasPermission(ctx, actor, action, resource)

				Expect(fakeRecorder.ObserveCallCount()).To(Equal(1))

				duration := fakeRecorder.ObserveArgsForCall(0)
				Expect(duration).To(Equal(duration))
			})

			It("returns an error if recording fails", func() {
				observeErr := errors.New("test err")
				fakeRecorder.ObserveStub = func(time.Duration) error {
					return observeErr
				}

				_, err := subject.HasPermission(ctx, actor, action, resource)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(FailedToObserveDurationError{Err: observeErr}))
			})
		})

		Context("when an error is encountered", func() {
			It("should return the error and not record the duration of the call", func() {
				returnedErr := errors.New("HasPermission error")
				fakeClient.HasPermissionReturns(false, returnedErr)

				hasPermission, err := subject.HasPermission(ctx, actor, action, resource)
				Expect(hasPermission).To(BeFalse())
				Expect(err).To(MatchError(returnedErr))

				Expect(fakeRecorder.ObserveCallCount()).To(Equal(0))
			})
		})
	})
})
