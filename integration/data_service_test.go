package integration_test

import (
	. "code.cloudfoundry.org/perm/db"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"

	"os"

	"fmt"
	"strings"

	"strconv"

	"database/sql"

	"code.cloudfoundry.org/perm/cmd"
	. "code.cloudfoundry.org/perm/integration"
	"code.cloudfoundry.org/perm/models"
	. "code.cloudfoundry.org/perm/models/modelsbehaviors"
)

var _ = Describe("DataService", func() {
	var (
		flag        cmd.SQLFlag
		mySQLRunner *MySQLRunner

		store *DataService

		conn *sql.DB
	)

	BeforeSuite(func() {
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

		port, err := strconv.Atoi(p)
		Expect(err).NotTo(HaveOccurred())

		schema := os.Getenv("PERM_TEST_MYSQL_SCHEMA_NAME")
		if schema == "" {
			uuid := uuid.NewV4()
			schema = fmt.Sprintf("perm_test_%s", strings.Replace(uuid.String(), "-", "_", -1))
		}

		flag = cmd.SQLFlag{
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

		mySQLRunner.CreateTestDB()

	})

	BeforeEach(func() {
		var err error
		conn, err = flag.Open()
		Expect(err).NotTo(HaveOccurred())

		Expect(conn.Ping()).To(Succeed())

		store = NewDataService(conn)
	})

	AfterEach(func() {
		conn.Close()
		mySQLRunner.Truncate()
	})

	AfterSuite(func() {
		mySQLRunner.DropTestDB()
	})

	BehavesLikeARoleService(func() models.RoleService { return store })
	BehavesLikeAnActorService(func() models.ActorService { return store })
	BehavesLikeARoleAssignmentService(func() models.RoleAssignmentService { return store }, func() models.RoleService { return store }, func() models.ActorService { return store })
})