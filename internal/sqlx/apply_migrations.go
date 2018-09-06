package sqlx

import (
	"context"

	"time"

	"code.cloudfoundry.org/perm/logx"
	"github.com/Masterminds/squirrel"
)

func ApplyMigrations(
	ctx context.Context,
	logger logx.Logger,
	conn *DB,
	tableName string,
	migrations []Migration,
) error {
	createTableLogger := logger.WithName("create-migrations-table").WithData(logx.Data{Key: "table_name", Value: tableName})
	if err := createMigrationsTable(ctx, createTableLogger, conn, tableName); err != nil {
		return err
	}

	migrationsLogger := logger.WithName("apply-migrations").WithData(logx.Data{
		Key:   "table_name",
		Value: tableName,
	})

	if len(migrations) == 0 {
		return nil
	}

	appliedMigrations, err := RetrieveAppliedMigrations(ctx, migrationsLogger, conn, tableName)
	if err != nil {
		return err
	}
	migrationsLogger.Debug(retrievedAppliedMigrations, logx.Data{Key: "versions", Value: appliedMigrations})

	for i, migration := range migrations {
		version := i
		migrationLogger := logger.WithData(logx.Data{Key: "version", Value: version}, logx.Data{Key: "name", Value: migration.Name})

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
	logger logx.Logger,
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
		"` (id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY, version INTEGER, name VARCHAR(255), applied_at DATETIME)")

	return
}

func applyMigration(
	ctx context.Context,
	logger logx.Logger,
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
