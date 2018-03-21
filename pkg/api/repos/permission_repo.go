package repos

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/api/models"
)

type HasPermissionQuery struct {
	Actor           models.Actor
	PermissionName  models.PermissionName
	ResourcePattern models.PermissionResourcePattern
}

type ListResourcePatternsQuery struct {
	Actor          models.Actor
	PermissionName models.PermissionName
}

//go:generate counterfeiter . PermissionRepo

type PermissionRepo interface {
	HasPermission(
		ctx context.Context,
		logger lager.Logger,
		query HasPermissionQuery,
	) (bool, error)

	ListResourcePatterns(
		ctx context.Context,
		logger lager.Logger,
		query ListResourcePatternsQuery,
	) ([]models.PermissionResourcePattern, error)
}
