package rpc

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/models"
)

func (s *InMemoryStore) HasPermission(ctx context.Context, logger lager.Logger, query models.HasPermissionQuery) (bool, error) {
	actor := models.Actor{
		DomainID: query.ActorQuery.DomainID,
		Issuer:   query.ActorQuery.Issuer,
	}

	assignments, ok := s.assignments[actor]
	if !ok {
		return false, nil
	}

	var permissions []*models.Permission
	for _, roleName := range assignments {
		p, ok := s.permissions[roleName]
		if !ok {
			continue
		}

		permissions = append(permissions, p...)
	}

	var hasPermission bool
	for _, permission := range permissions {
		if permission.Name == query.PermissionQuery.PermissionDefinitionQuery.Name && permission.ResourcePattern == query.PermissionQuery.ResourceID {
			hasPermission = true
		}
	}

	return hasPermission, nil
}
