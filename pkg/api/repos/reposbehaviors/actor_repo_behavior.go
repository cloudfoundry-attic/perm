package reposbehaviors_test

import (
	"context"

	"time"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/models"
	"code.cloudfoundry.org/perm/pkg/api/repos"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

func BehavesLikeAnActorRepo(actorRepoCreator func() repos.ActorRepo) {
	var (
		subject repos.ActorRepo

		ctx    context.Context
		logger *lagertest.TestLogger

		cancelFunc context.CancelFunc
	)

	BeforeEach(func() {
		subject = actorRepoCreator()

		ctx, cancelFunc = context.WithTimeout(context.Background(), 1*time.Second)
		logger = lagertest.NewTestLogger("perm-test")
	})

	AfterEach(func() {
		cancelFunc()
	})

	Describe("#CreateActor", func() {
		It("saves the actor", func() {
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			actor, err := subject.CreateActor(ctx, logger, domainID, issuer)

			Expect(err).NotTo(HaveOccurred())

			Expect(actor).NotTo(BeNil())
			Expect(actor.DomainID).To(Equal(domainID))
			Expect(actor.Issuer).To(Equal(issuer))

			_, err = subject.CreateActor(ctx, logger, domainID, issuer)

			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(models.ErrActorAlreadyExists))
		})

		It("fails if an actor with the domain ID/issuer combo already exists", func() {
			domainID := models.ActorDomainID(uuid.NewV4().String())
			issuer := models.ActorIssuer(uuid.NewV4().String())

			_, err := subject.CreateActor(ctx, logger, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			_, err = subject.CreateActor(ctx, logger, domainID, issuer)
			Expect(err).To(Equal(models.ErrActorAlreadyExists))

			uniqueDomainID := models.ActorDomainID(uuid.NewV4().String())
			_, err = subject.CreateActor(ctx, logger, uniqueDomainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			uniqueIssuer := models.ActorIssuer(uuid.NewV4().String())
			_, err = subject.CreateActor(ctx, logger, domainID, uniqueIssuer)
			Expect(err).NotTo(HaveOccurred())
		})
	})
}
