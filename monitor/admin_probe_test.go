package monitor_test

import (
	. "code.cloudfoundry.org/perm/monitor"

	"context"

	"errors"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/monitor/monitorfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AdminProbe", func() {
	var (
		p *AdminProbe

		fakeRoleServiceClient *monitorfakes.FakeRoleServiceClient
		fakeLogger            *lagertest.TestLogger
		fakeContext           context.Context

		someError       error
		someStatsDError error
	)

	BeforeEach(func() {
		fakeRoleServiceClient = new(monitorfakes.FakeRoleServiceClient)

		fakeLogger = lagertest.NewTestLogger("admin-probe")
		fakeContext = context.Background()

		p = &AdminProbe{
			RoleServiceClient: fakeRoleServiceClient,
		}

		someError = errors.New("some-error")
		someStatsDError = errors.New("some-statsd-error")
	})

	Describe("Run", func() {
		It("creates a role, assigns a role, unassigns a role, and deletes the role", func() {
			err := p.Run(fakeContext, fakeLogger)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeRoleServiceClient.CreateRoleCallCount()).To(Equal(1))
			Expect(fakeRoleServiceClient.AssignRoleCallCount()).To(Equal(1))
			Expect(fakeRoleServiceClient.UnassignRoleCallCount()).To(Equal(1))
			Expect(fakeRoleServiceClient.DeleteRoleCallCount()).To(Equal(1))
		})

		Context("when creating a role fails", func() {
			BeforeEach(func() {
				fakeRoleServiceClient.CreateRoleReturns(nil, someError)
			})

			It("errors and does not assign, unassign, or delete", func() {
				err := p.Run(fakeContext, fakeLogger)
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
				err := p.Run(fakeContext, fakeLogger)
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
				err := p.Run(fakeContext, fakeLogger)
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
				err := p.Run(fakeContext, fakeLogger)
				Expect(err).To(MatchError(someError))

				Expect(fakeRoleServiceClient.CreateRoleCallCount()).To(Equal(1))
				Expect(fakeRoleServiceClient.AssignRoleCallCount()).To(Equal(1))
				Expect(fakeRoleServiceClient.UnassignRoleCallCount()).To(Equal(1))
				Expect(fakeRoleServiceClient.DeleteRoleCallCount()).To(Equal(1))
			})
		})
	})
})
