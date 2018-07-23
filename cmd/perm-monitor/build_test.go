package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBuild(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Perm Binary Suite")
}

// Having this file here will cause Ginkgo to compile the perm binary. This
// will catch compilation errors at test time. We add a Ginkgo test suite above
// to stop Ginkgo complaining about an empty suite.
