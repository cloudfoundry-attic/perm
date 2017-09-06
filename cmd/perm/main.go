package main

import (
	"log"
	"net"

	"context"

	"code.cloudfoundry.org/perm/protos"
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

func newServer() *roleServiceServer {
	s := &roleServiceServer{
		roles:        make(map[uuid.UUID]*protos.Role),
		roleBindings: make(map[string][]uuid.UUID),
	}
	return s
}

func main() {
	lis, err := net.Listen("tcp", "localhost:8888")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	protos.RegisterRoleServiceServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
