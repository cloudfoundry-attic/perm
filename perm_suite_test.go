package perm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPerm(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Perm Suite")
}
