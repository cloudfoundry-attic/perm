package inmemory

import (
	"context"

	"code.cloudfoundry.org/perm"
	"code.cloudfoundry.org/perm/api/internal/repos"
	"code.cloudfoundry.org/perm/logx"
)

func (s *Store) CreateRole(
	ctx context.Context,
	logger logx.Logger,
	name string,
	permissions ...perm.Permission,
) (perm.Role, error) {
	if _, exists := s.roles[name]; exists {
		return perm.Role{}, perm.ErrRoleAlreadyExists
	}

	role := perm.Role{
		Name: name,
	}
	s.roles[name] = role

	s.permissions[name] = permissions
	return role, nil
}

func (s *Store) DeleteRole(
	ctx context.Context,
	logger logx.Logger,
	roleName string,
) error {
	if _, exists := s.roles[roleName]; !exists {
		return perm.ErrRoleNotFound
	}

	delete(s.roles, roleName)

	// "Cascade"
	// Remove role assignments for role
	for actor, assignments := range s.assignments {
		for i, roleName := range assignments {
			if roleName == roleName {
				s.assignments[actor] = append(assignments[:i], assignments[i+1:]...)
				assignmentData := []logx.Data{
					{Key: "actor.id", Value: actor.ID},
					{Key: "actor.namespace", Value: actor.Namespace},
					{Key: "role.name", Value: roleName},
				}
				logger.Debug(success, assignmentData...)
				break
			}
		}
	}
	// "Cascade"
	// Remove permissions for role
	delete(s.permissions, roleName)

	logger.Debug(success)

	return nil
}

func (s *Store) AssignRole(
	ctx context.Context,
	logger logx.Logger,
	roleName,
	id,
	namespace string,
) error {
	if _, exists := s.roles[roleName]; !exists {
		return perm.ErrRoleNotFound
	}
	a := actor{
		ID:        id,
		Namespace: namespace,
	}

	assignments, ok := s.assignments[a]
	if !ok {
		assignments = []string{}
	}

	for _, role := range assignments {
		if role == roleName {
			err := perm.ErrAssignmentAlreadyExists
			logger.Error(errRoleAssignmentAlreadyExists, err)
			return err
		}
	}

	assignments = append(assignments, roleName)

	s.assignments[a] = assignments
	return nil
}

func (s *Store) AssignRoleToGroup(
	ctx context.Context,
	logger logx.Logger,
	roleName,
	id string,
) error {
	if _, exists := s.roles[roleName]; !exists {
		return perm.ErrRoleNotFound
	}
	group := perm.Group{
		ID: id,
	}

	groupAssignments, ok := s.groupAssignments[group]
	if !ok {
		groupAssignments = []string{}
	}

	for _, role := range groupAssignments {
		if role == roleName {
			err := perm.ErrAssignmentAlreadyExists
			logger.Error(errRoleAssignmentAlreadyExists, err)
			return err
		}
	}

	groupAssignments = append(groupAssignments, roleName)

	s.groupAssignments[group] = groupAssignments
	return nil
}

func (s *Store) UnassignRole(
	ctx context.Context,
	logger logx.Logger,
	roleName,
	id,
	namespace string,
) error {
	if _, exists := s.roles[roleName]; !exists {
		return perm.ErrRoleNotFound
	}

	a := actor{
		ID:        id,
		Namespace: namespace,
	}

	assignments, ok := s.assignments[a]
	if !ok {
		err := perm.ErrAssignmentNotFound
		logger.Error(errRoleAssignmentNotFound, err)
		return err
	}

	for i, assignment := range assignments {
		if assignment == roleName {
			s.assignments[a] = append(assignments[:i], assignments[i+1:]...)
			logger.Debug(success)
			return nil
		}
	}

	err := perm.ErrAssignmentNotFound
	logger.Error(errRoleAssignmentNotFound, err)

	return err
}

func (s *Store) UnassignRoleFromGroup(
	ctx context.Context,
	logger logx.Logger,
	roleName,
	groupID string,
) error {
	if _, exists := s.roles[roleName]; !exists {
		return perm.ErrRoleNotFound
	}

	group := perm.Group{
		ID: groupID,
	}

	groupAssignments, ok := s.groupAssignments[group]
	if !ok {
		err := perm.ErrAssignmentNotFound
		logger.Error(errRoleAssignmentNotFound, err)
		return err
	}

	for i, assignment := range groupAssignments {
		if assignment == roleName {
			s.groupAssignments[group] = append(groupAssignments[:i], groupAssignments[i+1:]...)
			logger.Debug(success)
			return nil
		}
	}

	err := perm.ErrAssignmentNotFound
	logger.Error(errRoleAssignmentNotFound, err)

	return err
}

func (s *Store) HasRole(
	ctx context.Context,
	logger logx.Logger,
	query repos.HasRoleQuery,
) (bool, error) {
	roleName := query.RoleName

	_, ok := s.roles[roleName]
	if !ok {
		return false, perm.ErrRoleNotFound
	}

	assignments, ok := s.assignments[actor{
		ID:        query.Actor.ID,
		Namespace: query.Actor.Namespace,
	}]
	if !ok {
		return false, nil
	}

	var found bool

	for _, name := range assignments {
		if name == roleName {
			found = true
			break
		}
	}

	return found, nil
}

func (s *Store) HasRoleForGroup(
	ctx context.Context,
	logger logx.Logger,
	query repos.HasRoleForGroupQuery,
) (bool, error) {
	roleName := query.RoleName

	_, ok := s.roles[roleName]
	if !ok {
		return false, perm.ErrRoleNotFound
	}

	groupAssignments, ok := s.groupAssignments[query.Group]
	if !ok {
		return false, nil
	}

	var found bool

	for _, name := range groupAssignments {
		if name == roleName {
			found = true
			break
		}
	}

	return found, nil
}

func (s *Store) ListRolePermissions(
	ctx context.Context,
	logger logx.Logger,
	query repos.ListRolePermissionsQuery,
) ([]perm.Permission, error) {
	permissions, exists := s.permissions[query.RoleName]
	if !exists {
		return nil, perm.ErrRoleNotFound
	}

	return permissions, nil
}
