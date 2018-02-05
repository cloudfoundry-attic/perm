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
	BehavesLikeAnActorService(func() models.ActorService { return store })
	BehavesLikeARoleAssignmentService(
		func() models.RoleAssignmentService { return store },
		func() models.RoleService { return store },
		func() models.ActorService { return store },
	)
	BehavesLikeAPermissionRepo(
		func() models.PermissionRepo { return store },
		func() models.RoleService { return store },
		func() models.ActorService { return store },
		func() models.RoleAssignmentService { return store },
	)
})
