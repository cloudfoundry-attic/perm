package monitor_test

import (
	. "code.cloudfoundry.org/perm/monitor"

	"context"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/monitor/monitorfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AdminProbe", func() {
	var (
		p *AdminProbe

		fakeRoleServiceClient *monitorfakes.FakeRoleServiceClient
		fakeStatsDClient      *monitorfakes.FakeStatter
		fakeLogger            *lagertest.TestLogger
		fakeContext           context.Context
	)

	BeforeEach(func() {
		fakeRoleServiceClient = new(monitorfakes.FakeRoleServiceClient)
		fakeStatsDClient = new(monitorfakes.FakeStatter)

		fakeLogger = lagertest.NewTestLogger("admin-probe")
		fakeContext = context.Background()

		p = &AdminProbe{
			RoleServiceClient: fakeRoleServiceClient,
			StatsDClient:      fakeStatsDClient,
		}
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
	})
})
