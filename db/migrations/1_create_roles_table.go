package migrations

import (
	"context"
	"database/sql"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
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

func CreateRolesTableUp(ctx context.Context, logger lager.Logger, tx *sql.Tx) error {
	logger = logger.Session("create-roles-table")
	logger.Debug(messages.Starting)
	defer logger.Debug(messages.Finished)

	_, err := tx.ExecContext(ctx,
		createRolesTable)

	return err
}

func CreateRolesTableDown(ctx context.Context, logger lager.Logger, tx *sql.Tx) error {
	logger = logger.Session("create-roles-table")
	logger.Debug(messages.Starting)
	defer logger.Debug(messages.Finished)

	_, err := tx.ExecContext(ctx,
		deleteRolesTable)

	return err
}
