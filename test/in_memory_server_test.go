package perm_test

import (
	"code.cloudfoundry.org/perm/api"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("In-memory server", func() {
	testAPI(func() []api.ServerOption {
		return nil
	})
})
