package migrations

import (
	"context"

	"code.cloudfoundry.org/perm/internal/sqlx"
	"code.cloudfoundry.org/perm/logx"
)

var createActorsTable = `
CREATE TABLE IF NOT EXISTS actor
(
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  uuid BINARY(16) NOT NULL UNIQUE,
  domain_id VARCHAR(511) NOT NULL,
  issuer VARCHAR(2047) NOT NULL,
  domain_id_issuer_hash VARCHAR(64) AS (SHA2(CONCAT(domain_id, issuer), 256)) STORED
)
`

var uniqueActorsConstraint = `
ALTER TABLE
	actor
ADD CONSTRAINT
	actor_unique_domain_id_issuer_hash
UNIQUE(domain_id_issuer_hash)
`

var deleteActorsTable = `DROP TABLE actor`

func createActorsTableUp(ctx context.Context, logger logx.Logger, tx *sqlx.Tx) error {
	logger = logger.WithName("create-actors-table")
	logger.Debug(starting)
	defer logger.Debug(finished)

	var err error

	_, err = tx.ExecContext(ctx, createActorsTable)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, uniqueActorsConstraint)

	return err
}

func createActorsTableDown(ctx context.Context, logger logx.Logger, tx *sqlx.Tx) error {
	logger = logger.WithName("create-actors-table")
	logger.Debug(starting)
	defer logger.Debug(finished)

	_, err := tx.ExecContext(ctx, deleteActorsTable)

	return err
}
