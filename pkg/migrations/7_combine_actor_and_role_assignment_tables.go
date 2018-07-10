package migrations

import (
	"context"
	"strings"

	"code.cloudfoundry.org/perm/pkg/logx"
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

var dropAssignmentTable = `DROP TABLE IF EXISTS assignment`

var dropRoleAssignmentTable = `DROP TABLE IF EXISTS role_assignment`

func combineActorAndRoleAssignmentTablesUp(ctx context.Context, logger logx.Logger, tx *sqlx.Tx) error {
	logger = logger.WithName("create-actor-and-role-assignment-tables")
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

	type roleAssignment struct {
		RoleID         int64
		ActorID        string
		ActorNamespace string
	}

	var roleAssignments []roleAssignment

	for rows.Next() {
		ra := roleAssignment{}
		err = rows.Scan(&ra.RoleID, &ra.ActorID, &ra.ActorNamespace)
		if err != nil {
			return err
		}
		roleAssignments = append(roleAssignments, ra)
	}
	for _, ra := range roleAssignments {
		u := uuid.NewV4().Bytes()
		_, err = squirrel.Insert("assignment").
			Columns("uuid", "role_id", "actor_id", "actor_namespace").
			Values(u, ra.RoleID, ra.ActorID, ra.ActorNamespace).
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

func combineActorAndRoleAssignmentTablesDown(ctx context.Context, logger logx.Logger, tx *sqlx.Tx) error {
	logger = logger.WithName("create-actor-and-role-assignment-tables")
	logger.Debug(starting)
	defer logger.Debug(finished)
	var err error

	err = createActorsTableUp(ctx, logger, tx)
	if err != nil {
		return err
	}

	err = createRoleAssignmentsTableUp(ctx, logger, tx)
	if err != nil {
		return err
	}

	var entityData [][]string
	rows, err := squirrel.Select("actor_id", "actor_namespace").
		From("assignment").
		RunWith(tx).
		QueryContext(ctx)
	if err != nil {
		return err
	}

	for rows.Next() {
		var (
			actorID        string
			actorNamespace string
		)

		err = rows.Scan(&actorID, &actorNamespace)
		entityData = append(entityData, []string{actorID, actorNamespace})
		if err != nil {
			return err
		}

	}
	rows.Close()

	for _, actor := range entityData {
		u := uuid.NewV4().Bytes()
		_, err = squirrel.Insert("actor").
			Columns("uuid", "domain_id", "issuer").
			Values(u, actor[0], actor[1]).
			RunWith(tx).
			ExecContext(ctx)
		if err != nil {
			return err
		}
	}

	var roleAssignmentData [][]int64
	rows, err = squirrel.Select("actor.id", "assignment.role_id").
		From("actor").
		JoinClause("INNER JOIN assignment ON actor.domain_id = assignment.actor_id and actor.issuer = assignment.actor_namespace").
		RunWith(tx).
		QueryContext(ctx)
	if err != nil {
		return err
	}

	for rows.Next() {
		var (
			actorID int64
			roleID  int64
		)

		err = rows.Scan(&actorID, &roleID)
		roleAssignmentData = append(roleAssignmentData, []int64{actorID, roleID})
		if err != nil {
			return err
		}

	}
	rows.Close()

	for _, roleAssignment := range roleAssignmentData {
		_, err = squirrel.Insert("role_assignment").
			Columns("actor_id", "role_id").
			Values(roleAssignment[0], roleAssignment[1]).
			RunWith(tx).
			ExecContext(ctx)
		if err != nil {
			return err
		}
	}

	_, err = tx.ExecContext(ctx, dropAssignmentTable)
	if err != nil {
		return err
	}

	return nil
}
