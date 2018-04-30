package flags

import (
	"context"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/cryptox"
	"code.cloudfoundry.org/perm/pkg/ioutilx"
	"code.cloudfoundry.org/perm/pkg/sqlx"
)

type DBFlag struct {
	Driver   sqlx.DBDriver `long:"driver" description:"Database driver to use for SQL backend (e.g. mysql, postgres, in-memory)" required:"true"`
	Host     string        `long:"host" description:"Host for SQL backend"`
	Port     int           `long:"port" description:"Port for SQL backend"`
	Schema   string        `long:"schema" description:"Database name to use for connecting to SQL backend"`
	Username string        `long:"username" description:"Username to use for connecting to SQL backend"`
	Password string        `long:"password" description:"Password to use for connecting to SQL backend"`

	TLS    SQLTLSFlag    `group:"TLS" namespace:"tls"`
	Tuning SQLTuningFlag `group:"Tuning" namespace:"tuning"`
}

type SQLTLSFlag struct {
	Required bool                   `long:"required" description:"Require TLS connections to the SQL backend"`
	RootCAs  []ioutilx.FileOrString `long:"root-ca" description:"CA certificate(s) for TLS connection to the SQL backend"`
}

type SQLTuningFlag struct {
	ConnMaxLifetime int `long:"connection-max-lifetime" description:"Limit the lifetime in milliseconds of a SQL connection"`
}

func (o *DBFlag) Connect(ctx context.Context, logger lager.Logger) (*sqlx.DB, error) {
	logger = logger.WithData(lager.Data{
		"db_driver":   o.Driver,
		"db_host":     o.Host,
		"db_port":     o.Port,
		"db_schema":   o.Schema,
		"db_username": o.Username,
	})

	dbOpts := []sqlx.DBOption{
		sqlx.DBUsername(o.Username),
		sqlx.DBPassword(o.Password),
		sqlx.DBDatabaseName(o.Schema),
		sqlx.DBHost(o.Host),
		sqlx.DBPort(o.Port),
		sqlx.DBConnectionMaxLifetime(time.Duration(o.Tuning.ConnMaxLifetime) * time.Millisecond),
	}

	if len(o.TLS.RootCAs) != 0 {
		tlsLogger := logger.Session("create-sql-root-ca-pool")

		var certs [][]byte
		for _, cert := range o.TLS.RootCAs {
			b, bErr := cert.Bytes(ioutilx.OS, ioutilx.IOReader)
			if bErr != nil {
				tlsLogger.Error(failedToReadFile, bErr)
				return nil, bErr
			}

			certs = append(certs, b)
		}

		rootCAPool, err := cryptox.NewCertPool(certs...)
		if err != nil {
			tlsLogger.Error(failedToParseTLSCredentials, err)
			return nil, err
		}

		dbOpts = append(dbOpts, sqlx.DBRootCAPool(rootCAPool))
	}

	conn, err := sqlx.Connect(o.Driver, dbOpts...)
	if err != nil {
		logger.Error(failedToOpenSQLConnection, err)
		return nil, err
	}

	return conn, nil
}
