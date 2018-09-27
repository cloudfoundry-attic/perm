package repos

import (
	"context"

	"code.cloudfoundry.org/perm"
	"code.cloudfoundry.org/perm/logx"
)

type HasPermissionQuery struct {
	Actor           perm.Actor
	Action          string
	ResourcePattern string
}

type ListResourcePatternsQuery struct {
	Actor  perm.Actor
	Action string
}

//go:generate counterfeiter . PermissionRepo

type PermissionRepo interface {
	HasPermission(
		ctx context.Context,
		logger logx.Logger,
		query HasPermissionQuery,
	) (bool, error)

	ListResourcePatterns(
		ctx context.Context,
		logger logx.Logger,
		query ListResourcePatternsQuery,
	) ([]string, error)
}
