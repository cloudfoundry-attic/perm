package repos

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/perm"
)

type ListRolePermissionsQuery struct {
	RoleName string
}

type ListActorRolesQuery struct {
	Actor perm.Actor
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
		logger lager.Logger,
		name string,
		permissions ...*perm.Permission,
	) (*perm.Role, error)

	DeleteRole(
		context.Context,
		lager.Logger,
		string,
	) error

	ListRolePermissions(
		ctx context.Context,
		logger lager.Logger,
		query ListRolePermissionsQuery,
	) ([]*perm.Permission, error)

	AssignRole(
		ctx context.Context,
		logger lager.Logger,
		roleName,
		domainID,
		namespace string,
	) error

	AssignRoleToGroup(
		ctx context.Context,
		logger lager.Logger,
		roleName,
		groupID string,
	) error

	UnassignRole(
		ctx context.Context,
		logger lager.Logger,
		roleName,
		domainID,
		namespace string,
	) error

	UnassignRoleFromGroup(
		ctx context.Context,
		logger lager.Logger,
		roleName,
		groupID string,
	) error

	HasRole(
		ctx context.Context,
		logger lager.Logger,
		query HasRoleQuery,
	) (bool, error)

	HasRoleForGroup(
		ctx context.Context,
		logger lager.Logger,
		query HasRoleForGroupQuery,
	) (bool, error)

	ListActorRoles(
		ctx context.Context,
		logger lager.Logger,
		query ListActorRolesQuery,
	) ([]*perm.Role, error)
}
