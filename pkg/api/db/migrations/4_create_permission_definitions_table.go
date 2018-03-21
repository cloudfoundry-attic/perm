package migrations

import (
	"context"

	"code.cloudfoundry.org/lager"
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

func CreatePermissionDefinitionsTableUp(ctx context.Context, logger lager.Logger, tx *sqlx.Tx) error {
	logger = logger.Session("create-permission-definitions-table")
	logger.Debug(starting)
	defer logger.Debug(finished)

	_, err := tx.ExecContext(ctx,
		createPermissionDefinitionsTable)

	return err
}

func CreatePermissionDefinitionsTableDown(ctx context.Context, logger lager.Logger, tx *sqlx.Tx) error {
	logger = logger.Session("create-permission-definitions-table")
	logger.Debug(starting)
	defer logger.Debug(finished)

	_, err := tx.ExecContext(ctx,
		deletePermissionDefinitionsTable)

	return err
}
