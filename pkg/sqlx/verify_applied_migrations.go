package sqlx

import (
	"context"

	"code.cloudfoundry.org/lager"
)

func VerifyAppliedMigrations(
	ctx context.Context,
	logger lager.Logger,
	conn *DB,
	tableName string,
	migrations []Migration,
) error {
	retrieveLogger := logger.Session("retrieve-applied-migrations")
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
			logger.Info(migrationNotFound, lager.Data{
				"name": migration.Name,
			})
			return ErrMigrationsOutOfSync
		}

		if migration.Name != appliedMigration.Name {
			logger.Info(migrationMismatch, lager.Data{
				"expected_name": migration.Name,
				"applied_name":  appliedMigration.Name,
			})
			return ErrMigrationsOutOfSync
		}
	}

	logger.Debug(success)
	return nil
}
