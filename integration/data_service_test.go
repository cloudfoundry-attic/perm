package integration_test

import (
	. "code.cloudfoundry.org/perm/pkg/api/db"
	"code.cloudfoundry.org/perm/pkg/api/repos"
	"code.cloudfoundry.org/perm/pkg/sqlx/sqlxtest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/perm/pkg/api/repos/reposbehaviors"
	"code.cloudfoundry.org/perm/pkg/sqlx"
)

var _ = Describe("DataService", func() {
	var (
		testDB *sqlxtest.TestMySQLDB

		store *DataService

		conn *sqlx.DB
	)

	BeforeSuite(func() {
		var err error

		testDB = sqlxtest.NewTestMySQLDB()
		err = testDB.Create(Migrations...)
		Expect(err).NotTo(HaveOccurred())
	})

	BeforeEach(func() {
		var err error

		conn, err = testDB.Connect()
		Expect(err).NotTo(HaveOccurred())

		store = NewDataService(conn)
	})

	AfterEach(func() {
		Expect(conn.Close()).To(Succeed())

		err := testDB.Truncate(
			"DELETE FROM role",
			"DELETE FROM actor",
		)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterSuite(func() {
		err := testDB.Drop()
		Expect(err).NotTo(HaveOccurred())
	})

	BehavesLikeARoleRepo(func() repos.RoleRepo { return store })
	BehavesLikeARoleAssignmentRepo(func() repos.RoleAssignmentRepo { return store }, func() repos.RoleRepo { return store })
	BehavesLikeAPermissionRepo(func() repos.PermissionRepo { return store }, func() repos.RoleRepo { return store }, func() repos.RoleAssignmentRepo { return store })
})
