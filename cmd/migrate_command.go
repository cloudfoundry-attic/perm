package cmd

import (
	"context"

	"code.cloudfoundry.org/perm/db"
	"code.cloudfoundry.org/perm/messages"
)

type MigrateCommand struct {
	Up UpCommand `command:"up" description:"Run migrations"`
}

type UpCommand struct {
	Logger LagerFlag

	SQL SQLFlag `group:"SQL" namespace:"sql"`
}

func (cmd UpCommand) Execute([]string) error {
	logger, _ := cmd.Logger.Logger("perm")
	logger = logger.Session("migrate")

	ctx := context.Background()

	conn, err := cmd.SQL.Open()
	if err != nil {
		logger.Error(messages.ErrFailedToOpenSQLConnection, err)
		return err
	}

	pingLogger := logger.Session(messages.PingSQLConnection, cmd.SQL.LagerData())
	pingLogger.Debug(messages.Starting)
	err = conn.PingContext(ctx)
	if err != nil {
		pingLogger.Error(messages.ErrFailedToPingSQLConnection, err, cmd.SQL.LagerData())
		return err
	}
	pingLogger.Debug(messages.Finished)

	defer conn.Close()

	return db.ApplyMigrations(ctx, logger, conn, MigrationsTableName, db.Migrations)
}

var MigrationsTableName = "perm_migrations"
