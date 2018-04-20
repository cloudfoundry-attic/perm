package db

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/api/repos"
	"code.cloudfoundry.org/perm/pkg/perm"
	"code.cloudfoundry.org/perm/pkg/sqlx"
)

func (s *DataService) AssignRole(
	ctx context.Context,
	logger lager.Logger,
	roleName,
	domainID,
	namespace string,
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

	err = assignRole(ctx, logger, tx, roleName, domainID, namespace)

	return
}

func (s *DataService) AssignRoleToGroup(
	ctx context.Context,
	logger lager.Logger,
	roleName,
	groupID string,
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

	err = assignRoleToGroup(ctx, logger, tx, roleName, groupID)

	return err
}

func (s *DataService) UnassignRole(
	ctx context.Context,
	logger lager.Logger,
	roleName,
	domainID,
	namespace string,
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

	err = unassignRole(ctx, logger, tx, roleName, domainID, namespace)

	return
}

func (s *DataService) HasRole(
	ctx context.Context,
	logger lager.Logger,
	query repos.HasRoleQuery,
) (bool, error) {
	return hasRole(ctx, logger.Session("data-service"), s.conn, query)
}

func (s *DataService) HasRoleForGroup(
	ctx context.Context,
	logger lager.Logger,
	query repos.HasRoleForGroupQuery,
) (bool, error) {
	return hasRoleForGroup(ctx, logger.Session("data-service"), s.conn, query)
}

func (s *DataService) ListActorRoles(
	ctx context.Context,
	logger lager.Logger,
	query repos.ListActorRolesQuery,
) ([]*perm.Role, error) {
	r, err := listActorRoles(ctx, logger.Session("data-service"), s.conn, query)
	if err != nil {
		return nil, err
	}

	var roles []*perm.Role
	for _, role := range r {
		roles = append(roles, role.Role)
	}

	return roles, nil
}
