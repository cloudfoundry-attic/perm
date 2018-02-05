package models

import (
	"context"

	"code.cloudfoundry.org/lager"
)

type RoleRepo interface {
	CreateRole(
		ctx context.Context,
		logger lager.Logger,
		name RoleName,
		permissions ...*Permission,
	) (*Role, error)

	FindRole(
		context.Context,
		lager.Logger,
		RoleQuery,
	) (*Role, error)

	DeleteRole(
		context.Context,
		lager.Logger,
		RoleQuery,
	) error

	ListRolePermissions(
		ctx context.Context,
		logger lager.Logger,
		query RoleQuery,
	) ([]*Permission, error)
}

type RoleQuery struct {
	Name RoleName
}
