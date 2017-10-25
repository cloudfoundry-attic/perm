package models

import (
	"context"

	"code.cloudfoundry.org/lager"
)

type RoleService interface {
	CreateRole(ctx context.Context, logger lager.Logger, name string) (*Role, error)
	FindRole(context.Context, lager.Logger, RoleQuery) (*Role, error)
	DeleteRole(context.Context, lager.Logger, RoleQuery) error
}

type RoleQuery struct {
	Name string
}
