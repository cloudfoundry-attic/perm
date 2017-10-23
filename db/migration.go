package db

import (
	"context"
	"database/sql"

	"code.cloudfoundry.org/lager"
)

type Migration struct {
	Name string
	Up   MigrationFunc
	Down MigrationFunc
}

type MigrationFunc func(context.Context, lager.Logger, *sql.Tx) error
