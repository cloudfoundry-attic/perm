package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/perm/pkg/api/db"
	"code.cloudfoundry.org/perm/pkg/api/repos"
	"code.cloudfoundry.org/perm/pkg/sqlx"

	. "code.cloudfoundry.org/perm/pkg/api/repos/reposbehaviors"
)

var _ = Describe("DataService", func() {
	var (
		store *db.DataService
		conn  *sqlx.DB
	)

	BeforeEach(func() {
		var err error

		conn, err = testDB.Connect()
		Expect(err).NotTo(HaveOccurred())

		store = db.NewDataService(conn)
	})

	AfterEach(func() {
		Expect(conn.Close()).To(Succeed())

		err := testDB.Truncate(
			"DELETE FROM role",
			"DELETE FROM action",
		)
		Expect(err).NotTo(HaveOccurred())
	})

	BehavesLikeARoleRepo(func() repos.RoleRepo { return store })
	BehavesLikeAPermissionRepo(func() repos.PermissionRepo { return store }, func() repos.RoleRepo { return store })
})
