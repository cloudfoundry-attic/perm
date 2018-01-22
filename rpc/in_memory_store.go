package rpc

import (
	"code.cloudfoundry.org/perm/models"
)

type InMemoryStore struct {
	roles       map[models.RoleName]*models.Role
	permissions map[models.RoleName][]*models.Permission

	assignments map[models.Actor][]models.RoleName
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		roles:       make(map[models.RoleName]*models.Role),
		assignments: make(map[models.Actor][]models.RoleName),
		permissions: make(map[models.RoleName][]*models.Permission),
	}
}
