package sqlx

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
)

func VerifyAppliedMigrations(ctx context.Context, logger lager.Logger, conn *DB, tableName string, migrations []Migration) (bool, error) {
	appliedMigrations, err := RetrieveAppliedMigrations(ctx, logger.Session("retrieve-applied-migrations"), conn, tableName)
	if err != nil {
		return false, err
	}

	if len(migrations) != len(appliedMigrations) {
		logger.Info(messages.MigrationCountMismatch)
		return false, nil
	}

	for i, migration := range migrations {
		appliedMigration, exists := appliedMigrations[i]
		if !exists {
			logger.Info(messages.MigrationNotFound, lager.Data{
				"name": migration.Name,
			})
			return false, nil
		}

		if migration.Name != appliedMigration.Name {
			logger.Info(messages.MigrationMismatch, lager.Data{
				"expected_name": migration.Name,
				"applied_name":  appliedMigration.Name,
			})
			return false, nil
		}
	}

	logger.Debug(messages.Success)
	return true, nil
}
