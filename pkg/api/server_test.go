package api_test

import (
	"net"

	. "code.cloudfoundry.org/perm/pkg/api"
	"code.cloudfoundry.org/perm/pkg/api/db"
	"code.cloudfoundry.org/perm/pkg/sqlx"
	"code.cloudfoundry.org/perm/pkg/sqlx/sqlxtest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server", func() {
	var (
		testDB *sqlxtest.TestMySQLDB

		conn *sqlx.DB

		subject *Server
	)

	BeforeSuite(func() {
		var err error

		testDB = sqlxtest.NewTestMySQLDB()
		err = testDB.Create(db.Migrations...)
		Expect(err).NotTo(HaveOccurred())
	})

	BeforeEach(func() {
		var err error

		conn, err = testDB.Connect()
		Expect(err).NotTo(HaveOccurred())

		subject = NewServer(conn)
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

	Describe("#Serve", func() {
		It("fails if the server has already been stopped", func() {
			listener, err := net.Listen("tcp", "localhost:0")
			Expect(err).NotTo(HaveOccurred())

			defer listener.Close()

			go subject.Serve(listener)
			subject.Stop()

			err = subject.Serve(listener)
			Expect(err).To(MatchError("perm: the server has been stopped"))
		})

		It("fails when the listener is unable to accept connections", func() {
			listener, err := net.Listen("tcp", "localhost:0")
			Expect(err).NotTo(HaveOccurred())

			listener.Close()

			err = subject.Serve(listener)
			Expect(err).To(MatchError("perm: the server failed to start"))
		})
	})
})
