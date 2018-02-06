package reposbehaviors_test

import (
	"context"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/models"
	"code.cloudfoundry.org/perm/repos"
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
			roleName := models.RoleName(uuid.NewV4().String())
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			permission := &models.Permission{
				Name:            "some-permission",
				ResourcePattern: "some-resource-ID",
			}
			_, err := roleRepo.CreateRole(ctx, logger, roleName, permission)
			Expect(err).NotTo(HaveOccurred())

			err = roleAssignmentRepo.AssignRole(ctx, logger, roleName, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			permissionQuery := repos.PermissionQuery{
				PermissionName:  "some-permission",
				ResourcePattern: "some-resource-ID",
			}

			actorQuery := repos.ActorQuery{
				DomainID: domainID,
				Issuer:   issuer,
			}

			query := repos.HasPermissionQuery{
				PermissionQuery: permissionQuery,
				ActorQuery:      actorQuery,
			}

			yes, err := subject.HasPermission(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeTrue())
		})

		It("returns false if they have not been assigned the role", func() {
			roleName := models.RoleName(uuid.NewV4().String())
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			permission := &models.Permission{
				Name:            "some-permission",
				ResourcePattern: "some-resource-ID",
			}
			_, err := roleRepo.CreateRole(ctx, logger, roleName, permission)
			Expect(err).NotTo(HaveOccurred())

			_, err = actorRepo.CreateActor(ctx, logger, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			permissionQuery := repos.PermissionQuery{
				PermissionName:  "some-permission",
				ResourcePattern: "some-resource-ID",
			}

			actorQuery := repos.ActorQuery{
				DomainID: domainID,
				Issuer:   issuer,
			}

			query := repos.HasPermissionQuery{
				PermissionQuery: permissionQuery,
				ActorQuery:      actorQuery,
			}
			yes, err := subject.HasPermission(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeFalse())
		})

		It("return false if the actor does not exist", func() {
			roleName := models.RoleName(uuid.NewV4().String())
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			_, err := roleRepo.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			permissionQuery := repos.PermissionQuery{
				PermissionName:  "some-permission",
				ResourcePattern: "some-resource-ID",
			}

			actorQuery := repos.ActorQuery{
				DomainID: domainID,
				Issuer:   issuer,
			}

			query := repos.HasPermissionQuery{
				PermissionQuery: permissionQuery,
				ActorQuery:      actorQuery,
			}
			yes, err := subject.HasPermission(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeFalse())
		})
	})

	Describe("#ListResourcePatterns", func() {
		It("returns the list of resource patterns for which the actor has that permission", func() {
			roleName := models.RoleName(uuid.NewV4().String())
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())
			permissionName := models.PermissionName(uuid.NewV4().String())
			resourcePattern1 := models.PermissionResourcePattern(uuid.NewV4().String())
			resourcePattern2 := models.PermissionResourcePattern(uuid.NewV4().String())
			resourcePattern3 := models.PermissionResourcePattern(uuid.NewV4().String())

			permission1 := &models.Permission{
				Name:            permissionName,
				ResourcePattern: resourcePattern1,
			}
			permission2 := &models.Permission{
				Name:            permissionName,
				ResourcePattern: resourcePattern2,
			}
			permission3 := &models.Permission{
				Name:            models.PermissionName("another-permission"),
				ResourcePattern: resourcePattern3,
			}

			_, err := roleRepo.CreateRole(ctx, logger, roleName, permission1, permission2, permission3)
			Expect(err).NotTo(HaveOccurred())

			err = roleAssignmentRepo.AssignRole(ctx, logger, roleName, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			permission4 := &models.Permission{
				Name:            permissionName,
				ResourcePattern: models.PermissionResourcePattern("should-not-have-this-resource-pattern"),
			}

			_, err = roleRepo.CreateRole(ctx, logger, models.RoleName("not-assigned-to-this-role"), permission4)
			Expect(err).NotTo(HaveOccurred())

			actorQuery := repos.ActorQuery{
				DomainID: domainID,
				Issuer:   issuer,
			}
			permissionDefinitionQuery := repos.PermissionDefinitionQuery{
				Name: permissionName,
			}
			query := repos.ListResourcePatternsQuery{
				ActorQuery:                actorQuery,
				PermissionDefinitionQuery: permissionDefinitionQuery,
			}

			resourcePatterns, err := subject.ListResourcePatterns(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(resourcePatterns).NotTo(BeNil())
			Expect(resourcePatterns).To(HaveLen(2))
			Expect(resourcePatterns).To(ConsistOf(resourcePattern1, resourcePattern2))
		})

		It("de-dupes the results if the user has access to the same resource pattern through multiple roles/permissions", func() {
			roleName1 := models.RoleName(uuid.NewV4().String())
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())
			permissionName := models.PermissionName(uuid.NewV4().String())
			resourcePattern := models.PermissionResourcePattern(uuid.NewV4().String())

			permission1 := &models.Permission{
				Name:            permissionName,
				ResourcePattern: resourcePattern,
			}

			_, err := roleRepo.CreateRole(ctx, logger, roleName1, permission1)
			Expect(err).NotTo(HaveOccurred())

			err = roleAssignmentRepo.AssignRole(ctx, logger, roleName1, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			roleName2 := models.RoleName(uuid.NewV4().String())
			permission2 := &models.Permission{
				Name:            permissionName,
				ResourcePattern: resourcePattern,
			}

			_, err = roleRepo.CreateRole(ctx, logger, roleName2, permission2)
			Expect(err).NotTo(HaveOccurred())

			err = roleAssignmentRepo.AssignRole(ctx, logger, roleName2, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			actorQuery := repos.ActorQuery{
				DomainID: domainID,
				Issuer:   issuer,
			}
			permissionDefinitionQuery := repos.PermissionDefinitionQuery{
				Name: permissionName,
			}
			query := repos.ListResourcePatternsQuery{
				ActorQuery:                actorQuery,
				PermissionDefinitionQuery: permissionDefinitionQuery,
			}

			resourcePatterns, err := subject.ListResourcePatterns(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(resourcePatterns).NotTo(BeNil())
			Expect(resourcePatterns).To(HaveLen(1))
			Expect(resourcePatterns).To(ConsistOf(resourcePattern))
		})

		It("returns empty if the actor is not assigned to any roles with that permission", func() {
			actorQuery := repos.ActorQuery{
				DomainID: "fake-actor",
				Issuer:   "fake-issuer",
			}
			permissionName := models.PermissionName("fake-permission-name")
			permissionDefinitionQuery := repos.PermissionDefinitionQuery{
				Name: permissionName,
			}
			query := repos.ListResourcePatternsQuery{
				ActorQuery:                actorQuery,
				PermissionDefinitionQuery: permissionDefinitionQuery,
			}

			resourcePatterns, err := subject.ListResourcePatterns(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(resourcePatterns).To(BeEmpty())
		})
	})
}
