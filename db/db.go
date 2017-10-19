package db

import (
	"context"
	"database/sql"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
	"github.com/Masterminds/squirrel"
)

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
		name    string
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
