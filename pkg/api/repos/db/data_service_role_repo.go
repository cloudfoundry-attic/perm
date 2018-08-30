package db

import (
	"context"

	"code.cloudfoundry.org/perm/pkg/api/repos"
	"code.cloudfoundry.org/perm/pkg/logx"
	"code.cloudfoundry.org/perm/pkg/perm"
	"code.cloudfoundry.org/perm/pkg/sqlx"
)

func (s *DataService) CreateRole(
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

func (s *DataService) DeleteRole(
	ctx context.Context,
	logger logx.Logger,
	roleName string,
) error {
	return deleteRole(ctx, logger.WithName("data-service"), s.conn, roleName)
}

func (s *DataService) ListRolePermissions(
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