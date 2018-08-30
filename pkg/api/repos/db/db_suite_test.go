package db_test

import (
	"code.cloudfoundry.org/perm/internal/migrations"
	"code.cloudfoundry.org/perm/pkg/sqlx/testsqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDB(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DB Suite")
}

var testDB *testsqlx.TestMySQLDB

var _ = BeforeSuite(func() {
	var err error

	testDB = testsqlx.NewTestMySQLDB()
	err = testDB.Create(migrations.Migrations...)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := testDB.Drop()
	Expect(err).NotTo(HaveOccurred())
})
