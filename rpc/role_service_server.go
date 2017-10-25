package rpc

import (
	"context"
	"errors"

	"database/sql"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
	"code.cloudfoundry.org/perm/models"
	"code.cloudfoundry.org/perm/protos"
	"google.golang.org/grpc/codes"
)

type RoleServiceServer struct {
	dbConn *sql.DB

	logger      lager.Logger
	assignments map[models.Actor][]string

	deps Deps
}

type Deps interface {
	models.RoleService
}

func NewRoleServiceServer(logger lager.Logger, dbConn *sql.DB, deps Deps) *RoleServiceServer {
	return &RoleServiceServer{
		logger:      logger,
		dbConn:      dbConn,
		assignments: make(map[models.Actor][]string),
		deps:        deps,
	}
}

func (s *RoleServiceServer) CreateRole(ctx context.Context, req *protos.CreateRoleRequest) (*protos.CreateRoleResponse, error) {
	logger := s.logger.Session("create-role")
	name := req.GetName()
	logData := lager.Data{"role.name": name}
	logger.Debug(messages.Starting, logData)

	role, err := s.deps.CreateRole(ctx, logger, name)

	if err != nil {
		return nil, togRPCErrorNew(err)
		//err := togRPCError(codes.AlreadyExists, errors.New(messages.ErrRoleAlreadyExists))
		//logger.Error(messages.ErrRoleAlreadyExists, err, logData)
		//return nil, err
	}

	logger.Debug(messages.Success, logData)
	return &protos.CreateRoleResponse{
		Role: role.ToProto(),
	}, nil
}

func (s *RoleServiceServer) GetRole(ctx context.Context, req *protos.GetRoleRequest) (*protos.GetRoleResponse, error) {
	logger := s.logger.Session("get-role")
	name := req.GetName()
	logData := lager.Data{"role.name": name}

	role, err := s.deps.FindRole(ctx, logger, models.RoleQuery{Name: name})
	if err != nil {
		return nil, togRPCErrorNew(err)
	}

	logger.Debug(messages.Success, logData)
	return &protos.GetRoleResponse{
		Role: role.ToProto(),
	}, nil
}

func (s *RoleServiceServer) DeleteRole(ctx context.Context, req *protos.DeleteRoleRequest) (*protos.DeleteRoleResponse, error) {
	logger := s.logger.Session("delete-role")
	name := req.GetName()
	logData := lager.Data{"role.name": name}

	err := s.deps.DeleteRole(ctx, logger, models.RoleQuery{Name: name})
	if err != nil {
		return nil, togRPCErrorNew(err)
	}

	// "Cascade"
	// Remove role assignments for role
	for actor, assignments := range s.assignments {
		for i, roleName := range assignments {
			if roleName == name {
				s.assignments[actor] = append(assignments[:i], assignments[i+1:]...)
				assignmentData := lager.Data{
					"actor.id":     actor.DomainID,
					"actor.issuer": actor.Issuer,
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
	pActor := req.GetActor()

	actor := models.Actor{
		DomainID: pActor.GetID(),
		Issuer:   pActor.GetIssuer(),
	}

	logData := lager.Data{
		"actor.id":     actor.DomainID,
		"actor.issuer": actor.Issuer,
		"role.name":    roleName,
	}

	_, err := s.deps.FindRole(ctx, logger, models.RoleQuery{Name: roleName})
	if err != nil {
		return nil, togRPCErrorNew(err)
	}

	assignments, ok := s.assignments[actor]
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

	s.assignments[actor] = assignments
	logger.Debug(messages.Success, logData)

	return &protos.AssignRoleResponse{}, nil
}

func (s *RoleServiceServer) UnassignRole(ctx context.Context, req *protos.UnassignRoleRequest) (*protos.UnassignRoleResponse, error) {
	logger := s.logger.Session("unassign-role")
	roleName := req.GetRoleName()
	pActor := req.GetActor()
	actor := models.Actor{
		DomainID: pActor.GetID(),
		Issuer:   pActor.GetIssuer(),
	}
	logData := lager.Data{
		"actor.id":     actor.DomainID,
		"actor.issuer": actor.Issuer,
		"role.name":    roleName,
	}

	_, err := s.deps.FindRole(ctx, logger, models.RoleQuery{Name: roleName})
	if err != nil {
		return nil, togRPCErrorNew(err)
	}

	assignments, ok := s.assignments[actor]
	if !ok {
		assignments = []string{}
	}

	for i, assignment := range assignments {
		if assignment == roleName {
			s.assignments[actor] = append(assignments[:i], assignments[i+1:]...)
			logger.Debug(messages.Success, logData)
			return &protos.UnassignRoleResponse{}, nil
		}
	}

	err = togRPCError(codes.NotFound, errors.New(messages.ErrRoleAssignmentNotFound))
	logger.Error(messages.ErrRoleAssignmentNotFound, err, logData)

	return nil, togRPCError(codes.NotFound, err)
}

func (s *RoleServiceServer) HasRole(ctx context.Context, req *protos.HasRoleRequest) (*protos.HasRoleResponse, error) {
	role := req.GetRoleName()
	pActor := req.GetActor()
	actor := models.Actor{
		DomainID: pActor.GetID(),
		Issuer:   pActor.GetIssuer(),
	}
	assignments, ok := s.assignments[actor]

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
	pActor := req.GetActor()
	actor := models.Actor{
		DomainID: pActor.GetID(),
		Issuer:   pActor.GetIssuer(),
	}
	assignments, ok := s.assignments[actor]
	if !ok {
		return &protos.ListActorRolesResponse{
			Roles: []*protos.Role{},
		}, nil
	}

	var roles []*models.Role

	for _, id := range assignments {
		role, err := s.deps.FindRole(ctx, s.logger, models.RoleQuery{Name: id})
		if err != nil {
			return nil, togRPCErrorNew(err)
		}

		roles = append(roles, role)
	}

	var pRoles []*protos.Role

	for _, r := range roles {
		pRoles = append(pRoles, r.ToProto())
	}

	return &protos.ListActorRolesResponse{
		Roles: pRoles,
	}, nil
}
