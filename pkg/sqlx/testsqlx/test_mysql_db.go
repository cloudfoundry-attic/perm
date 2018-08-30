package testsqlx

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/pkg/logx/lagerx"
	"code.cloudfoundry.org/perm/pkg/sqlx"
	uuid "github.com/satori/go.uuid"
)

const (
	TestMySQLHost     = "TEST_MYSQL_HOST"
	TestMySQLPort     = "TEST_MYSQL_PORT"
	TestMySQLDatabase = "TEST_MYSQL_DATABASE"
	TestMySQLUsername = "TEST_MYSQL_USERNAME"
	TestMySQLPassword = "TEST_MYSQL_PASSWORD"
)

type TestMySQLDB struct {
	options *options
}

func NewTestMySQLDB(opts ...TestDBOption) *TestMySQLDB {
	host := os.Getenv(TestMySQLHost)
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv(TestMySQLPort)
	if port == "" {
		port = "3306"
	}
	database := os.Getenv(TestMySQLDatabase)
	if database == "" {
		database = fmt.Sprintf("test_%s", strings.Replace(uuid.NewV4().String(), "-", "_", -1))
	}
	user := os.Getenv(TestMySQLUsername)
	if user == "" {
		user = "root"
	}
	password := os.Getenv(TestMySQLPassword)

	config := &options{
		host:     host,
		port:     port,
		database: database,
		user:     user,
		password: password,
	}

	for _, o := range opts {
		o(config)
	}

	return &TestMySQLDB{
		options: config,
	}
}

func (db *TestMySQLDB) Create(migrations ...sqlx.Migration) error {
	stmt := fmt.Sprintf("CREATE DATABASE %s", db.options.database)
	err := db.exec(stmt)
	if err != nil {
		return err
	}

	conn, err := db.Connect()
	if err != nil {
		_ = db.Drop()
		return err
	}
	defer conn.Close()

	logger := lagerx.NewLogger(lagertest.NewTestLogger("test-db"))
	err = sqlx.ApplyMigrations(context.Background(), logger, conn, "migrations", migrations)
	if err != nil {
		_ = db.Drop()
		return err
	}

	return nil
}

func (db *TestMySQLDB) Drop() error {
	stmt := fmt.Sprintf("DROP DATABASE %s", db.options.database)

	return db.exec(stmt)
}

func (db *TestMySQLDB) Connect() (*sqlx.DB, error) {
	iPort, err := strconv.ParseInt(db.options.port, 0, 0)
	if err != nil {
		return nil, err
	}

	dbOpts := []sqlx.DBOption{
		sqlx.DBUsername(db.options.user),
		sqlx.DBPassword(db.options.password),
		sqlx.DBHost(db.options.host),
		sqlx.DBPort(int(iPort)),
		sqlx.DBDatabaseName(db.options.database),
	}

	return sqlx.Connect(sqlx.DBDriverMySQL, dbOpts...)
}

func (db *TestMySQLDB) Truncate(truncateStmts ...string) error {
	conn, err := db.Connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	for _, stmt := range truncateStmts {
		if _, err = conn.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}

func (db *TestMySQLDB) exec(stmt string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		"mysql",
		"--user", db.options.user,
		fmt.Sprintf("--password=%s", db.options.password),
		"--port", db.options.port,
		"-e", stmt,
	)

	_, err := cmd.Output()

	return err
}
