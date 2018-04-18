package repos

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/perm"
)

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
