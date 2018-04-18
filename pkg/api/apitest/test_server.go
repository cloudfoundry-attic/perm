package apitest

import (
	"context"
	"crypto/tls"
	"net"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/api"
	"code.cloudfoundry.org/perm/pkg/api/logging"
	"code.cloudfoundry.org/perm/pkg/api/rpc"
	"code.cloudfoundry.org/perm/protos/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type TestServer struct {
	logger lager.Logger
	server *grpc.Server
}

func NewTestServer(opts ...TestServerOption) *TestServer {
	config := &options{
		logger:         &emptyLogger{},
		securityLogger: &emptySecurityLogger{},
	}

	for _, o := range opts {
		o(config)
	}

	var serverOpts []grpc.ServerOption
	if config.credentials != nil {
		serverOpts = append(serverOpts, grpc.Creds(config.credentials))
	}

	server := grpc.NewServer(serverOpts...)
	store := rpc.NewInMemoryStore()
	logger := config.logger
	securityLogger := config.securityLogger

	roleServiceServer := rpc.NewRoleServiceServer(logger, securityLogger, store, store)
	protos.RegisterRoleServiceServer(server, roleServiceServer)

	permissionServiceServer := rpc.NewPermissionServiceServer(logger, securityLogger, store)
	protos.RegisterPermissionServiceServer(server, permissionServiceServer)

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

func WithLogger(logger lager.Logger) TestServerOption {
	return func(o *options) {
		o.logger = logger
	}
}

func WithSecurityLogger(logger rpc.SecurityLogger) TestServerOption {
	return func(o *options) {
		o.securityLogger = logger
	}
}
func WithTLSConfig(config *tls.Config) TestServerOption {
	return func(o *options) {
		o.credentials = credentials.NewTLS(config)
	}
}

type options struct {
	logger         lager.Logger
	securityLogger rpc.SecurityLogger
	credentials    credentials.TransportCredentials
}

type emptyLogger struct{}

func (l *emptyLogger) RegisterSink(lager.Sink) {}

func (l *emptyLogger) SessionName() string {
	return ""
}

func (l *emptyLogger) Session(string, ...lager.Data) lager.Logger {
	return l
}

func (l *emptyLogger) WithData(lager.Data) lager.Logger {
	return l
}

func (l *emptyLogger) Debug(string, ...lager.Data) {}

func (l *emptyLogger) Info(string, ...lager.Data) {}

func (l *emptyLogger) Error(string, error, ...lager.Data) {}

func (l *emptyLogger) Fatal(string, error, ...lager.Data) {}

type emptySecurityLogger struct{}

func (l *emptySecurityLogger) Log(context.Context, string, string, ...logging.CustomExtension) {}
