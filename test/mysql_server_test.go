package perm_test

import (
	"code.cloudfoundry.org/perm/internal/sqlx"
	"code.cloudfoundry.org/perm/pkg/api"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MySQL server", func() {
	var (
		conn *sqlx.DB
	)

	BeforeEach(func() {
		var err error
		conn, err = testMySQLDB.Connect()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := testMySQLDB.Truncate(
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
