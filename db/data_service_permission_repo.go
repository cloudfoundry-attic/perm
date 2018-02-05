package db

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/models"
)

func (s *DataService) HasPermission(
	ctx context.Context,
	logger lager.Logger,
	query models.HasPermissionQuery,
) (bool, error) {
	return hasPermission(ctx, logger.Session("data-service"), s.conn, query)
}

func (s *DataService) ListResourcePatterns(
	ctx context.Context,
	logger lager.Logger,
	query models.ListResourcePatternsQuery,
) ([]models.PermissionResourcePattern, error) {
	return listResourcePatterns(ctx, logger.Session("data-service"), s.conn, query)
}
