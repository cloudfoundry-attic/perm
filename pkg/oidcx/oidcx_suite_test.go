package oidcx_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestOidcx(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Oidcx Suite")
}
