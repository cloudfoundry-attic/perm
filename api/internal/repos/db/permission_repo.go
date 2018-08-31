package db

import (
	"context"

	"code.cloudfoundry.org/perm/api/internal/repos"
	"code.cloudfoundry.org/perm/pkg/logx"
)

func (s *Store) HasPermission(
	ctx context.Context,
	logger logx.Logger,
	query repos.HasPermissionQuery,
) (bool, error) {
	return hasPermission(ctx, logger.WithName("data-service"), s.conn, query)
}

func (s *Store) ListResourcePatterns(
	ctx context.Context,
	logger logx.Logger,
	query repos.ListResourcePatternsQuery,
) ([]string, error) {
	return listResourcePatterns(ctx, logger.WithName("data-service"), s.conn, query)
}
