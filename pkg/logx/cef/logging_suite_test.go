package cef_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCEF(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CEF Suite")
}
