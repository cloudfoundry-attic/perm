package inmemory

import (
	"context"

	"code.cloudfoundry.org/perm/pkg/api/repos"
	"code.cloudfoundry.org/perm/pkg/logx"
	"code.cloudfoundry.org/perm/pkg/perm"
)

func (s *InMemoryStore) CreateRole(
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

func (s *InMemoryStore) DeleteRole(
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
					logx.Data{"actor.id", actor.ID},
					logx.Data{"actor.namespace", actor.Namespace},
					logx.Data{"role.name", roleName},
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

func (s *InMemoryStore) AssignRole(
	ctx context.Context,
	logger logx.Logger,
	roleName,
	id,
	namespace string,
) error {
	if _, exists := s.roles[roleName]; !exists {
		return perm.ErrRoleNotFound
	}
	actor := perm.Actor{
		ID:        id,
		Namespace: namespace,
	}

	assignments, ok := s.assignments[actor]
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

	s.assignments[actor] = assignments
	return nil
}

func (s *InMemoryStore) AssignRoleToGroup(
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

func (s *InMemoryStore) UnassignRole(
	ctx context.Context,
	logger logx.Logger,
	roleName,
	id,
	namespace string,
) error {
	if _, exists := s.roles[roleName]; !exists {
		return perm.ErrRoleNotFound
	}

	actor := perm.Actor{
		ID:        id,
		Namespace: namespace,
	}

	assignments, ok := s.assignments[actor]
	if !ok {
		err := perm.ErrAssignmentNotFound
		logger.Error(errRoleAssignmentNotFound, err)
		return err
	}

	for i, assignment := range assignments {
		if assignment == roleName {
			s.assignments[actor] = append(assignments[:i], assignments[i+1:]...)
			logger.Debug(success)
			return nil
		}
	}

	err := perm.ErrAssignmentNotFound
	logger.Error(errRoleAssignmentNotFound, err)

	return err
}

func (s *InMemoryStore) UnassignRoleFromGroup(
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

func (s *InMemoryStore) HasRole(
	ctx context.Context,
	logger logx.Logger,
	query repos.HasRoleQuery,
) (bool, error) {
	roleName := query.RoleName

	_, ok := s.roles[roleName]
	if !ok {
		return false, perm.ErrRoleNotFound
	}

	assignments, ok := s.assignments[query.Actor]
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

func (s *InMemoryStore) HasRoleForGroup(
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

func (s *InMemoryStore) ListRolePermissions(
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
