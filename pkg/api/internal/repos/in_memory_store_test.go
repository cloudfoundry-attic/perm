package repos_test

import (
	"code.cloudfoundry.org/perm/pkg/api/internal/repos/inmemory"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("InMemoryStore", func() {
	var (
		store *inmemory.Store
	)

	BeforeEach(func() {
		store = inmemory.NewStore()
	})

	testRepo(func() repo { return store })
})
