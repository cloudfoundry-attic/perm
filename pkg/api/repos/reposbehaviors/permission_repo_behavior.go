package reposbehaviors_test

import (
	"context"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/pkg/api/repos"
	"code.cloudfoundry.org/perm/pkg/perm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	"github.com/satori/go.uuid"
)

func BehavesLikeAPermissionRepo(
	subjectCreator func() repos.PermissionRepo,
	roleRepoCreator func() repos.RoleRepo,
) {
	var (
		subject repos.PermissionRepo

		roleRepo repos.RoleRepo

		ctx    context.Context
		logger *lagertest.TestLogger

		cancelFunc context.CancelFunc

		roleName                                             string
		groups                                               []perm.Group
		sameAction                                           string
		resourcePattern1, resourcePattern2, resourcePattern3 string

		actor                    perm.Actor
		permission1, permission2 *perm.Permission
	)

	BeforeEach(func() {
		subject = subjectCreator()

		roleRepo = roleRepoCreator()

		ctx, cancelFunc = context.WithTimeout(context.Background(), 1*time.Second)
		logger = lagertest.NewTestLogger("perm-test")

		roleName = uuid.NewV4().String()
		sameAction = "some-action"
		resourcePattern1 = "resource-pattern-1"
		resourcePattern2 = "resource-pattern-2"
		resourcePattern3 = "resource-pattern-3"

		permission1 = &perm.Permission{Action: sameAction, ResourcePattern: resourcePattern1}
		permission2 = &perm.Permission{Action: sameAction, ResourcePattern: resourcePattern2}
		actor = perm.Actor{
			ID:        uuid.NewV4().String(),
			Namespace: uuid.NewV4().String(),
		}
		groups = []perm.Group{
			{ID: uuid.NewV4().String()},
			{ID: uuid.NewV4().String()},
		}
	})

	AfterEach(func() {
		cancelFunc()
	})

	Describe("#HasPermission", func() {
		It("returns true if they have been assigned to a role that has the permission", func() {
			_, err := roleRepo.CreateRole(ctx, logger, roleName, permission1)
			Expect(err).NotTo(HaveOccurred())

			err = roleRepo.AssignRole(ctx, logger, roleName, actor.ID, actor.Namespace)
			Expect(err).NotTo(HaveOccurred())

			query := repos.HasPermissionQuery{
				Actor:           actor,
				Action:          permission1.Action,
				ResourcePattern: permission1.ResourcePattern,
			}

			yes, err := subject.HasPermission(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeTrue())
		})

		It("returns false if they have not been assigned the role", func() {
			_, err := roleRepo.CreateRole(ctx, logger, roleName, permission1)
			Expect(err).NotTo(HaveOccurred())

			query := repos.HasPermissionQuery{
				Actor:           actor,
				Action:          permission1.Action,
				ResourcePattern: permission1.ResourcePattern,
			}
			yes, err := subject.HasPermission(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeFalse())
		})

		It("return false if the actor does not exist", func() {
			_, err := roleRepo.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			query := repos.HasPermissionQuery{
				Actor:           actor,
				Action:          permission1.Action,
				ResourcePattern: permission1.ResourcePattern,
			}
			yes, err := subject.HasPermission(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeFalse())
		})

		Context("when the actor doesn't have permission but groups are supplied", func() {
			It("returns true if a group is assigned to a role with permission", func() {
				_, err := roleRepo.CreateRole(ctx, logger, roleName, permission1)
				Expect(err).NotTo(HaveOccurred())

				err = roleRepo.AssignRoleToGroup(ctx, logger, roleName, groups[0].ID)
				Expect(err).NotTo(HaveOccurred())

				query := repos.HasPermissionQuery{
					Actor:           actor,
					Action:          permission1.Action,
					ResourcePattern: permission1.ResourcePattern,
					Groups:          groups,
				}

				yes, err := subject.HasPermission(ctx, logger, query)

				Expect(err).NotTo(HaveOccurred())
				Expect(yes).To(BeTrue())
			})

			It("returns false if the group is not assigned to a role with permission", func() {
				_, err := roleRepo.CreateRole(ctx, logger, roleName, permission1)
				Expect(err).NotTo(HaveOccurred())

				query := repos.HasPermissionQuery{
					Actor:           actor,
					Action:          permission1.Action,
					ResourcePattern: permission1.ResourcePattern,
					Groups:          groups,
				}
				yes, err := subject.HasPermission(ctx, logger, query)

				Expect(err).NotTo(HaveOccurred())
				Expect(yes).To(BeFalse())
			})
		})
	})

	Describe("#ListResourcePatterns", func() {
		It("returns the list of resource patterns for which the actor has that permission", func() {
			permission3 := &perm.Permission{
				Action:          "different-action",
				ResourcePattern: resourcePattern3,
			}

			_, err := roleRepo.CreateRole(ctx, logger, roleName, permission1, permission2, permission3)
			Expect(err).NotTo(HaveOccurred())

			err = roleRepo.AssignRole(ctx, logger, roleName, actor.ID, actor.Namespace)
			Expect(err).NotTo(HaveOccurred())

			permissionForUnassignedRole := &perm.Permission{
				Action:          sameAction,
				ResourcePattern: "should-not-have-this-resource-pattern",
			}

			_, err = roleRepo.CreateRole(ctx, logger, "not-assigned-to-this-role", permissionForUnassignedRole)
			Expect(err).NotTo(HaveOccurred())

			query := repos.ListResourcePatternsQuery{
				Actor:  actor,
				Action: sameAction,
			}

			resourcePatterns, err := subject.ListResourcePatterns(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(resourcePatterns).NotTo(BeNil())
			Expect(resourcePatterns).To(HaveLen(2))
			Expect(resourcePatterns).To(ConsistOf(resourcePattern1, resourcePattern2))
		})

		It("de-dupes the results if the user has access to the same resource pattern through multiple roles/permissions", func() {
			_, err := roleRepo.CreateRole(ctx, logger, roleName, permission1)
			Expect(err).NotTo(HaveOccurred())

			err = roleRepo.AssignRole(ctx, logger, roleName, actor.ID, actor.Namespace)
			Expect(err).NotTo(HaveOccurred())

			roleName2 := uuid.NewV4().String()
			permission2 := &perm.Permission{
				Action:          sameAction,
				ResourcePattern: resourcePattern1,
			}

			_, err = roleRepo.CreateRole(ctx, logger, roleName2, permission2)
			Expect(err).NotTo(HaveOccurred())

			err = roleRepo.AssignRole(ctx, logger, roleName2, actor.ID, actor.Namespace)
			Expect(err).NotTo(HaveOccurred())

			query := repos.ListResourcePatternsQuery{
				Actor:  actor,
				Action: sameAction,
			}

			resourcePatterns, err := subject.ListResourcePatterns(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(resourcePatterns).NotTo(BeNil())
			Expect(resourcePatterns).To(HaveLen(1))
			Expect(resourcePatterns).To(ConsistOf(resourcePattern1))
		})

		It("returns empty if the actor is not assigned to any roles with that permission", func() {
			query := repos.ListResourcePatternsQuery{
				Actor:  actor,
				Action: sameAction,
			}

			resourcePatterns, err := subject.ListResourcePatterns(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(resourcePatterns).To(BeEmpty())
		})

		Context("when providing groups", func() {
			var (
				otherPermission *perm.Permission
				otherRoleName   string
			)

			BeforeEach(func() {
				otherRoleName = "other-role"
				otherPermission = &perm.Permission{
					Action:          "action-that-should-not-appear",
					ResourcePattern: "other-resource-pattern",
				}
			})
			Context("when both the actor and groups have assigned permissions", func() {
				It("returns a list of all permissions", func() {
					permission3 := &perm.Permission{
						Action:          sameAction,
						ResourcePattern: resourcePattern3,
					}
					_, err := roleRepo.CreateRole(ctx, logger, roleName, permission1)
					Expect(err).NotTo(HaveOccurred())
					err = roleRepo.AssignRole(ctx, logger, roleName, actor.ID, actor.Namespace)
					Expect(err).NotTo(HaveOccurred())

					_, err = roleRepo.CreateRole(ctx, logger, otherRoleName, permission2)
					Expect(err).NotTo(HaveOccurred())
					err = roleRepo.AssignRoleToGroup(ctx, logger, otherRoleName, groups[0].ID)
					Expect(err).NotTo(HaveOccurred())

					_, err = roleRepo.CreateRole(ctx, logger, "some-other-role-name", permission3)
					Expect(err).NotTo(HaveOccurred())
					err = roleRepo.AssignRoleToGroup(ctx, logger, "some-other-role-name", groups[1].ID)
					Expect(err).NotTo(HaveOccurred())

					query := repos.ListResourcePatternsQuery{
						Actor:  actor,
						Action: sameAction,
						Groups: groups,
					}

					resourcePatterns, err := subject.ListResourcePatterns(ctx, logger, query)

					Expect(err).NotTo(HaveOccurred())
					Expect(resourcePatterns).NotTo(BeNil())
					Expect(resourcePatterns).To(HaveLen(3))
					Expect(resourcePatterns).To(ConsistOf(resourcePattern1, resourcePattern2, resourcePattern3))
				})
			})

			Context("when the groups provided specify permissions for different actions", func() {
				It("does not list out the permissions for those different actions", func() {
					_, err := roleRepo.CreateRole(ctx, logger, roleName, permission1)
					Expect(err).NotTo(HaveOccurred())
					err = roleRepo.AssignRoleToGroup(ctx, logger, roleName, groups[0].ID)
					Expect(err).NotTo(HaveOccurred())

					_, err = roleRepo.CreateRole(ctx, logger, otherRoleName, otherPermission)
					Expect(err).NotTo(HaveOccurred())
					err = roleRepo.AssignRoleToGroup(ctx, logger, otherRoleName, groups[1].ID)
					Expect(err).NotTo(HaveOccurred())

					query := repos.ListResourcePatternsQuery{
						Actor:  actor,
						Action: sameAction,
						Groups: groups,
					}

					resourcePatterns, err := subject.ListResourcePatterns(ctx, logger, query)

					Expect(err).NotTo(HaveOccurred())
					Expect(resourcePatterns).NotTo(BeNil())
					Expect(resourcePatterns).To(HaveLen(1))
					Expect(resourcePatterns).To(ConsistOf(resourcePattern1))
				})
			})

			Context("when there are no actor roles", func() {
				It("still returns the group's roles", func() {
					_, err := roleRepo.CreateRole(ctx, logger, roleName, permission1, permission2)
					Expect(err).NotTo(HaveOccurred())
					err = roleRepo.AssignRoleToGroup(ctx, logger, roleName, groups[0].ID)
					Expect(err).NotTo(HaveOccurred())

					query := repos.ListResourcePatternsQuery{
						Actor:  actor,
						Action: sameAction,
						Groups: groups,
					}

					resourcePatterns, err := subject.ListResourcePatterns(ctx, logger, query)

					Expect(err).NotTo(HaveOccurred())
					Expect(resourcePatterns).NotTo(BeNil())
					Expect(resourcePatterns).To(HaveLen(2))
					Expect(resourcePatterns).To(ConsistOf(resourcePattern1, resourcePattern2))
				})
			})
		})
	})
}
