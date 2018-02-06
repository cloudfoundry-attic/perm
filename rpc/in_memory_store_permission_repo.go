package rpc

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/models"
	"code.cloudfoundry.org/perm/repos"
)

func (s *InMemoryStore) HasPermission(
	ctx context.Context,
	logger lager.Logger,
	query repos.HasPermissionQuery,
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

func (s *InMemoryStore) ListResourcePatterns(
	ctx context.Context,
	logger lager.Logger,
	query repos.ListResourcePatternsQuery,
) ([]models.PermissionResourcePattern, error) {
	actor := models.Actor{
		DomainID: query.ActorQuery.DomainID,
		Issuer:   query.ActorQuery.Issuer,
	}
	permissionName := query.PermissionDefinitionQuery.Name

	var resourcePatterns []models.PermissionResourcePattern

	assignments, ok := s.assignments[actor]
	if !ok {
		return resourcePatterns, nil
	}

	var permissions []*models.Permission
	for _, roleName := range assignments {
		p, ok := s.permissions[roleName]
		if !ok {
			continue
		}

		permissions = append(permissions, p...)
	}

	patternMap := make(map[models.PermissionResourcePattern]interface{})

	for _, permission := range permissions {
		if permission.Name == permissionName {
			patternMap[permission.ResourcePattern] = nil
		}
	}

	for k := range patternMap {
		resourcePatterns = append(resourcePatterns, k)
	}

	return resourcePatterns, nil
}

func hasPermission(permission *models.Permission, query repos.HasPermissionQuery) bool {
	namesMatch := permission.Name == query.PermissionQuery.PermissionName
	resourcesMatch := string(permission.ResourcePattern) == string(query.PermissionQuery.ResourcePattern)

	return namesMatch && resourcesMatch
}
