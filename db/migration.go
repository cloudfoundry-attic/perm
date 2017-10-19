package db

import (
	"context"
	"database/sql"

	"errors"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
)

func IrreversibleMigrationDown(context.Context, lager.Logger, *sql.Tx) error {
	return ErrIrreversibleMigration
}

var ErrIrreversibleMigration = errors.New(messages.ErrIrreversibleMigration)

type Migration struct {
	Name string
	Up   MigrationFunc
	Down MigrationFunc
}

type MigrationFunc func(context.Context, lager.Logger, *sql.Tx) error
