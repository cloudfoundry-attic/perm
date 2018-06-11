package repos

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/perm"
)

type HasPermissionQuery struct {
	Actor           perm.Actor
	Action          string
	ResourcePattern string
	Groups          []perm.Group
}

type Groups []perm.Group

func (g *Groups) GetIDs() []string {
	ids := []string{}
	for _, group := range *g {
		ids = append(ids, group.ID)
	}
	return ids
}

type ListResourcePatternsQuery struct {
	Actor  perm.Actor
	Groups Groups
	Action string
}

//go:generate counterfeiter . PermissionRepo

type PermissionRepo interface {
	HasPermission(
		ctx context.Context,
		logger lager.Logger,
		query HasPermissionQuery,
	) (bool, error)

	ListResourcePatterns(
		ctx context.Context,
		logger lager.Logger,
		query ListResourcePatternsQuery,
	) ([]string, error)
}
