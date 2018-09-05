package sqlx

import "code.cloudfoundry.org/perm/logx"

func Commit(logger logx.Logger, tx *Tx, err error) error {
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			logger.Error(failedToRollback, rollbackErr)
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		logger.Error(failedToCommit, err)
		return err
	}

	logger.Debug(committed)
	return nil
}
