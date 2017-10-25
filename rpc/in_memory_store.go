package rpc

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/models"
)

type InMemoryStore struct {
	roles map[string]*models.Role
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		roles: make(map[string]*models.Role),
	}
}

func (s *InMemoryStore) CreateRole(ctx context.Context, logger lager.Logger, name string) (*models.Role, error) {
	if _, exists := s.roles[name]; exists {
		return nil, models.ErrRoleAlreadyExists
	}

	role := &models.Role{
		Name: name,
	}
	s.roles[name] = role
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

	return nil
}
