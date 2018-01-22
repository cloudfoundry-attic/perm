package rpc

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/models"
)

func (s *InMemoryStore) HasPermission(
	ctx context.Context,
	logger lager.Logger,
	query models.HasPermissionQuery,
) (bool, error) {
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

	for _, permission := range permissions {
		if hasPermission(permission, query) {
			return true, nil
		}
	}

	return false, nil
}

func hasPermission(permission *models.Permission, query models.HasPermissionQuery) bool {
	namesMatch := permission.Name == query.PermissionQuery.PermissionName
	resourcesMatch := string(permission.ResourcePattern) == string(query.PermissionQuery.ResourceID)

	return namesMatch && resourcesMatch
}
