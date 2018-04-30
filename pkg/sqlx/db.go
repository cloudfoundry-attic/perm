package sqlx

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"net"
	"regexp"
	"strconv"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/go-sql-driver/mysql"
	uuid "github.com/satori/go.uuid"
)

type DBDriver string
type DBFlavor string

const (
	DBDriverMySQL DBDriver = "mysql"

	DBFlavorMySQL   = "mysql"
	DBFlavorMariaDB = "mariadb"
)

type DBOption interface {
	config(*dbConfig)
}

func DBUsername(username string) DBOption {
	return &dbUsernameOption{username: username}
}

func DBPassword(password string) DBOption {
	return &dbPasswordOption{password: password}
}

func DBDatabaseName(dbName string) DBOption {
	return &dbDatabaseNameOption{dbName: dbName}
}

func DBHost(host string) DBOption {
	return &dbHostOption{host: host}
}

func DBPort(port int) DBOption {
	return &dbPortOption{port: port}
}

func DBConnectionMaxLifetime(max time.Duration) DBOption {
	return &dbConnectionMaxLifetime{max: max}
}

func DBRootCAPool(rootCAPool *x509.CertPool) DBOption {
	return &dbTLSConfigOption{
		tlsConfig: &tls.Config{
			RootCAs:    rootCAPool,
			MinVersion: tls.VersionTLS12,
		},
	}
}

type DB struct {
	Conn *sql.DB

	driver  DBDriver
	flavor  DBFlavor
	version string
}

func Connect(driver DBDriver, options ...DBOption) (*DB, error) {
	cfg := &dbConfig{}

	for _, opt := range options {
		opt.config(cfg)
	}

	db, err := open(driver, cfg)
	if err != nil {
		return nil, err
	}

	db.Conn.SetConnMaxLifetime(cfg.connMaxLifetime)

	for attempt := 0; attempt < 10; attempt++ {
		err = db.Ping()
		if err == nil {
			return db, nil
		}
	}

	if err = db.Close(); err != nil {
		return nil, err
	}

	return nil, ErrFailedToEstablishConnection
}

func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.Conn.Exec(query, args...)
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.Conn.ExecContext(ctx, query, args...)
}

func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.Conn.Query(query, args...)
}

func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.Conn.QueryContext(ctx, query, args...)
}

func (db *DB) QueryRow(query string, args ...interface{}) squirrel.RowScanner {
	return db.Conn.QueryRow(query, args...)
}

func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) squirrel.RowScanner {
	return db.Conn.QueryRowContext(ctx, query, args...)
}

// BeginTx will generate a Database-aware transaction, with all database information
// duplicated from the Database-aware connection
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.Conn.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &Tx{
		tx:      tx,
		driver:  db.driver,
		flavor:  db.flavor,
		version: db.version,
	}, nil
}

func (tx *Tx) QueryRowContext(ctx context.Context, query string, args ...interface{}) squirrel.RowScanner {
	return tx.tx.QueryRowContext(ctx, query, args...)
}

func (tx *Tx) QueryRow(query string, args ...interface{}) squirrel.RowScanner {
	return tx.tx.QueryRow(query, args...)
}

func open(driver DBDriver, cfg *dbConfig) (*DB, error) {
	var (
		version string
		flavor  DBFlavor
	)

	switch driver {
	case DBDriverMySQL:
		dataSourceName, err := cfg.dataSourceNameMySQL()
		if err != nil {
			return nil, err
		}

		db, err := sql.Open(string(driver), dataSourceName)
		if err != nil {
			return nil, err
		}

		var (
			unused    string
			dbVersion string
		)

		err = db.QueryRow(`SHOW VARIABLES LIKE 'version'`).Scan(&unused, &dbVersion)
		// MySQL will error if the system table 'performance_schema.session_variables' doesn't exist
		if err == nil {
			mariadbVersionRegex := regexp.MustCompile("(.*)-MariaDB")

			matches := mariadbVersionRegex.FindStringSubmatch(dbVersion)
			if matches == nil {
				// Not MariaDB
				version = dbVersion
			} else {
				v := matches[1]

				flavor = DBFlavorMariaDB
				version = v
			}
		} else {
			flavor = DBFlavorMySQL
		}

		return &DB{
			Conn:    db,
			driver:  driver,
			flavor:  flavor,
			version: version,
		}, nil
	default:
		return nil, ErrUnsupportedSQLDriver
	}
}

func (db *DB) Close() error {
	return db.Conn.Close()
}

func (db *DB) Ping() error {
	return db.Conn.Ping()
}

type dbConfig struct {
	username string
	password string
	dbName   string
	host     string
	port     int

	tlsConfig *tls.Config

	connMaxLifetime time.Duration
}

func (c *dbConfig) dataSourceNameMySQL() (string, error) {
	cfg := mysql.NewConfig()
	cfg.User = c.username
	cfg.Passwd = c.password
	cfg.DBName = c.dbName
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort(c.host, strconv.Itoa(c.port))
	cfg.ParseTime = true

	if c.tlsConfig != nil {
		tlsConfigName := uuid.NewV4().String()
		if err := mysql.RegisterTLSConfig(tlsConfigName, c.tlsConfig); err != nil {
			return "", err
		}

		cfg.TLSConfig = tlsConfigName
	}

	return cfg.FormatDSN(), nil
}

type dbUsernameOption struct {
	username string
}

func (o *dbUsernameOption) config(c *dbConfig) {
	c.username = o.username
}

type dbPasswordOption struct {
	password string
}

func (o *dbPasswordOption) config(c *dbConfig) {
	c.password = o.password
}

type dbDatabaseNameOption struct {
	dbName string
}

func (o *dbDatabaseNameOption) config(c *dbConfig) {
	c.dbName = o.dbName
}

type dbHostOption struct {
	host string
}

func (o *dbHostOption) config(c *dbConfig) {
	c.host = o.host
}

type dbPortOption struct {
	port int
}

func (o *dbPortOption) config(c *dbConfig) {
	c.port = o.port
}

type dbTLSConfigOption struct {
	tlsConfig *tls.Config
}

func (o *dbTLSConfigOption) config(c *dbConfig) {
	c.tlsConfig = o.tlsConfig
}

type dbConnectionMaxLifetime struct {
	max time.Duration
}

func (o *dbConnectionMaxLifetime) config(c *dbConfig) {
	c.connMaxLifetime = o.max
}
