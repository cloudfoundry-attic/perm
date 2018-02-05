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

type ListResourcePatternsQuery struct {
	ActorQuery     ActorQuery
	PermissionName PermissionName
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
	) ([]PermissionResourcePattern, error)
}
