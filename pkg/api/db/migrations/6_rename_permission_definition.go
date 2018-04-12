package migrations

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/sqlx"
)

var renamePermissionDefinitionTable = `
ALTER TABLE
	permission_definition
RENAME TO
	action
`

var dropPermissionPermissionDefinitionIDForeignKey = `
ALTER TABLE
	permission
DROP FOREIGN KEY
	permission_permission_definition_id_fkey
`

var renamePermissionPermissionDefinitionID = `
ALTER TABLE
	permission
CHANGE COLUMN
	permission_definition_id action_id BIGINT NOT NULL
`

var addPermissionActionIDForeignKey = `
ALTER TABLE
	permission
ADD CONSTRAINT
	permission_action_id_fkey
FOREIGN KEY(action_id) REFERENCES action(id)
ON DELETE CASCADE
`

func RenamePermissionDefinitionToActionUp(ctx context.Context, logger lager.Logger, tx *sqlx.Tx) error {
	logger = logger.Session("rename-permission-definition-to-action")
	logger.Debug(starting)
	defer logger.Debug(finished)

	_, err := tx.ExecContext(ctx, renamePermissionDefinitionTable)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, dropPermissionPermissionDefinitionIDForeignKey)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, renamePermissionPermissionDefinitionID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, addPermissionActionIDForeignKey)
	return err
}

var dropPermissionActionIDForeignKey = `
ALTER TABLE
	permission
DROP FOREIGN KEY permission_action_id_fkey
`

var renamePermissionActionID = `
ALTER TABLE
	permission
CHANGE COLUMN
	action_id permission_definition_id BIGINT NOT NULL
`

var renameActionTable = `
ALTER TABLE
	action
RENAME TO
	permission_definition
`

func RenamePermissionDefinitionToActionDown(ctx context.Context, logger lager.Logger, tx *sqlx.Tx) error {
	logger = logger.Session("rename-action-to-permission-definition")
	logger.Debug(starting)
	defer logger.Debug(finished)

	_, err := tx.ExecContext(ctx, dropPermissionActionIDForeignKey)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, renamePermissionActionID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, addPermissionRoleIDForeignKey)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, renameActionTable)
	return err
}
