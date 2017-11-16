package db

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
	"code.cloudfoundry.org/perm/models"
	"code.cloudfoundry.org/perm/sqlx"
)

func (s *DataService) CreateRole(ctx context.Context, logger lager.Logger, name string, permissions ...*models.Permission) (r *models.Role, err error) {
	logger = logger.Session("data-service")

	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(messages.FailedToStartTransaction, err)
		return
	}

	defer func() {
		if err != nil {
			return
		}
		err = sqlx.Commit(logger, tx, err)
	}()

	var r2 *role
	r2, err = createRoleAndAssignPermissions(ctx, logger, tx, name, permissions...)
	if err != nil {
		return
	}
	r = r2.Role

	return
}

func (s *DataService) FindRole(ctx context.Context, logger lager.Logger, query models.RoleQuery) (*models.Role, error) {
	role, err := findRole(ctx, logger.Session("data-service"), s.conn, query)
	if err != nil {
		return nil, err
	}
	return role.Role, nil
}

func (s *DataService) DeleteRole(ctx context.Context, logger lager.Logger, query models.RoleQuery) error {
	return deleteRole(ctx, logger.Session("data-service"), s.conn, query)
}

func (s *DataService) ListRolePermissions(ctx context.Context, logger lager.Logger, query models.RoleQuery) ([]*models.Permission, error) {
	p, err := listRolePermissions(ctx, logger.Session("data-service"), s.conn, query)
	if err != nil {
		return nil, err
	}

	var permissions []*models.Permission
	for _, permission := range p {
		permissions = append(permissions, permission.Permission)
	}

	return permissions, nil
}
