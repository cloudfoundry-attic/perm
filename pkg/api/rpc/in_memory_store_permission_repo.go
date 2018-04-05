package rpc

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/api/repos"
	"code.cloudfoundry.org/perm/pkg/perm"
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

	var permissions []*perm.Permission
	action := query.Action

	for _, roleName := range assignments {
		p, ok := s.permissions[roleName]
		if !ok {
			continue
		}

		permissions = append(permissions, p...)
	}

	resourcePattern := query.ResourcePattern

	for _, permission := range permissions {
		if permission.Action == action && permission.ResourcePattern == resourcePattern {
			return true, nil
		}
	}

	return false, nil
}

func (s *InMemoryStore) ListResourcePatterns(
	ctx context.Context,
	logger lager.Logger,
	query repos.ListResourcePatternsQuery,
) ([]string, error) {
	var resourcePatterns []string

	assignments, ok := s.assignments[query.Actor]
	if !ok {
		return resourcePatterns, nil
	}

	var permissions []*perm.Permission
	for _, roleName := range assignments {
		p, ok := s.permissions[roleName]
		if !ok {
			continue
		}

		permissions = append(permissions, p...)
	}

	patternMap := make(map[string]interface{})
	action := query.Action

	for _, permission := range permissions {
		if permission.Action == action {
			patternMap[permission.ResourcePattern] = nil
		}
	}

	for k := range patternMap {
		resourcePatterns = append(resourcePatterns, k)
	}

	return resourcePatterns, nil
}
