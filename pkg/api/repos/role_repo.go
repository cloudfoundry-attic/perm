package repos

import (
	"context"

	"code.cloudfoundry.org/perm"
	"code.cloudfoundry.org/perm/pkg/logx"
)

type ListRolePermissionsQuery struct {
	RoleName string
}

type HasRoleQuery struct {
	Actor    perm.Actor
	RoleName string
}

type HasRoleForGroupQuery struct {
	Group    perm.Group
	RoleName string
}

type RoleRepo interface {
	CreateRole(
		ctx context.Context,
		logger logx.Logger,
		name string,
		permissions ...perm.Permission,
	) (perm.Role, error)

	DeleteRole(
		context.Context,
		logx.Logger,
		string,
	) error

	ListRolePermissions(
		ctx context.Context,
		logger logx.Logger,
		query ListRolePermissionsQuery,
	) ([]perm.Permission, error)

	AssignRole(
		ctx context.Context,
		logger logx.Logger,
		roleName,
		domainID,
		namespace string,
	) error

	AssignRoleToGroup(
		ctx context.Context,
		logger logx.Logger,
		roleName,
		groupID string,
	) error

	UnassignRole(
		ctx context.Context,
		logger logx.Logger,
		roleName,
		domainID,
		namespace string,
	) error

	UnassignRoleFromGroup(
		ctx context.Context,
		logger logx.Logger,
		roleName,
		groupID string,
	) error

	HasRole(
		ctx context.Context,
		logger logx.Logger,
		query HasRoleQuery,
	) (bool, error)

	HasRoleForGroup(
		ctx context.Context,
		logger logx.Logger,
		query HasRoleForGroupQuery,
	) (bool, error)
}
