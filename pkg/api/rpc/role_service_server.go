package rpc

import (
	"context"
	"errors"
	"strings"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm-go"
	"code.cloudfoundry.org/perm/pkg/api/logging"
	"code.cloudfoundry.org/perm/pkg/api/repos"
	"code.cloudfoundry.org/perm/pkg/perm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RoleServiceServer struct {
	logger         lager.Logger
	securityLogger SecurityLogger

	roleRepo           repos.RoleRepo
	roleAssignmentRepo repos.RoleAssignmentRepo
}

func NewRoleServiceServer(
	logger lager.Logger,
	securityLogger SecurityLogger,
	roleRepo repos.RoleRepo,
	roleAssignmentRepo repos.RoleAssignmentRepo,
) *RoleServiceServer {
	return &RoleServiceServer{
		logger:             logger,
		securityLogger:     securityLogger,
		roleRepo:           roleRepo,
		roleAssignmentRepo: roleAssignmentRepo,
	}
}

func (s *RoleServiceServer) CreateRole(
	ctx context.Context,
	req *protos.CreateRoleRequest,
) (*protos.CreateRoleResponse, error) {
	name := req.GetName()
	var permissions []*perm.Permission
	for _, p := range req.GetPermissions() {
		permissions = append(permissions, &perm.Permission{
			Action:          p.GetName(),
			ResourcePattern: p.GetResourcePattern(),
		})
	}
	logExtensions := logging.CustomExtension{Key: "roleName", Value: name}
	s.securityLogger.Log(ctx, "CreateRole", "Role creation", logExtensions)
	logger := s.logger.Session("create-role").WithData(lager.Data{"role.name": name, "permissions": permissions})
	logger.Debug(starting)

	role, err := s.roleRepo.CreateRole(ctx, logger, name, permissions...)

	if err != nil {
		return nil, togRPCError(err)
	}

	logger.Debug(success)
	return &protos.CreateRoleResponse{
		Role: &protos.Role{
			Name: role.Name,
		},
	}, nil
}

func (s *RoleServiceServer) GetRole(
	ctx context.Context,
	req *protos.GetRoleRequest,
) (*protos.GetRoleResponse, error) {
	name := req.GetName()
	logger := s.logger.Session("get-role").WithData(lager.Data{"role.name": name})
	logger.Debug(starting)

	query := repos.FindRoleQuery{
		RoleName: name,
	}
	role, err := s.roleRepo.FindRole(ctx, logger, query)
	if err != nil {
		return nil, togRPCError(err)
	}

	logger.Debug(success)
	return &protos.GetRoleResponse{
		Role: &protos.Role{
			Name: role.Name,
		},
	}, nil
}

func (s *RoleServiceServer) DeleteRole(
	ctx context.Context,
	req *protos.DeleteRoleRequest,
) (*protos.DeleteRoleResponse, error) {
	name := req.GetName()
	logExtensions := logging.CustomExtension{Key: "roleName", Value: name}
	s.securityLogger.Log(ctx, "DeleteRole", "Role deletion", logExtensions)
	logger := s.logger.Session("delete-role").WithData(lager.Data{
		"role.name": name,
	})
	logger.Debug(starting)

	err := s.roleRepo.DeleteRole(ctx, logger, name)
	if err != nil {
		return nil, togRPCError(err)
	}

	logger.Debug(success)
	return &protos.DeleteRoleResponse{}, nil
}

func validateAssignRoleRequest(req *protos.AssignRoleRequest) error {
	pActor := req.GetActor()
	namespace := pActor.GetNamespace()
	if strings.Trim(namespace, "\t \n") == "" {
		return errors.New("actor namespace cannot be empty")
	}

	return nil
}

func validateHasRoleRequest(req *protos.HasRoleRequest) error {
	pActor := req.GetActor()
	namespace := pActor.GetNamespace()
	if strings.Trim(namespace, "\t \n") == "" {
		return errors.New("actor namespace cannot be empty")
	}

	return nil
}

