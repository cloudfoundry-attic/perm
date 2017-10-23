package cmd

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/db"
	"code.cloudfoundry.org/perm/db/migrator"
	"code.cloudfoundry.org/perm/messages"
)

type MigrateCommand struct {
	Up   UpCommand   `command:"up" description:"Run migrations"`
	Down DownCommand `command:"down" description:"Revert migrations"`
}

type UpCommand struct {
	Logger LagerFlag

	SQL SQLFlag `group:"SQL" namespace:"sql"`
}

type DownCommand struct {
	Logger LagerFlag

	SQL SQLFlag `group:"SQL" namespace:"sql"`

	All bool `long:"all" description:"Revert all migrations"`
}

func (cmd UpCommand) Execute([]string) error {
	logger, _ := cmd.Logger.Logger("perm")
	logger = logger.Session("migrate-up")

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

	return migrator.ApplyMigrations(ctx, logger, conn, db.MigrationsTableName, db.Migrations)
}

func (cmd DownCommand) Execute([]string) error {
	logger, _ := cmd.Logger.Logger("perm")
	logger = logger.Session("migrate-down").WithData(lager.Data{
		"all": cmd.All,
	})

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

	return migrator.RollbackMigrations(ctx, logger, conn, db.MigrationsTableName, db.Migrations, cmd.All)
}
