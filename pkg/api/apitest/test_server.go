package apitest

import (
	"crypto/tls"
	"net"

	"code.cloudfoundry.org/perm/pkg/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type TestServer struct {
	server *grpc.Server
}

func NewTestServer(opts ...TestServerOption) *TestServer {
	config := &options{}
	for _, o := range opts {
		o(config)
	}

	var serverOpts []grpc.ServerOption
	if config.credentials != nil {
		serverOpts = append(serverOpts, grpc.Creds(config.credentials))
	}

	server := grpc.NewServer(serverOpts...)

	return &TestServer{
		server: server,
	}
}

func (s *TestServer) Serve(listener net.Listener) error {
	err := s.server.Serve(listener)

	switch err {
	case nil:
		return nil
	case grpc.ErrServerStopped:
		return api.ErrServerStopped
	default:
		return api.ErrServerFailedToStart
	}
}

func (s *TestServer) GracefulStop() {
	s.server.GracefulStop()
}

func (s *TestServer) Stop() {
	s.server.Stop()
}

type TestServerOption func(*options)

func WithTLSConfig(config *tls.Config) TestServerOption {
	return func(o *options) {
		o.credentials = credentials.NewTLS(config)
	}
}

type options struct {
	credentials credentials.TransportCredentials
}
