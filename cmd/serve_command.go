package cmd

import (
	"net"

	"strconv"

	"context"

	"time"

	"io/ioutil"
	"os"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm-go"
	"code.cloudfoundry.org/perm/cmd/contextx"
	"code.cloudfoundry.org/perm/ioutilx"
	"code.cloudfoundry.org/perm/pkg/api/db"
	"code.cloudfoundry.org/perm/pkg/api/logging"
	"code.cloudfoundry.org/perm/pkg/api/rpc"
	"code.cloudfoundry.org/perm/pkg/sqlx"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/stats"
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
	AuditFilePath     string        `long:"audit-file-path" default:""`
}

type StatsHandler struct{}

func (s *StatsHandler) TagRPC(ctx context.Context, st *stats.RPCTagInfo) context.Context {
	return contextx.WithReceiptTime(ctx, time.Now())
}

func (s *StatsHandler) HandleRPC(ctx context.Context, rpcStats stats.RPCStats) {

}

func (s *StatsHandler) TagConn(ctx context.Context, connTagInfo *stats.ConnTagInfo) context.Context {
	return ctx
}

func (s *StatsHandler) HandleConn(ctx context.Context, st stats.ConnStats) {
}

func (cmd ServeCommand) Execute([]string) error {
	//TODO Figure out version dynamically
	version := logging.Version("0.0.0")
	logger, _ := cmd.Logger.Logger("perm")
	logger = logger.Session("serve")

	var auditSink = ioutil.Discard
	if cmd.AuditFilePath != "" {
		securityLogFile, err := ioutilx.OpenLogFile(cmd.AuditFilePath)
		if err != nil {
			return err
		}

		defer securityLogFile.Close()
		auditSink = securityLogFile
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	securityLogger := logging.NewCEFLogger(auditSink, "cloud_foundry", "perm", version, logging.Hostname(hostname), cmd.Port)

	ctx := context.Background()

	listenInterface := cmd.Hostname
	port := cmd.Port
	lis, err := net.Listen("tcp", net.JoinHostPort(listenInterface, strconv.Itoa(port)))

	maxConnectionIdle := cmd.MaxConnectionIdle

	listeningLogData := lager.Data{
		"protocol":          "tcp",
		"hostname":          listenInterface,
		"port":              port,
		"maxConnectionIdle": maxConnectionIdle.String(),
	}
	if err != nil {
		logger.Error(failedToListen, err, listeningLogData)
		return err
	}

	keepaliveParams := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle: maxConnectionIdle,
	})

	tlsCreds, err := credentials.NewServerTLSFromFile(cmd.TLSCertificate, cmd.TLSKey)
	if err != nil {
		logger.Error(failedToParseTLSCredentials, err)
		return err
	}

	recoveryOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(func(p interface{}) error {
			grpcErr := status.Errorf(codes.Internal, "%s", p)
			logger.Error(errInternal, grpcErr)
			return grpcErr
		}),
	}

	streamMiddleware := grpc_middleware.ChainStreamServer(grpc_recovery.StreamServerInterceptor(recoveryOpts...))
	unaryMiddleware := grpc_middleware.ChainUnaryServer(grpc_recovery.UnaryServerInterceptor(recoveryOpts...))

	streamInterceptor := grpc.StreamInterceptor(streamMiddleware)
	unaryInterceptor := grpc.UnaryInterceptor(unaryMiddleware)

	serverOpts := []grpc.ServerOption{
		keepaliveParams,
		grpc.StatsHandler(&StatsHandler{}),
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
	permissionServiceServer := rpc.NewPermissionServiceServer(logger, securityLogger, store)
	protos.RegisterPermissionServiceServer(grpcServer, permissionServiceServer)

	logger.Debug(starting, listeningLogData)

	return grpcServer.Serve(lis)
}
