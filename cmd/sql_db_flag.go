package cmd

import (
	"context"
	"net"
	"strconv"
	"time"

	"crypto/tls"
	"crypto/x509"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/sqlx"
	"github.com/go-sql-driver/mysql"
)

type SQLFlag struct {
	DB     DBFlag        `group:"DB" namespace:"db"`
	TLS    SQLTLSFlag    `group:"TLS" namespace:"tls"`
	Tuning SQLTuningFlag `group:"Tuning" namespace:"tuning"`
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

func (o *SQLFlag) Connect(ctx context.Context, logger lager.Logger, statter Statter, reader FileReader) (*sqlx.DB, error) {
	logger = logger.WithData(lager.Data{
		"db_driver":   o.DB.Driver,
		"db_host":     o.DB.Host,
		"db_port":     o.DB.Port,
		"db_schema":   o.DB.Schema,
		"db_username": o.DB.Username,
	})

	conn, err := o.open(ctx, logger.Session(openSQLConnection), statter, reader)
	if err != nil {
		return nil, err
	}

	conn.SetConnMaxLifetime(time.Duration(o.Tuning.ConnMaxLifetime) * time.Millisecond)

	err = ping(ctx, logger.Session(pingSQLConnection), conn)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (o *SQLFlag) open(ctx context.Context, logger lager.Logger, statter Statter, reader FileReader) (*sqlx.DB, error) {
	logger.Debug(starting)

	var (
		conn *sqlx.DB
		err  error
	)

	defer func() {
		if err != nil {
			logger.Error(failedToOpenSQLConnection, err)

		} else {
			logger.Debug(finished)
		}
	}()

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
					return nil, err
				}
				if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
					err = ErrFailedToAppendCertsFromPem
					return nil, err
				}
			}

			tlsConfigName := "perm"
			err = mysql.RegisterTLSConfig(tlsConfigName, &tls.Config{
				MinVersion: tls.VersionTLS12,
				RootCAs:    rootCertPool,
			})
			if err != nil {
				return nil, err
			}
			cfg.TLSConfig = tlsConfigName
		}

		conn, err = sqlx.Connect(context.Background(), o.DB.Driver, cfg.FormatDSN())

		if err != nil {
			return nil, err
		}

		return conn, nil
	default:
		err = ErrUnsupportedSQLDriver
		return nil, err
	}
}

func ping(ctx context.Context, logger lager.Logger, conn *sqlx.DB) error {
	logger.Debug(starting)

	var attempt int
	for {
		attempt++

		if attempt > 10 {
			err := NewAttemptError(10)
			logger.Error(failedToPingSQLConnection, err)
			return err
		}

		err := conn.PingContext(ctx)
		if err != nil {
			logger.Error(failedToPingSQLConnection, err, lager.Data{
				"attempt": attempt,
			})

			time.Sleep(1 * time.Second)
		} else {
			logger.Debug(finished)
			break
		}
	}

	return nil
}
