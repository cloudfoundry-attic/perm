package migrations

import (
	"context"

	"code.cloudfoundry.org/perm/pkg/logx"
	"code.cloudfoundry.org/perm/pkg/sqlx"
)

var createRolesTable = `
CREATE TABLE IF NOT EXISTS role
(
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  uuid BINARY(16) NOT NULL UNIQUE,
  name VARCHAR(255) NOT NULL UNIQUE
)
`

var deleteRolesTable = `DROP TABLE role`

func createRolesTableUp(ctx context.Context, logger logx.Logger, tx *sqlx.Tx) error {
	logger = logger.WithName("create-roles-table")
	logger.Debug(starting)
	defer logger.Debug(finished)

	_, err := tx.ExecContext(ctx,
		createRolesTable)

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
