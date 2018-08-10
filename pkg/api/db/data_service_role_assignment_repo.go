package db

import (
	"context"

	"code.cloudfoundry.org/perm/pkg/api/repos"
	"code.cloudfoundry.org/perm/pkg/logx"
	"code.cloudfoundry.org/perm/pkg/sqlx"
)

func (s *DataService) AssignRole(
	ctx context.Context,
	logger logx.Logger,
	roleName,
	domainID,
	namespace string,
) (err error) {
	logger = logger.WithName("data-service")

	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(failedToStartTransaction, err)
		return
	}

	defer func() {
		if commitErr := sqlx.Commit(logger, tx, err); commitErr != nil {
			err = commitErr
		}
	}()

	err = assignRole(ctx, logger, tx, roleName, domainID, namespace)

	return
}

func (s *DataService) AssignRoleToGroup(
	ctx context.Context,
	logger logx.Logger,
	roleName,
	groupID string,
) (err error) {
	logger = logger.WithName("data-service")

	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(failedToStartTransaction, err)
		return
	}

	defer func() {
		if commitErr := sqlx.Commit(logger, tx, err); commitErr != nil {
			err = commitErr
		}
	}()

	err = assignRoleToGroup(ctx, logger, tx, roleName, groupID)

	return err
}

func (s *DataService) UnassignRole(
	ctx context.Context,
	logger logx.Logger,
	roleName,
	domainID,
	namespace string,
) (err error) {
	logger = logger.WithName("data-service")

	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(failedToStartTransaction, err)
		return
	}

	defer func() {
		if commitErr := sqlx.Commit(logger, tx, err); commitErr != nil {
			err = commitErr
		}
	}()

	err = unassignRole(ctx, logger, tx, roleName, domainID, namespace)

	return
}

func (s *DataService) UnassignRoleFromGroup(
	ctx context.Context,
	logger logx.Logger,
	roleName,
	groupID string,
) (err error) {
	logger = logger.WithName("data-service")

	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(failedToStartTransaction, err)
		return
	}

	defer func() {
		if commitErr := sqlx.Commit(logger, tx, err); commitErr != nil {
			err = commitErr
		}
	}()

	err = unassignRoleFromGroup(ctx, logger, tx, roleName, groupID)
	return
}

func (s *DataService) HasRole(
	ctx context.Context,
	logger logx.Logger,
	query repos.HasRoleQuery,
) (bool, error) {
	return hasRole(ctx, logger.WithName("data-service"), s.conn, query)
}

func (s *DataService) HasRoleForGroup(
	ctx context.Context,
	logger logx.Logger,
	query repos.HasRoleForGroupQuery,
) (bool, error) {
	return hasRoleForGroup(ctx, logger.WithName("data-service"), s.conn, query)
}
