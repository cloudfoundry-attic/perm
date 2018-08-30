package migrations

import (
	"context"

	"code.cloudfoundry.org/perm/internal/sqlx"
	"code.cloudfoundry.org/perm/pkg/logx"
)

var addUniqueIndexToAssignmentTable = `
ALTER TABLE
	assignment
ADD UNIQUE INDEX
	unique_assignment (role_id, actor_id, actor_namespace)
`

var dropRoleIDActorHashFromAssignmentTable = `
ALTER TABLE
	assignment
DROP COLUMN
	role_id_actor_hash
`

var dropUniqueIndexFromAssignmentTable = `
ALTER TABLE
	assignment
DROP INDEX
	unique_assignment
`

var addRoleIDActorHashToAssignmentTable = `
ALTER TABLE
	assignment
ADD COLUMN
  role_id_actor_hash VARCHAR(64) AS (SHA2(CONCAT(role_id, actor_id, actor_namespace), 256)) VIRTUAL UNIQUE
`

func alterAssignmentTableAddUniqueIndexUp(ctx context.Context, logger logx.Logger, tx *sqlx.Tx) error {
	logger = logger.WithName("alter-assignment-table-add-unique-index")
	logger.Debug(starting)
	defer logger.Debug(finished)

	_, err := tx.ExecContext(ctx, addUniqueIndexToAssignmentTable)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, dropRoleIDActorHashFromAssignmentTable)
	if err != nil {
		return err
	}
	return nil
}

func alterAssignmentTableAddUniqueIndexDown(ctx context.Context, logger logx.Logger, tx *sqlx.Tx) error {
	logger = logger.WithName("alter-assignment-table-add-unique-index")
	logger.Debug(starting)
	defer logger.Debug(finished)

	_, err := tx.ExecContext(ctx, addRoleIDActorHashToAssignmentTable)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, dropUniqueIndexFromAssignmentTable)
	if err != nil {
		return err
	}
	return nil
}
