package sqlx

import (
	"context"

	"time"

	"code.cloudfoundry.org/lager"
	"github.com/Masterminds/squirrel"
)

func ApplyMigrations(
	ctx context.Context,
	logger lager.Logger,
	conn *DB,
	tableName string,
	migrations []Migration,
) error {
	createTableLogger := logger.Session("create-migrations-table").WithData(lager.Data{
		"table_name": tableName,
	})
	if err := createMigrationsTable(ctx, createTableLogger, conn, tableName); err != nil {
		return err
	}

	migrationsLogger := logger.Session("apply-migrations").WithData(lager.Data{
		"table_name": tableName,
	})

	if len(migrations) == 0 {
		return nil
	}

	appliedMigrations, err := RetrieveAppliedMigrations(ctx, migrationsLogger, conn, tableName)
	if err != nil {
		return err
	}
	migrationsLogger.Debug(retrievedAppliedMigrations, lager.Data{
		"versions": appliedMigrations,
	})

	for i, migration := range migrations {
		version := i
		migrationLogger := logger.WithData(lager.Data{
			"version": version,
			"name":    migration.Name,
		})

		_, ok := appliedMigrations[version]
		if ok {
			migrationLogger.Debug(skippedAppliedMigration)
		} else {
			err = applyMigration(ctx, migrationLogger, conn, tableName, version, migration)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func createMigrationsTable(
	ctx context.Context,
	logger lager.Logger,
	conn *DB, tableName string,
) (err error) {
	var tx *Tx
	tx, err = conn.BeginTx(ctx, nil)

	if err != nil {
		logger.Error(failedToStartTransaction, err)
		return
	}

	defer func() {
		if err != nil {
			logger.Error(failedToCreateTable, err)
		}
		err = Commit(logger, tx, err)
	}()

	_, err = tx.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS `"+tableName+
		"` (version INTEGER, name VARCHAR(255), applied_at DATETIME)")

	return
}

func applyMigration(
	ctx context.Context,
	logger lager.Logger,
	conn *DB,
	tableName string,
	version int,
	migration Migration,
) (err error) {
	logger.Debug(starting)

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(failedToStartTransaction, err)
		return
	}

	defer func() {
		if err != nil {
			logger.Error(failedToApplyMigration, err)
		}
		err = Commit(logger, tx, err)
	}()

	err = migration.Up(ctx, logger, tx)
	if err != nil {
		return
	}

	_, err = squirrel.Insert(tableName).
		Columns("version", "name", "applied_at").
		Values(version, migration.Name, time.Now()).
		RunWith(tx).
		ExecContext(ctx)

	logger.Debug(finished)

	return
}
