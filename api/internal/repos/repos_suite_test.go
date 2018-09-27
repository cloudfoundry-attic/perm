package repos_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm"
	. "code.cloudfoundry.org/perm/api/internal/repos"
	"code.cloudfoundry.org/perm/internal/migrations"
	"code.cloudfoundry.org/perm/internal/sqlx/testsqlx"
	"code.cloudfoundry.org/perm/logx"
	"code.cloudfoundry.org/perm/logx/lagerx"

	"testing"
)

var (
	testDB *testsqlx.TestMySQLDB
)

func TestRepos(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Repos Suite")
}

var _ = BeforeSuite(func() {
	var err error

	testDB = testsqlx.NewTestMySQLDB()
	err = testDB.Create(migrations.Migrations...)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := testDB.Drop()
	Expect(err).NotTo(HaveOccurred())
})

func testRepo(subjectCreator func() repo) {
	var (
		subject repo
	)

	BeforeEach(func() {
		subject = subjectCreator()
	})

	testPermissionRepo(func() PermissionRepo { return subject }, func() RoleRepo { return subject })
	testRoleRepo(func() RoleRepo { return subject })
}

func testPermissionRepo(subjectCreator func() PermissionRepo, roleRepoCreator func() RoleRepo) {
	var (
		subject PermissionRepo

		roleRepo RoleRepo

		logger logx.Logger

		roleName                                             string
		sameAction                                           string
		resourcePattern1, resourcePattern2, resourcePattern3 string

		actor                    perm.Actor
		permission1, permission2 perm.Permission
	)

	BeforeEach(func() {
		subject = subjectCreator()

		roleRepo = roleRepoCreator()

		logger = lagerx.NewLogger(lagertest.NewTestLogger("perm-test"))

		roleName = uuid.NewV4().String()
		sameAction = "some-action"
		resourcePattern1 = "resource-pattern-1"
		resourcePattern2 = "resource-pattern-2"
		resourcePattern3 = "resource-pattern-3"

		permission1 = perm.Permission{Action: sameAction, ResourcePattern: resourcePattern1}
		permission2 = perm.Permission{Action: sameAction, ResourcePattern: resourcePattern2}
		actor = perm.Actor{
			ID:        uuid.NewV4().String(),
			Namespace: uuid.NewV4().String(),
			Groups: []perm.Group{
				{ID: uuid.NewV4().String()},
				{ID: uuid.NewV4().String()},
			},
		}
	})

	Describe("#HasPermission", func() {
		It("returns true if they have been assigned to a role that has the permission", func() {
			_, err := roleRepo.CreateRole(context.Background(), logger, roleName, permission1)
			Expect(err).NotTo(HaveOccurred())

			err = roleRepo.AssignRole(context.Background(), logger, roleName, actor.ID, actor.Namespace)
			Expect(err).NotTo(HaveOccurred())

			query := HasPermissionQuery{
				Actor:           actor,
				Action:          permission1.Action,
				ResourcePattern: permission1.ResourcePattern,
			}

			yes, err := subject.HasPermission(context.Background(), logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeTrue())
		})

		It("returns false if they have not been assigned the role", func() {
			_, err := roleRepo.CreateRole(context.Background(), logger, roleName, permission1)
			Expect(err).NotTo(HaveOccurred())

			query := HasPermissionQuery{
				Actor:           actor,
				Action:          permission1.Action,
				ResourcePattern: permission1.ResourcePattern,
			}
			yes, err := subject.HasPermission(context.Background(), logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeFalse())
		})

		It("return false if the actor does not exist", func() {
			_, err := roleRepo.CreateRole(context.Background(), logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			query := HasPermissionQuery{
				Actor:           actor,
				Action:          permission1.Action,
				ResourcePattern: permission1.ResourcePattern,
			}
			yes, err := subject.HasPermission(context.Background(), logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeFalse())
		})

		Context("when the actor doesn't have permission but groups are supplied", func() {
			It("returns true if a group is assigned to a role with permission", func() {
				_, err := roleRepo.CreateRole(context.Background(), logger, roleName, permission1)
				Expect(err).NotTo(HaveOccurred())

				err = roleRepo.AssignRoleToGroup(context.Background(), logger, roleName, actor.Groups[0].ID)
				Expect(err).NotTo(HaveOccurred())

				query := HasPermissionQuery{
					Actor:           actor,
					Action:          permission1.Action,
					ResourcePattern: permission1.ResourcePattern,
				}

				yes, err := subject.HasPermission(context.Background(), logger, query)

				Expect(err).NotTo(HaveOccurred())
				Expect(yes).To(BeTrue())
			})

			It("returns false if the group is not assigned to a role with permission", func() {
				_, err := roleRepo.CreateRole(context.Background(), logger, roleName, permission1)
				Expect(err).NotTo(HaveOccurred())

				query := HasPermissionQuery{
					Actor:           actor,
					Action:          permission1.Action,
					ResourcePattern: permission1.ResourcePattern,
				}
				yes, err := subject.HasPermission(context.Background(), logger, query)

				Expect(err).NotTo(HaveOccurred())
				Expect(yes).To(BeFalse())
			})
		})
	})

	Describe("#ListResourcePatterns", func() {
		It("returns the list of resource patterns for which the actor has that permission", func() {
			permission3 := perm.Permission{
				Action:          "different-action",
				ResourcePattern: resourcePattern3,
			}

			_, err := roleRepo.CreateRole(context.Background(), logger, roleName, permission1, permission2, permission3)
			Expect(err).NotTo(HaveOccurred())

			err = roleRepo.AssignRole(context.Background(), logger, roleName, actor.ID, actor.Namespace)
			Expect(err).NotTo(HaveOccurred())

			permissionForUnassignedRole := perm.Permission{
				Action:          sameAction,
				ResourcePattern: "should-not-have-this-resource-pattern",
			}

			_, err = roleRepo.CreateRole(context.Background(), logger, "not-assigned-to-this-role", permissionForUnassignedRole)
			Expect(err).NotTo(HaveOccurred())

			query := ListResourcePatternsQuery{
				Actor:  actor,
				Action: sameAction,
			}

			resourcePatterns, err := subject.ListResourcePatterns(context.Background(), logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(resourcePatterns).NotTo(BeNil())
			Expect(resourcePatterns).To(HaveLen(2))
			Expect(resourcePatterns).To(ConsistOf(resourcePattern1, resourcePattern2))
		})

		It("de-dupes the results if the user has access to the same resource pattern through multiple roles/permissions", func() {
			_, err := roleRepo.CreateRole(context.Background(), logger, roleName, permission1)
			Expect(err).NotTo(HaveOccurred())

			err = roleRepo.AssignRole(context.Background(), logger, roleName, actor.ID, actor.Namespace)
			Expect(err).NotTo(HaveOccurred())

			roleName2 := uuid.NewV4().String()
			permission2 = perm.Permission{
				Action:          sameAction,
				ResourcePattern: resourcePattern1,
			}

			_, err = roleRepo.CreateRole(context.Background(), logger, roleName2, permission2)
			Expect(err).NotTo(HaveOccurred())

			err = roleRepo.AssignRole(context.Background(), logger, roleName2, actor.ID, actor.Namespace)
			Expect(err).NotTo(HaveOccurred())

			query := ListResourcePatternsQuery{
				Actor:  actor,
				Action: sameAction,
			}

			resourcePatterns, err := subject.ListResourcePatterns(context.Background(), logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(resourcePatterns).NotTo(BeNil())
			Expect(resourcePatterns).To(HaveLen(1))
			Expect(resourcePatterns).To(ConsistOf(resourcePattern1))
		})

		It("returns empty if the actor is not assigned to any roles with that permission", func() {
			query := ListResourcePatternsQuery{
				Actor:  actor,
				Action: sameAction,
			}

			resourcePatterns, err := subject.ListResourcePatterns(context.Background(), logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(resourcePatterns).To(BeEmpty())
		})

		Context("when providing groups", func() {
			var (
				otherPermission perm.Permission
				otherRoleName   string
			)

			BeforeEach(func() {
				otherRoleName = "other-role"
				otherPermission = perm.Permission{
					Action:          "action-that-should-not-appear",
					ResourcePattern: "other-resource-pattern",
				}
			})
			Context("when both the actor and groups have assigned permissions", func() {
				It("returns a list of all permissions", func() {
					permission3 := perm.Permission{
						Action:          sameAction,
						ResourcePattern: resourcePattern3,
					}
					_, err := roleRepo.CreateRole(context.Background(), logger, roleName, permission1)
					Expect(err).NotTo(HaveOccurred())
					err = roleRepo.AssignRole(context.Background(), logger, roleName, actor.ID, actor.Namespace)
					Expect(err).NotTo(HaveOccurred())

					_, err = roleRepo.CreateRole(context.Background(), logger, otherRoleName, permission2)
					Expect(err).NotTo(HaveOccurred())
					err = roleRepo.AssignRoleToGroup(context.Background(), logger, otherRoleName, actor.Groups[0].ID)
					Expect(err).NotTo(HaveOccurred())

					_, err = roleRepo.CreateRole(context.Background(), logger, "some-other-role-name", permission3)
					Expect(err).NotTo(HaveOccurred())
					err = roleRepo.AssignRoleToGroup(context.Background(), logger, "some-other-role-name", actor.Groups[1].ID)
					Expect(err).NotTo(HaveOccurred())

					query := ListResourcePatternsQuery{
						Actor:  actor,
						Action: sameAction,
					}

					resourcePatterns, err := subject.ListResourcePatterns(context.Background(), logger, query)

					Expect(err).NotTo(HaveOccurred())
					Expect(resourcePatterns).NotTo(BeNil())
					Expect(resourcePatterns).To(HaveLen(3))
					Expect(resourcePatterns).To(ConsistOf(resourcePattern1, resourcePattern2, resourcePattern3))
				})
			})

			Context("when the groups provided specify permissions for different actions", func() {
				It("does not list out the permissions for those different actions", func() {
					_, err := roleRepo.CreateRole(context.Background(), logger, roleName, permission1)
					Expect(err).NotTo(HaveOccurred())
					err = roleRepo.AssignRoleToGroup(context.Background(), logger, roleName, actor.Groups[0].ID)
					Expect(err).NotTo(HaveOccurred())

					_, err = roleRepo.CreateRole(context.Background(), logger, otherRoleName, otherPermission)
					Expect(err).NotTo(HaveOccurred())
					err = roleRepo.AssignRoleToGroup(context.Background(), logger, otherRoleName, actor.Groups[1].ID)
					Expect(err).NotTo(HaveOccurred())

					query := ListResourcePatternsQuery{
						Actor:  actor,
						Action: sameAction,
					}

					resourcePatterns, err := subject.ListResourcePatterns(context.Background(), logger, query)

					Expect(err).NotTo(HaveOccurred())
					Expect(resourcePatterns).NotTo(BeNil())
					Expect(resourcePatterns).To(HaveLen(1))
					Expect(resourcePatterns).To(ConsistOf(resourcePattern1))
				})
			})

			Context("when there are no actor roles", func() {
				It("still returns the group's roles", func() {
					_, err := roleRepo.CreateRole(context.Background(), logger, roleName, permission1, permission2)
					Expect(err).NotTo(HaveOccurred())
					err = roleRepo.AssignRoleToGroup(context.Background(), logger, roleName, actor.Groups[0].ID)
					Expect(err).NotTo(HaveOccurred())

					query := ListResourcePatternsQuery{
						Actor:  actor,
						Action: sameAction,
					}

					resourcePatterns, err := subject.ListResourcePatterns(context.Background(), logger, query)

					Expect(err).NotTo(HaveOccurred())
					Expect(resourcePatterns).NotTo(BeNil())
					Expect(resourcePatterns).To(HaveLen(2))
					Expect(resourcePatterns).To(ConsistOf(resourcePattern1, resourcePattern2))
				})
			})
		})
	})
}

func testRoleRepo(subjectCreator func() RoleRepo) {
	var (
		subject RoleRepo

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

			listRolePermsQuery := ListRolePermissionsQuery{RoleName: name}
			permissions, err := subject.ListRolePermissions(context.Background(), logger, listRolePermsQuery)
			Expect(len(permissions)).To(Equal(1))
			Expect(err).ToNot(HaveOccurred())

			err = subject.DeleteRole(context.Background(), logger, name)
			Expect(err).NotTo(HaveOccurred())

			listRolePermsQuery = ListRolePermissionsQuery{RoleName: name}
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

			query := ListRolePermissionsQuery{
				RoleName: roleName,
			}

			permissions, err := subject.ListRolePermissions(context.Background(), logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(permissions).To(HaveLen(2))
			Expect(permissions).To(ContainElement(permission1))
			Expect(permissions).To(ContainElement(permission2))
		})

		It("fails if the actor does not exist", func() {
			query := ListRolePermissionsQuery{
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

			query := HasRoleQuery{
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

			query := HasRoleForGroupQuery{
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

			query := HasRoleQuery{
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

			query := HasRoleQuery{
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

			query := HasRoleQuery{
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

			query := HasRoleQuery{
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

			query := HasRoleQuery{
				Actor:    actor,
				RoleName: roleName,
			}
			hasRole, err := subject.HasRole(context.Background(), logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(hasRole).To(BeFalse())
		})

		It("fails if the role does not exist", func() {
			query := HasRoleQuery{
				Actor:    actor,
				RoleName: roleName,
			}
			_, err := subject.HasRole(context.Background(), logger, query)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(perm.ErrRoleNotFound))
		})
	})
}

type repo interface {
	PermissionRepo
	RoleRepo
}
