package repos

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/models"
)

type FindRoleQuery struct {
	RoleName models.RoleName
}

type ListRolePermissionsQuery struct {
	RoleName models.RoleName
}

type RoleRepo interface {
	CreateRole(
		ctx context.Context,
		logger lager.Logger,
		name models.RoleName,
		permissions ...*models.Permission,
	) (*models.Role, error)

	FindRole(
		context.Context,
		lager.Logger,
		FindRoleQuery,
	) (*models.Role, error)

	DeleteRole(
		context.Context,
		lager.Logger,
		models.RoleName,
	) error

	ListRolePermissions(
		ctx context.Context,
		logger lager.Logger,
		query ListRolePermissionsQuery,
	) ([]*models.Permission, error)
}
