package migrations

import (
	"context"

	"strings"

	"code.cloudfoundry.org/perm/pkg/logx"
	"code.cloudfoundry.org/perm/pkg/sqlx"
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

var createActorsTableMariaDB = `
CREATE TABLE IF NOT EXISTS actor
(
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  uuid BINARY(16) NOT NULL UNIQUE,
  domain_id VARCHAR(511) NOT NULL,
  issuer VARCHAR(2047) NOT NULL,
  domain_id_issuer_hash VARCHAR(64) AS (SHA2(CONCAT(domain_id, issuer), 256)) PERSISTENT
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

	if tx.Flavor() == sqlx.DBFlavorMariaDB && strings.HasPrefix(tx.Version(), "10.1") {
		_, err = tx.ExecContext(ctx, createActorsTableMariaDB)
	} else {
		_, err = tx.ExecContext(ctx, createActorsTable)
	}
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
