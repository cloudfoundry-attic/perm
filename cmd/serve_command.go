package cmd

import (
	"net"

	"strconv"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/db"
	"code.cloudfoundry.org/perm/messages"
	"code.cloudfoundry.org/perm/protos"
	"code.cloudfoundry.org/perm/rpc"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

type ServeCommand struct {
	Logger LagerFlag

	Hostname       string  `long:"listen-hostname" description:"Hostname on which to listen for gRPC traffic" default:"0.0.0.0"`
	Port           int     `long:"listen-port" description:"Port on which to listen for gRPC traffic" default:"6283"`
	TLSCertificate string  `long:"tls-certificate" description:"File path of TLS certificate" required:"true"`
	TLSKey         string  `long:"tls-key" description:"File path of TLS private key" required:"true"`
	SQL            SQLFlag `group:"SQL" namespace:"sql"`
}

func (cmd ServeCommand) Execute([]string) error {
	logger, _ := cmd.Logger.Logger("perm")
	logger = logger.Session("serve")

	hostname := cmd.Hostname
	port := cmd.Port
	lis, err := net.Listen("tcp", net.JoinHostPort(hostname, strconv.Itoa(port)))

	listeningLogData := lager.Data{
		"protocol": "tcp",
		"hostname": hostname,
		"port":     port,
	}
	if err != nil {
		logger.Error(messages.FailedToListen, err, listeningLogData)
		return err
	}

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
	serverOpts := []grpc.ServerOption{
		grpc.Creds(tlsCreds),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(grpc_recovery.StreamServerInterceptor(recoveryOpts...))),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(grpc_recovery.UnaryServerInterceptor(recoveryOpts...))),
	}

	grpcServer := grpc.NewServer(serverOpts...)

	conn, err := cmd.SQL.Open()
	if err != nil {
		logger.Error(messages.FailedToOpenSQLConnection, err)
		return err
	}

	pingLogger := logger.Session(messages.PingSQLConnection, cmd.SQL.LagerData())
	pingLogger.Debug(messages.Starting)
	err = conn.Ping()
	if err != nil {
		logger.Error(messages.FailedToPingSQLConnection, err, cmd.SQL.LagerData())
		return err
	}
	pingLogger.Debug(messages.Finished)

	defer conn.Close()

	logger = logger.Session("grpc-server")

	store := db.NewDataService(conn)
	roleServiceServer := rpc.NewRoleServiceServer(logger, store, store)
	protos.RegisterRoleServiceServer(grpcServer, roleServiceServer)
	logger.Info(messages.Starting, listeningLogData)

	return grpcServer.Serve(lis)
}
