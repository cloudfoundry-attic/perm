package rpc

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
	"code.cloudfoundry.org/perm/models"
	"code.cloudfoundry.org/perm/repos"
)

func (s *InMemoryStore) CreateRole(
	ctx context.Context,
	logger lager.Logger,
	name models.RoleName,
	permissions ...*models.Permission,
) (*models.Role, error) {
	if _, exists := s.roles[name]; exists {
		return nil, models.ErrRoleAlreadyExists
	}

	role := &models.Role{
		Name: name,
	}
	s.roles[name] = role

	s.permissions[name] = permissions
	return role, nil
}

func (s *InMemoryStore) FindRole(
	ctx context.Context,
	logger lager.Logger,
	query repos.FindRoleQuery,
) (*models.Role, error) {
	name := query.RoleName
	role, exists := s.roles[name]

	if !exists {
		return nil, models.ErrRoleNotFound
	}

	return role, nil
}

func (s *InMemoryStore) DeleteRole(
	ctx context.Context,
	logger lager.Logger,
	roleName models.RoleName,
) error {
	if _, exists := s.roles[roleName]; !exists {
		return models.ErrRoleNotFound
	}

	delete(s.roles, roleName)

	// "Cascade"
	// Remove role assignments for role
	for actor, assignments := range s.assignments {
		for i, roleName := range assignments {
			if roleName == roleName {
				s.assignments[actor] = append(assignments[:i], assignments[i+1:]...)
				assignmentData := lager.Data{
					"actor.id":     actor.DomainID,
					"actor.issuer": actor.Issuer,
					"role.name":    roleName,
				}
				logger.Debug(messages.Success, assignmentData)
				break
			}
		}
	}
	// "Cascade"
	// Remove permissions for role
	s.permissions[roleName] = []*models.Permission{}

	logger.Debug(messages.Success)

	return nil
}

func (s *InMemoryStore) AssignRole(
	ctx context.Context,
	logger lager.Logger,
	roleName models.RoleName,
	domainID models.ActorDomainID,
	issuer models.ActorIssuer,
) error {
	if _, exists := s.roles[roleName]; !exists {
		return models.ErrRoleNotFound
	}
	actor := models.Actor{
		DomainID: domainID,
		Issuer:   issuer,
	}

	assignments, ok := s.assignments[actor]
	if !ok {
		assignments = []models.RoleName{}
	}

	for _, role := range assignments {
		if role == roleName {
			err := models.ErrRoleAssignmentAlreadyExists
			logger.Error(messages.ErrRoleAssignmentAlreadyExists, err)
			return err
		}
	}

	assignments = append(assignments, roleName)

	s.assignments[actor] = assignments
	return nil
}

func (s *InMemoryStore) UnassignRole(
	ctx context.Context,
	logger lager.Logger,
	roleName models.RoleName,
	domainID models.ActorDomainID,
	issuer models.ActorIssuer,
) error {
	if _, exists := s.roles[roleName]; !exists {
		return models.ErrRoleNotFound
	}

	actor := models.Actor{
		DomainID: domainID,
		Issuer:   issuer,
	}

	assignments, ok := s.assignments[actor]
	if !ok {
		err := models.ErrRoleAssignmentNotFound
		logger.Error(messages.ErrRoleAssignmentNotFound, err)
		return err
	}

	for i, assignment := range assignments {
		if assignment == roleName {
			s.assignments[actor] = append(assignments[:i], assignments[i+1:]...)
			logger.Debug(messages.Success)
			return nil
		}
	}

	err := models.ErrRoleAssignmentNotFound
	logger.Error(messages.ErrRoleAssignmentNotFound, err)

	return err
}

func (s *InMemoryStore) HasRole(
	ctx context.Context,
	logger lager.Logger,
	query repos.HasRoleQuery,
) (bool, error) {
	roleName := query.RoleName

	_, ok := s.roles[roleName]
	if !ok {
		return false, models.ErrRoleNotFound
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

func (s *InMemoryStore) ListActorRoles(
	ctx context.Context,
	logger lager.Logger,
	query repos.ListActorRolesQuery,
) ([]*models.Role, error) {
	actor := query.Actor

	var roles []*models.Role

	assignments, ok := s.assignments[actor]
	if !ok {
		return roles, nil
	}

	for _, name := range assignments {
		role, found := s.roles[name]
		if !found {
			return nil, models.ErrRoleNotFound
		}

		roles = append(roles, role)
	}

	return roles, nil
}

func (s *InMemoryStore) CreateActor(
	ctx context.Context,
	logger lager.Logger,
	domainID models.ActorDomainID,
	issuer models.ActorIssuer,
) (*models.Actor, error) {
	actor := models.Actor{
		DomainID: domainID,
		Issuer:   issuer,
	}

	if _, exists := s.assignments[actor]; exists {
		return nil, models.ErrActorAlreadyExists
	}

	s.assignments[actor] = []models.RoleName{}

	return &actor, nil
}

func (s *InMemoryStore) ListRolePermissions(
	ctx context.Context,
	logger lager.Logger,
	query repos.ListRolePermissionsQuery,
) ([]*models.Permission, error) {
	permissions, exists := s.permissions[query.RoleName]
	if !exists {
		return nil, models.ErrRoleNotFound
	}

	return permissions, nil
}
