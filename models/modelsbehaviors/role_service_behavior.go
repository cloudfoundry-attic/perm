package modelsbehaviors_test

import (
	"context"

	"time"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

func BehavesLikeARoleService(subjectCreator func() models.RoleService) {
	var (
		subject models.RoleService

		ctx    context.Context
		logger *lagertest.TestLogger

		cancelFunc context.CancelFunc
	)

	BeforeEach(func() {
		subject = subjectCreator()

		ctx, cancelFunc = context.WithTimeout(context.Background(), 1*time.Second)
		logger = lagertest.NewTestLogger("perm-test")
	})

	AfterEach(func() {
		cancelFunc()
	})

	Describe("#CreateRole", func() {
		It("saves the role", func() {
			name := uuid.NewV4().String()

			role, err := subject.CreateRole(ctx, logger, name)

			Expect(err).NotTo(HaveOccurred())

			Expect(role).NotTo(BeNil())
			Expect(role.Name).To(Equal(name))

			expectedRole := role
			role, err = subject.FindRole(ctx, logger, models.RoleQuery{Name: name})

			Expect(err).NotTo(HaveOccurred())
			Expect(role).To(Equal(expectedRole))
		})

		It("fails if a role with the name already exists", func() {
			name := uuid.NewV4().String()

			_, err := subject.CreateRole(ctx, logger, name)
			Expect(err).NotTo(HaveOccurred())

			_, err = subject.CreateRole(ctx, logger, name)

			Expect(err).To(Equal(models.ErrRoleAlreadyExists))
		})
	})

	Describe("#FindRole", func() {
		It("fails if the role does not exist", func() {
			name := uuid.NewV4().String()

			role, err := subject.FindRole(ctx, logger, models.RoleQuery{Name: name})

			Expect(role).To(BeNil())
			Expect(err).To(Equal(models.ErrRoleNotFound))
		})
	})

	Describe("#DeleteRole", func() {
		It("deletes the role if it exists", func() {
			name := uuid.NewV4().String()

			_, err := subject.CreateRole(ctx, logger, name)
			Expect(err).NotTo(HaveOccurred())

			err = subject.DeleteRole(ctx, logger, models.RoleQuery{Name: name})
			Expect(err).NotTo(HaveOccurred())

			role, err := subject.FindRole(ctx, logger, models.RoleQuery{Name: name})

			Expect(role).To(BeNil())
			Expect(err).To(Equal(models.ErrRoleNotFound))
		})

		It("fails if the role does not exist", func() {
			name := uuid.NewV4().String()

			err := subject.DeleteRole(ctx, logger, models.RoleQuery{Name: name})

			Expect(err).To(Equal(models.ErrRoleNotFound))
		})
	})

	Describe("#ListRolePermissions", func() {
		It("returns a list of all permissions that the role has been created with", func() {
			roleName := uuid.NewV4().String()

			permission1 := &models.Permission{Name: "permission-1", ResourcePattern: "resource-pattern-1"}
			permission2 := &models.Permission{Name: "permission-2", ResourcePattern: "resource-pattern-2"}
			_, err := subject.CreateRole(ctx, logger, roleName, permission1, permission2)
			Expect(err).NotTo(HaveOccurred())

			query := models.RoleQuery{
				Name: roleName,
			}

			permissions, err := subject.ListRolePermissions(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(permissions).To(HaveLen(2))
			Expect(permissions).To(ContainElement(permission1))
			Expect(permissions).To(ContainElement(permission2))
		})

		It("fails if the actor does not exist", func() {
			query := models.RoleQuery{
				Name: "foobar",
			}
			_, err := subject.ListRolePermissions(ctx, logger, query)
			Expect(err).To(MatchError(models.ErrRoleNotFound))
		})
	})
}
