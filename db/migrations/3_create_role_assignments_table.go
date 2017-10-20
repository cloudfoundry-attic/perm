package migrations

import (
	"context"
	"database/sql"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
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

func CreateRoleAssignmentsTableUp(ctx context.Context, logger lager.Logger, tx *sql.Tx) error {
	logger = logger.Session("create-role-assignments-table")
	logger.Debug(messages.Starting)
	defer logger.Debug(messages.Finished)

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

func CreateRoleAssignmentsTableDown(ctx context.Context, logger lager.Logger, tx *sql.Tx) error {
	logger = logger.Session("create-role-assignments-table")
	logger.Debug(messages.Starting)
	defer logger.Debug(messages.Finished)

	_, err := tx.ExecContext(ctx,
		deleteRoleAssignmentsTable)

	return err
}
