package perm_test

import (
	"code.cloudfoundry.org/perm/internal/migrations"
	"code.cloudfoundry.org/perm/pkg/sqlx/sqlxtest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	testMySQLDB *sqlxtest.TestMySQLDB
)

func TestPerm(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Perm Suite")
}

var _ = BeforeSuite(func() {
	var err error

	testMySQLDB = sqlxtest.NewTestMySQLDB()

	err = testMySQLDB.Create(migrations.Migrations...)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := testMySQLDB.Drop()
	Expect(err).NotTo(HaveOccurred())
})
