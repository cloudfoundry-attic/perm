package repos

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/perm"
)

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

type RoleAssignmentRepo interface {
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
