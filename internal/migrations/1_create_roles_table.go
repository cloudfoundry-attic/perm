package migrations

import (
	"context"

	"code.cloudfoundry.org/perm/internal/sqlx"
	"code.cloudfoundry.org/perm/logx"
)

var createRolesTableMySQL = `
CREATE TABLE IF NOT EXISTS role
(
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  uuid BINARY(16) NOT NULL UNIQUE,
  name VARCHAR(255) NOT NULL UNIQUE
)
`

var createRolesTablePostgres = `
CREATE TABLE IF NOT EXISTS role
(
	id BIGSERIAL NOT NULL PRIMARY KEY,
	uuid BYTEA NOT NULL UNIQUE,
	name VARCHAR(255) NOT NULL UNIQUE
)`

var deleteRolesTable = `DROP TABLE role`

func createRolesTableUp(ctx context.Context, logger logx.Logger, tx *sqlx.Tx) error {
	logger = logger.WithName("create-roles-table")
	logger.Debug(starting)
	defer logger.Debug(finished)

	var err error

	if tx.Driver() == sqlx.DBDriverMySQL {
		_, err = tx.ExecContext(ctx, createRolesTableMySQL)
	} else {
		_, err = tx.ExecContext(ctx, createRolesTablePostgres)
	}
	return err
}

func createRolesTableDown(ctx context.Context, logger logx.Logger, tx *sqlx.Tx) error {
	logger = logger.WithName("create-roles-table")
	logger.Debug(starting)
	defer logger.Debug(finished)

	_, err := tx.ExecContext(ctx,
		deleteRolesTable)

	return err
}
