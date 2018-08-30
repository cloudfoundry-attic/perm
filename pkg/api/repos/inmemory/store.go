package inmemory

import "code.cloudfoundry.org/perm/pkg/perm"

type Store struct {
	roles       map[string]perm.Role
	permissions map[string][]perm.Permission

	assignments      map[perm.Actor][]string
	groupAssignments map[perm.Group][]string
}

func NewStore() *Store {
	return &Store{
		roles:            make(map[string]perm.Role),
		assignments:      make(map[perm.Actor][]string),
		groupAssignments: make(map[perm.Group][]string),
		permissions:      make(map[string][]perm.Permission),
	}
}
