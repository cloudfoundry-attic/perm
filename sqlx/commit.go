package sqlx

import (
	"database/sql"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
)

func Commit(logger lager.Logger, tx *sql.Tx, err error) error {
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
