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

	Describe("#DeleteRole", func() {
		It("deletes the role if it exists", func() {
			name := uuid.NewV4().String()

			permission := &perm.Permission{"a", "b"}
			_, err := subject.CreateRole(ctx, logger, name, permission)
			Expect(err).NotTo(HaveOccurred())

			listRolePermsQuery := repos.ListRolePermissionsQuery{RoleName: name}
			permissions, err := subject.ListRolePermissions(ctx, logger, listRolePermsQuery)
			Expect(len(permissions)).To(Equal(1))
			Expect(err).ToNot(HaveOccurred())

			err = subject.DeleteRole(ctx, logger, name)
			Expect(err).NotTo(HaveOccurred())

			listRolePermsQuery = repos.ListRolePermissionsQuery{RoleName: name}
			_, err = subject.ListRolePermissions(ctx, logger, listRolePermsQuery)
			Expect(err).To(HaveOccurred())
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

	Describe("#AssignRole", func() {
		It("saves the role assignment, saving the actor if it does not exist", func() {
			actor := perm.Actor{
				ID:        uuid.NewV4().String(),
				Namespace: uuid.NewV4().String(),
			}
			roleName := uuid.NewV4().String()

			_, err := subject.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRole(ctx, logger, roleName, actor.ID, actor.Namespace)

			Expect(err).NotTo(HaveOccurred())

			query := repos.HasRoleQuery{
				Actor:    actor,
				RoleName: roleName,
			}
			yes, err := subject.HasRole(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeTrue())
		})

		It("fails if the role assignment already exists", func() {
			actor := perm.Actor{
				ID:        uuid.NewV4().String(),
				Namespace: uuid.NewV4().String(),
			}
			roleName := uuid.NewV4().String()

			_, err := subject.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRole(ctx, logger, roleName, actor.ID, actor.Namespace)

			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRole(ctx, logger, roleName, actor.ID, actor.Namespace)
			Expect(err).To(Equal(perm.ErrAssignmentAlreadyExists))
		})

		It("fails if the role does not exist", func() {
			id := uuid.NewV4().String()
			namespace := uuid.NewV4().String()
			roleName := uuid.NewV4().String()

			err := subject.AssignRole(ctx, logger, roleName, id, namespace)
			Expect(err).To(Equal(perm.ErrRoleNotFound))
		})
	})

	Describe("#AssignRoleToGroup", func() {
		It("saves the role assignment, saving the group if it does not exist", func() {
			group := perm.Group{
				ID: uuid.NewV4().String(),
			}
			roleName := uuid.NewV4().String()

			_, err := subject.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRoleToGroup(ctx, logger, roleName, group.ID)

			Expect(err).NotTo(HaveOccurred())

			query := repos.HasRoleForGroupQuery{
				Group:    group,
				RoleName: roleName,
			}
			yes, err := subject.HasRoleForGroup(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeTrue())
		})

		It("fails if the role assignment already exists", func() {
			group := perm.Group{
				ID: uuid.NewV4().String(),
			}
			roleName := uuid.NewV4().String()

			_, err := subject.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRoleToGroup(ctx, logger, roleName, group.ID)

			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRoleToGroup(ctx, logger, roleName, group.ID)
			Expect(err).To(Equal(perm.ErrAssignmentAlreadyExists))
		})

		It("fails if the role does not exist", func() {
			id := uuid.NewV4().String()
			roleName := uuid.NewV4().String()

			err := subject.AssignRoleToGroup(ctx, logger, roleName, id)
			Expect(err).To(Equal(perm.ErrRoleNotFound))
		})
	})

	Describe("#UnassignRole", func() {
		It("removes the role assignment", func() {
			actor := perm.Actor{
				ID:        uuid.NewV4().String(),
				Namespace: uuid.NewV4().String(),
			}
			roleName := uuid.NewV4().String()

			_, err := subject.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRole(ctx, logger, roleName, actor.ID, actor.Namespace)
			Expect(err).NotTo(HaveOccurred())

			err = subject.UnassignRole(ctx, logger, roleName, actor.ID, actor.Namespace)
			Expect(err).NotTo(HaveOccurred())

			query := repos.HasRoleQuery{
				Actor:    actor,
				RoleName: roleName,
			}
			yes, err := subject.HasRole(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeFalse())
		})

		It("fails if the role does not exist", func() {
			id := uuid.NewV4().String()
			namespace := uuid.NewV4().String()
			roleName := uuid.NewV4().String()

			err := subject.UnassignRole(ctx, logger, roleName, id, namespace)
			Expect(err).To(Equal(perm.ErrRoleNotFound))
		})

		It("fails if the actor does not exist", func() {
			id := uuid.NewV4().String()
			namespace := uuid.NewV4().String()
			roleName := uuid.NewV4().String()

			_, err := subject.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.UnassignRole(ctx, logger, roleName, id, namespace)
			Expect(err).To(MatchError(perm.ErrAssignmentNotFound))
		})

		It("fails if the role assignment does not exist", func() {
			id := uuid.NewV4().String()
			namespace := uuid.NewV4().String()
			roleName := uuid.NewV4().String()

			_, err := subject.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.UnassignRole(ctx, logger, roleName, id, namespace)
			Expect(err).To(Equal(perm.ErrAssignmentNotFound))
		})
	})

	Describe("#HasRole", func() {
		It("returns true if they have been assigned the role", func() {
			actor := perm.Actor{
				ID:        uuid.NewV4().String(),
				Namespace: uuid.NewV4().String(),
			}
			roleName := uuid.NewV4().String()

			_, err := subject.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRole(ctx, logger, roleName, actor.ID, actor.Namespace)
			Expect(err).NotTo(HaveOccurred())

			query := repos.HasRoleQuery{
				Actor:    actor,
				RoleName: roleName,
			}
			yes, err := subject.HasRole(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeTrue())
		})

		It("returns false if they have not been assigned the role", func() {
			actor := perm.Actor{
				ID:        uuid.NewV4().String(),
				Namespace: uuid.NewV4().String(),
			}
			roleName := uuid.NewV4().String()

			_, err := subject.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			query := repos.HasRoleQuery{
				Actor:    actor,
				RoleName: roleName,
			}
			yes, err := subject.HasRole(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeFalse())
		})

		It("returns false if the actor does not exist", func() {
			actor := perm.Actor{
				ID:        uuid.NewV4().String(),
				Namespace: uuid.NewV4().String(),
			}
			roleName := uuid.NewV4().String()

			_, err := subject.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			query := repos.HasRoleQuery{
				Actor:    actor,
				RoleName: roleName,
			}
			hasRole, err := subject.HasRole(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(hasRole).To(BeFalse())
		})

		It("fails if the role does not exist", func() {
			actor := perm.Actor{
				ID:        uuid.NewV4().String(),
				Namespace: uuid.NewV4().String(),
			}
			roleName := uuid.NewV4().String()

			query := repos.HasRoleQuery{
				Actor:    actor,
				RoleName: roleName,
			}
			_, err := subject.HasRole(ctx, logger, query)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(perm.ErrRoleNotFound))
		})
	})

	Describe("#ListActorRoles", func() {
		It("returns a list of all roles that the actor has been assigned to", func() {
			actor := perm.Actor{
				ID:        uuid.NewV4().String(),
				Namespace: uuid.NewV4().String(),
			}

			roleName1 := uuid.NewV4().String()
			role1, err := subject.CreateRole(ctx, logger, roleName1)
			Expect(err).NotTo(HaveOccurred())

			roleName2 := uuid.NewV4().String()
			role2, err := subject.CreateRole(ctx, logger, roleName2)
			Expect(err).NotTo(HaveOccurred())

			roleName3 := uuid.NewV4().String()
			role3, err := subject.CreateRole(ctx, logger, roleName3)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRole(ctx, logger, roleName1, actor.ID, actor.Namespace)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRole(ctx, logger, roleName2, actor.ID, actor.Namespace)
			Expect(err).NotTo(HaveOccurred())

			query := repos.ListActorRolesQuery{
				Actor: actor,
			}
			roles, err := subject.ListActorRoles(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(roles).To(HaveLen(2))
			Expect(roles).To(ContainElement(role1))
			Expect(roles).To(ContainElement(role2))
			Expect(roles).NotTo(ContainElement(role3))
		})

		It("returns an empty list if the actor does not exist", func() {
			actor := perm.Actor{
				ID:        uuid.NewV4().String(),
				Namespace: uuid.NewV4().String(),
			}
			roleName := uuid.NewV4().String()

			_, err := subject.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			query := repos.ListActorRolesQuery{
				Actor: actor,
			}
			roles, err := subject.ListActorRoles(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(roles).To(BeEmpty())
		})
	})
}
