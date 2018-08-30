package db

import (
	"context"

	"code.cloudfoundry.org/perm"
	"code.cloudfoundry.org/perm/internal/sqlx"
	"code.cloudfoundry.org/perm/pkg/api/repos"
	"code.cloudfoundry.org/perm/pkg/logx"
)

func (s *Store) CreateRole(
	ctx context.Context,
	logger logx.Logger,
	name string,
	permissions ...perm.Permission,
) (r perm.Role, err error) {
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

	var r2 role
	r2, err = createRoleAndAssignPermissions(ctx, logger, tx, name, permissions...)
	if err != nil {
		return
	}
	r = r2.Role

	return
}

func (s *Store) DeleteRole(
	ctx context.Context,
	logger logx.Logger,
	roleName string,
) error {
	return deleteRole(ctx, logger.WithName("data-service"), s.conn, roleName)
}

func (s *Store) ListRolePermissions(
	ctx context.Context,
	logger logx.Logger,
	query repos.ListRolePermissionsQuery,
) ([]perm.Permission, error) {
	p, err := listRolePermissions(ctx, logger.WithName("data-service"), s.conn, query)
	if err != nil {
		return nil, err
	}

	var permissions []perm.Permission
	for _, permission := range p {
		permissions = append(permissions, permission.Permission)
	}

	return permissions, nil
}

func (s *Store) AssignRole(
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

func (s *Store) AssignRoleToGroup(
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

func (s *Store) UnassignRole(
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

func (s *Store) UnassignRoleFromGroup(
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

func (s *Store) HasRole(
	ctx context.Context,
	logger logx.Logger,
	query repos.HasRoleQuery,
) (bool, error) {
	return hasRole(ctx, logger.WithName("data-service"), s.conn, query)
}

func (s *Store) HasRoleForGroup(
	ctx context.Context,
	logger logx.Logger,
	query repos.HasRoleForGroupQuery,
) (bool, error) {
	return hasRoleForGroup(ctx, logger.WithName("data-service"), s.conn, query)
}
