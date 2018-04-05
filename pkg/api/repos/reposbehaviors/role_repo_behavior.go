package reposbehaviors_test

import (
	"context"

	"time"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/pkg/api/repos"
	"code.cloudfoundry.org/perm/pkg/perm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

func BehavesLikeARoleRepo(subjectCreator func() repos.RoleRepo) {
	var (
		subject repos.RoleRepo

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
		})

		It("fails if a role with the name already exists", func() {
			name := uuid.NewV4().String()

			_, err := subject.CreateRole(ctx, logger, name)
			Expect(err).NotTo(HaveOccurred())

			_, err = subject.CreateRole(ctx, logger, name)
			Expect(err).To(Equal(perm.ErrRoleAlreadyExists))
		})
	})

	Describe("#FindRole", func() {
		It("fails if the role does not exist", func() {
			name := uuid.NewV4().String()

			role, err := subject.FindRole(ctx, logger, repos.FindRoleQuery{RoleName: name})

			Expect(role).To(BeNil())
			Expect(err).To(Equal(perm.ErrRoleNotFound))
		})
	})

	Describe("#DeleteRole", func() {
		It("deletes the role if it exists", func() {
			name := uuid.NewV4().String()

			_, err := subject.CreateRole(ctx, logger, name)
			Expect(err).NotTo(HaveOccurred())

			err = subject.DeleteRole(ctx, logger, name)
			Expect(err).NotTo(HaveOccurred())

			role, err := subject.FindRole(ctx, logger, repos.FindRoleQuery{RoleName: name})

			Expect(role).To(BeNil())
			Expect(err).To(Equal(perm.ErrRoleNotFound))
		})

		It("fails if the role does not exist", func() {
			name := uuid.NewV4().String()

			err := subject.DeleteRole(ctx, logger, name)

			Expect(err).To(Equal(perm.ErrRoleNotFound))
		})
	})

	Describe("#ListRolePermissions", func() {
		It("returns a list of all permissions that the role has been created with", func() {
			roleName := uuid.NewV4().String()

			permission1 := &perm.Permission{Action: "permission-1", ResourcePattern: "resource-pattern-1"}
			permission2 := &perm.Permission{Action: "permission-2", ResourcePattern: "resource-pattern-2"}
			_, err := subject.CreateRole(ctx, logger, roleName, permission1, permission2)
			Expect(err).NotTo(HaveOccurred())

			query := repos.ListRolePermissionsQuery{
				RoleName: roleName,
			}

			permissions, err := subject.ListRolePermissions(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(permissions).To(HaveLen(2))
			Expect(permissions).To(ContainElement(permission1))
			Expect(permissions).To(ContainElement(permission2))
		})

		It("fails if the actor does not exist", func() {
			query := repos.ListRolePermissionsQuery{
				RoleName: "foobar",
			}
			_, err := subject.ListRolePermissions(ctx, logger, query)
			Expect(err).To(MatchError(perm.ErrRoleNotFound))
		})
	})
}
