package db

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/models"
	"code.cloudfoundry.org/perm/repos"
	"code.cloudfoundry.org/perm/sqlx"
)

func (s *DataService) AssignRole(
	ctx context.Context,
	logger lager.Logger,
	roleName models.RoleName,
	domainID models.ActorDomainID,
	issuer models.ActorIssuer,
) (err error) {
	logger = logger.Session("data-service")

	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(failedToStartTransaction, err)
		return
	}

	defer func() {
		if err != nil {
			return
		}
		err = sqlx.Commit(logger, tx, err)
	}()

	err = assignRole(ctx, logger, tx, roleName, domainID, issuer)

	return
}

func (s *DataService) UnassignRole(
	ctx context.Context,
	logger lager.Logger,
	roleName models.RoleName,
	domainID models.ActorDomainID,
	issuer models.ActorIssuer,
) (err error) {
	logger = logger.Session("data-service")

	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(failedToStartTransaction, err)
		return
	}

	defer func() {
		if err != nil {
			return
		}
		err = sqlx.Commit(logger, tx, err)
	}()

	err = unassignRole(ctx, logger, tx, roleName, domainID, issuer)

	return
}

func (s *DataService) HasRole(
	ctx context.Context,
	logger lager.Logger,
	query repos.HasRoleQuery,
) (bool, error) {
	return hasRole(ctx, logger.Session("data-service"), s.conn, query)
}

func (s *DataService) ListActorRoles(
	ctx context.Context,
	logger lager.Logger,
	query repos.ListActorRolesQuery,
) ([]*models.Role, error) {
	r, err := listActorRoles(ctx, logger.Session("data-service"), s.conn, query)
	if err != nil {
		return nil, err
	}

	var roles []*models.Role
	for _, role := range r {
		roles = append(roles, role.Role)
	}

	return roles, nil
}
