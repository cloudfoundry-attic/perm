package integration_test

import (
	. "code.cloudfoundry.org/perm/db"

	"database/sql"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"

	"os"

	"fmt"
	"strings"

	"strconv"

	"code.cloudfoundry.org/perm/cmd"
	. "code.cloudfoundry.org/perm/integration"
)

var _ = Describe("DataService", func() {
	var (
		conn *sql.DB

		mySQLRunner *MySQLRunner

		store *DataService
	)

	BeforeEach(func() {
		var err error
		conn, _, err = sqlmock.New()

		Expect(err).NotTo(HaveOccurred())

		driver := "mysql"
		hostname := "localhost"

		username := os.Getenv("PERM_TEST_MYSQL_USERNAME")
		if username == "" {
			username = "root"
		}

		password := os.Getenv("PERM_TEST_MYSQL_PASSWORD")
		if password == "" {
			password = "password"
		}

		p := os.Getenv("PERM_TEST_MYSQL_PORT")
		if p == "" {
			p = "3306"
		}

		var port int
		port, err = strconv.Atoi(p)
		Expect(err).NotTo(HaveOccurred())

		schema := os.Getenv("PERM_TEST_MYSQL_SCHEMA_NAME")
		if schema == "" {
			uuid := uuid.NewV4()
			schema = fmt.Sprintf("perm_test_%s", strings.Replace(uuid.String(), "-", "_", -1))
		}

		flag := cmd.SQLFlag{
			DB: cmd.DBFlag{
				Driver:   driver,
				Host:     hostname,
				Port:     port,
				Username: username,
				Password: password,
				Schema:   schema,
			},
		}

		mySQLRunner = NewRunner(flag)

		conn, err = flag.Open()
		Expect(err).NotTo(HaveOccurred())

		mySQLRunner.CreateTestDB()

		Expect(conn.Ping()).To(Succeed())

		store = NewDataService(conn)
	})

	AfterEach(func() {
		mySQLRunner.DropTestDB()
	})
})
