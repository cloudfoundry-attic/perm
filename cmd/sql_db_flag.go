package cmd

import (
	"context"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/cryptox"
	"code.cloudfoundry.org/perm/pkg/sqlx"
)

type SQLFlag struct {
	DB     DBFlag        `group:"DB" namespace:"db"`
	TLS    SQLTLSFlag    `group:"TLS" namespace:"tls"`
	Tuning SQLTuningFlag `group:"Tuning" namespace:"tuning"`
}

type DBFlag struct {
	Driver   sqlx.DBDriver `long:"driver" description:"Database driver to use for SQL backend (e.g. mysql, postgres)" required:"true"`
	Host     string        `long:"host" description:"Host for SQL backend" required:"true"`
	Port     int           `long:"port" description:"Port for SQL backend" required:"true"`
	Schema   string        `long:"schema" description:"Database name to use for connecting to SQL backend" required:"true"`
	Username string        `long:"username" description:"Username to use for connecting to SQL backend" required:"true"`
	Password string        `long:"password" description:"Password to use for connecting to SQL backend" required:"true"`
}

type SQLTLSFlag struct {
	Required bool               `long:"required" description:"Require TLS connections to the SQL backend"`
	RootCAs  []FileOrStringFlag `long:"root-ca" description:"CA certificate(s) for TLS connection to the SQL backend"`
}

type SQLTuningFlag struct {
	ConnMaxLifetime int `long:"connection-max-lifetime" description:"Limit the lifetime in milliseconds of a SQL connection"`
}

func (o *SQLFlag) Connect(ctx context.Context, logger lager.Logger, statter Statter, reader FileReader) (*sqlx.DB, error) {
	logger = logger.WithData(lager.Data{
		"db_driver":   o.DB.Driver,
		"db_host":     o.DB.Host,
		"db_port":     o.DB.Port,
		"db_schema":   o.DB.Schema,
		"db_username": o.DB.Username,
	})

	dbOpts := []sqlx.DBOption{
		sqlx.DBUsername(o.DB.Username),
		sqlx.DBPassword(o.DB.Password),
		sqlx.DBDatabaseName(o.DB.Schema),
		sqlx.DBHost(o.DB.Host),
		sqlx.DBPort(o.DB.Port),
		sqlx.DBConnectionMaxLifetime(time.Duration(o.Tuning.ConnMaxLifetime) * time.Millisecond),
	}

	if len(o.TLS.RootCAs) != 0 {
		tlsLogger := logger.Session("create-sql-root-ca-pool")

		var certs [][]byte
		for _, cert := range o.TLS.RootCAs {
			b, bErr := cert.Bytes(OS, IOReader)
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

	conn, err := sqlx.Connect(o.DB.Driver, dbOpts...)
	if err != nil {
		logger.Error(failedToOpenSQLConnection, err)
		return nil, err
	}

	return conn, nil
}
