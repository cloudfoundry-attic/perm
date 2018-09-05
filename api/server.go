package api

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"

	"code.cloudfoundry.org/perm/api/internal/repos"
	"code.cloudfoundry.org/perm/api/internal/repos/db"
	"code.cloudfoundry.org/perm/api/internal/repos/inmemory"
	"code.cloudfoundry.org/perm/api/internal/rpc"
	"code.cloudfoundry.org/perm/api/internal/rpc/interceptors"
	"code.cloudfoundry.org/perm/internal/protos"
	"code.cloudfoundry.org/perm/internal/sqlx"
	"code.cloudfoundry.org/perm/logx"
	"code.cloudfoundry.org/perm/metrics"
	"code.cloudfoundry.org/perm/oidcx"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
)

type Server struct {
	logger         logx.Logger
	securityLogger logx.SecurityLogger
	server         *grpc.Server
}

type store interface {
	repos.PermissionRepo
	repos.RoleRepo
}

func NewServer(opts ...ServerOption) *Server {
	config := &serverConfig{
		logger:         &emptyLogger{},
		securityLogger: &emptySecurityLogger{},
	}

	for _, opt := range opts {
		opt(config)
	}

	logger := config.logger

	recoveryOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(func(p interface{}) error {
			grpcErr := status.Errorf(codes.Internal, "%s", p)
			logger.Error(internal, grpcErr)
			return grpcErr
		}),
	}
	unaryServerInterceptors := []grpc.UnaryServerInterceptor{
		grpc_recovery.UnaryServerInterceptor(recoveryOpts...),
	}

	if config.oidcProvider != nil {
		unaryServerInterceptors = append(unaryServerInterceptors, interceptors.OIDCInterceptor(config.oidcProvider, config.securityLogger))
	}

	if config.statter != nil {
		unaryServerInterceptors = append(unaryServerInterceptors, interceptors.MetricsInterceptor(config.statter))
	}

	unaryMiddleware := grpc_middleware.ChainUnaryServer(unaryServerInterceptors...)

	unaryInterceptor := grpc.UnaryInterceptor(unaryMiddleware)

	serverOpts := []grpc.ServerOption{
		grpc.KeepaliveParams(config.keepalive),
		unaryInterceptor,
	}

	if config.credentials != nil {
		serverOpts = append(serverOpts, grpc.Creds(config.credentials))
	}

	server := grpc.NewServer(serverOpts...)

	var s store
	if config.conn == nil {
		s = inmemory.NewStore()
	} else {
		s = db.NewStore(config.conn)
	}

	roleServiceServer := rpc.NewRoleServiceServer(logger, config.securityLogger, s)
	protos.RegisterRoleServiceServer(server, roleServiceServer)

	permissionServiceServer := rpc.NewPermissionServiceServer(logger, config.securityLogger, s)
	protos.RegisterPermissionServiceServer(server, permissionServiceServer)

	return &Server{
		logger:         logger,
		securityLogger: config.securityLogger,
		server:         server,
	}
}

func (s *Server) Serve(listener net.Listener) error {
	err := s.server.Serve(listener)

	switch err {
	case nil:
		return nil
	case grpc.ErrServerStopped:
		return ErrServerStopped
	default:
		return ErrServerFailedToStart
	}
}

func (s *Server) GracefulStop() {
	s.server.GracefulStop()
}

func (s *Server) Stop() {
	s.server.Stop()
}

type ServerOption func(*serverConfig)

func WithLogger(logger logx.Logger) ServerOption {
	return func(o *serverConfig) {
		o.logger = logger
	}
}

func WithSecurityLogger(logger logx.SecurityLogger) ServerOption {
	return func(o *serverConfig) {
		o.securityLogger = logger
	}
}

func WithTLSConfig(config *tls.Config) ServerOption {
	return func(o *serverConfig) {
		o.credentials = credentials.NewTLS(config)
	}
}

func WithMaxConnectionIdle(duration time.Duration) ServerOption {
	return func(o *serverConfig) {
		o.keepalive.MaxConnectionIdle = duration
	}
}

func WithOIDCProvider(provider oidcx.Provider) ServerOption {
	return func(o *serverConfig) {
		o.oidcProvider = provider
	}
}

func WithDBConn(conn *sqlx.DB) ServerOption {
	return func(o *serverConfig) {
		o.conn = conn
	}
}

func WithStatter(statter metrics.Statter) ServerOption {
	return func(o *serverConfig) {
		o.statter = statter
	}
}

type serverConfig struct {
	logger         logx.Logger
	securityLogger logx.SecurityLogger

	credentials credentials.TransportCredentials
	keepalive   keepalive.ServerParameters
	statter     metrics.Statter

	oidcProvider oidcx.Provider

	conn *sqlx.DB
}

type emptyLogger struct{}

func (l *emptyLogger) WithName(string) logx.Logger {
	return l
}

func (l *emptyLogger) WithData(...logx.Data) logx.Logger {
	return l
}

func (l *emptyLogger) Debug(string, ...logx.Data) {}

func (l *emptyLogger) Info(string, ...logx.Data) {}

func (l *emptyLogger) Error(string, error, ...logx.Data) {}

type emptySecurityLogger struct{}

func (l *emptySecurityLogger) Log(context.Context, string, string, ...logx.SecurityData) {}
