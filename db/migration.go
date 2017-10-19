package db

import (
	"context"
	"database/sql"

	"code.cloudfoundry.org/lager"
)

type Migration struct {
	Name string
	Up   func(context.Context, lager.Logger, *sql.Tx) error
	Down func(context.Context, lager.Logger, *sql.Tx) error
}
