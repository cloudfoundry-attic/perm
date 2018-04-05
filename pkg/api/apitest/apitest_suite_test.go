package apitest_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAPITest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "APITest Suite")
}
