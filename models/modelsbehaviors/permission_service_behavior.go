package modelsbehaviors_test

import (
	"context"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	"github.com/satori/go.uuid"
)

func BehavesLikeAPermissionService(subjectCreator func() models.PermissionService, roleServiceCreator func() models.RoleService, actorServiceCreator func() models.ActorService, roleAssignmentServiceCreator func() models.RoleAssignmentService) {
	var (
		subject models.PermissionService

		roleService           models.RoleService
		actorService          models.ActorService
		roleAssignmentService models.RoleAssignmentService

		ctx    context.Context
		logger *lagertest.TestLogger

		cancelFunc context.CancelFunc
	)

	BeforeEach(func() {
		subject = subjectCreator()

		roleService = roleServiceCreator()
		actorService = actorServiceCreator()
		roleAssignmentService = roleAssignmentServiceCreator()

		ctx, cancelFunc = context.WithTimeout(context.Background(), 1*time.Second)
		logger = lagertest.NewTestLogger("perm-test")
	})

	AfterEach(func() {
		cancelFunc()
	})

	Describe("#HasPermission", func() {
		It("returns true if they have been assigned a role with a permission with a name matching the permission name and a resource pattern that matches the resourceID of the query", func() {
			roleName := uuid.NewV4().String()
			domainID := uuid.NewV4().String()
			issuer := uuid.NewV4().String()

			permission := &models.Permission{
				Name:            "some-permission",
				ResourcePattern: "some-resource-ID",
			}
			_, err := roleService.CreateRole(ctx, logger, roleName, permission)
			Expect(err).NotTo(HaveOccurred())

			err = roleAssignmentService.AssignRole(ctx, logger, roleName, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			permissionQuery := models.PermissionQuery{
				PermissionDefinitionQuery: models.PermissionDefinitionQuery{
					Name: "some-permission",
				},
				ResourceID: "some-resource-ID",
			}

			actorQuery := models.ActorQuery{
				DomainID: domainID,
				Issuer:   issuer,
			}

			query := models.HasPermissionQuery{
				PermissionQuery: permissionQuery,
				ActorQuery:      actorQuery,
			}

			yes, err := subject.HasPermission(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeTrue())
		})

		It("returns false if they have not been assigned the role", func() {
			roleName := uuid.NewV4().String()
			domainID := uuid.NewV4().String()
			issuer := uuid.NewV4().String()

			permission := &models.Permission{
				Name:            "some-permission",
				ResourcePattern: "some-resource-ID",
			}
			_, err := roleService.CreateRole(ctx, logger, roleName, permission)
			Expect(err).NotTo(HaveOccurred())

			_, err = actorService.CreateActor(ctx, logger, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			permissionQuery := models.PermissionQuery{
				PermissionDefinitionQuery: models.PermissionDefinitionQuery{
					Name: "some-permission",
				},
				ResourceID: "some-resource-ID",
			}

			actorQuery := models.ActorQuery{
				DomainID: domainID,
				Issuer:   issuer,
			}

			query := models.HasPermissionQuery{
				PermissionQuery: permissionQuery,
				ActorQuery:      actorQuery,
			}
			yes, err := subject.HasPermission(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeFalse())
		})

		It("return false if the actor does not exist", func() {
			roleName := uuid.NewV4().String()
			domainID := uuid.NewV4().String()
			issuer := uuid.NewV4().String()

			_, err := roleService.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			permissionQuery := models.PermissionQuery{
				PermissionDefinitionQuery: models.PermissionDefinitionQuery{
					Name: "some-permission",
				},
				ResourceID: "some-resource-ID",
			}

			actorQuery := models.ActorQuery{
				DomainID: domainID,
				Issuer:   issuer,
			}

			query := models.HasPermissionQuery{
				PermissionQuery: permissionQuery,
				ActorQuery:      actorQuery,
			}
			yes, err := subject.HasPermission(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeFalse())
		})
	})
}
