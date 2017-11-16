package migrations

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
	"code.cloudfoundry.org/perm/sqlx"
)

var createPermissionsTable = `
CREATE TABLE IF NOT EXISTS permission
(
	id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  	uuid BINARY(16) NOT NULL UNIQUE,
	role_id BIGINT NOT NULL,
	permission_definition_id BIGINT NOT NULL,
	resource_pattern VARCHAR(255) NOT NULL
)
`

var addPermissionRoleIDForeignKey = `
ALTER TABLE
	permission
ADD CONSTRAINT
	permission_role_id_fkey
FOREIGN KEY(role_id) REFERENCES role(id)
ON DELETE CASCADE
`

var addPermissionPermissionDefinitionIDForeignKey = `
ALTER TABLE
	permission
ADD CONSTRAINT
	permission_permission_definition_id_fkey
FOREIGN KEY(permission_definition_id) REFERENCES permission_definition(id)
ON DELETE CASCADE
`

var deletePermissionsTable = `DROP TABLE permission`

func CreatePermissionsTableUp(ctx context.Context, logger lager.Logger, tx *sqlx.Tx) error {
	logger = logger.Session("create-permissions-table")
	logger.Debug(messages.Starting)
	defer logger.Debug(messages.Finished)

	_, err := tx.ExecContext(ctx,
		createPermissionsTable)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, addPermissionRoleIDForeignKey)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, addPermissionPermissionDefinitionIDForeignKey)
	return err
}

func CreatePermissionsTableDown(ctx context.Context, logger lager.Logger, tx *sqlx.Tx) error {
	logger = logger.Session("create-permissions-table")
	logger.Debug(messages.Starting)
	defer logger.Debug(messages.Finished)

	_, err := tx.ExecContext(ctx,
		deletePermissionsTable)

	return err
}
