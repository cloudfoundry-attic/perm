package rpc

import (
	"context"
	"errors"

	"database/sql"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
	"code.cloudfoundry.org/perm/protos"
	"google.golang.org/grpc/codes"
)

type RoleServiceServer struct {
	dbConn *sql.DB

	logger      lager.Logger
	roles       map[string]*protos.Role
	assignments map[protos.Actor][]string
}

func NewRoleServiceServer(logger lager.Logger, dbConn *sql.DB) *RoleServiceServer {
	return &RoleServiceServer{
		logger:      logger,
		roles:       make(map[string]*protos.Role),
		assignments: make(map[protos.Actor][]string),
	}
}

func (s *RoleServiceServer) CreateRole(ctx context.Context, req *protos.CreateRoleRequest) (*protos.CreateRoleResponse, error) {
	logger := s.logger.Session("create-role")
	name := req.GetName()
	logData := lager.Data{"role.name": name}
	logger.Debug(messages.Starting, logData)

	if _, exists := s.roles[name]; exists {
		err := togRPCError(codes.AlreadyExists, errors.New(messages.ErrRoleAlreadyExists))
		logger.Error(messages.ErrRoleAlreadyExists, err, logData)
		return nil, err
	}

	role := &protos.Role{
		Name: name,
	}
	s.roles[name] = role

	logger.Debug(messages.Success, logData)
	return &protos.CreateRoleResponse{
		Role: role,
	}, nil
}

func (s *RoleServiceServer) GetRole(ctx context.Context, req *protos.GetRoleRequest) (*protos.GetRoleResponse, error) {
	logger := s.logger.Session("get-role")
	name := req.GetName()
	logData := lager.Data{"role.name": name}

	for _, role := range s.roles {
		if role.GetName() == name {
			logger.Debug(messages.Success, logData)
			return &protos.GetRoleResponse{
				Role: role,
			}, nil
		}
	}

	err := togRPCError(codes.NotFound, errors.New(messages.ErrRoleNotFound))
	logger.Error(messages.ErrRoleNotFound, err, logData)
	return nil, err
}

func (s *RoleServiceServer) DeleteRole(ctx context.Context, req *protos.DeleteRoleRequest) (*protos.DeleteRoleResponse, error) {
	logger := s.logger.Session("delete-role")
	name := req.GetName()
	logData := lager.Data{"role.name": name}

	_, ok := s.roles[name]

	if !ok {
		err := togRPCError(codes.NotFound, errors.New(messages.ErrRoleNotFound))
		s.logger.Error(messages.ErrRoleNotFound, err, logData)
		return nil, err
	}

	delete(s.roles, name)

	// "Cascade"
	// Remove role assignments for role
	for actor, assignments := range s.assignments {
		for i, roleName := range assignments {
			if roleName == name {
				s.assignments[actor] = append(assignments[:i], assignments[i+1:]...)
				assignmentData := lager.Data{
					"actor.id":     actor.GetID(),
					"actor.issuer": actor.GetIssuer(),
					"role.name":    name,
				}
				logger.Debug(messages.Success, assignmentData)
				break
			}
		}
	}

	logger.Debug(messages.Success, logData)

	return &protos.DeleteRoleResponse{}, nil
}

func (s *RoleServiceServer) AssignRole(ctx context.Context, req *protos.AssignRoleRequest) (*protos.AssignRoleResponse, error) {
	logger := s.logger.Session("assign-role")
	roleName := req.GetRoleName()
	actor := req.GetActor()
	logData := lager.Data{
		"actor.id":     actor.GetID(),
		"actor.issuer": actor.GetIssuer(),
		"role.name":    roleName,
	}
	_, exists := s.roles[roleName]

	if !exists {
		err := togRPCError(codes.NotFound, errors.New(messages.ErrRoleNotFound))
		logger.Error(messages.ErrRoleNotFound, err, logData)
		return nil, err
	}

	assignments, ok := s.assignments[*actor]
	if !ok {
		assignments = []string{}
	}

	for _, role := range assignments {
		if role == roleName {
			err := togRPCError(codes.AlreadyExists, errors.New(messages.ErrRoleAssignmentAlreadyExists))
			logger.Error(messages.ErrRoleAssignmentAlreadyExists, err, logData)
			return nil, err
		}
	}

	assignments = append(assignments, roleName)

	s.assignments[*actor] = assignments
	logger.Debug(messages.Success, logData)

	return &protos.AssignRoleResponse{}, nil
}

func (s *RoleServiceServer) UnassignRole(ctx context.Context, req *protos.UnassignRoleRequest) (*protos.UnassignRoleResponse, error) {
	logger := s.logger.Session("unassign-role")
	roleName := req.GetRoleName()
	actor := req.GetActor()
	logData := lager.Data{
		"actor.id":     actor.GetID(),
		"actor.issuer": actor.GetIssuer(),
		"role.name":    roleName,
	}
	_, exists := s.roles[roleName]

	if !exists {
		err := togRPCError(codes.NotFound, errors.New(messages.ErrRoleNotFound))
		logger.Error(messages.ErrRoleNotFound, err, logData)
		return nil, togRPCError(codes.NotFound, err)
	}

	assignments, ok := s.assignments[*actor]
	if !ok {
		assignments = []string{}
	}

	for i, assignment := range assignments {
		if assignment == roleName {
			s.assignments[*actor] = append(assignments[:i], assignments[i+1:]...)
			logger.Debug(messages.Success, logData)
			return &protos.UnassignRoleResponse{}, nil
		}
	}

	err := togRPCError(codes.NotFound, errors.New(messages.ErrRoleAssignmentNotFound))
	logger.Error(messages.ErrRoleAssignmentNotFound, err, logData)
	return nil, togRPCError(codes.NotFound, err)
}

func (s *RoleServiceServer) HasRole(ctx context.Context, req *protos.HasRoleRequest) (*protos.HasRoleResponse, error) {
	actor := req.GetActor()
	role := req.GetRoleName()
	assignments, ok := s.assignments[*actor]

	if !ok {
		return &protos.HasRoleResponse{HasRole: false}, nil
	}

	var found bool

	for _, name := range assignments {
		if name == role {
			found = true
			break
		}
	}

	return &protos.HasRoleResponse{HasRole: found}, nil
}

func (s *RoleServiceServer) ListActorRoles(ctx context.Context, req *protos.ListActorRolesRequest) (*protos.ListActorRolesResponse, error) {
	actor := req.GetActor()
	assignments, ok := s.assignments[*actor]
	if !ok {
		return &protos.ListActorRolesResponse{
			Roles: []*protos.Role{},
		}, nil
	}

	var roles []*protos.Role

	for _, id := range assignments {
		role, found := s.roles[id]
		if !found {
			return nil, togRPCError(codes.Unknown, errors.New("found an assignment for non-existent role"))
		}

		roles = append(roles, role)
	}

	return &protos.ListActorRolesResponse{
		Roles: roles,
	}, nil
}
