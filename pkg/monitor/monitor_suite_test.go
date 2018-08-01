package monitor_test

import (
	"context"
	"errors"
	"time"

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
		delta        time.Duration
		zeroDuration time.Duration
		fakeClient   *monitorfakes.FakeClient

		subject *Probe
	)

	BeforeEach(func() {
		delta = time.Duration(10)
		zeroDuration = time.Duration(0)
		fakeClient = new(monitorfakes.FakeClient)

		fakeClient.HasPermissionReturnsOnCall(0, true, zeroDuration, nil)
		fakeClient.HasPermissionReturnsOnCall(1, false, zeroDuration, nil)

		subject = NewProbe(fakeClient, opts...)
	})

	Describe("#Run", func() {
		It("runs through the basic functionality", func() {
			err := subject.Run()
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			_, roleName, permissions := fakeClient.CreateRoleArgsForCall(0)
			Expect(permissions).To(HaveLen(1))

			permission := permissions[0]

			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			_, assignedRole, assignedActor := fakeClient.AssignRoleArgsForCall(0)
			Expect(assignedRole).To(Equal(roleName))

			Expect(fakeClient.HasPermissionCallCount()).To(Equal(2))

			_, actorWithPermission, action, resource := fakeClient.HasPermissionArgsForCall(0)
			Expect(actorWithPermission).To(Equal(assignedActor))
			Expect(perm.Permission{
				Action:          action,
				ResourcePattern: resource,
			}).To(Equal(permission))

			_, actorWithoutPermission, action, resource := fakeClient.HasPermissionArgsForCall(1)
			Expect(actorWithoutPermission).NotTo(Equal(assignedActor))
			Expect(perm.Permission{
				Action:          action,
				ResourcePattern: resource,
			}).To(Equal(permission))

			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(1))
			_, unassignedRole, unassignedActor := fakeClient.UnassignRoleArgsForCall(0)
			Expect(unassignedRole).To(Equal(roleName))
			Expect(unassignedActor).To(Equal(assignedActor))

			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))
			_, deletedRoleName := fakeClient.DeleteRoleArgsForCall(0)
			Expect(deletedRoleName).To(Equal(roleName))
		})

		It("uses the correct timeout for all calls", func() {
			start := time.Now()

			err := subject.Run()
			Expect(err).NotTo(HaveOccurred())

			end := time.Now()

			ctx, _, _ := fakeClient.CreateRoleArgsForCall(0)
			deadline, ok := ctx.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedTimeout)))

			ctx, _, _ = fakeClient.AssignRoleArgsForCall(0)
			deadline, ok = ctx.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedTimeout)))

			ctx, _, _, _ = fakeClient.HasPermissionArgsForCall(0)
			deadline, ok = ctx.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedTimeout)))

			ctx, _, _, _ = fakeClient.HasPermissionArgsForCall(1)
			deadline, ok = ctx.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedTimeout)))

			ctx, _, _ = fakeClient.UnassignRoleArgsForCall(0)
			deadline, ok = ctx.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedTimeout)))

			ctx, _ = fakeClient.DeleteRoleArgsForCall(0)
			deadline, ok = ctx.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedTimeout)))
		})

		It("uses a unique role each time", func() {
			err := subject.Run()
			Expect(err).NotTo(HaveOccurred())

			_, firstRole, _ := fakeClient.CreateRoleArgsForCall(0)

			fakeClient.HasPermissionReturnsOnCall(2, true, zeroDuration, nil)
			fakeClient.HasPermissionReturnsOnCall(3, false, zeroDuration, nil)

			err = subject.Run()
			Expect(err).NotTo(HaveOccurred())

			_, secondRole, _ := fakeClient.CreateRoleArgsForCall(1)

			Expect(firstRole).NotTo(Equal(secondRole))
		})

		It("uses a unique permission each time", func() {
			err := subject.Run()
			Expect(err).NotTo(HaveOccurred())

			_, _, firstPermissions := fakeClient.CreateRoleArgsForCall(0)
			firstPermission := firstPermissions[0]

			fakeClient.HasPermissionReturnsOnCall(2, true, zeroDuration, nil)
			fakeClient.HasPermissionReturnsOnCall(3, false, zeroDuration, nil)

			err = subject.Run()
			Expect(err).NotTo(HaveOccurred())

			_, _, secondPermissions := fakeClient.CreateRoleArgsForCall(1)
			secondPermission := secondPermissions[0]

			Expect(firstPermission).NotTo(Equal(secondPermission))
		})

		It("runs all other calls but returns an error if CreateRole takes an unacceptable amount of time", func() {
			fakeClient.CreateRoleReturns(perm.Role{}, allowedLatency+delta, nil)

			err := subject.Run()
			Expect(err).To(MatchError(ErrExceededMaxLatency))

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(2))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))
		})

		It("runs all other calls but returns an error if AssignRole takes an unacceptable amount of time", func() {
			fakeClient.AssignRoleReturns(allowedLatency+delta, nil)

			err := subject.Run()
			Expect(err).To(MatchError(ErrExceededMaxLatency))

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(2))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))
		})

		It("runs all other calls but returns an error if the first HasPermission call takes an unacceptable amount of time", func() {
			var called bool
			fakeClient.HasPermissionStub = func(context.Context, perm.Actor, string, string) (bool, time.Duration, error) {
				if !called {
					called = true
					return true, allowedLatency + delta, nil
				} else {
					return false, time.Duration(0), nil
				}
			}

			err := subject.Run()
			Expect(err).To(MatchError(ErrExceededMaxLatency))

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(2))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))
		})

		It("runs all other calls but returns an error if the second HasPermission call takes an unacceptable amount of time", func() {
			var called bool
			fakeClient.HasPermissionStub = func(context.Context, perm.Actor, string, string) (bool, time.Duration, error) {
				if !called {
					called = true
					return true, time.Duration(0), nil
				} else {
					return false, allowedLatency + delta, nil
				}
			}

			err := subject.Run()
			Expect(err).To(MatchError(ErrExceededMaxLatency))

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(2))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))
		})

		It("runs all other calls but returns an error if UnassignRole takes an unacceptable amount of time", func() {
			fakeClient.UnassignRoleReturns(allowedLatency+delta, nil)

			err := subject.Run()
			Expect(err).To(MatchError(ErrExceededMaxLatency))

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(2))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))
		})

		It("runs all other calls but returns an error if DeleteRole takes an unacceptable amount of time", func() {
			fakeClient.DeleteRoleReturns(allowedLatency+delta, nil)

			err := subject.Run()
			Expect(err).To(MatchError(ErrExceededMaxLatency))

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(2))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))
		})

		It("stops early and attempts to cleanup if CreateRole fails", func() {
			start := time.Now()

			createRoleErr := errors.New("error")
			fakeClient.CreateRoleReturns(perm.Role{}, zeroDuration, createRoleErr)

			err := subject.Run()
			Expect(err).To(Equal(createRoleErr))

			end := time.Now()

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			_, roleName, _ := fakeClient.CreateRoleArgsForCall(0)

			Expect(fakeClient.AssignRoleCallCount()).To(Equal(0))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(0))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(0))

			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))

			ctx, deletedRoleName := fakeClient.DeleteRoleArgsForCall(0)
			Expect(deletedRoleName).To(Equal(roleName))

			deadline, ok := ctx.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedCleanuptTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedCleanuptTimeout)))
		})

		It("stops early and attempts to cleanup if AssignRole fails", func() {
			start := time.Now()

			assignRoleErr := errors.New("error")
			fakeClient.AssignRoleReturns(zeroDuration, assignRoleErr)

			err := subject.Run()
			Expect(err).To(Equal(assignRoleErr))

			end := time.Now()

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			_, roleName, _ := fakeClient.CreateRoleArgsForCall(0)

			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(0))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(0))

			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))

			ctx, deletedRoleName := fakeClient.DeleteRoleArgsForCall(0)
			Expect(deletedRoleName).To(Equal(roleName))

			deadline, ok := ctx.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedCleanuptTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedCleanuptTimeout)))
		})

		It("stops early and attempts to cleanup if the first HasPermission call fails", func() {
			start := time.Now()

			hasPermissionErr := errors.New("error")
			fakeClient.HasPermissionReturnsOnCall(0, false, zeroDuration, hasPermissionErr)

			err := subject.Run()
			Expect(err).To(Equal(hasPermissionErr))

			end := time.Now()

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			_, roleName, _ := fakeClient.CreateRoleArgsForCall(0)

			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(1))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(0))

			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))

			ctx, deletedRoleName := fakeClient.DeleteRoleArgsForCall(0)
			Expect(deletedRoleName).To(Equal(roleName))

			deadline, ok := ctx.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedCleanuptTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedCleanuptTimeout)))
		})

		It("stops early and attempts to cleanup if the second HasPermission call fails", func() {
			start := time.Now()

			hasPermissionErr := errors.New("error")
			fakeClient.HasPermissionReturnsOnCall(1, false, zeroDuration, hasPermissionErr)

			err := subject.Run()
			Expect(err).To(Equal(hasPermissionErr))

			end := time.Now()

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			_, roleName, _ := fakeClient.CreateRoleArgsForCall(0)

			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(2))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(0))

			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))

			ctx, deletedRoleName := fakeClient.DeleteRoleArgsForCall(0)
			Expect(deletedRoleName).To(Equal(roleName))

			deadline, ok := ctx.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedCleanuptTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedCleanuptTimeout)))
		})

		It("stops early and attempts to cleanup if UnassignRole fails", func() {
			start := time.Now()

			unassignRoleErr := errors.New("error")
			fakeClient.UnassignRoleReturns(zeroDuration, unassignRoleErr)

			err := subject.Run()
			Expect(err).To(Equal(unassignRoleErr))

			end := time.Now()

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			_, roleName, _ := fakeClient.CreateRoleArgsForCall(0)

			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(2))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))

			ctx, deletedRoleName := fakeClient.DeleteRoleArgsForCall(0)
			Expect(deletedRoleName).To(Equal(roleName))

			deadline, ok := ctx.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedCleanuptTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedCleanuptTimeout)))
		})

		It("stops and attempts to cleanup if DeleteRole fails", func() {
			start := time.Now()

			deleteRoleErr := errors.New("error")
			fakeClient.DeleteRoleReturnsOnCall(0, zeroDuration, deleteRoleErr)

			err := subject.Run()
			Expect(err).To(Equal(deleteRoleErr))

			end := time.Now()

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			_, roleName, _ := fakeClient.CreateRoleArgsForCall(0)

			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(2))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(1))

			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(2))

			ctx, deletedRoleName := fakeClient.DeleteRoleArgsForCall(1)
			Expect(deletedRoleName).To(Equal(roleName))

			deadline, ok := ctx.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedCleanuptTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedCleanuptTimeout)))
		})

		It("stops early and attempts to cleanup if the first HasPermission succeeds but has an incorrect result", func() {
			start := time.Now()

			fakeClient.HasPermissionReturnsOnCall(0, false, zeroDuration, nil)

			err := subject.Run()
			Expect(err).To(MatchError(ErrHasAssignedPermission))

			end := time.Now()

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			_, roleName, _ := fakeClient.CreateRoleArgsForCall(0)

			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(1))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(0))

			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))

			ctx, deletedRoleName := fakeClient.DeleteRoleArgsForCall(0)
			Expect(deletedRoleName).To(Equal(roleName))

			deadline, ok := ctx.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedCleanuptTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedCleanuptTimeout)))
		})

		It("stops early and attempts to cleanup if the second HasPermission succeeds but has an incorrect result", func() {
			start := time.Now()

			fakeClient.HasPermissionReturnsOnCall(1, true, zeroDuration, nil)

			err := subject.Run()
			Expect(err).To(MatchError(ErrHasUnassignedPermission))

			end := time.Now()

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			_, roleName, _ := fakeClient.CreateRoleArgsForCall(0)

			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(2))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(0))

			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))

			ctx, deletedRoleName := fakeClient.DeleteRoleArgsForCall(0)
			Expect(deletedRoleName).To(Equal(roleName))

			deadline, ok := ctx.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally(">=", start.Add(expectedCleanuptTimeout)))
			Expect(deadline).To(BeTemporally("<=", end.Add(expectedCleanuptTimeout)))
		})
	})
}
