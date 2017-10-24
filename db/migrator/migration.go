package migrator

import (
	"context"
	"database/sql"

	"time"

	"code.cloudfoundry.org/lager"
)

type Migration struct {
	Name string
	Up   MigrationFunc
	Down MigrationFunc
}

type AppliedMigration struct {
	Version   int
	Name      string
	AppliedAt time.Time
}

type MigrationFunc func(context.Context, lager.Logger, *sql.Tx) error
