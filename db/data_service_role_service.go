package db

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/models"
)

func (s *DataService) CreateRole(ctx context.Context, logger lager.Logger, name string) (*models.Role, error) {
	role, err := createRole(ctx, logger.Session("data-service"), s.conn, name)
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
