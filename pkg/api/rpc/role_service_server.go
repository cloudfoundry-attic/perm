package rpc

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/api/logging"
	"code.cloudfoundry.org/perm/pkg/api/repos"
	"code.cloudfoundry.org/perm/pkg/perm"
	"code.cloudfoundry.org/perm/protos/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RoleServiceServer struct {
	logger         lager.Logger
	securityLogger SecurityLogger

	roleRepo repos.RoleRepo
}

func NewRoleServiceServer(
	logger lager.Logger,
	securityLogger SecurityLogger,
	roleRepo repos.RoleRepo,
) *RoleServiceServer {
	return &RoleServiceServer{
		logger:         logger,
		securityLogger: securityLogger,
		roleRepo:       roleRepo,
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
			Action:          p.GetAction(),
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

func (s *RoleServiceServer) AssignRole(
	ctx context.Context,
	req *protos.AssignRoleRequest,
) (*protos.AssignRoleResponse, error) {
	roleName := req.GetRoleName()
	pActor := req.GetActor()
	err := validateActor(pActor)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

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

	err = s.roleRepo.AssignRole(ctx, logger, roleName, domainID, namespace)
	if err != nil {
		return nil, togRPCError(err)
	}

	logger.Debug(success)
	return &protos.AssignRoleResponse{}, nil
}

func (s *RoleServiceServer) AssignRoleToGroup(
	ctx context.Context,
	req *protos.AssignRoleToGroupRequest,
) (*protos.AssignRoleToGroupResponse, error) {
	roleName := req.GetRoleName()
	pGroup := req.GetGroup()

	groupID := pGroup.GetID()
	logExtensions := []logging.CustomExtension{
		{Key: "roleName", Value: roleName},
		{Key: "groupID", Value: pGroup.ID},
	}

	s.securityLogger.Log(ctx, "AssignRoleToGroup", "Role assignment", logExtensions...)
	logger := s.logger.Session("assign-role").WithData(lager.Data{
		"actor.id":  groupID,
		"role.name": roleName,
	})
	logger.Debug(starting)

	err := s.roleRepo.AssignRoleToGroup(ctx, logger, roleName, groupID)
	if err != nil {
		return nil, togRPCError(err)
	}

	logger.Debug(success)
	return &protos.AssignRoleToGroupResponse{}, nil
}

func (s *RoleServiceServer) UnassignRole(
	ctx context.Context,
	req *protos.UnassignRoleRequest,
) (*protos.UnassignRoleResponse, error) {
	roleName := req.GetRoleName()
	pActor := req.GetActor()
	err := validateActor(pActor)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

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

	err = s.roleRepo.UnassignRole(ctx, logger, roleName, domainID, namespace)
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
	roleName := req.GetRoleName()
	pActor := req.GetActor()
	err := validateActor(pActor)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

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

	found, err := s.roleRepo.HasRole(ctx, logger, query)
	if err != nil {
		if err == perm.ErrRoleNotFound {
			return &protos.HasRoleResponse{HasRole: false}, nil
		}

		return nil, togRPCError(err)
	}

	logger.Debug(success)
	return &protos.HasRoleResponse{HasRole: found}, nil
}

func (s *RoleServiceServer) HasRoleForGroup(
	ctx context.Context,
	req *protos.HasRoleForGroupRequest,
) (*protos.HasRoleForGroupResponse, error) {
	roleName := req.GetRoleName()
	pGroup := req.GetGroup()

	group := perm.Group{
		ID: pGroup.GetID(),
	}

	logger := s.logger.Session("has-role").WithData(lager.Data{
		"group.id":  group.ID,
		"role.name": roleName,
	})
	logger.Debug(starting)

	query := repos.HasRoleForGroupQuery{
		Group:    group,
		RoleName: roleName,
	}

	found, err := s.roleRepo.HasRoleForGroup(ctx, logger, query)
	if err != nil {
		if err == perm.ErrRoleNotFound {
			return &protos.HasRoleForGroupResponse{HasRole: false}, nil
		}

		return nil, togRPCError(err)
	}

	logger.Debug(success)
	return &protos.HasRoleForGroupResponse{HasRole: found}, nil
}

func (s *RoleServiceServer) ListActorRoles(
	ctx context.Context,
	req *protos.ListActorRolesRequest,
) (*protos.ListActorRolesResponse, error) {
	pActor := req.GetActor()
	err := validateActor(pActor)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

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
	roles, err := s.roleRepo.ListActorRoles(ctx, logger, query)
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
			Action:          p.Action,
			ResourcePattern: p.ResourcePattern,
		})
	}

	logger.Debug(success)
	return &protos.ListRolePermissionsResponse{
		Permissions: pPermissions,
	}, nil
}
