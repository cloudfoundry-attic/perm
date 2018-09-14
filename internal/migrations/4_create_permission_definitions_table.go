package migrations

import (
	"context"

	"code.cloudfoundry.org/perm/internal/sqlx"
	"code.cloudfoundry.org/perm/logx"
)

var createPermissionDefinitionsTableMySQL = `
CREATE TABLE IF NOT EXISTS permission_definition
(
	id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  	uuid BINARY(16) NOT NULL UNIQUE,
  	name VARCHAR(255) NOT NULL UNIQUE
)
`

var createPermissionDefinitionsTablePostgres = `
CREATE TABLE IF NOT EXISTS permission_definition
(
	id BIGSERIAL NOT NULL PRIMARY KEY,
  uuid BYTEA NOT NULL UNIQUE,
  name VARCHAR(255) NOT NULL UNIQUE
)
`

var deletePermissionDefinitionsTable = `DROP TABLE permission_definition`

func createPermissionDefinitionsTableUp(ctx context.Context, logger logx.Logger, tx *sqlx.Tx) error {
	logger = logger.WithName("create-permission-definitions-table")
	logger.Debug(starting)
	defer logger.Debug(finished)

	var err error

	if tx.Driver() == sqlx.DBDriverMySQL {
		_, err = tx.ExecContext(ctx, createPermissionDefinitionsTableMySQL)
	} else {
		_, err = tx.ExecContext(ctx, createPermissionDefinitionsTablePostgres)
	}

	return err
}

func createPermissionDefinitionsTableDown(ctx context.Context, logger logx.Logger, tx *sqlx.Tx) error {
	logger = logger.WithName("create-permission-definitions-table")
	logger.Debug(starting)
	defer logger.Debug(finished)

	_, err := tx.ExecContext(ctx, deletePermissionDefinitionsTable)

	return err
}
