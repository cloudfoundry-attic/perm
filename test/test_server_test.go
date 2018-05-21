package perm_test

import (
	"crypto/tls"
	"crypto/x509"
	"net"

	"code.cloudfoundry.org/perm/pkg/api"
	"code.cloudfoundry.org/perm/pkg/api/rpc"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test server", func() {
	var (
		listener net.Listener

		subject *api.Server

		listenerWithAuth net.Listener

		subjectWithAuth *api.Server
	)

	BeforeEach(func() {
		var err error

		cert, err := tls.X509KeyPair([]byte(testCert), []byte(testCertKey))
		Expect(err).NotTo(HaveOccurred())

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		store := rpc.NewInMemoryStore()

		subject = api.NewServer(store, api.WithTLSConfig(tlsConfig))

		// Port 0 should find a random open port
		listener, err = net.Listen("tcp", "localhost:0")
		Expect(err).NotTo(HaveOccurred())

		go func() {
			err = subject.Serve(listener)
			Expect(err).NotTo(HaveOccurred())
		}()

		subjectWithAuth = api.NewServer(rpc.NewInMemoryStore(), api.WithTLSConfig(tlsConfig), api.WithRequireAuth(true))

		listenerWithAuth, err = net.Listen("tcp", "localhost:0")
		Expect(err).NotTo(HaveOccurred())

		go func() {
			err = subjectWithAuth.Serve(listenerWithAuth)
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
	}, func() serverConfig {
		rootCAPool := x509.NewCertPool()

		ok := rootCAPool.AppendCertsFromPEM([]byte(testCA))
		Expect(ok).To(BeTrue())

		return serverConfig{
			addr: listenerWithAuth.Addr().String(),
			tlsConfig: &tls.Config{
				RootCAs: rootCAPool,
			},
		}
	})
})
