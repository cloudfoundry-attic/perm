package repos

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/perm"
)

type FindRoleQuery struct {
	RoleName string
}

type ListRolePermissionsQuery struct {
	RoleName string
}

type RoleRepo interface {
	CreateRole(
		ctx context.Context,
		logger lager.Logger,
		name string,
		permissions ...*perm.Permission,
	) (*perm.Role, error)

	FindRole(
		context.Context,
		lager.Logger,
		FindRoleQuery,
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
}
