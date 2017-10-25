package rpc_test

import (
	. "code.cloudfoundry.org/perm/rpc"

	"code.cloudfoundry.org/perm/models"
	"code.cloudfoundry.org/perm/models/modelsbehaviors"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("InMemoryStore", func() {
	modelsbehaviors.BehavesLikeARoleService(func() models.RoleService { return NewInMemoryStore() })
})
