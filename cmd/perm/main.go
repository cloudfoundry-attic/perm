package main

import (
	"net"

	"fmt"
	"os"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
	"code.cloudfoundry.org/perm/protos"
	"code.cloudfoundry.org/perm/rpc"
	"github.com/jessevdk/go-flags"
	"google.golang.org/grpc"
)

type options struct {
	Hostname string `long:"listen-hostname" description:"Hostname on which to listen for gRPC traffic" default:"0.0.0.0"`
	Port     int    `long:"listen-port" description:"Port on which to listen for gRPC traffic" default:"6283"`
	Logger   LagerFlag
}

func main() {
	parserOpts := &options{}
	parser := flags.NewParser(parserOpts, flags.Default)

	_, err := parser.Parse()
	if err != nil {
		lager.NewLogger("perm").Error(messages.ErrFailedToParseOptions, err)
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
		logger.Error(messages.ErrFailedToListen, err, listeningLogData)
		os.Exit(1)
	}

	var serverOpts []grpc.ServerOption

	grpcServer := grpc.NewServer(serverOpts...)

	logger = logger.Session("grpc-server")

	protos.RegisterRoleServiceServer(grpcServer, rpc.NewRoleServiceServer(logger))
	logger.Debug(messages.StartingServer, listeningLogData)
	grpcServer.Serve(lis)
}
