package migrations

import (
	"context"

	"code.cloudfoundry.org/perm/pkg/logx"
	"code.cloudfoundry.org/perm/pkg/sqlx"
)

var createPermissionDefinitionsTable = `
CREATE TABLE IF NOT EXISTS permission_definition
(
	id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  	uuid BINARY(16) NOT NULL UNIQUE,
  	name VARCHAR(255) NOT NULL UNIQUE
)
`

var deletePermissionDefinitionsTable = `DROP TABLE permission_definition`

func createPermissionDefinitionsTableUp(ctx context.Context, logger logx.Logger, tx *sqlx.Tx) error {
	logger = logger.WithName("create-permission-definitions-table")
	logger.Debug(starting)
	defer logger.Debug(finished)

	_, err := tx.ExecContext(ctx, createPermissionDefinitionsTable)

	return err
}

func createPermissionDefinitionsTableDown(ctx context.Context, logger logx.Logger, tx *sqlx.Tx) error {
	logger = logger.WithName("create-permission-definitions-table")
	logger.Debug(starting)
	defer logger.Debug(finished)

	_, err := tx.ExecContext(ctx, deletePermissionDefinitionsTable)

	return err
}
