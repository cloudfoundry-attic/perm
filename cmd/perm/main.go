package main

import (
	"net"

	"context"

	"errors"

	"fmt"
	"os"

	"code.cloudfoundry.org/perm/protos"
	"github.com/jessevdk/go-flags"
	"github.com/satori/go.uuid"
	"google.golang.org/grpc"
)

type roleServiceServer struct {
	roles        map[uuid.UUID]*protos.Role
	roleBindings map[string][]uuid.UUID
}

func (s *roleServiceServer) AssignRole(ctx context.Context, req *protos.AssignRoleRequest) (*protos.AssignRoleResponse, error) {
	actor := req.GetActor()
	roleID := req.GetRoleID()

	roleBindings, ok := s.roleBindings[actor]
	if !ok {
		roleBindings = nil
	}

	u, err := uuid.FromString(roleID)
	if err != nil {
		return &protos.AssignRoleResponse{}, err
	}

	roleBindings = append(roleBindings, u)

	s.roleBindings[actor] = roleBindings

	return &protos.AssignRoleResponse{}, nil
}

func (s *roleServiceServer) HasRole(ctx context.Context, req *protos.HasRoleRequest) (*protos.HasRoleResponse, error) {
	actor := req.GetActor()
	roleID, err := uuid.FromString(req.GetRoleID())
	if err != nil {
		return &protos.HasRoleResponse{}, err
	}

	roleBindings, ok := s.roleBindings[actor]
	if !ok {
		return &protos.HasRoleResponse{HasRole: false}, nil
	}

	var found bool

	for _, id := range roleBindings {
		if uuid.Equal(id, roleID) {
			found = true
		}
	}

	return &protos.HasRoleResponse{HasRole: found}, nil
}

func (s *roleServiceServer) CreateRole(ctx context.Context, req *protos.CreateRoleRequest) (*protos.CreateRoleResponse, error) {
	id := uuid.NewV4()
	role := &protos.Role{
		Name: req.GetName(),
		ID:   id.String(),
	}
	s.roles[id] = role

	return &protos.CreateRoleResponse{
		Role: role,
	}, nil
}

func (s *roleServiceServer) ListActorRoles(ctx context.Context, req *protos.ListActorRolesRequest) (*protos.ListActorRolesResponse, error) {
	roleBindings, ok := s.roleBindings[req.GetActor()]
	if !ok {
		return &protos.ListActorRolesResponse{
			Roles: []*protos.Role{},
		}, nil
	}

	var roles []*protos.Role

	for _, id := range roleBindings {
		role, found := s.roles[id]
		if !found {
			return &protos.ListActorRolesResponse{
				Roles: []*protos.Role{},
			}, errors.New("could not find assigned role")
		}

		roles = append(roles, role)
	}
	return &protos.ListActorRolesResponse{
		Roles: roles,
	}, nil
}

func newServer() *roleServiceServer {
	s := &roleServiceServer{
		roles:        make(map[uuid.UUID]*protos.Role),
		roleBindings: make(map[string][]uuid.UUID),
	}
	return s
}

type options struct {
	Hostname string `long:"listen-hostname" description:"Hostname on which to listen for gRPC traffic" default:"0.0.0.0"`
	Port     int    `long:"listen-port" description:"Port on which to listen for gRPC traffic" default:"6283"`
}

func main() {
	parserOpts := &options{}
	parser := flags.NewParser(parserOpts, flags.Default)

	_, err := parser.Parse()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", parserOpts.Hostname, parserOpts.Port))
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to listen: %v", err)
		os.Exit(1)
	}

	var serverOpts []grpc.ServerOption

	grpcServer := grpc.NewServer(serverOpts...)
	protos.RegisterRoleServiceServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
