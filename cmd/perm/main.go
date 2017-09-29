package main

import (
	"net"

	"fmt"
	"os"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
	"code.cloudfoundry.org/perm/protos"
	"code.cloudfoundry.org/perm/rpc"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/jessevdk/go-flags"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

type options struct {
	Hostname       string `long:"listen-hostname" description:"Hostname on which to listen for gRPC traffic" default:"0.0.0.0"`
	Port           int    `long:"listen-port" description:"Port on which to listen for gRPC traffic" default:"6283"`
	TLSCertificate string `long:"tls-certificate" description:"File path of TLS certificate" required:"true"`
	TLSKey         string `long:"tls-key" description:"File path of TLS private key" required:"true"`
	Logger         LagerFlag
}

func main() {
	parserOpts := &options{}
	parser := flags.NewParser(parserOpts, flags.Default)

	_, err := parser.Parse()
	if err != nil {
		lager.NewLogger("perm").Fatal(messages.ErrFailedToParseOptions, err)
		os.Exit(1)
	}

	logger, _ := parserOpts.Logger.Logger("perm")

	hostname := parserOpts.Hostname
	port := parserOpts.Port
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", hostname, port))

	listeningLogData := lager.Data{
		"protocol": "tcp",
		"hostname": hostname,
		"port":     port,
	}
	if err != nil {
		logger.Fatal(messages.ErrFailedToListen, err, listeningLogData)
		os.Exit(1)
	}

	tlsCreds, err := credentials.NewServerTLSFromFile(parserOpts.TLSCertificate, parserOpts.TLSKey)

	if err != nil {
		logger.Fatal(messages.ErrInvalidTLSCredentials, err)
		os.Exit(1)
	}

	recoveryOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(func(p interface{}) error {
			err := status.Errorf(codes.Internal, "%s", p)
			logger.Error(messages.ErrInternal, err)
			return err
		}),
	}
	serverOpts := []grpc.ServerOption{
		grpc.Creds(tlsCreds),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(grpc_recovery.StreamServerInterceptor(recoveryOpts...))),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(grpc_recovery.UnaryServerInterceptor(recoveryOpts...))),
	}

	grpcServer := grpc.NewServer(serverOpts...)

	logger = logger.Session("grpc-server")

	protos.RegisterRoleServiceServer(grpcServer, rpc.NewRoleServiceServer(logger))
	logger.Info(messages.StartingServer, listeningLogData)

	grpcServer.Serve(lis)
}
