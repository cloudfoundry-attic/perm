package flags

import (
	"context"
	"errors"
	"fmt"
	"time"

	"code.cloudfoundry.org/perm/internal/sqlx"
	"code.cloudfoundry.org/perm/pkg/cryptox"
	"code.cloudfoundry.org/perm/pkg/ioutilx"
	"code.cloudfoundry.org/perm/pkg/logx"
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

func (o *DBFlag) IsInMemory() bool {
	return o.Driver == "in-memory"
}

func (o *DBFlag) Connect(ctx context.Context, logger logx.Logger) (*sqlx.DB, error) {
	if o.IsInMemory() {
		return nil, errors.New("Connect() unsupported for in-memory driver")
	}

	if err := o.validate(); err != nil {
		return nil, err
	}

	data := []logx.Data{
		logx.Data{Key: "driver", Value: o.Driver},
		logx.Data{Key: "host", Value: o.Host},
		logx.Data{Key: "port", Value: o.Port},
		logx.Data{Key: "schema", Value: o.Schema},
		logx.Data{Key: "username", Value: o.Username},
	}
	logger = logger.WithData(data...)

	dbOpts := []sqlx.DBOption{
		sqlx.DBUsername(o.Username),
		sqlx.DBPassword(o.Password),
		sqlx.DBDatabaseName(o.Schema),
		sqlx.DBHost(o.Host),
		sqlx.DBPort(o.Port),
		sqlx.DBConnectionMaxLifetime(time.Duration(o.Tuning.ConnMaxLifetime) * time.Millisecond),
	}

	if len(o.TLS.RootCAs) != 0 {
		tlsLogger := logger.WithName("create-sql-root-ca-pool")

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

func (flag *DBFlag) validate() error {
	if flag.Host == "" {
		return &missingFlagError{param: "host"}
	}
	if flag.Port == 0 {
		return &missingFlagError{param: "port"}
	}
	if flag.Schema == "" {
		return &missingFlagError{param: "schema"}
	}
	if flag.Username == "" {
		return &missingFlagError{param: "username"}
	}
	return nil
}

type missingFlagError struct {
	param string
}

func (e *missingFlagError) Error() string {
	return fmt.Sprintf("the required %s parameter was not specified; see --help", e.param)
}
