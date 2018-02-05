package models

import (
	"context"

	"code.cloudfoundry.org/lager"
)

type RoleAssignmentRepo interface {
	AssignRole(
		ctx context.Context,
		logger lager.Logger,
		roleName RoleName,
		domainID ActorDomainID,
		issuer ActorIssuer,
	) error

	UnassignRole(
		ctx context.Context,
		logger lager.Logger,
		roleName RoleName,
		domainID ActorDomainID,
		issuer ActorIssuer,
	) error

	HasRole(
		ctx context.Context,
		logger lager.Logger,
		query RoleAssignmentQuery,
	) (bool, error)

	ListActorRoles(
		ctx context.Context,
		logger lager.Logger,
		query ActorQuery,
	) ([]*Role, error)
}

type RoleAssignmentQuery struct {
	RoleQuery  RoleQuery
	ActorQuery ActorQuery
}
