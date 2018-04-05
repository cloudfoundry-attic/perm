package rpc

import "code.cloudfoundry.org/perm/pkg/perm"

type InMemoryStore struct {
	roles       map[string]*perm.Role
	permissions map[string][]*perm.Permission

	assignments map[perm.Actor][]string
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		roles:       make(map[string]*perm.Role),
		assignments: make(map[perm.Actor][]string),
		permissions: make(map[string][]*perm.Permission),
	}
}
