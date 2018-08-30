package perm_test

import (
	. "code.cloudfoundry.org/perm"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Client", func() {
	Describe("#Dial", func() {
		It("succeeds when TLS config is supplied", func() {
			server := ghttp.NewTLSServer()
			defer server.Close()

			client, err := Dial(server.Addr(), WithTLSConfig(server.HTTPTestServer.TLS))
			Expect(err).NotTo(HaveOccurred())

			Expect(client).NotTo(BeNil())
		})

		It("fails when no transport security is supplied", func() {
			server := ghttp.NewTLSServer()
			defer server.Close()

			_, err := Dial(server.Addr())

			Expect(err).To(MatchError("perm: no transport security set (use perm.WithTLSConfig() to set)"))
		})
	})

	Describe("#Close", func() {
		It("succeeds on the first call only", func() {
			server := ghttp.NewTLSServer()
			defer server.Close()

			client, err := Dial(server.Addr(), WithTLSConfig(server.HTTPTestServer.TLS))
			Expect(err).NotTo(HaveOccurred())

			err = client.Close()
			Expect(err).NotTo(HaveOccurred())

			err = client.Close()
			Expect(err).To(MatchError("perm: the client connection is already closing or closed"))
		})
	})
})
