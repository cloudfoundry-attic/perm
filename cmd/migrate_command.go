package cmd

import (
	"context"
	"database/sql"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
)

type MigrateCommand struct {
	Logger LagerFlag

	MigrationsTableName string `long:"migrations-table-name" description:"Name of the table which holds migration information" default:"perm_db_migrations"`

	SQL SQLFlag `group:"SQL" namespace:"sql"`
}

func (cmd MigrateCommand) Execute([]string) error {
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

	logger = logger.Session("migrations")
	logger.Info(messages.Starting)

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(messages.ErrFailedToStartTransaction, err)
	}

	createTableLogger := logger.Session("create-migrations-table").WithData(lager.Data{
		"table_name": cmd.MigrationsTableName,
	})
	_, err = tx.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS `"+cmd.MigrationsTableName+"` (version INTEGER)")
	if err != nil {
		createTableLogger.Error("failed-to-create-table", err)
		err = tx.Rollback()
		if err != nil {
			createTableLogger.Error("failed-to-rollback", err)
			return err
		}
	} else {
		err = tx.Commit()
		if err != nil {
			createTableLogger.Error("failed-to-commit", err)
			return err
		}
	}

	for i, migration := range Migrations {
		migrationLogger := logger.Session("up").WithData(lager.Data{
			"Version": i,
			"Name":    migration.Name,
		})

		migrationLogger.Debug(messages.Starting)
		err = func(m Migration) (err error) {
			defer func() {
				if err != nil {
					tx.Rollback()
					return
				}
				err = tx.Commit()
			}()

			err = migration.Up(ctx, migrationLogger, tx)

			return
		}(migration)
		migrationLogger.Debug(messages.Finished)

		if err != nil {
			migrationLogger.Error(messages.ErrFailedToRunMigration, err)
			return err
		}

		migrationLogger.Debug(messages.Committed)
	}
	logger.Info(messages.Finished)

	return nil
}

var Migrations = []Migration{}

type Migration struct {
	Name string
	Up   func(context.Context, lager.Logger, *sql.Tx) error
	Down func(context.Context, lager.Logger, *sql.Tx) error
}
