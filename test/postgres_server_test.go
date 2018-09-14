package perm_test

import (
	"code.cloudfoundry.org/perm/api"
	"code.cloudfoundry.org/perm/internal/sqlx"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Postgres server", func() {
	var (
		conn *sqlx.DB
	)

	BeforeEach(func() {
		var err error
		conn, err = testPostgresDB.Connect()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := testPostgresDB.Truncate(
			"DELETE FROM role",
			"DELETE FROM action",
		)
		Expect(err).NotTo(HaveOccurred())

		err = conn.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	testAPI(func() []api.ServerOption {
		return []api.ServerOption{api.WithDBConn(conn)}
	})
})
