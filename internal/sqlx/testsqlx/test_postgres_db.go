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
	"code.cloudfoundry.org/perm/internal/sqlx"
	"code.cloudfoundry.org/perm/logx/lagerx"
	uuid "github.com/satori/go.uuid"
)

const (
	bootstrapDB          = "postgres"
	TestPostgresHost     = "TEST_POSTGRES_HOST"
	TestPostgresPort     = "TEST_POSTGRES_PORT"
	TestPostgresDatabase = "TEST_POSTGRES_DATABASE"
	TestPostgresUsername = "TEST_POSTGRES_USERNAME"
	TestPostgresPassword = "TEST_POSTGRES_PASSWORD"
)

type TestPostgresDB struct {
	options *options
}

func NewTestPostgresDB(opts ...TestDBOption) *TestPostgresDB {
	host := os.Getenv(TestPostgresHost)
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv(TestPostgresPort)
	if port == "" {
		port = "5432"
	}
	database := os.Getenv(TestPostgresDatabase)
	if database == "" {
		database = fmt.Sprintf("test_%s", strings.Replace(uuid.NewV4().String(), "-", "_", -1))
	}
	user := os.Getenv(TestPostgresUsername)
	password := os.Getenv(TestPostgresPassword)

	config := &options{
		bootstrapDB: bootstrapDB,
		host:        host,
		port:        port,
		database:    database,
		user:        user,
		password:    password,
	}

	for _, o := range opts {
		o(config)
	}

	return &TestPostgresDB{
		options: config,
	}
}

func (db *TestPostgresDB) Create(migrations ...sqlx.Migration) error {
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

func (db *TestPostgresDB) Drop() error {
	stmt := fmt.Sprintf("DROP DATABASE %s", db.options.database)

	return db.exec(stmt)
}

func (db *TestPostgresDB) Connect() (*sqlx.DB, error) {
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

	return sqlx.Connect(sqlx.DBDriverPostgres, dbOpts...)
}

func (db *TestPostgresDB) Truncate(truncateStmts ...string) error {
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

func (db *TestPostgresDB) exec(stmt string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		"psql",
		"-d", db.options.bootstrapDB,
		"--user", db.options.user,
		"--port", db.options.port,
		"-c", stmt,
	)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", db.options.password))

	_, err := cmd.Output()

	return err
}
