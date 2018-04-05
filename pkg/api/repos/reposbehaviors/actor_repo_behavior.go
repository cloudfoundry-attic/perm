package reposbehaviors_test

import (
	"context"

	"time"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/pkg/api/repos"
	"code.cloudfoundry.org/perm/pkg/perm"
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
			id := uuid.NewV4().String()
			namespace := uuid.NewV4().String()

			actor, err := subject.CreateActor(ctx, logger, id, namespace)

			Expect(err).NotTo(HaveOccurred())

			Expect(actor).NotTo(BeNil())
			Expect(actor.ID).To(Equal(id))
			Expect(actor.Namespace).To(Equal(namespace))

			_, err = subject.CreateActor(ctx, logger, id, namespace)

			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(perm.ErrActorAlreadyExists))
		})

		It("fails if an actor with the domain ID/namespace combo already exists", func() {
			id := uuid.NewV4().String()
			namespace := uuid.NewV4().String()

			_, err := subject.CreateActor(ctx, logger, id, namespace)
			Expect(err).NotTo(HaveOccurred())

			_, err = subject.CreateActor(ctx, logger, id, namespace)
			Expect(err).To(Equal(perm.ErrActorAlreadyExists))

			uniqueID := uuid.NewV4().String()
			_, err = subject.CreateActor(ctx, logger, uniqueID, namespace)
			Expect(err).NotTo(HaveOccurred())

			uniqueNamespace := uuid.NewV4().String()
			_, err = subject.CreateActor(ctx, logger, id, uniqueNamespace)
			Expect(err).NotTo(HaveOccurred())
		})
	})
}
