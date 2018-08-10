package reposbehaviors_test

import (
	"context"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/pkg/api/repos"
	"code.cloudfoundry.org/perm/pkg/logx"
	"code.cloudfoundry.org/perm/pkg/logx/lagerx"
	"code.cloudfoundry.org/perm/pkg/perm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

func BehavesLikeARoleRepo(subjectCreator func() repos.RoleRepo) {
	var (
		subject repos.RoleRepo

		logger logx.Logger

		name      string
		roleName  string
		namespace string

		actor                    perm.Actor
		permission1, permission2 perm.Permission
		group                    perm.Group
	)

	BeforeEach(func() {
		subject = subjectCreator()

		logger = lagerx.NewLogger(lagertest.NewTestLogger("perm-test"))

		name = uuid.NewV4().String()
		roleName = uuid.NewV4().String()
		namespace = uuid.NewV4().String()

		permission1 = perm.Permission{Action: "permission-1", ResourcePattern: "resource-pattern-1"}
		permission2 = perm.Permission{Action: "permission-2", ResourcePattern: "resource-pattern-2"}
		actor = perm.Actor{
			ID:        uuid.NewV4().String(),
			Namespace: uuid.NewV4().String(),
		}
		group = perm.Group{ID: uuid.NewV4().String()}
	})

	Describe("#CreateRole", func() {
		It("saves the role", func() {
			role, err := subject.CreateRole(context.Background(), logger, name)

			Expect(err).NotTo(HaveOccurred())

			Expect(role).NotTo(BeNil())
			Expect(role.Name).To(Equal(name))
		})

		It("fails if a role with the name already exists", func() {
			_, err := subject.CreateRole(context.Background(), logger, name)
			Expect(err).NotTo(HaveOccurred())

			_, err = subject.CreateRole(context.Background(), logger, name)
			Expect(err).To(Equal(perm.ErrRoleAlreadyExists))
		})
	})

	Describe("#DeleteRole", func() {
		It("deletes the role if it exists", func() {
			permission := perm.Permission{Action: "a", ResourcePattern: "b"}
			_, err := subject.CreateRole(context.Background(), logger, name, permission)
			Expect(err).NotTo(HaveOccurred())

			listRolePermsQuery := repos.ListRolePermissionsQuery{RoleName: name}
			permissions, err := subject.ListRolePermissions(context.Background(), logger, listRolePermsQuery)
			Expect(len(permissions)).To(Equal(1))
			Expect(err).ToNot(HaveOccurred())

			err = subject.DeleteRole(context.Background(), logger, name)
			Expect(err).NotTo(HaveOccurred())

			listRolePermsQuery = repos.ListRolePermissionsQuery{RoleName: name}
			_, err = subject.ListRolePermissions(context.Background(), logger, listRolePermsQuery)
			Expect(err).To(HaveOccurred())
		})

		It("fails if the role does not exist", func() {
			err := subject.DeleteRole(context.Background(), logger, name)

			Expect(err).To(Equal(perm.ErrRoleNotFound))
		})
	})

	Describe("#ListRolePermissions", func() {
		It("returns a list of all permissions that the role has been created with", func() {
			_, err := subject.CreateRole(context.Background(), logger, roleName, permission1, permission2)
			Expect(err).NotTo(HaveOccurred())

			query := repos.ListRolePermissionsQuery{
				RoleName: roleName,
			}

			permissions, err := subject.ListRolePermissions(context.Background(), logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(permissions).To(HaveLen(2))
			Expect(permissions).To(ContainElement(permission1))
			Expect(permissions).To(ContainElement(permission2))
		})

		It("fails if the actor does not exist", func() {
			query := repos.ListRolePermissionsQuery{
				RoleName: "foobar",
			}
			_, err := subject.ListRolePermissions(context.Background(), logger, query)
			Expect(err).To(MatchError(perm.ErrRoleNotFound))
		})
	})

	Describe("#AssignRole", func() {
		It("saves the role assignment, saving the actor if it does not exist", func() {
			_, err := subject.CreateRole(context.Background(), logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRole(context.Background(), logger, roleName, actor.ID, actor.Namespace)

			Expect(err).NotTo(HaveOccurred())

			query := repos.HasRoleQuery{
				Actor:    actor,
				RoleName: roleName,
			}
			yes, err := subject.HasRole(context.Background(), logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeTrue())
		})

		It("fails if the role assignment already exists", func() {
			_, err := subject.CreateRole(context.Background(), logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRole(context.Background(), logger, roleName, actor.ID, actor.Namespace)

			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRole(context.Background(), logger, roleName, actor.ID, actor.Namespace)
			Expect(err).To(Equal(perm.ErrAssignmentAlreadyExists))
		})

		It("fails if the role does not exist", func() {
			err := subject.AssignRole(context.Background(), logger, roleName, actor.ID, namespace)
			Expect(err).To(Equal(perm.ErrRoleNotFound))
		})
	})

	Describe("#AssignRoleToGroup", func() {
		It("saves the role assignment, saving the group if it does not exist", func() {
			_, err := subject.CreateRole(context.Background(), logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRoleToGroup(context.Background(), logger, roleName, group.ID)

			Expect(err).NotTo(HaveOccurred())

			query := repos.HasRoleForGroupQuery{
				Group:    group,
				RoleName: roleName,
			}
			yes, err := subject.HasRoleForGroup(context.Background(), logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeTrue())
		})

		It("fails if the role assignment already exists", func() {
			_, err := subject.CreateRole(context.Background(), logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRoleToGroup(context.Background(), logger, roleName, group.ID)

			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRoleToGroup(context.Background(), logger, roleName, group.ID)
			Expect(err).To(Equal(perm.ErrAssignmentAlreadyExists))
		})

		It("fails if the role does not exist", func() {
			err := subject.AssignRoleToGroup(context.Background(), logger, roleName, uuid.NewV4().String())
			Expect(err).To(Equal(perm.ErrRoleNotFound))
		})
	})

	Describe("#UnassignRole", func() {
		It("removes the role assignment", func() {
			_, err := subject.CreateRole(context.Background(), logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRole(context.Background(), logger, roleName, actor.ID, actor.Namespace)
			Expect(err).NotTo(HaveOccurred())

			err = subject.UnassignRole(context.Background(), logger, roleName, actor.ID, actor.Namespace)
			Expect(err).NotTo(HaveOccurred())

			query := repos.HasRoleQuery{
				Actor:    actor,
				RoleName: roleName,
			}
			yes, err := subject.HasRole(context.Background(), logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeFalse())
		})

		It("fails if the role does not exist", func() {
			err := subject.UnassignRole(context.Background(), logger, roleName, actor.ID, namespace)
			Expect(err).To(Equal(perm.ErrRoleNotFound))
		})

		It("fails if the role assignment does not exist", func() {
			_, err := subject.CreateRole(context.Background(), logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.UnassignRole(context.Background(), logger, roleName, actor.ID, namespace)
			Expect(err).To(MatchError(perm.ErrAssignmentNotFound))
		})
	})

	Describe("#UnassignRoleFromGroup", func() {
		It("removes the role assignment", func() {
			_, err := subject.CreateRole(context.Background(), logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRoleToGroup(context.Background(), logger, roleName, group.ID)
			Expect(err).NotTo(HaveOccurred())

			err = subject.UnassignRoleFromGroup(context.Background(), logger, roleName, group.ID)
			Expect(err).NotTo(HaveOccurred())

			query := repos.HasRoleQuery{
				Actor:    actor,
				RoleName: roleName,
			}
			yes, err := subject.HasRole(context.Background(), logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeFalse())
		})

		It("fails if the role does not exist", func() {
			err := subject.UnassignRoleFromGroup(context.Background(), logger, roleName, group.ID)
			Expect(err).To(Equal(perm.ErrRoleNotFound))
		})

		It("fails if the role assignment does not exist", func() {
			_, err := subject.CreateRole(context.Background(), logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.UnassignRoleFromGroup(context.Background(), logger, roleName, group.ID)
			Expect(err).To(Equal(perm.ErrAssignmentNotFound))
		})
	})

	Describe("#HasRole", func() {
		It("returns true if they have been assigned the role", func() {
			_, err := subject.CreateRole(context.Background(), logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRole(context.Background(), logger, roleName, actor.ID, actor.Namespace)
			Expect(err).NotTo(HaveOccurred())

			query := repos.HasRoleQuery{
				Actor:    actor,
				RoleName: roleName,
			}
			yes, err := subject.HasRole(context.Background(), logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeTrue())
		})

		It("returns false if they have not been assigned the role", func() {
			_, err := subject.CreateRole(context.Background(), logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			query := repos.HasRoleQuery{
				Actor:    actor,
				RoleName: roleName,
			}
			yes, err := subject.HasRole(context.Background(), logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeFalse())
		})

		It("returns false if the actor does not exist", func() {
			_, err := subject.CreateRole(context.Background(), logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			query := repos.HasRoleQuery{
				Actor:    actor,
				RoleName: roleName,
			}
			hasRole, err := subject.HasRole(context.Background(), logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(hasRole).To(BeFalse())
		})

		It("fails if the role does not exist", func() {
			query := repos.HasRoleQuery{
				Actor:    actor,
				RoleName: roleName,
			}
			_, err := subject.HasRole(context.Background(), logger, query)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(perm.ErrRoleNotFound))
		})
	})
}
