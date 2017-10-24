package migrator

import (
	"context"
	"database/sql"

	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
	"github.com/Masterminds/squirrel"
)

func RetrieveAppliedMigrations(ctx context.Context, logger lager.Logger, conn *sql.DB, tableName string) (map[int]AppliedMigration, error) {
	rows, err := squirrel.Select("version", "name", "applied_at").
		From(tableName).
		RunWith(conn).
		QueryContext(ctx)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var (
		version   int
		name      string
		appliedAt time.Time
	)

	versions := make(map[int]AppliedMigration)
	for rows.Next() {
		err = rows.Scan(&version, &name, &appliedAt)
		if err != nil {
			logger.Error(messages.ErrFailedToParseAppliedMigration, err)

			return nil, err
		}
		versions[version] = AppliedMigration{
			Version:   version,
			Name:      name,
			AppliedAt: appliedAt,
		}
	}

	err = rows.Err()
	if err != nil {
		logger.Error(messages.ErrFailedToQueryMigrations, err)
		return nil, err
	}

	return versions, nil
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
