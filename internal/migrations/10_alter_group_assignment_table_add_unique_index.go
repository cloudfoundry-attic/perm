package migrations

import (
	"context"

	"code.cloudfoundry.org/perm/pkg/logx"
	"code.cloudfoundry.org/perm/pkg/sqlx"
)

var addUniqueIndexToGroupAssignmentTable = `
ALTER TABLE
	group_assignment
ADD UNIQUE INDEX
	unique_group_assignment (role_id, group_id)
`

var dropRoleIDGroupHashFromGroupAssignmentTable = `
ALTER TABLE
	group_assignment
DROP COLUMN
	role_id_group_hash
`

var dropUniqueIndexFromGroupAssignmentTable = `
ALTER TABLE
	group_assignment
DROP INDEX
	unique_group_assignment
`

var addRoleIDGroupHashToGroupAssignmentTable = `
ALTER TABLE
	group_assignment
ADD COLUMN
  role_id_group_hash VARCHAR(64) AS (SHA2(CONCAT(role_id, group_id), 256)) VIRTUAL UNIQUE
`

func alterGroupAssignmentTableAddUniqueIndexUp(ctx context.Context, logger logx.Logger, tx *sqlx.Tx) error {
	logger = logger.WithName("alter-group-assignment-table-add-unique-index")
	logger.Debug(starting)
	defer logger.Debug(finished)

	_, err := tx.ExecContext(ctx, addUniqueIndexToGroupAssignmentTable)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, dropRoleIDGroupHashFromGroupAssignmentTable)
	if err != nil {
		return err
	}
	return nil
}

func alterGroupAssignmentTableAddUniqueIndexDown(ctx context.Context, logger logx.Logger, tx *sqlx.Tx) error {
	logger = logger.WithName("alter-group-assignment-table-add-unique-index")
	logger.Debug(starting)
	defer logger.Debug(finished)

	_, err := tx.ExecContext(ctx, addRoleIDGroupHashToGroupAssignmentTable)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, dropUniqueIndexFromGroupAssignmentTable)
	if err != nil {
		return err
	}
	return nil
}
