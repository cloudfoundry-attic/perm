package modelsbehaviors_test

import (
	"context"

	"time"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

func BehavesLikeAnActorService(actorServiceCreator func() models.ActorService) {
	var (
		subject models.ActorService

		ctx    context.Context
		logger *lagertest.TestLogger

		cancelFunc context.CancelFunc
	)

	BeforeEach(func() {
		subject = actorServiceCreator()

		ctx, cancelFunc = context.WithTimeout(context.Background(), 1*time.Second)
		logger = lagertest.NewTestLogger("perm-test")
	})

	AfterEach(func() {
		cancelFunc()
	})

	Describe("#CreateActor", func() {
		It("saves the actor", func() {
			domainID := uuid.NewV4().String()
			issuer := uuid.NewV4().String()

			actor, err := subject.CreateActor(ctx, logger, domainID, issuer)

			Expect(err).NotTo(HaveOccurred())

			Expect(actor).NotTo(BeNil())
			Expect(actor.DomainID).To(Equal(domainID))
			Expect(actor.Issuer).To(Equal(issuer))

			expectedActor := actor
			actor, err = subject.FindActor(ctx, logger, models.ActorQuery{DomainID: domainID, Issuer: issuer})

			Expect(err).NotTo(HaveOccurred())
			Expect(actor).To(Equal(expectedActor))
		})

		It("fails if an actor with the domain ID/issuer combo already exists", func() {
			domainID := uuid.NewV4().String()
			issuer := uuid.NewV4().String()

			_, err := subject.CreateActor(ctx, logger, domainID, issuer)
			Expect(err).NotTo(HaveOccurred())

			_, err = subject.CreateActor(ctx, logger, domainID, issuer)
			Expect(err).To(Equal(models.ErrActorAlreadyExists))

			_, err = subject.CreateActor(ctx, logger, uuid.NewV4().String(), issuer)
			Expect(err).NotTo(HaveOccurred())

			_, err = subject.CreateActor(ctx, logger, domainID, uuid.NewV4().String())
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("#FindActor", func() {
		It("fails if the actor does not exist", func() {
			domainID := uuid.NewV4().String()
			issuer := uuid.NewV4().String()
			_, err := subject.FindActor(ctx, logger, models.ActorQuery{DomainID: domainID, Issuer: issuer})

			Expect(err).To(Equal(models.ErrActorNotFound))
		})
	})
}
