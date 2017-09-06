package main

import (
	"log"
	"net"

	"context"

	"reflect"

	"code.cloudfoundry.org/perm/protos"
	"google.golang.org/grpc"
)

type roleServiceServer struct {
	roleBindings map[string]map[string][]map[string]string
}

func (s *roleServiceServer) AssignRole(ctx context.Context, req *protos.AssignRoleRequest) (*protos.AssignRoleResponse, error) {
	actor := req.GetActor()
	role := req.GetRole()

	roleBinding, ok := s.roleBindings[actor]
	if !ok {
		roleBinding = make(map[string][]map[string]string)
		s.roleBindings[actor] = roleBinding
	}

	roleBinding[role] = append(roleBinding[role], req.GetContext().GetContext())

	return &protos.AssignRoleResponse{}, nil
}

func (s *roleServiceServer) HasRole(ctx context.Context, req *protos.HasRoleRequest) (*protos.HasRoleResponse, error) {
	actor := req.GetActor()
	role := req.GetRole()
	roleContext := req.GetContext().GetContext()

	roleBinding, ok := s.roleBindings[actor]
	if !ok {
		return &protos.HasRoleResponse{HasRole: false}, nil
	}

	contexts, ok := roleBinding[role]
	if !ok {
		return &protos.HasRoleResponse{HasRole: false}, nil
	}

	var found bool
	for _, c := range contexts {
		if reflect.DeepEqual(c, roleContext) {
			found = true
			break
		}
	}

	return &protos.HasRoleResponse{HasRole: found}, nil
}

func newServer() *roleServiceServer {
	s := &roleServiceServer{
		roleBindings: make(map[string]map[string][]map[string]string),
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
