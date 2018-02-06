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

func BehavesLikeARoleAssignmentRepo(
	subjectCreator func() repos.RoleAssignmentRepo,
	roleRepoCreator func() repos.RoleRepo,
	actorRepoCreator func() repos.ActorRepo,
) {
	var (
		subject repos.RoleAssignmentRepo

		roleRepo  repos.RoleRepo
		actorRepo repos.ActorRepo

		ctx    context.Context
		logger *lagertest.TestLogger

		cancelFunc context.CancelFunc

		roleName models.RoleName
	)

	BeforeEach(func() {
		roleName = models.RoleName(uuid.NewV4().String())

		subject = subjectCreator()

		roleRepo = roleRepoCreator()
		actorRepo = actorRepoCreator()

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

			_, err := roleRepo.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRole(ctx, logger, roleName, domainID, issuer)

			Expect(err).NotTo(HaveOccurred())

			roleQuery := repos.RoleQuery{Name: roleName}
			actorQuery := repos.ActorQuery{
				DomainID: domainID,
				Issuer:   issuer,
			}
			query := repos.RoleAssignmentQuery{
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

			_, err := roleRepo.CreateRole(ctx, logger, roleName)
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

			_, err := actorRepo.CreateActor(ctx, logger, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			_, err = roleRepo.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRole(ctx, logger, roleName, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			err = subject.UnassignRole(ctx, logger, roleName, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			roleQuery := repos.RoleQuery{Name: roleName}
			actorQuery := repos.ActorQuery{
				DomainID: domainID,
				Issuer:   issuer,
			}
			query := repos.RoleAssignmentQuery{
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

			err := subject.UnassignRole(ctx, logger, roleName, domainID, issuer)
			Expect(err).To(Equal(models.ErrRoleNotFound))
		})

		It("fails if the actor does not exist", func() {
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			_, err := roleRepo.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.UnassignRole(ctx, logger, roleName, domainID, issuer)
			Expect(err).To(MatchError(models.ErrRoleAssignmentNotFound))
		})

		It("fails if the role assignment does not exist", func() {
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			_, err := actorRepo.CreateActor(ctx, logger, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			_, err = roleRepo.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.UnassignRole(ctx, logger, roleName, domainID, issuer)
			Expect(err).To(Equal(models.ErrRoleAssignmentNotFound))
		})
	})

	Describe("#HasRole", func() {
		It("returns true if they have been assigned the role", func() {
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			_, err := roleRepo.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRole(ctx, logger, roleName, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			roleQuery := repos.RoleQuery{Name: roleName}
			actorQuery := repos.ActorQuery{
				DomainID: domainID,
				Issuer:   issuer,
			}
			query := repos.RoleAssignmentQuery{
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

			_, err := actorRepo.CreateActor(ctx, logger, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			_, err = roleRepo.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			roleQuery := repos.RoleQuery{Name: roleName}
			actorQuery := repos.ActorQuery{
				DomainID: domainID,
				Issuer:   issuer,
			}
			query := repos.RoleAssignmentQuery{
				RoleQuery:  roleQuery,
				ActorQuery: actorQuery,
			}
			yes, err := subject.HasRole(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(yes).To(BeFalse())
		})

		It("returns false if the actor does not exist", func() {
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			_, err := roleRepo.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			roleQuery := repos.RoleQuery{Name: roleName}
			actorQuery := repos.ActorQuery{
				DomainID: domainID,
				Issuer:   issuer,
			}
			query := repos.RoleAssignmentQuery{
				RoleQuery:  roleQuery,
				ActorQuery: actorQuery,
			}
			hasRole, err := subject.HasRole(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(hasRole).To(BeFalse())
		})

		It("fails if the role does not exist", func() {
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			_, err := actorRepo.CreateActor(ctx, logger, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			roleQuery := repos.RoleQuery{Name: roleName}
			actorQuery := repos.ActorQuery{
				DomainID: domainID,
				Issuer:   issuer,
			}
			query := repos.RoleAssignmentQuery{
				RoleQuery:  roleQuery,
				ActorQuery: actorQuery,
			}
			_, err = subject.HasRole(ctx, logger, query)

			Expect(err).To(MatchError(models.ErrRoleNotFound))
		})
	})

	Describe("#ListActorRoles", func() {
		It("returns a list of all roles that the actor has been assigned to", func() {
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			roleName1 := models.RoleName(uuid.NewV4().String())
			role1, err := roleRepo.CreateRole(ctx, logger, roleName1)
			Expect(err).NotTo(HaveOccurred())

			roleName2 := models.RoleName(uuid.NewV4().String())
			role2, err := roleRepo.CreateRole(ctx, logger, roleName2)
			Expect(err).NotTo(HaveOccurred())

			roleName3 := models.RoleName(uuid.NewV4().String())
			role3, err := roleRepo.CreateRole(ctx, logger, roleName3)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRole(ctx, logger, roleName1, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			err = subject.AssignRole(ctx, logger, roleName2, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			query := repos.ActorQuery{
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

		It("returns an empty list if the actor does not exist", func() {
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			_, err := roleRepo.CreateRole(ctx, logger, roleName)
			Expect(err).NotTo(HaveOccurred())

			query := repos.ActorQuery{
				DomainID: domainID,
				Issuer:   issuer,
			}
			roles, err := subject.ListActorRoles(ctx, logger, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(roles).To(BeEmpty())
		})
	})
}
