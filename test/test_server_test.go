package perm_test

import (
	"crypto/tls"
	"crypto/x509"
	"net"

	"code.cloudfoundry.org/perm/pkg/api/apitest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test server", func() {
	var (
		listener net.Listener

		subject *apitest.TestServer
	)

	BeforeEach(func() {
		var err error

		// Port 0 should find a random open port
		listener, err = net.Listen("tcp", "localhost:0")
		Expect(err).NotTo(HaveOccurred())

		cert, err := tls.X509KeyPair([]byte(testCert), []byte(testCertKey))
		Expect(err).NotTo(HaveOccurred())

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		subject = apitest.NewTestServer(apitest.WithTLSConfig(tlsConfig))

		go func() {
			err = subject.Serve(listener)
			Expect(err).NotTo(HaveOccurred())
		}()
	})

	AfterEach(func() {
		subject.Stop()
	})

	testAPI(func() serverConfig {
		rootCAPool := x509.NewCertPool()

		ok := rootCAPool.AppendCertsFromPEM([]byte(testCA))
		Expect(ok).To(BeTrue())

		return serverConfig{
			addr: listener.Addr().String(),
			tlsConfig: &tls.Config{
				RootCAs: rootCAPool,
			},
		}
	})
})
