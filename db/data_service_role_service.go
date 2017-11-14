package db

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/models"
)

func (s *DataService) CreateRole(ctx context.Context, logger lager.Logger, name string, permissions ...*models.Permission) (*models.Role, error) {
	role, err := createRole(ctx, logger.Session("data-service"), s.conn, name, permissions...)
	if err != nil {
		return nil, err
	}

	return role.Role, nil
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
