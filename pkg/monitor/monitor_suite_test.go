package monitor_test

import (
	"context"
	"errors"
	"time"

	"code.cloudfoundry.org/clock/fakeclock"
	. "code.cloudfoundry.org/perm/pkg/monitor"
	"code.cloudfoundry.org/perm/pkg/monitor/monitorfakes"
	"code.cloudfoundry.org/perm/pkg/perm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMonitor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Monitor Suite")
}

func testProbe(expectedTimeout time.Duration, expectedCleanuptTimeout time.Duration, allowedLatency time.Duration, opts ...Option) {
	var (
		client *monitorfakes.FakeClient

		subject *Probe
	)

	BeforeEach(func() {
		client = new(monitorfakes.FakeClient)

		client.HasPermissionReturnsOnCall(0, true, nil)
		client.HasPermissionReturnsOnCall(1, false, nil)

		subject = NewProbe(client, opts...)
	})

	Describe("#Run", func() {
		It("runs through the basic functionality", func() {
			err := subject.Run()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.CreateRoleCallCount()).To(Equal(1))

			_, roleName, permissions := client.CreateRoleArgsForCall(0)
			Expect(client.AssignRoleCallCount()).To(Equal(1))
			Expect(permissions).To(HaveLen(1))

			permission := permissions[0]

			_, assignedRole, assignedActor := client.AssignRoleArgsForCall(0)
			Expect(assignedRole).To(Equal(roleName))

			Expect(client.HasPermissionCallCount()).To(Equal(2))

			_, actorWithPermission, action, resource := client.HasPermissionArgsForCall(0)
			Expect(actorWithPermission).To(Equal(assignedActor))
			Expect(perm.Permission{
				Action:          action,
				ResourcePattern: resource,
			}).To(Equal(permission))

			_, actorWithoutPermission, action, resource := client.HasPermissionArgsForCall(1)
			Expect(actorWithoutPermission).NotTo(Equal(assignedActor))
			Expect(perm.Permission{
				Action:          action,
				ResourcePattern: resource,
			}).To(Equal(permission))

			Expect(client.UnassignRoleCallCount()).To(Equal(1))
			_, unassignedRole, unassignedActor := client.UnassignRoleArgsForCall(0)
			Expect(unassignedRole).To(Equal(roleName))
			Expect(unassignedActor).To(Equal(assignedActor))

			Expect(client.DeleteRoleCallCount()).To(Equal(1))
			_, deletedRoleName := client.DeleteRoleArgsForCall(0)
			Expect(deletedRoleName).To(Equal(roleName))
		})

		It("uses the correct timeout for all calls", func() {
			start := time.Now()

			err := subject.Run()
			Expect(err).NotTo(HaveOccurred())

			end := time.Now()

			ctx, _, _ := client.CreateRoleArgsForCall(0)
			deadline, ok := ctx.Deadline()

			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedTimeout)))

			ctx, _, _ = client.AssignRoleArgsForCall(0)
			deadline, ok = ctx.Deadline()

			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedTimeout)))

			ctx, _, _, _ = client.HasPermissionArgsForCall(0)
			deadline, ok = ctx.Deadline()

			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedTimeout)))

			ctx, _, _, _ = client.HasPermissionArgsForCall(1)
			deadline, ok = ctx.Deadline()

			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedTimeout)))

			ctx, _, _ = client.UnassignRoleArgsForCall(0)
			deadline, ok = ctx.Deadline()

			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedTimeout)))

			ctx, _ = client.DeleteRoleArgsForCall(0)
			deadline, ok = ctx.Deadline()

			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedTimeout)))
		})

		It("uses a unique role each time", func() {
			err := subject.Run()
			Expect(err).NotTo(HaveOccurred())

			_, firstRole, _ := client.CreateRoleArgsForCall(0)

			client.HasPermissionReturnsOnCall(2, true, nil)
			client.HasPermissionReturnsOnCall(3, false, nil)

			err = subject.Run()
			Expect(err).NotTo(HaveOccurred())

			_, secondRole, _ := client.CreateRoleArgsForCall(1)

			Expect(firstRole).NotTo(Equal(secondRole))
		})

		It("uses a unique permission each time", func() {
			err := subject.Run()
			Expect(err).NotTo(HaveOccurred())

			_, _, firstPermissions := client.CreateRoleArgsForCall(0)
			firstPermission := firstPermissions[0]

			client.HasPermissionReturnsOnCall(2, true, nil)
			client.HasPermissionReturnsOnCall(3, false, nil)

			err = subject.Run()
			Expect(err).NotTo(HaveOccurred())

			_, _, secondPermissions := client.CreateRoleArgsForCall(1)
			secondPermission := secondPermissions[0]

			Expect(firstPermission).NotTo(Equal(secondPermission))
		})

		It("runs all other calls but returns an error if CreateRole takes an unacceptable amount of time", func() {
			now := time.Now()
			clock := fakeclock.NewFakeClock(now)

			subject = NewProbe(client, append(opts, WithClock(clock))...)

			client.CreateRoleStub = func(context.Context, string, ...perm.Permission) (perm.Role, error) {
				clock.Increment(allowedLatency * 2)
				return perm.Role{}, nil
			}

			err := subject.Run()
			Expect(err).To(MatchError(ErrExceededMaxLatency))

			Expect(client.CreateRoleCallCount()).To(Equal(1))
			Expect(client.AssignRoleCallCount()).To(Equal(1))
			Expect(client.HasPermissionCallCount()).To(Equal(2))
			Expect(client.UnassignRoleCallCount()).To(Equal(1))
			Expect(client.DeleteRoleCallCount()).To(Equal(1))
		})

		It("runs all other calls but returns an error if AssignRole takes an unacceptable amount of time", func() {
			now := time.Now()
			clock := fakeclock.NewFakeClock(now)

			subject = NewProbe(client, append(opts, WithClock(clock))...)

			client.AssignRoleStub = func(context.Context, string, perm.Actor) error {
				clock.Increment(allowedLatency * 2)
				return nil
			}

			err := subject.Run()
			Expect(err).To(MatchError(ErrExceededMaxLatency))

			Expect(client.CreateRoleCallCount()).To(Equal(1))
			Expect(client.AssignRoleCallCount()).To(Equal(1))
			Expect(client.HasPermissionCallCount()).To(Equal(2))
			Expect(client.UnassignRoleCallCount()).To(Equal(1))
			Expect(client.DeleteRoleCallCount()).To(Equal(1))
		})

		It("runs all other calls but returns an error if the first HasPermission call takes an unacceptable amount of time", func() {
			now := time.Now()
			clock := fakeclock.NewFakeClock(now)

			subject = NewProbe(client, append(opts, WithClock(clock))...)

			var called bool
			client.HasPermissionStub = func(context.Context, perm.Actor, string, string) (bool, error) {
				if !called {
					called = true
					clock.Increment(allowedLatency * 2)
					return true, nil
				} else {
					return false, nil
				}
			}

			err := subject.Run()
			Expect(err).To(MatchError(ErrExceededMaxLatency))

			Expect(client.CreateRoleCallCount()).To(Equal(1))
			Expect(client.AssignRoleCallCount()).To(Equal(1))
			Expect(client.HasPermissionCallCount()).To(Equal(2))
			Expect(client.UnassignRoleCallCount()).To(Equal(1))
			Expect(client.DeleteRoleCallCount()).To(Equal(1))
		})

		It("runs all other calls but returns an error if the second HasPermission call takes an unacceptable amount of time", func() {
			now := time.Now()
			clock := fakeclock.NewFakeClock(now)

			subject = NewProbe(client, append(opts, WithClock(clock))...)

			var called bool
			client.HasPermissionStub = func(context.Context, perm.Actor, string, string) (bool, error) {
				if !called {
					called = true
					return true, nil
				} else {
					clock.Increment(allowedLatency * 2)
					return false, nil
				}
			}

			err := subject.Run()
			Expect(err).To(MatchError(ErrExceededMaxLatency))

			Expect(client.CreateRoleCallCount()).To(Equal(1))
			Expect(client.AssignRoleCallCount()).To(Equal(1))
			Expect(client.HasPermissionCallCount()).To(Equal(2))
			Expect(client.UnassignRoleCallCount()).To(Equal(1))
			Expect(client.DeleteRoleCallCount()).To(Equal(1))
		})

		It("runs all other calls but returns an error if UnassignRole takes an unacceptable amount of time", func() {
			now := time.Now()
			clock := fakeclock.NewFakeClock(now)

			subject = NewProbe(client, append(opts, WithClock(clock))...)

			client.UnassignRoleStub = func(context.Context, string, perm.Actor) error {
				clock.Increment(allowedLatency * 2)
				return nil
			}

			err := subject.Run()
			Expect(err).To(MatchError(ErrExceededMaxLatency))

			Expect(client.CreateRoleCallCount()).To(Equal(1))
			Expect(client.AssignRoleCallCount()).To(Equal(1))
			Expect(client.HasPermissionCallCount()).To(Equal(2))
			Expect(client.UnassignRoleCallCount()).To(Equal(1))
			Expect(client.DeleteRoleCallCount()).To(Equal(1))
		})

		It("runs all other calls but returns an error if DeleteRole takes an unacceptable amount of time", func() {
			now := time.Now()
			clock := fakeclock.NewFakeClock(now)

			subject = NewProbe(client, append(opts, WithClock(clock))...)

			client.DeleteRoleStub = func(context.Context, string) error {
				clock.Increment(allowedLatency * 2)
				return nil
			}

			err := subject.Run()
			Expect(err).To(MatchError(ErrExceededMaxLatency))

			Expect(client.CreateRoleCallCount()).To(Equal(1))
			Expect(client.AssignRoleCallCount()).To(Equal(1))
			Expect(client.HasPermissionCallCount()).To(Equal(2))
			Expect(client.UnassignRoleCallCount()).To(Equal(1))
			Expect(client.DeleteRoleCallCount()).To(Equal(1))
		})

		It("stops early and attempts to cleanup if CreateRole fails", func() {
			start := time.Now()

			createRoleErr := errors.New("error")

			client.CreateRoleReturns(perm.Role{}, createRoleErr)

			err := subject.Run()
			Expect(err).To(Equal(createRoleErr))

			end := time.Now()

			_, roleName, _ := client.CreateRoleArgsForCall(0)

			Expect(client.AssignRoleCallCount()).To(Equal(0))
			Expect(client.HasPermissionCallCount()).To(Equal(0))
			Expect(client.UnassignRoleCallCount()).To(Equal(0))

			Expect(client.DeleteRoleCallCount()).To(Equal(1))

			ctx, deletedRoleName := client.DeleteRoleArgsForCall(0)
			Expect(deletedRoleName).To(Equal(roleName))

			deadline, ok := ctx.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedCleanuptTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedCleanuptTimeout)))
		})

		It("stops early and attempts to cleanup if AssignRole fails", func() {
			start := time.Now()

			assignRoleErr := errors.New("error")

			client.AssignRoleReturns(assignRoleErr)

			err := subject.Run()
			Expect(err).To(Equal(assignRoleErr))

			end := time.Now()

			_, roleName, _ := client.CreateRoleArgsForCall(0)

			Expect(client.HasPermissionCallCount()).To(Equal(0))
			Expect(client.UnassignRoleCallCount()).To(Equal(0))

			Expect(client.DeleteRoleCallCount()).To(Equal(1))

			ctx, deletedRoleName := client.DeleteRoleArgsForCall(0)
			Expect(deletedRoleName).To(Equal(roleName))

			deadline, ok := ctx.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedCleanuptTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedCleanuptTimeout)))
		})

		It("stops early and attempts to cleanup if the first HasPermission call fails", func() {
			start := time.Now()

			hasPermissionErr := errors.New("error")

			client.HasPermissionReturnsOnCall(0, true, hasPermissionErr)

			err := subject.Run()
			Expect(err).To(Equal(hasPermissionErr))

			end := time.Now()

			_, roleName, _ := client.CreateRoleArgsForCall(0)

			Expect(client.HasPermissionCallCount()).To(Equal(1))
			Expect(client.UnassignRoleCallCount()).To(Equal(0))

			Expect(client.DeleteRoleCallCount()).To(Equal(1))

			ctx, deletedRoleName := client.DeleteRoleArgsForCall(0)
			Expect(deletedRoleName).To(Equal(roleName))

			deadline, ok := ctx.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedCleanuptTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedCleanuptTimeout)))
		})

		It("stops early and attempts to cleanup if the second HasPermission call fails", func() {
			start := time.Now()

			hasPermissionErr := errors.New("error")

			client.HasPermissionReturnsOnCall(1, false, hasPermissionErr)

			err := subject.Run()
			Expect(err).To(Equal(hasPermissionErr))

			end := time.Now()

			_, roleName, _ := client.CreateRoleArgsForCall(0)

			Expect(client.UnassignRoleCallCount()).To(Equal(0))

			Expect(client.DeleteRoleCallCount()).To(Equal(1))

			ctx, deletedRoleName := client.DeleteRoleArgsForCall(0)
			Expect(deletedRoleName).To(Equal(roleName))

			deadline, ok := ctx.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedCleanuptTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedCleanuptTimeout)))
		})

		It("stops early and attempts to cleanup if UnassignRole fails", func() {
			start := time.Now()

			unassignRoleErr := errors.New("error")

			client.UnassignRoleReturns(unassignRoleErr)

			err := subject.Run()
			Expect(err).To(Equal(unassignRoleErr))

			end := time.Now()

			_, roleName, _ := client.CreateRoleArgsForCall(0)

			Expect(client.DeleteRoleCallCount()).To(Equal(1))

			ctx, deletedRoleName := client.DeleteRoleArgsForCall(0)
			Expect(deletedRoleName).To(Equal(roleName))

			deadline, ok := ctx.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedCleanuptTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedCleanuptTimeout)))
		})

		It("stops and attempts to cleanup if DeleteRole fails", func() {
			start := time.Now()

			deleteRoleErr := errors.New("error")

			client.DeleteRoleReturnsOnCall(0, deleteRoleErr)

			err := subject.Run()
			Expect(err).To(Equal(deleteRoleErr))

			end := time.Now()

			_, roleName, _ := client.CreateRoleArgsForCall(0)

			Expect(client.DeleteRoleCallCount()).To(Equal(2))

			ctx, deletedRoleName := client.DeleteRoleArgsForCall(1)
			Expect(deletedRoleName).To(Equal(roleName))

			deadline, ok := ctx.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedCleanuptTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedCleanuptTimeout)))
		})

		It("stops early and attempts to cleanup if the first HasPermission has an incorrect result", func() {
			start := time.Now()

			client.HasPermissionReturnsOnCall(0, false, nil)

			err := subject.Run()
			Expect(err).To(MatchError("probe: incorrect HasPermission result"))

			end := time.Now()

			_, roleName, _ := client.CreateRoleArgsForCall(0)

			Expect(client.HasPermissionCallCount()).To(Equal(1))
			Expect(client.UnassignRoleCallCount()).To(Equal(0))

			Expect(client.DeleteRoleCallCount()).To(Equal(1))

			ctx, deletedRoleName := client.DeleteRoleArgsForCall(0)
			Expect(deletedRoleName).To(Equal(roleName))

			deadline, ok := ctx.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedCleanuptTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedCleanuptTimeout)))
		})

		It("stops early and attempts to cleanup if the second HasPermission has an incorrect result", func() {
			start := time.Now()

			client.HasPermissionReturnsOnCall(1, true, nil)

			err := subject.Run()
			Expect(err).To(MatchError("probe: incorrect HasPermission result"))

			end := time.Now()

			_, roleName, _ := client.CreateRoleArgsForCall(0)

			Expect(client.UnassignRoleCallCount()).To(Equal(0))

			Expect(client.DeleteRoleCallCount()).To(Equal(1))

			ctx, deletedRoleName := client.DeleteRoleArgsForCall(0)
			Expect(deletedRoleName).To(Equal(roleName))

			deadline, ok := ctx.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedCleanuptTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedCleanuptTimeout)))
		})
	})
}
