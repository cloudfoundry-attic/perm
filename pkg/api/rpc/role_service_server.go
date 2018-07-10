package rpc

import (
	"context"

	"code.cloudfoundry.org/perm/pkg/api/protos"
	"code.cloudfoundry.org/perm/pkg/api/repos"
	"code.cloudfoundry.org/perm/pkg/logx"
	"code.cloudfoundry.org/perm/pkg/perm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RoleServiceServer struct {
	logger         logx.Logger
	securityLogger logx.SecurityLogger

	roleRepo repos.RoleRepo
}

func NewRoleServiceServer(
	logger logx.Logger,
	securityLogger logx.SecurityLogger,
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
	var permissions []perm.Permission
	for _, p := range req.GetPermissions() {
		permissions = append(permissions, perm.Permission{
			Action:          p.GetAction(),
			ResourcePattern: p.GetResourcePattern(),
		})
	}
	logExtensions := logx.SecurityData{Key: "roleName", Value: name}
	s.securityLogger.Log(ctx, "CreateRole", "Role creation", logExtensions)

	logger := s.logger.WithName("create-role").WithData(
		logx.Data{"role.name", name},
		logx.Data{"permissions", permissions},
	)
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
	logExtensions := logx.SecurityData{Key: "roleName", Value: name}
	s.securityLogger.Log(ctx, "DeleteRole", "Role deletion", logExtensions)
	logger := s.logger.WithName("delete-role").WithData(logx.Data{"role.name", name})
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
	logExtensions := []logx.SecurityData{
		{Key: "roleName", Value: roleName},
		{Key: "userID", Value: pActor.ID},
	}

	s.securityLogger.Log(ctx, "AssignRole", "Role assignment", logExtensions...)
	logger := s.logger.WithName("assign-role").WithData(
		logx.Data{"actor.id", domainID},
		logx.Data{"actor.namespace", namespace},
		logx.Data{"role.name", roleName},
	)
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
	logExtensions := []logx.SecurityData{
		{Key: "roleName", Value: roleName},
		{Key: "groupID", Value: pGroup.ID},
	}

	s.securityLogger.Log(ctx, "AssignRoleToGroup", "Role assignment", logExtensions...)
	logger := s.logger.WithName("assign-role").WithData(
		logx.Data{"actor.id", groupID},
		logx.Data{"role.name", roleName},
	)
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
	logExtensions := []logx.SecurityData{
		{Key: "roleName", Value: roleName},
		{Key: "userID", Value: pActor.ID},
	}
	s.securityLogger.Log(ctx, "UnassignRole", "Role unassignment", logExtensions...)
	logger := s.logger.WithName("unassign-role").WithData(
		logx.Data{"actor.id", actor.ID},
		logx.Data{"actor.namespace", actor.Namespace},
		logx.Data{"role.name", roleName},
	)
	logger.Debug(starting)

	err = s.roleRepo.UnassignRole(ctx, logger, roleName, domainID, namespace)
	if err != nil {
		return nil, togRPCError(err)
	}

	logger.Debug(success)
	return &protos.UnassignRoleResponse{}, nil
}

func (s *RoleServiceServer) UnassignRoleFromGroup(
	ctx context.Context,
	req *protos.UnassignRoleFromGroupRequest,
) (*protos.UnassignRoleFromGroupResponse, error) {
	roleName := req.GetRoleName()
	pGroup := req.GetGroup()

	domainID := pGroup.GetID()
	group := perm.Group{
		ID: domainID,
	}
	logExtensions := []logx.SecurityData{
		{Key: "roleName", Value: roleName},
		{Key: "userID", Value: pGroup.ID},
	}
	s.securityLogger.Log(ctx, "UnassignRoleFromGroup", "Role group unassignment", logExtensions...)
	logger := s.logger.WithName("unassign-role-from-group").WithData(
		logx.Data{"group.id", group.ID},
		logx.Data{"role.name", roleName},
	)
	logger.Debug(starting)

	err := s.roleRepo.UnassignRoleFromGroup(ctx, logger, roleName, pGroup.ID)
	if err != nil {
		return nil, togRPCError(err)
	}

	logger.Debug(success)
	return &protos.UnassignRoleFromGroupResponse{}, nil
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

	logger := s.logger.WithName("has-role").WithData(
		logx.Data{"actor.id", actor.ID},
		logx.Data{"actor.namespace", actor.Namespace},
		logx.Data{"role.name", roleName},
	)
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

	logger := s.logger.WithName("has-role").WithData(
		logx.Data{"group.id", group.ID},
		logx.Data{"role.name", roleName},
	)
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

func (s *RoleServiceServer) ListRolePermissions(
	ctx context.Context,
	req *protos.ListRolePermissionsRequest,
) (*protos.ListRolePermissionsResponse, error) {
	roleName := req.GetRoleName()
	logger := s.logger.WithName("list-role-permissions").WithData(logx.Data{"role.name", roleName})
	logger.Debug(starting)

	query := repos.ListRolePermissionsQuery{
		RoleName: roleName,
	}
	permissions, err := s.roleRepo.ListRolePermissions(ctx, logger, query)
	if err != nil {
		if err == perm.ErrRoleNotFound {
			permissions = []perm.Permission{}
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
