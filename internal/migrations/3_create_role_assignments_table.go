package migrations

import (
	"context"

	"code.cloudfoundry.org/perm/pkg/logx"
	"code.cloudfoundry.org/perm/pkg/sqlx"
)

var createRoleAssignmentsTable = `
CREATE TABLE IF NOT EXISTS role_assignment
(
	role_id BIGINT NOT NULL,
	actor_id BIGINT NOT NULL
)
`

var addRoleAssignmentPrimaryKey = `
ALTER TABLE
	role_assignment
ADD CONSTRAINT
	role_assignment_pkey
PRIMARY KEY(role_id, actor_id)
`

var addRoleIDForeignKey = `
ALTER TABLE
	role_assignment
ADD CONSTRAINT
	role_assignment_role_id_fkey
FOREIGN KEY(role_id) REFERENCES role(id)
ON DELETE CASCADE
`

var addActorIDForeignKey = `
ALTER TABLE
	role_assignment
ADD CONSTRAINT
	role_assignment_actor_id_fkey
FOREIGN KEY(actor_id) REFERENCES actor(id)
ON DELETE CASCADE
`

var deleteRoleAssignmentsTable = `DROP TABLE role_assignment`

func createRoleAssignmentsTableUp(ctx context.Context, logger logx.Logger, tx *sqlx.Tx) error {
	logger = logger.WithName("create-role-assignments-table")
	logger.Debug(starting)
	defer logger.Debug(finished)

	_, err := tx.ExecContext(ctx,
		createRoleAssignmentsTable)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, addRoleAssignmentPrimaryKey)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, addRoleIDForeignKey)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, addActorIDForeignKey)

	return err
}

func createRoleAssignmentsTableDown(ctx context.Context, logger logx.Logger, tx *sqlx.Tx) error {
	logger = logger.WithName("create-role-assignments-table")
	logger.Debug(starting)
	defer logger.Debug(finished)

	_, err := tx.ExecContext(ctx,
		deleteRoleAssignmentsTable)

	return err
}
