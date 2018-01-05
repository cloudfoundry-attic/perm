package monitor_test

import (
	. "code.cloudfoundry.org/perm/monitor"

	"context"

	"errors"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/monitor/monitorfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("AdminProbe", func() {
	var (
		p *AdminProbe

		fakeRoleServiceClient *monitorfakes.FakeRoleServiceClient
		fakeLogger            *lagertest.TestLogger
		fakeContext           context.Context

		uniqueSuffix string

		someError error
	)

	BeforeEach(func() {
		fakeRoleServiceClient = new(monitorfakes.FakeRoleServiceClient)

		fakeLogger = lagertest.NewTestLogger("admin-probe")
		fakeContext = context.Background()

		uniqueSuffix = "foobar"

		p = &AdminProbe{
			RoleServiceClient: fakeRoleServiceClient,
		}

		someError = errors.New("some-error")
	})

	Describe("Cleanup", func() {
		It("deletes the role", func() {
			err := p.Cleanup(fakeContext, fakeLogger, uniqueSuffix)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeRoleServiceClient.DeleteRoleCallCount()).To(Equal(1))
			_, deleteRoleRequest, _ := fakeRoleServiceClient.DeleteRoleArgsForCall(0)
			Expect(deleteRoleRequest.GetName()).To(Equal("system.admin-probe.foobar"))
		})

		Context("when the role doesn't exist", func() {
			BeforeEach(func() {
				fakeRoleServiceClient.DeleteRoleReturns(nil, status.Error(codes.NotFound, "role-not-found"))
			})

			It("swallows the error", func() {
				err := p.Cleanup(fakeContext, fakeLogger, uniqueSuffix)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when any other grpc error occurs", func() {
			BeforeEach(func() {
				fakeRoleServiceClient.DeleteRoleReturns(nil, status.Error(codes.Unavailable, "server-not-available"))
			})

			It("errors", func() {
				err := p.Cleanup(fakeContext, fakeLogger, uniqueSuffix)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Run", func() {
		It("creates a role, assigns a role, unassigns a role, and deletes the role", func() {
			err := p.Run(fakeContext, fakeLogger, uniqueSuffix)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeRoleServiceClient.CreateRoleCallCount()).To(Equal(1))
			_, createRoleRequest, _ := fakeRoleServiceClient.CreateRoleArgsForCall(0)
			Expect(createRoleRequest.GetName()).To(Equal("system.admin-probe.foobar"))

			Expect(fakeRoleServiceClient.AssignRoleCallCount()).To(Equal(1))
			_, assignRoleRequest, _ := fakeRoleServiceClient.AssignRoleArgsForCall(0)
			Expect(assignRoleRequest.GetRoleName()).To(Equal("system.admin-probe.foobar"))
			Expect(assignRoleRequest.GetActor().GetIssuer()).To(Equal("system"))
			Expect(assignRoleRequest.GetActor().GetID()).To(Equal("admin-probe"))

			Expect(fakeRoleServiceClient.UnassignRoleCallCount()).To(Equal(1))
			_, unassignRoleRequest, _ := fakeRoleServiceClient.UnassignRoleArgsForCall(0)
			Expect(unassignRoleRequest.GetRoleName()).To(Equal("system.admin-probe.foobar"))
			Expect(unassignRoleRequest.GetActor().GetIssuer()).To(Equal("system"))
			Expect(unassignRoleRequest.GetActor().GetID()).To(Equal("admin-probe"))

			Expect(fakeRoleServiceClient.DeleteRoleCallCount()).To(Equal(1))
			_, deleteRoleRequest, _ := fakeRoleServiceClient.DeleteRoleArgsForCall(0)
			Expect(deleteRoleRequest.GetName()).To(Equal("system.admin-probe.foobar"))
		})

		Context("when creating a role fails", func() {
			BeforeEach(func() {
				fakeRoleServiceClient.CreateRoleReturns(nil, someError)
			})

			It("errors and does not assign, unassign, or delete", func() {
				err := p.Run(fakeContext, fakeLogger, uniqueSuffix)
				Expect(err).To(MatchError(someError))

				Expect(fakeRoleServiceClient.CreateRoleCallCount()).To(Equal(1))
				Expect(fakeRoleServiceClient.AssignRoleCallCount()).To(Equal(0))
				Expect(fakeRoleServiceClient.UnassignRoleCallCount()).To(Equal(0))
				Expect(fakeRoleServiceClient.DeleteRoleCallCount()).To(Equal(0))
			})
		})

		Context("when assigning a role fails", func() {
			BeforeEach(func() {
				fakeRoleServiceClient.AssignRoleReturns(nil, someError)
			})

			It("errors and does not unassign or delete", func() {
				err := p.Run(fakeContext, fakeLogger, uniqueSuffix)
				Expect(err).To(MatchError(someError))

				Expect(fakeRoleServiceClient.CreateRoleCallCount()).To(Equal(1))
				Expect(fakeRoleServiceClient.AssignRoleCallCount()).To(Equal(1))
				Expect(fakeRoleServiceClient.UnassignRoleCallCount()).To(Equal(0))
				Expect(fakeRoleServiceClient.DeleteRoleCallCount()).To(Equal(0))
			})
		})

		Context("when unassigning a role fails", func() {
			BeforeEach(func() {
				fakeRoleServiceClient.UnassignRoleReturns(nil, someError)
			})

			It("errors and does not unassign or delete", func() {
				err := p.Run(fakeContext, fakeLogger, uniqueSuffix)
				Expect(err).To(MatchError(someError))

				Expect(fakeRoleServiceClient.CreateRoleCallCount()).To(Equal(1))
				Expect(fakeRoleServiceClient.AssignRoleCallCount()).To(Equal(1))
				Expect(fakeRoleServiceClient.UnassignRoleCallCount()).To(Equal(1))
				Expect(fakeRoleServiceClient.DeleteRoleCallCount()).To(Equal(0))
			})
		})

		Context("when deleting a role fails", func() {
			BeforeEach(func() {
				fakeRoleServiceClient.DeleteRoleReturns(nil, someError)
			})

			It("errors and does not unassign or delete", func() {
				err := p.Run(fakeContext, fakeLogger, uniqueSuffix)
				Expect(err).To(MatchError(someError))

				Expect(fakeRoleServiceClient.CreateRoleCallCount()).To(Equal(1))
				Expect(fakeRoleServiceClient.AssignRoleCallCount()).To(Equal(1))
				Expect(fakeRoleServiceClient.UnassignRoleCallCount()).To(Equal(1))
				Expect(fakeRoleServiceClient.DeleteRoleCallCount()).To(Equal(1))
			})
		})
	})
})
