package rpc

import (
	"context"
	"errors"

	"code.cloudfoundry.org/perm/protos"
	"github.com/satori/go.uuid"
)

type RoleServiceServer struct {
	roles        map[uuid.UUID]*protos.Role
	roleBindings map[protos.Actor][]uuid.UUID
}

func NewRoleServiceServer() *RoleServiceServer {
	return &RoleServiceServer{
		roles:        make(map[uuid.UUID]*protos.Role),
		roleBindings: make(map[protos.Actor][]uuid.UUID),
	}
}

func (s *RoleServiceServer) AssignRole(ctx context.Context, req *protos.AssignRoleRequest) (*protos.AssignRoleResponse, error) {
	roleID := req.GetRoleID()
	actor := req.GetActor()

	roleBindings, ok := s.roleBindings[*actor]
	if !ok {
		roleBindings = nil
	}

	u, err := uuid.FromString(roleID)
	if err != nil {
		return nil, togRPCError(err)
	}

	roleBindings = append(roleBindings, u)

	s.roleBindings[*actor] = roleBindings

	return &protos.AssignRoleResponse{}, nil
}

func (s *RoleServiceServer) HasRole(ctx context.Context, req *protos.HasRoleRequest) (*protos.HasRoleResponse, error) {
	actor := req.GetActor()
	roleID, err := uuid.FromString(req.GetRoleID())
	if err != nil {
		return nil, togRPCError(err)
	}

	roleBindings, ok := s.roleBindings[*actor]
	if !ok {
		return &protos.HasRoleResponse{HasRole: false}, nil
	}

	var found bool

	for _, id := range roleBindings {
		if uuid.Equal(id, roleID) {
			found = true
			break
		}
	}

	return &protos.HasRoleResponse{HasRole: found}, nil
}

func (s *RoleServiceServer) CreateRole(ctx context.Context, req *protos.CreateRoleRequest) (*protos.CreateRoleResponse, error) {
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

func (s *RoleServiceServer) ListActorRoles(ctx context.Context, req *protos.ListActorRolesRequest) (*protos.ListActorRolesResponse, error) {
	actor := req.GetActor()
	roleBindings, ok := s.roleBindings[*actor]
	if !ok {
		return &protos.ListActorRolesResponse{
			Roles: []*protos.Role{},
		}, nil
	}

	var roles []*protos.Role

	for _, id := range roleBindings {
		role, found := s.roles[id]
		if !found {
			return nil, togRPCError(errors.New("found a role-binding for non-existent role"))
		}

		roles = append(roles, role)
	}

	return &protos.ListActorRolesResponse{
		Roles: roles,
	}, nil
}
