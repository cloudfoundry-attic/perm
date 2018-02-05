package repos

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/models"
)

type RoleAssignmentRepo interface {
	AssignRole(
		ctx context.Context,
		logger lager.Logger,
		roleName models.RoleName,
		domainID models.ActorDomainID,
		issuer models.ActorIssuer,
	) error

	UnassignRole(
		ctx context.Context,
		logger lager.Logger,
		roleName models.RoleName,
		domainID models.ActorDomainID,
		issuer models.ActorIssuer,
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
	) ([]*models.Role, error)
}

type RoleAssignmentQuery struct {
	RoleQuery  RoleQuery
	ActorQuery ActorQuery
}
