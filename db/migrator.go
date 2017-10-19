package db

import (
	"context"
	"database/sql"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
	"github.com/Masterminds/squirrel"
)

func ApplyMigrations(ctx context.Context, logger lager.Logger, conn *sql.DB, tableName string, migrations []Migration) error {
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

	appliedMigrations, err := retrieveAppliedMigrations(ctx, migrationsLogger, conn, tableName)
	if err != nil {
		return err
	}
	migrationsLogger.Debug(messages.RetrievedAppliedMigrations, lager.Data{
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
			migrationLogger.Debug(messages.SkippedAppliedMigration)
		} else {
			err = applyMigration(ctx, migrationLogger, conn, tableName, version, migration)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func createMigrationsTable(ctx context.Context, logger lager.Logger, conn *sql.DB, tableName string) (err error) {
	var tx *sql.Tx
	tx, err = conn.BeginTx(ctx, nil)

	if err != nil {
		logger.Error(messages.FailedToStartTransaction, err)
		return
	}

	defer func() {
		if err != nil {
			logger.Error(messages.ErrFailedToCreateTable, err)
		}
		err = commit(logger, tx, err)
	}()

	_, err = tx.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS `"+tableName+"` (version INTEGER, name VARCHAR(255))")

	return
}

func retrieveAppliedMigrations(ctx context.Context, logger lager.Logger, conn *sql.DB, tableName string) (map[int]string, error) {
	rows, err := squirrel.Select("version", "name").
		From(tableName).
		RunWith(conn).
		QueryContext(ctx)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var (
		version int
		name string
	)

	versions := make(map[int]string)
	for rows.Next() {
		err = rows.Scan(&version, &name)
		if err != nil {
			return nil, err
		}
		versions[version] = name
	}

	err = rows.Err()
	if err != nil {
		logger.Error(messages.ErrFailedToQueryMigrations, err)
		return nil, err
	}

	return versions, nil
}

func applyMigration(ctx context.Context, logger lager.Logger, conn *sql.DB, tableName string, version int, migration Migration) (err error) {
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

	err = migration.Up(ctx, logger, tx)
	if err != nil {
		return
	}

	_, err = squirrel.Insert(tableName).
		Columns("version", "name").
		Values(version, migration.Name).
		RunWith(tx).
		ExecContext(ctx)

	logger.Debug(messages.Finished)

	return
}

func commit(logger lager.Logger, tx *sql.Tx, err error) error {
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			logger.Error(messages.FailedToRollback, rollbackErr)
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		logger.Error(messages.FailedToCommit, err)
		return err
	}

	logger.Debug(messages.Committed)
	return nil
}
