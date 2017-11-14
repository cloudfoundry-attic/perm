package rpc

import (
	"code.cloudfoundry.org/perm/models"
)

type InMemoryStore struct {
	roles       map[string]*models.Role
	permissions map[string][]*models.Permission

	assignments map[models.Actor][]string
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		roles:       make(map[string]*models.Role),
		assignments: make(map[models.Actor][]string),
		permissions: make(map[string][]*models.Permission),
	}
}
