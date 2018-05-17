package permauth_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPermauth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Permauth Suite")
}
