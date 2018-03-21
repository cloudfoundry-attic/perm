package rpc

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/api/models"
	"code.cloudfoundry.org/perm/pkg/api/repos"
)

func (s *InMemoryStore) HasPermission(
	ctx context.Context,
	logger lager.Logger,
	query repos.HasPermissionQuery,
) (bool, error) {
	assignments, ok := s.assignments[query.Actor]
	if !ok {
		return false, nil
	}

	var permissions []*models.Permission
	permissionName := query.PermissionName

	for _, roleName := range assignments {
		p, ok := s.permissions[roleName]
		if !ok {
			continue
		}

		permissions = append(permissions, p...)
	}

	resourcePattern := query.ResourcePattern

	for _, permission := range permissions {
		if permission.Name == permissionName && permission.ResourcePattern == resourcePattern {
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
	var resourcePatterns []models.PermissionResourcePattern

	assignments, ok := s.assignments[query.Actor]
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
	permissionName := query.PermissionName

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
