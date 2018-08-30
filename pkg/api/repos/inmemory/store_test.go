package inmemory_test

import (
	. "code.cloudfoundry.org/perm/pkg/api/repos/inmemory"

	"code.cloudfoundry.org/perm/pkg/api/repos"
	. "code.cloudfoundry.org/perm/pkg/api/repos/reposbehaviors"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("Store", func() {
	var (
		store *Store
	)

	BeforeEach(func() {
		store = NewStore()
	})

	BehavesLikeARoleRepo(func() repos.RoleRepo { return store })
	BehavesLikeAPermissionRepo(
		func() repos.PermissionRepo { return store },
		func() repos.RoleRepo { return store },
	)
})
