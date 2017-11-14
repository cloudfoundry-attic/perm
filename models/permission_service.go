package models

import (
	"context"

	"code.cloudfoundry.org/lager"
)

type PermissionQuery struct {
	Name       string
	ResourceID string
}

type HasPermissionQuery struct {
	PermissionQuery PermissionQuery
	ActorQuery      ActorQuery
}

type PermissionService interface {
	HasPermission(ctx context.Context, logger lager.Logger, query HasPermissionQuery) (bool, error)
}
