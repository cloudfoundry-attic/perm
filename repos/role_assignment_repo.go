package repos

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/models"
)

type ListActorRolesQuery struct {
	Actor models.Actor
}

type HasRoleQuery struct {
	Actor    models.Actor
	RoleName models.RoleName
}

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
		query HasRoleQuery,
	) (bool, error)

	ListActorRoles(
		ctx context.Context,
		logger lager.Logger,
		query ListActorRolesQuery,
	) ([]*models.Role, error)
}
