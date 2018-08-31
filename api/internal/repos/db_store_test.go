package repos_test

import (
	"code.cloudfoundry.org/perm/api/internal/repos/db"
	"code.cloudfoundry.org/perm/internal/sqlx"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DBStore", func() {
	var (
		store *db.Store
		conn  *sqlx.DB
	)

	BeforeEach(func() {
		var err error

		conn, err = testDB.Connect()
		Expect(err).NotTo(HaveOccurred())

		store = db.NewStore(conn)
	})

	AfterEach(func() {
		Expect(conn.Close()).To(Succeed())

		err := testDB.Truncate(
			"DELETE FROM role",
			"DELETE FROM action",
		)
		Expect(err).NotTo(HaveOccurred())
	})

	testRepo(func() repo { return store })
})
