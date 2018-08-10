package monitor_test

import (
	"context"
	"errors"
	"time"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/pkg/logx"
	"code.cloudfoundry.org/perm/pkg/logx/lagerx"
	. "code.cloudfoundry.org/perm/pkg/monitor"
	"code.cloudfoundry.org/perm/pkg/monitor/monitorfakes"
	"code.cloudfoundry.org/perm/pkg/monitor/recording"
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
		fakeClient *monitorfakes.FakeClient
		fakeSender *monitorfakes.FakeSender
		fakeLogger logx.Logger

		subject *Probe

		zeroDuration time.Duration
		delta        time.Duration

		testProbeRunsCorrect     string
		testProbeRunsSuccess     string
		testProbeAPICallsSuccess string
	)

	BeforeEach(func() {
		fakeClient = new(monitorfakes.FakeClient)
		fakeSender = new(monitorfakes.FakeSender)
		fakeLogger = lagerx.NewLogger(lagertest.NewTestLogger("probe-test"))

		subject = NewProbe(fakeClient, fakeSender, fakeLogger, opts...)

		zeroDuration = time.Duration(0)
		delta = time.Duration(10)

		fakeClient.HasPermissionReturnsOnCall(0, true, zeroDuration, nil)
		fakeClient.HasPermissionReturnsOnCall(1, false, zeroDuration, nil)

		testProbeRunsCorrect = "perm.probe.runs.correct"
		testProbeRunsSuccess = "perm.probe.runs.success"
		testProbeAPICallsSuccess = "perm.probe.api.calls.success"
	})

	Describe("#Run", func() {
		It("runs through the basic functionality", func() {
			subject.Run()

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

			Expect(fakeSender.GaugeCallCount()).To(Equal(8))

			metric, value, alwaysSend := fakeSender.GaugeArgsForCall(0)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))
			Expect(alwaysSend).To(Equal(float32(1)))

			metric, value, alwaysSend = fakeSender.GaugeArgsForCall(1)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))
			Expect(alwaysSend).To(Equal(float32(1)))

			metric, value, alwaysSend = fakeSender.GaugeArgsForCall(2)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))
			Expect(alwaysSend).To(Equal(float32(1)))

			metric, value, alwaysSend = fakeSender.GaugeArgsForCall(3)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))
			Expect(alwaysSend).To(Equal(float32(1)))

			metric, value, alwaysSend = fakeSender.GaugeArgsForCall(4)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))
			Expect(alwaysSend).To(Equal(float32(1)))

			metric, value, alwaysSend = fakeSender.GaugeArgsForCall(5)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))
			Expect(alwaysSend).To(Equal(float32(1)))
		})

		It("uses the correct timeout for all calls", func() {
			start := time.Now()

			subject.Run()

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
			subject.Run()

			_, firstRole, _ := fakeClient.CreateRoleArgsForCall(0)

			fakeClient.HasPermissionReturnsOnCall(2, true, zeroDuration, nil)
			fakeClient.HasPermissionReturnsOnCall(3, false, zeroDuration, nil)

			subject.Run()

			_, secondRole, _ := fakeClient.CreateRoleArgsForCall(1)

			Expect(firstRole).NotTo(Equal(secondRole))
		})

		It("uses a unique permission each time", func() {
			subject.Run()

			_, _, firstPermissions := fakeClient.CreateRoleArgsForCall(0)
			firstPermission := firstPermissions[0]

			fakeClient.HasPermissionReturnsOnCall(2, true, zeroDuration, nil)
			fakeClient.HasPermissionReturnsOnCall(3, false, zeroDuration, nil)

			subject.Run()

			_, _, secondPermissions := fakeClient.CreateRoleArgsForCall(1)
			secondPermission := secondPermissions[0]

			Expect(firstPermission).NotTo(Equal(secondPermission))
		})

		It("runs all other calls but returns an error if CreateRole takes an unacceptable amount of time", func() {
			fakeClient.CreateRoleReturns(perm.Role{}, allowedLatency+delta, nil)

			subject.Run()

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(2))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))

			Expect(fakeSender.GaugeCallCount()).To(Equal(8))

			metric, value, _ := fakeSender.GaugeArgsForCall(0)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(0)))

			metric, value, _ = fakeSender.GaugeArgsForCall(1)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(2)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(3)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(4)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(5)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))
		})

		It("runs all other calls but returns an error if AssignRole takes an unacceptable amount of time", func() {
			fakeClient.AssignRoleReturns(allowedLatency+delta, nil)

			subject.Run()

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(2))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))

			Expect(fakeSender.GaugeCallCount()).To(Equal(8))

			metric, value, _ := fakeSender.GaugeArgsForCall(0)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(1)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(0)))

			metric, value, _ = fakeSender.GaugeArgsForCall(2)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(3)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(4)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(5)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))
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

			subject.Run()

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(2))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))

			Expect(fakeSender.GaugeCallCount()).To(Equal(8))

			metric, value, _ := fakeSender.GaugeArgsForCall(0)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(1)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(2)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(0)))

			metric, value, _ = fakeSender.GaugeArgsForCall(3)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(4)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(5)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))
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

			subject.Run()

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(2))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))

			Expect(fakeSender.GaugeCallCount()).To(Equal(8))

			metric, value, _ := fakeSender.GaugeArgsForCall(0)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(1)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(2)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(3)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(0)))

			metric, value, _ = fakeSender.GaugeArgsForCall(4)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(5)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))
		})

		It("runs all other calls but returns an error if UnassignRole takes an unacceptable amount of time", func() {
			fakeClient.UnassignRoleReturns(allowedLatency+delta, nil)

			subject.Run()

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(2))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))

			Expect(fakeSender.GaugeCallCount()).To(Equal(8))

			metric, value, _ := fakeSender.GaugeArgsForCall(0)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(1)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(2)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(3)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(4)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(0)))

			metric, value, _ = fakeSender.GaugeArgsForCall(5)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))
		})

		It("runs all other calls but returns an error if DeleteRole takes an unacceptable amount of time", func() {
			fakeClient.DeleteRoleReturns(allowedLatency+delta, nil)

			subject.Run()

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(2))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))

			Expect(fakeSender.GaugeCallCount()).To(Equal(8))

			metric, value, _ := fakeSender.GaugeArgsForCall(0)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(1)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(2)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(3)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(4)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(5)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(0)))
		})

		It("stops early and attempts to cleanup if CreateRole fails", func() {
			start := time.Now()

			createRoleErr := errors.New("error")
			fakeClient.CreateRoleReturns(perm.Role{}, zeroDuration, createRoleErr)

			subject.Run()

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

			Expect(fakeSender.GaugeCallCount()).To(Equal(2))

			metric, value, _ := fakeSender.GaugeArgsForCall(0)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(0)))

			metric, value, _ = fakeSender.GaugeArgsForCall(1)
			Expect(metric).To(Equal(testProbeRunsSuccess))
			Expect(value).To(Equal(int64(0)))
		})

		It("stops early and attempts to cleanup if AssignRole fails", func() {
			start := time.Now()

			assignRoleErr := errors.New("error")
			fakeClient.AssignRoleReturns(zeroDuration, assignRoleErr)

			subject.Run()

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

			Expect(fakeSender.GaugeCallCount()).To(Equal(3))

			metric, value, _ := fakeSender.GaugeArgsForCall(0)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(1)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(0)))

			metric, value, _ = fakeSender.GaugeArgsForCall(2)
			Expect(metric).To(Equal(testProbeRunsSuccess))
			Expect(value).To(Equal(int64(0)))
		})

		It("stops early and attempts to cleanup if the first HasPermission call fails", func() {
			start := time.Now()

			hasPermissionErr := errors.New("error")
			fakeClient.HasPermissionReturnsOnCall(0, false, zeroDuration, hasPermissionErr)

			subject.Run()

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

			Expect(fakeSender.GaugeCallCount()).To(Equal(4))

			metric, value, _ := fakeSender.GaugeArgsForCall(0)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(1)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(2)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(0)))

			metric, value, _ = fakeSender.GaugeArgsForCall(3)
			Expect(metric).To(Equal(testProbeRunsSuccess))
			Expect(value).To(Equal(int64(0)))
		})

		It("stops early and attempts to cleanup if the second HasPermission call fails", func() {
			start := time.Now()

			hasPermissionErr := errors.New("error")
			fakeClient.HasPermissionReturnsOnCall(1, false, zeroDuration, hasPermissionErr)

			subject.Run()

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

			Expect(fakeSender.GaugeCallCount()).To(Equal(5))

			metric, value, _ := fakeSender.GaugeArgsForCall(0)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(1)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(2)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(3)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(0)))

			metric, value, _ = fakeSender.GaugeArgsForCall(4)
			Expect(metric).To(Equal(testProbeRunsSuccess))
			Expect(value).To(Equal(int64(0)))
		})

		It("stops early and attempts to cleanup if UnassignRole fails", func() {
			start := time.Now()

			unassignRoleErr := errors.New("error")
			fakeClient.UnassignRoleReturns(zeroDuration, unassignRoleErr)

			subject.Run()

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

			Expect(fakeSender.GaugeCallCount()).To(Equal(6))

			metric, value, _ := fakeSender.GaugeArgsForCall(0)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(1)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(2)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(3)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(4)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(0)))

			metric, value, _ = fakeSender.GaugeArgsForCall(5)
			Expect(metric).To(Equal(testProbeRunsSuccess))
			Expect(value).To(Equal(int64(0)))
		})

		It("stops and attempts to cleanup if DeleteRole fails", func() {
			start := time.Now()

			deleteRoleErr := errors.New("error")
			fakeClient.DeleteRoleReturnsOnCall(0, zeroDuration, deleteRoleErr)

			subject.Run()

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

			Expect(fakeSender.GaugeCallCount()).To(Equal(7))

			metric, value, _ := fakeSender.GaugeArgsForCall(0)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(1)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(2)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(3)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(4)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(5)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(0)))

			metric, value, _ = fakeSender.GaugeArgsForCall(6)
			Expect(metric).To(Equal(testProbeRunsSuccess))
			Expect(value).To(Equal(int64(0)))
		})

		It("stops early and attempts to cleanup if the first HasPermission call succeeds but has an incorrect result", func() {
			start := time.Now()

			fakeClient.HasPermissionReturnsOnCall(0, false, zeroDuration, nil)

			subject.Run()

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

			Expect(fakeSender.GaugeCallCount()).To(Equal(5))

			metric, value, _ := fakeSender.GaugeArgsForCall(0)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(1)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(2)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(0)))

			metric, value, _ = fakeSender.GaugeArgsForCall(3)
			Expect(metric).To(Equal(testProbeRunsCorrect))
			Expect(value).To(Equal(int64(0)))

			metric, value, _ = fakeSender.GaugeArgsForCall(4)
			Expect(metric).To(Equal(testProbeRunsSuccess))
			Expect(value).To(Equal(int64(0)))
		})

		It("stops early and attempts to cleanup if the second HasPermission succeeds but has an incorrect result", func() {
			start := time.Now()

			fakeClient.HasPermissionReturnsOnCall(1, true, zeroDuration, nil)

			subject.Run()

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

			Expect(fakeSender.GaugeCallCount()).To(Equal(6))

			metric, value, _ := fakeSender.GaugeArgsForCall(0)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(1)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(2)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(1)))

			metric, value, _ = fakeSender.GaugeArgsForCall(3)
			Expect(metric).To(Equal(testProbeAPICallsSuccess))
			Expect(value).To(Equal(int64(0)))

			metric, value, _ = fakeSender.GaugeArgsForCall(4)
			Expect(metric).To(Equal(testProbeRunsCorrect))
			Expect(value).To(Equal(int64(0)))

			metric, value, _ = fakeSender.GaugeArgsForCall(5)
			Expect(metric).To(Equal(testProbeRunsSuccess))
			Expect(value).To(Equal(int64(0)))
		})

		It("continues if CreateRole fails to record duration", func() {
			fakeClient.CreateRoleReturns(perm.Role{}, zeroDuration, recording.FailedToObserveDurationError{})

			subject.Run()

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(2))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))
			Expect(fakeSender.GaugeCallCount()).To(Equal(8))
		})

		It("continues if AssignRole fails to record duration", func() {
			fakeClient.AssignRoleReturns(zeroDuration, recording.FailedToObserveDurationError{})

			subject.Run()

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(2))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))
			Expect(fakeSender.GaugeCallCount()).To(Equal(8))
		})

		It("continues if the first HasPermission call fails to record duration", func() {
			fakeClient.HasPermissionReturnsOnCall(0, true, zeroDuration, recording.FailedToObserveDurationError{})

			subject.Run()

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(2))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))
			Expect(fakeSender.GaugeCallCount()).To(Equal(8))
		})

		It("continues if the second HasPermission call fails to record duration", func() {
			fakeClient.HasPermissionReturnsOnCall(1, false, zeroDuration, recording.FailedToObserveDurationError{})

			subject.Run()

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(2))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))
			Expect(fakeSender.GaugeCallCount()).To(Equal(8))
		})

		It("continues if UnassignRole fails to record duration", func() {
			fakeClient.UnassignRoleReturns(zeroDuration, recording.FailedToObserveDurationError{})

			subject.Run()

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(2))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))
			Expect(fakeSender.GaugeCallCount()).To(Equal(8))
		})

		It("continues if DeleteRole fails to record duration", func() {
			fakeClient.DeleteRoleReturns(zeroDuration, recording.FailedToObserveDurationError{})

			subject.Run()

			Expect(fakeClient.CreateRoleCallCount()).To(Equal(1))
			Expect(fakeClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.HasPermissionCallCount()).To(Equal(2))
			Expect(fakeClient.UnassignRoleCallCount()).To(Equal(1))
			Expect(fakeClient.DeleteRoleCallCount()).To(Equal(1))
			Expect(fakeSender.GaugeCallCount()).To(Equal(8))
		})
	})
}
