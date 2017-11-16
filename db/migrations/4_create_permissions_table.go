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
	role_id BIGINT NOT NULL,
	name VARCHAR(255) NOT NULL UNIQUE,
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
