package cmd

import (
	"net"

	"strconv"

	"context"

	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm-go"
	"code.cloudfoundry.org/perm/db"
	"code.cloudfoundry.org/perm/logging"
	"code.cloudfoundry.org/perm/messages"
	"code.cloudfoundry.org/perm/rpc"
	"code.cloudfoundry.org/perm/sqlx"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

type ServeCommand struct {
	Logger            LagerFlag
	Hostname          string        `long:"listen-hostname" description:"Hostname on which to listen for gRPC traffic" default:"0.0.0.0"`
	Port              int           `long:"listen-port" description:"Port on which to listen for gRPC traffic" default:"6283"`
	MaxConnectionIdle time.Duration `long:"max-connection-idle" description:"The amount of time before an idle connection will be closed with a GoAway." default:"10s"`
	TLSCertificate    string        `long:"tls-certificate" description:"File path of TLS certificate" required:"true"`
	TLSKey            string        `long:"tls-key" description:"File path of TLS private key" required:"true"`
	SQL               SQLFlag       `group:"SQL" namespace:"sql"`
}

func (cmd ServeCommand) Execute([]string) error {
	//TODO Figure out version dynamically
	version := logging.Version("0.0.0")
	logger, _ := cmd.Logger.Logger("perm")
	logger = logger.Session("serve")
	securityLogger := logging.NewCEFLogger("cloud_foundry", "perm", version)

	ctx := context.Background()

	hostname := cmd.Hostname
	port := cmd.Port
	lis, err := net.Listen("tcp", net.JoinHostPort(hostname, strconv.Itoa(port)))

	maxConnectionIdle := cmd.MaxConnectionIdle

	listeningLogData := lager.Data{
		"protocol":          "tcp",
		"hostname":          hostname,
		"port":              port,
		"maxConnectionIdle": maxConnectionIdle.String(),
	}
	if err != nil {
		logger.Error(messages.FailedToListen, err, listeningLogData)
		return err
	}

	keepaliveParams := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle: maxConnectionIdle,
	})

	tlsCreds, err := credentials.NewServerTLSFromFile(cmd.TLSCertificate, cmd.TLSKey)
	if err != nil {
		logger.Error(messages.FailedToParseTLSCredentials, err)
		return err
	}

	recoveryOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(func(p interface{}) error {
			grpcErr := status.Errorf(codes.Internal, "%s", p)
			logger.Error(messages.ErrInternal, grpcErr)
			return grpcErr
		}),
	}
	streamMiddleware := grpc_middleware.ChainStreamServer(grpc_recovery.StreamServerInterceptor(recoveryOpts...))
	unaryMiddleware := grpc_middleware.ChainUnaryServer(grpc_recovery.UnaryServerInterceptor(recoveryOpts...))

	streamInterceptor := grpc.StreamInterceptor(streamMiddleware)
	unaryInterceptor := grpc.UnaryInterceptor(unaryMiddleware)

	serverOpts := []grpc.ServerOption{
		keepaliveParams,
		grpc.Creds(tlsCreds),
		streamInterceptor,
		unaryInterceptor,
	}

	grpcServer := grpc.NewServer(serverOpts...)

	conn, err := cmd.SQL.Connect(ctx, logger, OS, IOReader)
	if err != nil {
		return err
	}
	defer conn.Close()

	migrationLogger := logger.Session("verify-migrations")
	appliedCorrectly, err := sqlx.VerifyAppliedMigrations(
		context.Background(),
		migrationLogger,
		conn,
		db.MigrationsTableName,
		db.Migrations,
	)
	if err != nil {
		return err
	}
	if !appliedCorrectly {
		return ErrMigrationsOutOfSync
	}

	logger = logger.Session("grpc-server")
	store := db.NewDataService(conn)

	roleServiceServer := rpc.NewRoleServiceServer(logger, securityLogger, store, store)
	protos.RegisterRoleServiceServer(grpcServer, roleServiceServer)

	// TODO
	permissionServiceServer := rpc.NewPermissionServiceServer(logger, nil, store)
	protos.RegisterPermissionServiceServer(grpcServer, permissionServiceServer)

	logger.Debug(messages.Starting, listeningLogData)

	return grpcServer.Serve(lis)
}