func (s *RoleServiceServer) AssignRole(
	ctx context.Context,
	req *protos.AssignRoleRequest,
) (*protos.AssignRoleResponse, error) {

	err := validateAssignRoleRequest(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	roleName := req.GetRoleName()
	pActor := req.GetActor()

	domainID := pActor.GetID()
	namespace := pActor.GetNamespace()
	logExtensions := []logging.CustomExtension{
		{Key: "roleName", Value: roleName},
		{Key: "userID", Value: pActor.ID},
	}

	s.securityLogger.Log(ctx, "AssignRole", "Role assignment", logExtensions...)
	logger := s.logger.Session("assign-role").WithData(lager.Data{
		"actor.id":        domainID,
		"actor.namespace": namespace,
		"role.name":       roleName,
	})
	logger.Debug(starting)

	err = s.roleAssignmentRepo.AssignRole(ctx, logger, roleName, domainID, namespace)
	if err != nil {
		return nil, togRPCError(err)
	}

	logger.Debug(success)
	return &protos.AssignRoleResponse{}, nil
}

func (s *RoleServiceServer) UnassignRole(
	ctx context.Context,
	req *protos.UnassignRoleRequest,
) (*protos.UnassignRoleResponse, error) {
	roleName := req.GetRoleName()
	pActor := req.GetActor()

	domainID := pActor.GetID()
	namespace := pActor.GetNamespace()
	actor := perm.Actor{
		ID:        domainID,
		Namespace: namespace,
	}
	logExtensions := []logging.CustomExtension{
		{Key: "roleName", Value: roleName},
		{Key: "userID", Value: pActor.ID},
	}
	s.securityLogger.Log(ctx, "UnassignRole", "Role unassignment", logExtensions...)
	logger := s.logger.Session("unassign-role").WithData(lager.Data{
		"actor.id":        actor.ID,
		"actor.namespace": actor.Namespace,
		"role.name":       roleName,
	})
	logger.Debug(starting)

	err := s.roleAssignmentRepo.UnassignRole(ctx, logger, roleName, domainID, namespace)
	if err != nil {
		return nil, togRPCError(err)
	}

	logger.Debug(success)
	return &protos.UnassignRoleResponse{}, nil
}

func (s *RoleServiceServer) HasRole(
	ctx context.Context,
	req *protos.HasRoleRequest,
) (*protos.HasRoleResponse, error) {
	err := validateHasRoleRequest(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	roleName := req.GetRoleName()
	pActor := req.GetActor()
	actor := perm.Actor{
		ID:        pActor.GetID(),
		Namespace: pActor.GetNamespace(),
	}

	logger := s.logger.Session("has-role").WithData(lager.Data{
		"actor.id":        actor.ID,
		"actor.namespace": actor.Namespace,
		"role.name":       roleName,
	})
	logger.Debug(starting)

	query := repos.HasRoleQuery{
		Actor:    actor,
		RoleName: roleName,
	}

	found, err := s.roleAssignmentRepo.HasRole(ctx, logger, query)
	if err != nil {
		if err == perm.ErrRoleNotFound {
			return &protos.HasRoleResponse{HasRole: false}, nil
		}

		return nil, togRPCError(err)
	}

	logger.Debug(success)
	return &protos.HasRoleResponse{HasRole: found}, nil
}

func (s *RoleServiceServer) ListActorRoles(
	ctx context.Context,
	req *protos.ListActorRolesRequest,
) (*protos.ListActorRolesResponse, error) {
	pActor := req.GetActor()
	actor := perm.Actor{
		ID:        pActor.GetID(),
		Namespace: pActor.GetNamespace(),
	}
	logger := s.logger.Session("list-actor-roles").WithData(lager.Data{
		"actor.id":        actor.ID,
		"actor.namespace": actor.Namespace,
	})
	logger.Debug(starting)

	query := repos.ListActorRolesQuery{Actor: actor}
	roles, err := s.roleAssignmentRepo.ListActorRoles(ctx, logger, query)
	if err != nil {
		return nil, togRPCError(err)
	}

	var pRoles []*protos.Role

	for _, r := range roles {
		pRoles = append(pRoles, &protos.Role{
			Name: r.Name,
		})
	}

	logger.Debug(success)
	return &protos.ListActorRolesResponse{
		Roles: pRoles,
	}, nil
}

func (s *RoleServiceServer) ListRolePermissions(
	ctx context.Context,
	req *protos.ListRolePermissionsRequest,
) (*protos.ListRolePermissionsResponse, error) {
	roleName := req.GetRoleName()
	logger := s.logger.Session("list-role-permissions").WithData(lager.Data{
		"role.name": roleName,
	})
	logger.Debug(starting)

	query := repos.ListRolePermissionsQuery{
		RoleName: roleName,
	}
	permissions, err := s.roleRepo.ListRolePermissions(ctx, logger, query)
	if err != nil {
		if err == perm.ErrRoleNotFound {
			permissions = []*perm.Permission{}
		} else {
			return nil, togRPCError(err)
		}
	}

	var pPermissions []*protos.Permission

	for _, p := range permissions {
		pPermissions = append(pPermissions, &protos.Permission{
			Name:            p.Action,
			ResourcePattern: p.ResourcePattern,
		})
	}

	logger.Debug(success)
	return &protos.ListRolePermissionsResponse{
		Permissions: pPermissions,
	}, nil
}
