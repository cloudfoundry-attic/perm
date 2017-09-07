package main

import (
	"net"

	"fmt"
	"os"

	"code.cloudfoundry.org/perm"
	"code.cloudfoundry.org/perm/protos"
	"code.cloudfoundry.org/perm/rpc"
	"github.com/jessevdk/go-flags"
	"google.golang.org/grpc"
)

type options struct {
	Hostname string `long:"listen-hostname" description:"Hostname on which to listen for gRPC traffic" default:"0.0.0.0"`
	Port     int    `long:"listen-port" description:"Port on which to listen for gRPC traffic" default:"6283"`
	Version  bool   `short:"V" long:"version" description:"Prints the current Perm version"`
}

func main() {
	parserOpts := &options{}
	parser := flags.NewParser(parserOpts, flags.Default)

	_, err := parser.Parse()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if parserOpts.Version {
		fmt.Println(perm.Version)

		os.Exit(0)
		return
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", parserOpts.Hostname, parserOpts.Port))
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to listen: %v", err)
		os.Exit(1)
	}

	var serverOpts []grpc.ServerOption

	grpcServer := grpc.NewServer(serverOpts...)
	protos.RegisterRoleServiceServer(grpcServer, rpc.NewRoleServiceServer())
	grpcServer.Serve(lis)
}
