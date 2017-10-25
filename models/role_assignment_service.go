package models

import (
	"context"

	"code.cloudfoundry.org/lager"
)

type RoleAssignmentService interface {
	AssignRole(ctx context.Context, logger lager.Logger, roleName string, domainID string, issuer string) error
	UnassignRole(ctx context.Context, logger lager.Logger, roleName string, domainID string, issuer string) error
	HasRole(ctx context.Context, logger lager.Logger, query RoleAssignmentQuery) (bool, error)
	ListActorRoles(ctx context.Context, logger lager.Logger, query ActorQuery) ([]*Role, error)
}

type RoleAssignmentQuery struct {
	RoleQuery  RoleQuery
	ActorQuery ActorQuery
}
