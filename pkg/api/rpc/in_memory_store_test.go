package rpc_test

import (
	. "code.cloudfoundry.org/perm/pkg/api/rpc"

	"code.cloudfoundry.org/perm/pkg/api/repos"
	. "code.cloudfoundry.org/perm/pkg/api/repos/reposbehaviors"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("InMemoryStore", func() {
	var (
		store *InMemoryStore
	)
	BeforeEach(func() {
		store = NewInMemoryStore()
	})

	BehavesLikeARoleRepo(func() repos.RoleRepo { return store })
	BehavesLikeARoleAssignmentRepo(
		func() repos.RoleAssignmentRepo { return store },
		func() repos.RoleRepo { return store },
	)
	BehavesLikeAPermissionRepo(
		func() repos.PermissionRepo { return store },
		func() repos.RoleRepo { return store },
		func() repos.RoleAssignmentRepo { return store },
	)
})
