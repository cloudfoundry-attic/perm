package sqlx

import (
	"context"

	"code.cloudfoundry.org/perm/logx"
)

func VerifyAppliedMigrations(
	ctx context.Context,
	logger logx.Logger,
	conn *DB,
	tableName string,
	migrations []Migration,
) error {
	retrieveLogger := logger.WithName("retrieve-applied-migrations")
	appliedMigrations, err := RetrieveAppliedMigrations(ctx, retrieveLogger, conn, tableName)
	if err != nil {
		return err
	}

	if len(migrations) != len(appliedMigrations) {
		logger.Info(migrationCountMismatch)
		return ErrMigrationsOutOfSync
	}

	for i, migration := range migrations {
		appliedMigration, exists := appliedMigrations[i]
		if !exists {
			logger.Info(migrationNotFound, logx.Data{Key: "name", Value: migration.Name})
			return ErrMigrationsOutOfSync
		}

		if migration.Name != appliedMigration.Name {
			logger.Info(migrationMismatch, logx.Data{Key: "expected_name", Value: migration.Name}, logx.Data{Key: "applied_name", Value: appliedMigration.Name})
			return ErrMigrationsOutOfSync
		}
	}

	logger.Debug(success)
	return nil
}
