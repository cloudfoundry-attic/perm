package rpc_test

import (
	. "code.cloudfoundry.org/perm/rpc"

	"code.cloudfoundry.org/perm/models"
	. "code.cloudfoundry.org/perm/models/modelsbehaviors"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("InMemoryStore", func() {
	var (
		store *InMemoryStore
	)
	BeforeEach(func() {
		store = NewInMemoryStore()
	})

	BehavesLikeARoleService(func() models.RoleService { return store })
	BehavesLikeAnActorRepo(func() models.ActorRepo { return store })
	BehavesLikeARoleAssignmentRepo(
		func() models.RoleAssignmentRepo { return store },
		func() models.RoleService { return store },
		func() models.ActorRepo { return store },
	)
	BehavesLikeAPermissionRepo(
		func() models.PermissionRepo { return store },
		func() models.RoleService { return store },
		func() models.ActorRepo { return store },
		func() models.RoleAssignmentRepo { return store },
	)
})
