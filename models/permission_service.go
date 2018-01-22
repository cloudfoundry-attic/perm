package models

import (
	"context"

	"code.cloudfoundry.org/lager"
)

type PermissionDefinitionQuery struct {
	Name PermissionDefinitionName
}

type ResourceID string

type PermissionQuery struct {
	PermissionName PermissionName
	ResourceID     ResourceID
}

type HasPermissionQuery struct {
	PermissionQuery PermissionQuery
	ActorQuery      ActorQuery
}

type PermissionService interface {
	HasPermission(
		ctx context.Context,
		logger lager.Logger,
		query HasPermissionQuery,
	) (bool, error)
}
