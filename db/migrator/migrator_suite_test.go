package migrator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMigrator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Migrator Suite")
}
