package sqlx

import (
	"context"

	"time"

	"code.cloudfoundry.org/perm/pkg/logx"
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

type MigrationFunc func(context.Context, logx.Logger, *Tx) error
