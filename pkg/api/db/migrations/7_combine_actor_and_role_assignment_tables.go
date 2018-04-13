package migrations

import (
	"context"
	"strings"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/sqlx"
	"github.com/Masterminds/squirrel"
	uuid "github.com/satori/go.uuid"
)

var createAssignmentTable = `
CREATE TABLE IF NOT EXISTS assignment
(
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  uuid BINARY(16) NOT NULL UNIQUE,
	role_id BIGINT NOT NULL,
	actor_id VARCHAR(511) NOT NULL,
	actor_namespace VARCHAR(2047) NOT NULL,
  role_id_actor_hash VARCHAR(64) AS (SHA2(CONCAT(role_id, actor_id, actor_namespace), 256)) VIRTUAL UNIQUE
)
`

var createAssignmentTableMariaDB = `
CREATE TABLE IF NOT EXISTS assignment
(
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  uuid BINARY(16) NOT NULL UNIQUE,
	role_id BIGINT NOT NULL,
	actor_id VARCHAR(511) NOT NULL,
	actor_namespace VARCHAR(2047) NOT NULL,
  role_id_actor_hash VARCHAR(64) AS (SHA2(CONCAT(role_id, actor_id, actor_namespace), 256)) PERSISTENT UNIQUE
)
`

var addAssignmentRoleIDForeignKey = `
ALTER TABLE
	assignment
ADD CONSTRAINT
	assignment_role_id_fkey
FOREIGN KEY(role_id) REFERENCES role(id)
ON DELETE CASCADE
`
var dropActorTable = `DROP TABLE IF EXISTS actor`

var dropRoleAssignmentTable = `DROP TABLE IF EXISTS role_assignment`

func CombineActorAndRoleAssignmentTablesUp(ctx context.Context, logger lager.Logger, tx *sqlx.Tx) error {
	logger = logger.Session("create-actor-and-role-assignment-tables")
	logger.Debug(starting)
	defer logger.Debug(finished)
	var err error

	if tx.Flavor() == sqlx.DBFlavorMariaDB && strings.HasPrefix(tx.Version(), "10.1") {
		_, err = tx.ExecContext(ctx, createAssignmentTableMariaDB)
	} else {
		_, err = tx.ExecContext(ctx, createAssignmentTable)
	}
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, addAssignmentRoleIDForeignKey)
	if err != nil {
		return err
	}

	rows, err := squirrel.Select("role_assignment.role_id", "actor.domain_id", "actor.issuer").
		From("role_assignment").
		JoinClause("INNER JOIN actor ON role_assignment.actor_id = actor.id").
		RunWith(tx).
		QueryContext(ctx)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			roleID         int64
			actorID        string
			actorNamespace string
		)

		err = rows.Scan(&roleID, &actorID, &actorNamespace)
		if err != nil {
			return err
		}

		u := uuid.NewV4().Bytes()
		_, err = squirrel.Insert("assignment").
			Columns("uuid", "role_id", "actor_id", "actor_namespace").
			Values(u, roleID, actorID, actorNamespace).
			RunWith(tx).
			ExecContext(ctx)
		if err != nil {
			return err
		}
	}

	_, err = tx.ExecContext(ctx, dropRoleAssignmentTable)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, dropActorTable)
	return err
}

func CombineActorAndRoleAssignmentTablesDown(ctx context.Context, logger lager.Logger, tx *sqlx.Tx) error {
	logger = logger.Session("create-actor-and-role-assignment-tables")
	logger.Debug(starting)
	defer logger.Debug(finished)

	return nil
}