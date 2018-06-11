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
	)

	BeforeEach(func() {
		subject = subjectCreator()

		roleRepo = roleRepoCreator()

		ctx, cancelFunc = context.WithTimeout(context.Background(), 1*time.Second)
		logger = lagertest.NewTestLogger("perm-test")
	})

	AfterEach(func() {
		cancelFunc()
	})

	Describe("#HasPermission", func() {
		It("returns true if they have been assigned to a role that has the permission", func() {
			roleName := uuid.NewV4().String()
			actor := perm.Actor{
				ID:        uuid.NewV4().String(),
				Namespace: uuid.NewV4().String(),
			}
			permission := perm.Permission{
				Action:          "some-action",
				ResourcePattern: "some-resource",
			}

			_, err := roleRepo.CreateRole(ctx, logger, roleName, &permission)
			Expect(err).NotTo(HaveOccurred())

			err = roleRepo.AssignRole(ctx, logger, roleName, actor.ID, actor.Namespace)
			Expect(err).NotTo(HaveOccurred())

			query := repos.HasPermissionQuery{
				Actor:           actor,
				Action:          permission.Action,
				ResourcePattern: permission.ResourcePattern,
			}

			yes, err := subject.HasPermission(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeTrue())
		})

		It("returns false if they have not been assigned the role", func() {
			roleName := uuid.NewV4().String()
			actor := perm.Actor{
				ID:        uuid.NewV4().String(),
				Namespace: uuid.NewV4().String(),
			}
			permission := perm.Permission{
				Action:          "some-action",
				ResourcePattern: "some-resource",
			}

			_, err := roleRepo.CreateRole(ctx, logger, roleName, &permission)
			Expect(err).NotTo(HaveOccurred())

			query := repos.HasPermissionQuery{
				Actor:           actor,
				Action:          permission.Action,
				ResourcePattern: permission.ResourcePattern,
			}
			yes, err := subject.HasPermission(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeFalse())
		})

		It("return false if the actor does not exist", func() {
			roleName := uuid.NewV4().String()
			actor := perm.Actor{
				ID:        uuid.NewV4().String(),
				Namespace: uuid.NewV4().String(),
			}

			_, err := roleRepo.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			permission := perm.Permission{
				Action:          "some-action",
				ResourcePattern: "some-resource",
			}

			query := repos.HasPermissionQuery{
				Actor:           actor,
				Action:          permission.Action,
				ResourcePattern: permission.ResourcePattern,
			}
			yes, err := subject.HasPermission(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeFalse())
		})

		Context("when the actor doesn't have permission but groups are supplied", func() {
			It("returns true if a group is assigned to a role with permission", func() {
				roleName := uuid.NewV4().String()
				actor := perm.Actor{
					ID:        uuid.NewV4().String(),
					Namespace: uuid.NewV4().String(),
				}
				group := perm.Group{
					ID: uuid.NewV4().String(),
				}
				permission := perm.Permission{
					Action:          "some-action",
					ResourcePattern: "some-resource",
				}

				_, err := roleRepo.CreateRole(ctx, logger, roleName, &permission)
				Expect(err).NotTo(HaveOccurred())

				err = roleRepo.AssignRoleToGroup(ctx, logger, roleName, group.ID)
				Expect(err).NotTo(HaveOccurred())

				query := repos.HasPermissionQuery{
					Actor:           actor,
					Action:          permission.Action,
					ResourcePattern: permission.ResourcePattern,
					Groups:          []perm.Group{group},
				}

				yes, err := subject.HasPermission(ctx, logger, query)

				Expect(err).NotTo(HaveOccurred())
				Expect(yes).To(BeTrue())
			})

			It("returns false if the group is not assigned to a role with permission", func() {
				roleName := uuid.NewV4().String()
				actor := perm.Actor{
					ID:        uuid.NewV4().String(),
					Namespace: uuid.NewV4().String(),
				}
				group := perm.Group{
					ID: uuid.NewV4().String(),
				}
				permission := perm.Permission{
					Action:          "some-action",
					ResourcePattern: "some-resource",
				}

				_, err := roleRepo.CreateRole(ctx, logger, roleName, &permission)
				Expect(err).NotTo(HaveOccurred())

				query := repos.HasPermissionQuery{
					Actor:           actor,
					Action:          permission.Action,
					ResourcePattern: permission.ResourcePattern,
					Groups:          []perm.Group{group},
				}
				yes, err := subject.HasPermission(ctx, logger, query)

				Expect(err).NotTo(HaveOccurred())
				Expect(yes).To(BeFalse())
			})
		})
	})

	Describe("#ListResourcePatterns", func() {
		It("returns the list of resource patterns for which the actor has that permission", func() {
			roleName := uuid.NewV4().String()
			actor := perm.Actor{
				ID:        uuid.NewV4().String(),
				Namespace: uuid.NewV4().String(),
			}
			action := uuid.NewV4().String()
			resourcePattern1 := uuid.NewV4().String()
			resourcePattern2 := uuid.NewV4().String()
			resourcePattern3 := uuid.NewV4().String()

			permission1 := &perm.Permission{
				Action:          action,
				ResourcePattern: resourcePattern1,
			}
			permission2 := &perm.Permission{
				Action:          action,
				ResourcePattern: resourcePattern2,
			}
			permission3 := &perm.Permission{
				Action:          "another-action",
				ResourcePattern: resourcePattern3,
			}

			_, err := roleRepo.CreateRole(ctx, logger, roleName, permission1, permission2, permission3)
			Expect(err).NotTo(HaveOccurred())

			err = roleRepo.AssignRole(ctx, logger, roleName, actor.ID, actor.Namespace)
			Expect(err).NotTo(HaveOccurred())

			permission4 := &perm.Permission{
				Action:          action,
				ResourcePattern: "should-not-have-this-resource-pattern",
			}

			_, err = roleRepo.CreateRole(ctx, logger, "not-assigned-to-this-role", permission4)
			Expect(err).NotTo(HaveOccurred())

			query := repos.ListResourcePatternsQuery{
				Actor:  actor,
				Action: action,
			}

			resourcePatterns, err := subject.ListResourcePatterns(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(resourcePatterns).NotTo(BeNil())
			Expect(resourcePatterns).To(HaveLen(2))
			Expect(resourcePatterns).To(ConsistOf(resourcePattern1, resourcePattern2))
		})

		It("de-dupes the results if the user has access to the same resource pattern through multiple roles/permissions", func() {
			roleName1 := uuid.NewV4().String()
			actor := perm.Actor{
				ID:        uuid.NewV4().String(),
				Namespace: uuid.NewV4().String(),
			}
			action := uuid.NewV4().String()
			resourcePattern := uuid.NewV4().String()

			permission1 := &perm.Permission{
				Action:          action,
				ResourcePattern: resourcePattern,
			}

			_, err := roleRepo.CreateRole(ctx, logger, roleName1, permission1)
			Expect(err).NotTo(HaveOccurred())

			err = roleRepo.AssignRole(ctx, logger, roleName1, actor.ID, actor.Namespace)
			Expect(err).NotTo(HaveOccurred())

			roleName2 := uuid.NewV4().String()
			permission2 := &perm.Permission{
				Action:          action,
				ResourcePattern: resourcePattern,
			}

			_, err = roleRepo.CreateRole(ctx, logger, roleName2, permission2)
			Expect(err).NotTo(HaveOccurred())

			err = roleRepo.AssignRole(ctx, logger, roleName2, actor.ID, actor.Namespace)
			Expect(err).NotTo(HaveOccurred())

			query := repos.ListResourcePatternsQuery{
				Actor:  actor,
				Action: action,
			}

			resourcePatterns, err := subject.ListResourcePatterns(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(resourcePatterns).NotTo(BeNil())
			Expect(resourcePatterns).To(HaveLen(1))
			Expect(resourcePatterns).To(ConsistOf(resourcePattern))
		})

		It("returns empty if the actor is not assigned to any roles with that permission", func() {
			actor := perm.Actor{
				ID:        uuid.NewV4().String(),
				Namespace: uuid.NewV4().String(),
			}
			action := uuid.NewV4().String()
			query := repos.ListResourcePatternsQuery{
				Actor:  actor,
				Action: action,
			}

			resourcePatterns, err := subject.ListResourcePatterns(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(resourcePatterns).To(BeEmpty())
		})

		Context("when providing groups", func() {
			var (
				actor                                                  perm.Actor
				groups                                                 []perm.Group
				action                                                 string
				role1Name, role2Name, role3Name                        string
				resourcePattern1, resourcePattern2, resourcePattern3   string
				permission1, permission2, permission3, otherPermission *perm.Permission
			)

			BeforeEach(func() {
				actor = perm.Actor{
					ID:        uuid.NewV4().String(),
					Namespace: uuid.NewV4().String(),
				}
				groups = []perm.Group{
					{ID: uuid.NewV4().String()},
					{ID: uuid.NewV4().String()},
				}
				action = uuid.NewV4().String()

				role1Name = uuid.NewV4().String()
				role2Name = uuid.NewV4().String()
				role3Name = uuid.NewV4().String()

				resourcePattern1 = uuid.NewV4().String()
				resourcePattern2 = uuid.NewV4().String()
				resourcePattern3 = uuid.NewV4().String()

				permission1 = &perm.Permission{
					Action:          action,
					ResourcePattern: resourcePattern1,
				}
				permission2 = &perm.Permission{
					Action:          action,
					ResourcePattern: resourcePattern2,
				}
				permission3 = &perm.Permission{
					Action:          action,
					ResourcePattern: resourcePattern3,
				}
				otherPermission = &perm.Permission{
					Action:          "action-that-should-not-appear",
					ResourcePattern: "some-resource-pattern",
				}
			})

			Context("when both the actor and groups have assigned permissions", func() {
				It("returns a list of all permissions", func() {
					_, err := roleRepo.CreateRole(ctx, logger, role1Name, permission1)
					Expect(err).NotTo(HaveOccurred())
					err = roleRepo.AssignRole(ctx, logger, role1Name, actor.ID, actor.Namespace)
					Expect(err).NotTo(HaveOccurred())

					_, err = roleRepo.CreateRole(ctx, logger, role2Name, permission2)
					Expect(err).NotTo(HaveOccurred())
					err = roleRepo.AssignRoleToGroup(ctx, logger, role2Name, groups[0].ID)
					Expect(err).NotTo(HaveOccurred())

					_, err = roleRepo.CreateRole(ctx, logger, role3Name, permission3)
					Expect(err).NotTo(HaveOccurred())
					err = roleRepo.AssignRoleToGroup(ctx, logger, role3Name, groups[1].ID)
					Expect(err).NotTo(HaveOccurred())

					query := repos.ListResourcePatternsQuery{
						Actor:  actor,
						Action: action,
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
					_, err := roleRepo.CreateRole(ctx, logger, role1Name, permission1)
					Expect(err).NotTo(HaveOccurred())
					err = roleRepo.AssignRoleToGroup(ctx, logger, role1Name, groups[0].ID)
					Expect(err).NotTo(HaveOccurred())

					_, err = roleRepo.CreateRole(ctx, logger, role2Name, otherPermission)
					Expect(err).NotTo(HaveOccurred())
					err = roleRepo.AssignRoleToGroup(ctx, logger, role2Name, groups[1].ID)
					Expect(err).NotTo(HaveOccurred())

					query := repos.ListResourcePatternsQuery{
						Actor:  actor,
						Action: action,
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
					_, err := roleRepo.CreateRole(ctx, logger, role1Name, permission1, permission2)
					Expect(err).NotTo(HaveOccurred())
					err = roleRepo.AssignRoleToGroup(ctx, logger, role1Name, groups[0].ID)
					Expect(err).NotTo(HaveOccurred())

					query := repos.ListResourcePatternsQuery{
						Actor:  actor,
						Action: action,
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
