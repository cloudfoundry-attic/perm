package testsqlx

import (
	"strconv"

	"code.cloudfoundry.org/perm/internal/sqlx"
)

type TestDB interface {
	Create(migrations ...sqlx.Migration) error
	Drop() error
	Connect() (*sqlx.DB, error)
	Truncate(truncateStmts ...string) error
}

type TestDBOption func(*options)

func DBHost(host string) TestDBOption {
	return func(o *options) {
		o.host = host
	}
}

func DBPort(port int) TestDBOption {
	return func(o *options) {
		o.port = strconv.Itoa(port)
	}
}

func DBDatabase(database string) TestDBOption {
	return func(o *options) {
		o.database = database
	}
}

func DBUser(user string) TestDBOption {
	return func(o *options) {
		o.user = user
	}
}

func DBPassword(password string) TestDBOption {
	return func(o *options) {
		o.password = password
	}
}

type options struct {
	host     string
	port     string
	database string
	user     string
	password string
}
