package perm_test

import (
	"code.cloudfoundry.org/perm/pkg/api"
	"code.cloudfoundry.org/perm/pkg/permstats"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("In-memory server", func() {
	testAPI(func() []api.ServerOption {
		return []api.ServerOption{
			api.WithStats(&permstats.Handler{}),
		}
	})
})
