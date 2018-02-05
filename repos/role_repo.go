package repos

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/models"
)

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
		RoleQuery,
	) (*models.Role, error)

	DeleteRole(
		context.Context,
		lager.Logger,
		RoleQuery,
	) error

	ListRolePermissions(
		ctx context.Context,
		logger lager.Logger,
		query RoleQuery,
	) ([]*models.Permission, error)
}
