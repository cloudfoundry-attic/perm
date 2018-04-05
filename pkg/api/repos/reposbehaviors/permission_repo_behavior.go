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
	actorRepoCreator func() repos.ActorRepo,
	roleAssignmentRepoCreator func() repos.RoleAssignmentRepo,
) {
	var (
		subject repos.PermissionRepo

		roleRepo           repos.RoleRepo
		actorRepo          repos.ActorRepo
		roleAssignmentRepo repos.RoleAssignmentRepo

		ctx    context.Context
		logger *lagertest.TestLogger

		cancelFunc context.CancelFunc
	)

	BeforeEach(func() {
		subject = subjectCreator()

		roleRepo = roleRepoCreator()
		actorRepo = actorRepoCreator()
		roleAssignmentRepo = roleAssignmentRepoCreator()

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
				Action:          "some-permission",
				ResourcePattern: "some-resource-ID",
			}

			_, err := roleRepo.CreateRole(ctx, logger, roleName, &permission)
			Expect(err).NotTo(HaveOccurred())

			err = roleAssignmentRepo.AssignRole(ctx, logger, roleName, actor.ID, actor.Namespace)
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
				Action:          "some-permission",
				ResourcePattern: "some-resource-ID",
			}

			_, err := roleRepo.CreateRole(ctx, logger, roleName, &permission)
			Expect(err).NotTo(HaveOccurred())

			_, err = actorRepo.CreateActor(ctx, logger, actor.ID, actor.Namespace)
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
				Action:          "some-permission",
				ResourcePattern: "some-resource-ID",
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
				Action:          "another-permission",
				ResourcePattern: resourcePattern3,
			}

			_, err := roleRepo.CreateRole(ctx, logger, roleName, permission1, permission2, permission3)
			Expect(err).NotTo(HaveOccurred())

			err = roleAssignmentRepo.AssignRole(ctx, logger, roleName, actor.ID, actor.Namespace)
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

			err = roleAssignmentRepo.AssignRole(ctx, logger, roleName1, actor.ID, actor.Namespace)
			Expect(err).NotTo(HaveOccurred())

			roleName2 := uuid.NewV4().String()
			permission2 := &perm.Permission{
				Action:          action,
				ResourcePattern: resourcePattern,
			}

			_, err = roleRepo.CreateRole(ctx, logger, roleName2, permission2)
			Expect(err).NotTo(HaveOccurred())

			err = roleAssignmentRepo.AssignRole(ctx, logger, roleName2, actor.ID, actor.Namespace)
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
	})
}
