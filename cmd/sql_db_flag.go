package cmd

import (
	"database/sql"
	"errors"
	"net"
	"strconv"

	"code.cloudfoundry.org/lager"
	"github.com/go-sql-driver/mysql"
)

type SQLFlag struct {
	DB DBFlag `group:"DB" namespace:"db"`
}

type DBFlag struct {
	Driver   string `long:"driver" description:"Database driver to use for SQL backend (e.g. mysql, postgres)" required:"true"`
	Host     string `long:"host" description:"Host for SQL backend" required:"true"`
	Port     int    `long:"port" description:"Port for SQL backend" required:"true"`
	Schema   string `long:"schema" description:"Database name to use for connecting to SQL backend" required:"true"`
	Username string `long:"username" description:"Username to use for connecting to SQL backend" required:"true"`
	Password string `long:"password" description:"Password to use for connecting to SQL backend" required:"true"`
}

func (o *SQLFlag) Open() (*sql.DB, error) {
	switch o.DB.Driver {
	case "mysql":
		cfg := mysql.NewConfig()

		cfg.User = o.DB.Username
		cfg.Passwd = o.DB.Password
		cfg.Net = "tcp"
		cfg.Addr = net.JoinHostPort(o.DB.Host, strconv.Itoa(o.DB.Port))
		cfg.DBName = o.DB.Schema
		cfg.ParseTime = true

		return sql.Open(o.DB.Driver, cfg.FormatDSN())
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
