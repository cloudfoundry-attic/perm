package cmd

import (
	"context"
	"errors"
	"net"
	"strconv"

	"crypto/tls"
	"crypto/x509"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/sqlx"
	"github.com/go-sql-driver/mysql"
)

type SQLFlag struct {
	DB  DBFlag     `group:"DB" namespace:"db"`
	TLS SQLTLSFlag `group:"TLS" namespace:"tls"`
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

func (o *SQLFlag) Open(statter Statter, reader FileReader) (*sqlx.DB, error) {
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
				pem, err := rootCA.Bytes(statter, reader)
				if err != nil {
					return nil, err
				}
				if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
					return nil, FailedToAppendCertsFromPem
				}
			}

			tlsConfigName := "perm"
			mysql.RegisterTLSConfig(tlsConfigName, &tls.Config{
				MinVersion: tls.VersionTLS12,
				RootCAs:    rootCertPool,
			})
			cfg.TLSConfig = tlsConfigName
		}

		return sqlx.Connect(context.Background(), o.DB.Driver, cfg.FormatDSN())
	default:
		return nil, errors.New("unsupported sql driver")
	}
}

func (o *SQLFlag) LagerData() lager.Data {
	return lager.Data{
		"db_driver":   o.DB.Driver,
		"db_host":     o.DB.Host,
		"db_port":     o.DB.Port,
		"db_schema":   o.DB.Schema,
		"db_username": o.DB.Username,
	}
}
