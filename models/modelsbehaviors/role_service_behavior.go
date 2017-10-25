package modelsbehaviors

import (
	"context"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/models"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

type CreateRoleService func() models.RoleService

func BehavesLikeARoleService(subjectCreator CreateRoleService) {
	var (
		subject models.RoleService

		ctx    context.Context
		logger *lagertest.TestLogger
	)

	ginkgo.BeforeEach(func() {
		subject = subjectCreator()

		ctx = context.Background()
		logger = lagertest.NewTestLogger("perm-test")
	})

	ginkgo.Describe("#CreateRole", func() {
		ginkgo.It("saves the role", func() {
			name := uuid.NewV4().String()

			role, err := subject.CreateRole(ctx, logger, name)

			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(role).NotTo(gomega.BeNil())
			gomega.Expect(role.Name).To(gomega.Equal(name))

			expectedRole := role
			role, err = subject.FindRole(ctx, logger, models.RoleQuery{Name: name})

			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(role).To(gomega.Equal(expectedRole))
		})

		ginkgo.It("fails if a role with the name already exists", func() {
			name := uuid.NewV4().String()

			_, err := subject.CreateRole(ctx, logger, name)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			role, err := subject.CreateRole(ctx, logger, name)

			gomega.Expect(role).To(gomega.BeNil())
			gomega.Expect(err).To(gomega.Equal(models.ErrRoleAlreadyExists))
		})
	})

	ginkgo.Describe("#FindRole", func() {
		ginkgo.It("fails if the role does not exist", func() {
			name := uuid.NewV4().String()

			role, err := subject.FindRole(ctx, logger, models.RoleQuery{Name: name})

			gomega.Expect(role).To(gomega.BeNil())
			gomega.Expect(err).To(gomega.Equal(models.ErrRoleNotFound))
		})
	})

	ginkgo.Describe("#DeleteRole", func() {
		ginkgo.It("deletes the role if it exists", func() {
			name := uuid.NewV4().String()

			_, err := subject.CreateRole(ctx, logger, name)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = subject.DeleteRole(ctx, logger, models.RoleQuery{Name: name})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			role, err := subject.FindRole(ctx, logger, models.RoleQuery{Name: name})

			gomega.Expect(role).To(gomega.BeNil())
			gomega.Expect(err).To(gomega.Equal(models.ErrRoleNotFound))
		})

		ginkgo.It("fails if the role does not exist", func() {
			name := uuid.NewV4().String()

			err := subject.DeleteRole(ctx, logger, models.RoleQuery{Name: name})

			gomega.Expect(err).To(gomega.Equal(models.ErrRoleNotFound))
		})
	})
}
