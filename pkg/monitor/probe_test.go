package monitor_test

import (
	"time"

	. "code.cloudfoundry.org/perm/pkg/monitor"
	"code.cloudfoundry.org/perm/pkg/monitor/monitorfakes"
	"code.cloudfoundry.org/perm/pkg/perm"

	"context"

	"errors"

	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Probe", func() {
	var (
		subject *Probe

		client      *monitorfakes.FakeClient
		fakeLogger  *lagertest.TestLogger
		fakeContext context.Context

		uniqueSuffix string

		someError error
	)

	BeforeEach(func() {
		client = new(monitorfakes.FakeClient)

		fakeLogger = lagertest.NewTestLogger("probe")
		fakeContext = context.Background()

		uniqueSuffix = "foobar"

		subject = NewProbe(client)

		someError = errors.New("some-error")
	})

	Describe("Setup", func() {
		It("creates a role with a permission and assigns it to a test user", func() {
			_, err := subject.Setup(fakeContext, fakeLogger, uniqueSuffix)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.CreateRoleCallCount()).To(Equal(1))

			_, roleName, permissions := client.CreateRoleArgsForCall(0)

			Expect(roleName).To(Equal("system.probe.foobar"))

			Expect(permissions).To(HaveLen(1))
			Expect(permissions[0].Action).To(Equal("system.probe.assigned-permission.action"))
			Expect(permissions[0].ResourcePattern).To(Equal("system.probe.assigned-permission.resource"))

			Expect(client.AssignRoleCallCount()).To(Equal(1))

			_, roleName, actor := client.AssignRoleArgsForCall(0)

			Expect(roleName).To(Equal("system.probe.foobar"))
			Expect(actor.ID).To(Equal("probe"))
			Expect(actor.Namespace).To(Equal("system"))
		})

		It("records the durations of both the creation and assignments", func() {
			durations, err := subject.Setup(fakeContext, fakeLogger, uniqueSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(durations)).To(Equal(2))
		})

		Context("when the timeout deadline is exceeded", func() {
			It("respects the timeout and exits with an error", func() {
				contextWithExceededDeadline, cancelFunc := context.WithTimeout(context.Background(), time.Second)
				cancelFunc()
				client.CreateRoleStub = func(ctx context.Context, roleName string, permissions ...perm.Permission) (perm.Role, error) {
					time.Sleep(10 * time.Millisecond)
					return perm.Role{
						Name: roleName,
					}, nil
				}

				_, err := subject.Setup(contextWithExceededDeadline, fakeLogger, uniqueSuffix)
				Expect(err).To(MatchError("context canceled"))
			})
		})

		Context("when creating the role", func() {
			Context("when the role already exists", func() {
				BeforeEach(func() {
					client.CreateRoleReturns(perm.Role{}, perm.ErrRoleAlreadyExists)
				})

				It("swallows the error", func() {
					_, err := subject.Setup(fakeContext, fakeLogger, uniqueSuffix)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when any other error occurs", func() {
				BeforeEach(func() {
					client.CreateRoleReturns(perm.Role{}, errors.New("other error"))
				})

				It("errors", func() {
					_, err := subject.Setup(fakeContext, fakeLogger, uniqueSuffix)
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Context("when assigning the role", func() {
			Context("when the role has already been assigned", func() {
				BeforeEach(func() {
					client.AssignRoleReturns(perm.ErrAssignmentAlreadyExists)
				})

				It("swallows the error", func() {
					_, err := subject.Setup(fakeContext, fakeLogger, uniqueSuffix)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when any other error occurs", func() {
				BeforeEach(func() {
					client.AssignRoleReturns(errors.New("other error"))
				})

				It("errors", func() {
					_, err := subject.Setup(fakeContext, fakeLogger, uniqueSuffix)
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})

	Describe("Cleanup", func() {
		It("deletes the role and returns the time to complete", func() {
			durations, err := subject.Cleanup(fakeContext, time.Second, fakeLogger, uniqueSuffix)

			Expect(err).NotTo(HaveOccurred())
			Expect(len(durations)).To(Equal(1))

			Expect(client.DeleteRoleCallCount()).To(Equal(1))

			_, roleName := client.DeleteRoleArgsForCall(0)

			Expect(roleName).To(Equal("system.probe.foobar"))
		})

		Context("when the context timeout is exceeded", func() {
			It("respects the timeout and exits with an error", func() {
				contextWithExceededDeadline, cancelFunc := context.WithTimeout(context.Background(), time.Second)
				cancelFunc()

				client.DeleteRoleStub = func(ctx context.Context, roleName string) error {
					time.Sleep(20 * time.Millisecond)
					return nil
				}

				_, err := subject.Cleanup(contextWithExceededDeadline, time.Second, fakeLogger, uniqueSuffix)
				Expect(err).To(MatchError("context canceled"))
			})
		})

		Context("when cleanup timeout is exceeded", func() {
			It("respects the timeout and exits with an error", func() {
				contextWithExceededDeadline, cancelFunc := context.WithTimeout(context.Background(), time.Second)
				cancelFunc()

				client.DeleteRoleStub = func(ctx context.Context, roleNames string) error {
					time.Sleep(time.Second)
					return nil
				}

				_, err := subject.Cleanup(contextWithExceededDeadline, time.Duration(1*time.Millisecond), fakeLogger, uniqueSuffix)
				Expect(err).To(MatchError("context deadline exceeded"))
			})
		})

		Context("when the role doesn't exist", func() {
			BeforeEach(func() {
				client.DeleteRoleReturns(perm.ErrRoleNotFound)
			})

			It("swallows the error", func() {
				_, err := subject.Cleanup(fakeContext, time.Second, fakeLogger, uniqueSuffix)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when any other error occurs", func() {
			BeforeEach(func() {
				client.DeleteRoleReturns(errors.New("other error"))
			})

			It("errors", func() {
				_, err := subject.Cleanup(fakeContext, time.Second, fakeLogger, uniqueSuffix)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Run", func() {
		BeforeEach(func() {
			client.HasPermissionReturnsOnCall(0, true, nil)
			client.HasPermissionReturnsOnCall(1, false, nil)
		})

		It("asks if the actor has a permission it should have, and a permission it shouldn't", func() {
			correct, durations, err := subject.Run(fakeContext, fakeLogger, uniqueSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(correct).To(BeTrue())
			Expect(durations).To(HaveLen(2))

			Expect(client.HasPermissionCallCount()).To(Equal(2))

			_, actor, action, resource := client.HasPermissionArgsForCall(0)
			Expect(actor.ID).To(Equal("probe"))
			Expect(actor.Namespace).To(Equal("system"))
			Expect(action).To(Equal("system.probe.assigned-permission.action"))
			Expect(resource).To(Equal("system.probe.assigned-permission.resource"))

			_, actor, action, resource = client.HasPermissionArgsForCall(1)
			Expect(actor.ID).To(Equal("probe"))
			Expect(actor.Namespace).To(Equal("system"))
			Expect(action).To(Equal("system.probe.unassigned-permission.action"))
			Expect(resource).To(Equal("system.probe.unassigned-permission.resource"))
		})

		Context("when the timeout deadline is exceeded", func() {
			It("respects the timeout and exits with an error", func() {
				contextWithExceededDeadline, cancelFunc := context.WithTimeout(context.Background(), time.Second)
				cancelFunc()

				client.HasPermissionStub = func(ctx context.Context, actor perm.Actor, action, resource string) (bool, error) {
					time.Sleep(10 * time.Millisecond)
					return false, nil
				}

				_, _, err := subject.Run(contextWithExceededDeadline, fakeLogger, uniqueSuffix)
				Expect(err).To(MatchError("context canceled"))
			})
		})

		Context("when checking for the permission it should have errors", func() {
			BeforeEach(func() {
				client.HasPermissionReturnsOnCall(0, false, someError)
			})

			It("errors and does not ask for the next permission", func() {
				_, durations, err := subject.Run(fakeContext, fakeLogger, uniqueSuffix)
				Expect(err).To(MatchError(someError))
				Expect(durations).To(HaveLen(1))

				Expect(client.HasPermissionCallCount()).To(Equal(1))
			})
		})

		Context("when checking for the permission it should have returns that the actor doesn't have permission", func() {
			BeforeEach(func() {
				client.HasPermissionReturnsOnCall(0, false, nil)
			})

			It("returns that it's incorrect and does not ask for the next permission", func() {
				correct, durations, err := subject.Run(fakeContext, fakeLogger, uniqueSuffix)
				Expect(err).NotTo(HaveOccurred())
				Expect(correct).To(BeFalse())
				Expect(durations).To(HaveLen(1))

				Expect(client.HasPermissionCallCount()).To(Equal(1))
			})
		})

		Context("when checking for the permission it should not have errors", func() {
			BeforeEach(func() {
				client.HasPermissionReturnsOnCall(1, false, someError)
			})

			It("errors", func() {
				_, durations, err := subject.Run(fakeContext, fakeLogger, uniqueSuffix)
				Expect(err).To(MatchError(someError))
				Expect(durations).To(HaveLen(2))

				Expect(client.HasPermissionCallCount()).To(Equal(2))
			})
		})

		Context("when checking for the permission it should not have returns that the actor does have permission", func() {
			BeforeEach(func() {
				client.HasPermissionReturnsOnCall(1, true, nil)
			})

			It("returns that it's incorrect", func() {
				correct, durations, err := subject.Run(fakeContext, fakeLogger, uniqueSuffix)
				Expect(err).NotTo(HaveOccurred())
				Expect(correct).To(BeFalse())
				Expect(durations).To(HaveLen(2))

				Expect(client.HasPermissionCallCount()).To(Equal(2))
			})
		})
	})
})
