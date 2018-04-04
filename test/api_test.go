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
}

type serverConfig struct {
	addr      string
	tlsConfig *tls.Config
}
