package sqlx

import (
	"context"
	"database/sql"
	"regexp"

	"github.com/Masterminds/squirrel"
)

type DB struct {
	*sql.DB

	DriverName DBDriverName
	Type       DBType
	Flavor     DBFlavor
	Version    string
}

type Tx struct {
	*sql.Tx

	Type    DBType
	Flavor  DBFlavor
	Version string
}

type DBDriverName string

type DBType string

type DBFlavor string

const (
	DBDriverNameMySQL DBDriverName = "mysql"

	DBTypeMySQL DBType = "mysql"

	DBFlavorMariaDBMySQL DBFlavor = "mariadb"
)

func Connect(ctx context.Context, driverName DBDriverName, dataSourceName string) (*DB, error) {
	db, err := sql.Open(string(driverName), dataSourceName)
	if err != nil {
		return nil, err
	}

	var (
		ty      DBType
		flavor  DBFlavor
		version string
	)

	switch driverName {
	case DBDriverNameMySQL:
		ty = DBTypeMySQL

		var (
			variableName string
			dbVersion    string
		)

		err := db.QueryRowContext(ctx, `SHOW VARIABLES LIKE 'version'`).Scan(&variableName, &dbVersion)
		if err != nil {
			return nil, err
		}

		mariadbVersionRegex := regexp.MustCompile("(.*)-MariaDB")

		matches := mariadbVersionRegex.FindStringSubmatch(dbVersion)
		if matches == nil {
			// Not MariaDB
			version = dbVersion
		} else {
			v := matches[1]

			flavor = DBFlavorMariaDBMySQL
			version = v
		}
	}

	return &DB{
		DriverName: driverName,
		DB:         db,
		Type:       ty,
		Flavor:     flavor,
		Version:    version,
	}, nil
}

func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) squirrel.RowScanner {
	return db.DB.QueryRowContext(ctx, query, args...)
}

func (db *DB) QueryRow(query string, args ...interface{}) squirrel.RowScanner {
	return db.DB.QueryRow(query, args...)
}

// BeginTx will generate a Database-aware transaction, with all database information
// duplicated from the Database-aware connection
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &Tx{
		Tx:      tx,
		Type:    db.Type,
		Flavor:  db.Flavor,
		Version: db.Version,
	}, nil
}

func (tx *Tx) QueryRowContext(ctx context.Context, query string, args ...interface{}) squirrel.RowScanner {
	return tx.Tx.QueryRowContext(ctx, query, args...)
}

func (tx *Tx) QueryRow(query string, args ...interface{}) squirrel.RowScanner {
	return tx.Tx.QueryRow(query, args...)
}
