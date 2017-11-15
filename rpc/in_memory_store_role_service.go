package rpc

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
	"code.cloudfoundry.org/perm/models"
)

func (s *InMemoryStore) CreateRole(ctx context.Context, logger lager.Logger, name string, permissions ...*models.Permission) (*models.Role, error) {
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

func (s *InMemoryStore) FindRole(ctx context.Context, logger lager.Logger, query models.RoleQuery) (*models.Role, error) {
	name := query.Name
	role, exists := s.roles[name]

	if !exists {
		return nil, models.ErrRoleNotFound
	}

	return role, nil
}

func (s *InMemoryStore) DeleteRole(ctx context.Context, logger lager.Logger, query models.RoleQuery) error {
	name := query.Name

	if _, exists := s.roles[name]; !exists {
		return models.ErrRoleNotFound
	}

	delete(s.roles, name)

	// "Cascade"
	// Remove role assignments for role
	for actor, assignments := range s.assignments {
		for i, roleName := range assignments {
			if roleName == name {
				s.assignments[actor] = append(assignments[:i], assignments[i+1:]...)
				assignmentData := lager.Data{
					"actor.id":     actor.DomainID,
					"actor.issuer": actor.Issuer,
					"role.name":    name,
				}
				logger.Debug(messages.Success, assignmentData)
				break
			}
		}
	}
	// "Cascade"
	// Remove permissions for role
	s.permissions[name] = []*models.Permission{}

	logger.Debug(messages.Success)

	return nil
}

func (s *InMemoryStore) AssignRole(ctx context.Context, logger lager.Logger, roleName string, domainID string, issuer string) error {
	if _, exists := s.roles[roleName]; !exists {
		return models.ErrRoleNotFound
	}
	actor := models.Actor{
		DomainID: domainID,
		Issuer:   issuer,
	}

	assignments, ok := s.assignments[actor]
	if !ok {
		assignments = []string{}
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

func (s *InMemoryStore) UnassignRole(ctx context.Context, logger lager.Logger, roleName string, domainID string, issuer string) error {
	if _, exists := s.roles[roleName]; !exists {
		return models.ErrRoleNotFound
	}

	actor := models.Actor{
		DomainID: domainID,
		Issuer:   issuer,
	}

	assignments, ok := s.assignments[actor]
	if !ok {
		return models.ErrActorNotFound
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

func (s *InMemoryStore) HasRole(ctx context.Context, logger lager.Logger, query models.RoleAssignmentQuery) (bool, error) {
	actor := models.Actor{
		DomainID: query.ActorQuery.DomainID,
		Issuer:   query.ActorQuery.Issuer,
	}

	roleName := query.RoleQuery.Name

	_, ok := s.roles[roleName]
	if !ok {
		return false, models.ErrRoleNotFound
	}

	assignments, ok := s.assignments[actor]
	if !ok {
		return false, models.ErrActorNotFound
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

func (s *InMemoryStore) ListActorRoles(ctx context.Context, logger lager.Logger, query models.ActorQuery) ([]*models.Role, error) {
	actor := models.Actor(query)

	assignments, ok := s.assignments[actor]
	if !ok {
		return nil, models.ErrActorNotFound
	}

	var roles []*models.Role

	for _, name := range assignments {
		role, found := s.roles[name]
		if !found {
			return nil, models.ErrRoleNotFound
		}

		roles = append(roles, role)
	}

	return roles, nil
}

func (s *InMemoryStore) CreateActor(ctx context.Context, logger lager.Logger, domainID, issuer string) (*models.Actor, error) {
	actor := models.Actor{
		DomainID: domainID,
		Issuer:   issuer,
	}

	if _, exists := s.assignments[actor]; exists {
		return nil, models.ErrActorAlreadyExists
	}

	s.assignments[actor] = []string{}

	return &actor, nil
}

func (s *InMemoryStore) FindActor(ctx context.Context, logger lager.Logger, query models.ActorQuery) (*models.Actor, error) {
	actor := models.Actor(query)

	if _, exists := s.assignments[actor]; !exists {
		return nil, models.ErrActorNotFound
	}

	return &actor, nil
}

func (s *InMemoryStore) ListRolePermissions(ctx context.Context, logger lager.Logger, query models.RoleQuery) ([]*models.Permission, error) {
	roleName := query.Name

	permissions, exists := s.permissions[roleName]
	if !exists {
		return nil, models.ErrRoleNotFound
	}

	return permissions, nil
}
