package cmd

import (
	"context"
	"net"
	"strconv"
	"time"

	"crypto/tls"
	"crypto/x509"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
	"code.cloudfoundry.org/perm/sqlx"
	"github.com/go-sql-driver/mysql"
)

type SQLFlag struct {
	DB     DBFlag        `group:"DB" namespace:"db"`
	TLS    SQLTLSFlag    `group:"TLS" namespace:"tls"`
	Tuning SQLTuningFlag `group:"tuning" namespace:"tuning"`
}

type DBFlag struct {
	Driver   sqlx.DBDriverName `long:"driver" description:"Database driver to use for SQL backend (e.g. mysql, postgres)" required:"true"`
	Host     string            `long:"host" description:"Host for SQL backend" required:"true"`
	Port     int               `long:"port" description:"Port for SQL backend" required:"true"`
	Schema   string            `long:"schema" description:"Database name to use for connecting to SQL backend" required:"true"`
	Username string            `long:"username" description:"Username to use for connecting to SQL backend" required:"true"`
	Password string            `long:"password" description:"Password to use for connecting to SQL backend" required:"true"`
}

type SQLTLSFlag struct {
	Required bool               `long:"required" description:"Require TLS connections to the SQL backend"`
	RootCAs  []FileOrStringFlag `long:"root-ca" description:"CA certificate(s) for TLS connection to the SQL backend"`
}

type SQLTuningFlag struct {
	ConnMaxLifetime int `long:"connection-max-lifetime" description:"Limit the lifetime in milliseconds of a SQL connection"`
}

func (o *SQLFlag) Open(ctx context.Context, logger lager.Logger, statter Statter, reader FileReader) (*sqlx.DB, error) {
	logger = logger.WithData(lager.Data{
		"db_driver":   o.DB.Driver,
		"db_host":     o.DB.Host,
		"db_port":     o.DB.Port,
		"db_schema":   o.DB.Schema,
		"db_username": o.DB.Username,
	})

	openLogger := logger.Session(messages.OpenSQLConnection)
	openLogger.Debug(messages.Starting)

	var (
		conn *sqlx.DB
		err  error
	)

	switch o.DB.Driver {
	case "mysql":
		cfg := mysql.NewConfig()

		cfg.User = o.DB.Username
		cfg.Passwd = o.DB.Password
		cfg.Net = "tcp"
		cfg.Addr = net.JoinHostPort(o.DB.Host, strconv.Itoa(o.DB.Port))
		cfg.DBName = o.DB.Schema
		cfg.ParseTime = true

		if o.TLS.Required {
			rootCertPool := x509.NewCertPool()
			for _, rootCA := range o.TLS.RootCAs {
				var pem []byte
				pem, err = rootCA.Bytes(statter, reader)
				if err != nil {
					openLogger.Error(messages.FailedToOpenSQLConnection, err)
					return nil, err
				}
				if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
					err = ErrFailedToAppendCertsFromPem
					openLogger.Error(messages.FailedToOpenSQLConnection, err)
					return nil, err
				}
			}

			tlsConfigName := "perm"
			err = mysql.RegisterTLSConfig(tlsConfigName, &tls.Config{
				MinVersion: tls.VersionTLS12,
				RootCAs:    rootCertPool,
			})
			if err != nil {
				openLogger.Error(messages.FailedToOpenSQLConnection, err)
				return nil, err
			}
			cfg.TLSConfig = tlsConfigName
		}

		conn, err = sqlx.Connect(context.Background(), o.DB.Driver, cfg.FormatDSN())
		if err != nil {
			openLogger.Error(messages.FailedToOpenSQLConnection, err)
			return nil, err
		}

	default:
		err = ErrUnsupportedSQLDriver
		openLogger.Error(messages.FailedToOpenSQLConnection, err)
		return nil, err
	}

	conn.SetConnMaxLifetime(time.Duration(o.Tuning.ConnMaxLifetime) * time.Millisecond)

	pingLogger := logger.Session(messages.PingSQLConnection)
	pingLogger.Debug(messages.Starting)

	var attempt int
	for {
		attempt++

		if attempt > 10 {
			err = NewAttemptError(10)
			pingLogger.Error(messages.FailedToPingSQLConnection, err)
			return nil, err
		}

		err = conn.PingContext(ctx)
		if err != nil {
			pingLogger.Error(messages.FailedToPingSQLConnection, err, lager.Data{
				"attempt": attempt,
			})

			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}

	pingLogger.Debug(messages.Finished)
	openLogger.Debug(messages.Finished)

	return conn, err
}
