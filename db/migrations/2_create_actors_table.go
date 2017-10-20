package migrations

import (
	"context"
	"database/sql"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
)

var createActorsTable = `
CREATE TABLE IF NOT EXISTS actor
(
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  uuid BINARY(16) NOT NULL UNIQUE,
  domain_id VARCHAR(511) NOT NULL,
  issuer VARCHAR(2047) NOT NULL
)
`

var deleteActorsTable = `DROP TABLE actor`

func CreateActorsTableUp(ctx context.Context, logger lager.Logger, tx *sql.Tx) error {
	logger = logger.Session("create-actors-table")
	logger.Debug(messages.Starting)
	defer logger.Debug(messages.Finished)

	_, err := tx.ExecContext(ctx,
		createActorsTable)

	return err
}

func CreateActorsTableDown(ctx context.Context, logger lager.Logger, tx *sql.Tx) error {
	logger = logger.Session("create-actors-table")
	logger.Debug(messages.Starting)
	defer logger.Debug(messages.Finished)

	_, err := tx.ExecContext(ctx,
		deleteActorsTable)

	return err
}
