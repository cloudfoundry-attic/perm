package inmemory

import (
	"context"

	"code.cloudfoundry.org/perm"
	"code.cloudfoundry.org/perm/api/internal/repos"
	"code.cloudfoundry.org/perm/logx"
)

func (s *Store) HasPermission(
	ctx context.Context,
	logger logx.Logger,
	query repos.HasPermissionQuery,
) (bool, error) {
	// Actor-based check
	assignments, _ := s.assignments[actor{
		ID:        query.Actor.ID,
		Namespace: query.Actor.Namespace,
	}]

	var permissions []perm.Permission
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

	// Group-based check
	for _, group := range query.Actor.Groups {
		assignments, _ := s.groupAssignments[group]

		var permissions []perm.Permission
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
	}
	return false, nil
}

func (s *Store) ListResourcePatterns(
	ctx context.Context,
	logger logx.Logger,
	query repos.ListResourcePatternsQuery,
) ([]string, error) {
	var resourcePatterns []string

	assignments, _ := s.assignments[actor{
		ID:        query.Actor.ID,
		Namespace: query.Actor.Namespace,
	}]

	for _, group := range query.Actor.Groups {
		gAssignment, ok := s.groupAssignments[group]
		if ok {
			assignments = append(assignments, gAssignment...)
		}
	}

	if len(assignments) == 0 {
		return resourcePatterns, nil
	}

	var permissions []perm.Permission
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
