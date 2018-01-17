package models

import (
	"context"

	"code.cloudfoundry.org/lager"
)

type PermissionDefinitionQuery struct {
	Name string
}

type PermissionQuery struct {
	PermissionDefinitionQuery PermissionDefinitionQuery
	ResourceID                string
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
