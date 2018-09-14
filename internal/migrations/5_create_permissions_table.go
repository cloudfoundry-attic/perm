package migrations

import (
	"context"

	"code.cloudfoundry.org/perm/internal/sqlx"
	"code.cloudfoundry.org/perm/logx"
)

var createPermissionsTableMySQL = `
CREATE TABLE IF NOT EXISTS permission
(
	id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  	uuid BINARY(16) NOT NULL UNIQUE,
	role_id BIGINT NOT NULL,
	permission_definition_id BIGINT NOT NULL,
	resource_pattern VARCHAR(255) NOT NULL
)
`

var createPermissionsTablePostgres = `
CREATE TABLE IF NOT EXISTS permission
(
	id BIGSERIAL NOT NULL PRIMARY KEY,
  uuid BYTEA NOT NULL UNIQUE,
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

func createPermissionsTableUp(ctx context.Context, logger logx.Logger, tx *sqlx.Tx) error {
	logger = logger.WithName("create-permissions-table")
	logger.Debug(starting)
	defer logger.Debug(finished)

	var err error

	if tx.Driver() == sqlx.DBDriverMySQL {
		_, err = tx.ExecContext(ctx, createPermissionsTableMySQL)
	} else {
		_, err = tx.ExecContext(ctx, createPermissionsTablePostgres)
	}
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

func createPermissionsTableDown(ctx context.Context, logger logx.Logger, tx *sqlx.Tx) error {
	logger = logger.WithName("create-permissions-table")
	logger.Debug(starting)
	defer logger.Debug(finished)

	_, err := tx.ExecContext(ctx, deletePermissionsTable)

	return err
}
