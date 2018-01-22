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

func BehavesLikeARoleAssignmentService(
	subjectCreator func() models.RoleAssignmentService,
	roleServiceCreator func() models.RoleService,
	actorServiceCreator func() models.ActorService,
) {
	var (
		subject models.RoleAssignmentService

		roleService  models.RoleService
		actorService models.ActorService

		ctx    context.Context
		logger *lagertest.TestLogger

		cancelFunc context.CancelFunc

		roleName models.RoleName
	)

	BeforeEach(func() {
		roleName = models.RoleName(uuid.NewV4().String())

		subject = subjectCreator()

		roleService = roleServiceCreator()
		actorService = actorServiceCreator()

		ctx, cancelFunc = context.WithTimeout(context.Background(), 1*time.Second)
		logger = lagertest.NewTestLogger("perm-test")
	})

	AfterEach(func() {
		cancelFunc()
	})

	Describe("#AssignRole", func() {
		It("saves the role assignment, saving the actor if it does not exist", func() {
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			_, err := roleService.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRole(ctx, logger, roleName, domainID, issuer)

			Expect(err).NotTo(HaveOccurred())

			roleQuery := models.RoleQuery{Name: roleName}
			actorQuery := models.ActorQuery{
				DomainID: domainID,
				Issuer:   issuer,
			}
			query := models.RoleAssignmentQuery{
				RoleQuery:  roleQuery,
				ActorQuery: actorQuery,
			}
			yes, err := subject.HasRole(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeTrue())
		})

		It("fails if the role assignment already exists", func() {
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			_, err := roleService.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRole(ctx, logger, roleName, domainID, issuer)

			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRole(ctx, logger, roleName, domainID, issuer)
			Expect(err).To(Equal(models.ErrRoleAssignmentAlreadyExists))
		})

		It("fails if the role does not exist", func() {
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			err := subject.AssignRole(ctx, logger, roleName, domainID, issuer)
			Expect(err).To(Equal(models.ErrRoleNotFound))
		})
	})

	Describe("#UnassignRole", func() {
		It("removes the role assignment", func() {
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			_, err := actorService.CreateActor(ctx, logger, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			_, err = roleService.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRole(ctx, logger, roleName, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			err = subject.UnassignRole(ctx, logger, roleName, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			roleQuery := models.RoleQuery{Name: roleName}
			actorQuery := models.ActorQuery{
				DomainID: domainID,
				Issuer:   issuer,
			}
			query := models.RoleAssignmentQuery{
				RoleQuery:  roleQuery,
				ActorQuery: actorQuery,
			}
			yes, err := subject.HasRole(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeFalse())
		})

		It("fails if the actor does not exist", func() {
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			_, err := roleService.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.UnassignRole(ctx, logger, roleName, domainID, issuer)
			Expect(err).To(MatchError(models.ErrActorNotFound))
		})

		It("fails if the role does not exist", func() {
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			err := subject.UnassignRole(ctx, logger, roleName, domainID, issuer)
			Expect(err).To(Equal(models.ErrRoleNotFound))
		})

		It("fails if the role assignment does not exist", func() {
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			_, err := actorService.CreateActor(ctx, logger, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			_, err = roleService.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.UnassignRole(ctx, logger, roleName, domainID, issuer)
			Expect(err).To(Equal(models.ErrRoleAssignmentNotFound))
		})
	})

	Describe("#HasRole", func() {
		It("returns true if they have been assigned the role", func() {
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			_, err := roleService.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRole(ctx, logger, roleName, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			roleQuery := models.RoleQuery{Name: roleName}
			actorQuery := models.ActorQuery{
				DomainID: domainID,
				Issuer:   issuer,
			}
			query := models.RoleAssignmentQuery{
				RoleQuery:  roleQuery,
				ActorQuery: actorQuery,
			}
			yes, err := subject.HasRole(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeTrue())
		})

		It("returns false if they have not been assigned the role", func() {
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			_, err := actorService.CreateActor(ctx, logger, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			_, err = roleService.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			roleQuery := models.RoleQuery{Name: roleName}
			actorQuery := models.ActorQuery{
				DomainID: domainID,
				Issuer:   issuer,
			}
			query := models.RoleAssignmentQuery{
				RoleQuery:  roleQuery,
				ActorQuery: actorQuery,
			}
			yes, err := subject.HasRole(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeFalse())
		})

		It("fails if the role does not exist", func() {
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			_, err := actorService.CreateActor(ctx, logger, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			roleQuery := models.RoleQuery{Name: roleName}
			actorQuery := models.ActorQuery{
				DomainID: domainID,
				Issuer:   issuer,
			}
			query := models.RoleAssignmentQuery{
				RoleQuery:  roleQuery,
				ActorQuery: actorQuery,
			}
			_, err = subject.HasRole(ctx, logger, query)

			Expect(err).To(MatchError(models.ErrRoleNotFound))
		})

		It("fails if the actor does not exist", func() {
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			_, err := roleService.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			roleQuery := models.RoleQuery{Name: roleName}
			actorQuery := models.ActorQuery{
				DomainID: domainID,
				Issuer:   issuer,
			}
			query := models.RoleAssignmentQuery{
				RoleQuery:  roleQuery,
				ActorQuery: actorQuery,
			}
			_, err = subject.HasRole(ctx, logger, query)

			Expect(err).To(MatchError(models.ErrActorNotFound))
		})
	})

	Describe("#ListActorRoles", func() {
		It("returns a list of all roles that the actor has been assigned to", func() {
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			roleName1 := models.RoleName(uuid.NewV4().String())
			role1, err := roleService.CreateRole(ctx, logger, roleName1)
			Expect(err).NotTo(HaveOccurred())

			roleName2 := models.RoleName(uuid.NewV4().String())
			role2, err := roleService.CreateRole(ctx, logger, roleName2)
			Expect(err).NotTo(HaveOccurred())

			roleName3 := models.RoleName(uuid.NewV4().String())
			role3, err := roleService.CreateRole(ctx, logger, roleName3)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRole(ctx, logger, roleName1, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRole(ctx, logger, roleName2, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			query := models.ActorQuery{
				DomainID: domainID,
				Issuer:   issuer,
			}
			roles, err := subject.ListActorRoles(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(roles).To(HaveLen(2))
			Expect(roles).To(ContainElement(role1))
			Expect(roles).To(ContainElement(role2))
			Expect(roles).NotTo(ContainElement(role3))
		})

		It("fails if the actor does not exist", func() {
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			_, err := roleService.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			query := models.ActorQuery{
				DomainID: domainID,
				Issuer:   issuer,
			}
			_, err = subject.ListActorRoles(ctx, logger, query)
			Expect(err).To(MatchError(models.ErrActorNotFound))
		})
	})
}
