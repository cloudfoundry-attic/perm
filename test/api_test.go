package perm_test

import (
	"context"
	"crypto/tls"

	"code.cloudfoundry.org/perm/pkg/perm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
)

func testAPI(serverConfigFactory func() serverConfig) {
	var (
		client *perm.Client
	)

	BeforeEach(func() {
		var err error

		serverConfig := serverConfigFactory()

		client, err = perm.Dial(serverConfig.addr, perm.WithTLSConfig(serverConfig.tlsConfig))
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := client.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("#CreateRole", func() {
		It("returns the new role", func() {
			name := uuid.NewV4().String()

			role, err := client.CreateRole(context.Background(), name)
			Expect(err).NotTo(HaveOccurred())

			Expect(role.Name).To(Equal(name))
		})

		It("fails when a role with the same name already exists", func() {
			name := uuid.NewV4().String()

			_, err := client.CreateRole(context.Background(), name)
			Expect(err).NotTo(HaveOccurred())

			_, err = client.CreateRole(context.Background(), name)
			Expect(err).To(MatchError("role already exists"))
		})
	})

	Describe("#DeleteRole", func() {
		It("succeeds when the role exists", func() {
			role, err := client.CreateRole(context.Background(), uuid.NewV4().String())
			Expect(err).NotTo(HaveOccurred())

			err = client.DeleteRole(context.Background(), role.Name)
			Expect(err).NotTo(HaveOccurred())
		})

		It("fails when the role does not exist", func() {
			err := client.DeleteRole(context.Background(), uuid.NewV4().String())
			Expect(err).To(MatchError("role not found"))
		})
	})

	Describe("#AssignRole", func() {
		It("succeeds when the role exists and the actor has not yet been assigned to it", func() {
			role, err := client.CreateRole(context.Background(), uuid.NewV4().String())
			Expect(err).NotTo(HaveOccurred())

			actor := perm.Actor{
				ID:        uuid.NewV4().String(),
				Namespace: uuid.NewV4().String(),
			}

			err = client.AssignRole(context.Background(), role.Name, actor)
			Expect(err).NotTo(HaveOccurred())
		})

		It("allows assignments with the same actor ID but different namespaces", func() {
			role, err := client.CreateRole(context.Background(), uuid.NewV4().String())
			Expect(err).NotTo(HaveOccurred())

			id := uuid.NewV4().String()
			actor1 := perm.Actor{
				ID:        id,
				Namespace: uuid.NewV4().String(),
			}
			actor2 := perm.Actor{
				ID:        id,
				Namespace: uuid.NewV4().String(),
			}

			err = client.AssignRole(context.Background(), role.Name, actor1)
			Expect(err).NotTo(HaveOccurred())

			err = client.AssignRole(context.Background(), role.Name, actor2)
			Expect(err).NotTo(HaveOccurred())
		})

		It("allows assignments with the same namespace but different actor IDs", func() {
			role, err := client.CreateRole(context.Background(), uuid.NewV4().String())
			Expect(err).NotTo(HaveOccurred())

			namespace := uuid.NewV4().String()
			actor1 := perm.Actor{
				ID:        uuid.NewV4().String(),
				Namespace: namespace,
			}
			actor2 := perm.Actor{
				ID:        uuid.NewV4().String(),
				Namespace: namespace,
			}

			err = client.AssignRole(context.Background(), role.Name, actor1)
			Expect(err).NotTo(HaveOccurred())

			err = client.AssignRole(context.Background(), role.Name, actor2)
			Expect(err).NotTo(HaveOccurred())
		})

		It("fails when the role does not exist", func() {
			actor := perm.Actor{
				ID:        uuid.NewV4().String(),
				Namespace: uuid.NewV4().String(),
			}

			err := client.AssignRole(context.Background(), uuid.NewV4().String(), actor)
			Expect(err).To(MatchError("role not found"))
		})

		It("fails when the actor has already been assigned to the role", func() {
			role, err := client.CreateRole(context.Background(), uuid.NewV4().String())
			Expect(err).NotTo(HaveOccurred())

			actor := perm.Actor{
				ID:        uuid.NewV4().String(),
				Namespace: uuid.NewV4().String(),
			}

			err = client.AssignRole(context.Background(), role.Name, actor)
			Expect(err).NotTo(HaveOccurred())

			err = client.AssignRole(context.Background(), role.Name, actor)
			Expect(err).To(MatchError("assignment already exists"))
		})
	})
}

type serverConfig struct {
	addr      string
	tlsConfig *tls.Config
}
