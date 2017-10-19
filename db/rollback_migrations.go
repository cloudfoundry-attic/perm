package db

import (
	"context"
	"database/sql"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
	"github.com/Masterminds/squirrel"
)

func RollbackMigrations(ctx context.Context, logger lager.Logger, conn *sql.DB, tableName string, migrations []Migration, all bool) error {
	migrationsLogger := logger.Session("rollback-migrations").WithData(lager.Data{
		"table_name": tableName,
	})

	migrationsLogger.Info("starting")
	if len(migrations) == 0 {
		return nil
	}

	appliedMigrations, err := retrieveAppliedMigrations(ctx, migrationsLogger, conn, tableName)
	if err != nil {
		return err
	}
	migrationsLogger.Debug(messages.RetrievedAppliedMigrations, lager.Data{
		"versions": appliedMigrations,
	})

	for version := len(migrations) - 1; version >= 0; version-- {
		migration := migrations[version]
		_, ok := appliedMigrations[version]

		migrationLogger := logger.WithData(lager.Data{
			"version": version,
			"name":    migration.Name,
		})

		if !ok {
			migrationLogger.Debug("skipping")
			continue
		}

		err := rollbackMigration(ctx, migrationLogger, conn, tableName, version, migration)
		if err != nil {
			return err
		}
		if !all {
			return nil
		}
	}

	return nil
}

func rollbackMigration(ctx context.Context, logger lager.Logger, conn *sql.DB, tableName string, version int, migration Migration) (err error) {
	logger.Debug(messages.Starting)

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(messages.FailedToStartTransaction, err)
		return
	}

	defer func() {
		if err != nil {
			logger.Error(messages.ErrFailedToApplyMigration, err)
		}
		err = commit(logger, tx, err)
	}()

	err = migration.Down(ctx, logger, tx)
	if err != nil {
		return
	}

	_, err = squirrel.Delete(tableName).
		Where(squirrel.Eq{"version": version}).
		RunWith(tx).
		ExecContext(ctx)

	logger.Debug(messages.Finished)

	return
}
