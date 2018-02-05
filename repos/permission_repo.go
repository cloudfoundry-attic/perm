package repos

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/models"
)

type HasPermissionQuery struct {
	PermissionQuery PermissionQuery
	ActorQuery      ActorQuery
}

type ListResourcePatternsQuery struct {
	ActorQuery     ActorQuery
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
