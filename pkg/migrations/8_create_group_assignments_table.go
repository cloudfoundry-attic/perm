package migrations

import (
	"context"
	"strings"

	"code.cloudfoundry.org/perm/pkg/logx"
	"code.cloudfoundry.org/perm/pkg/sqlx"
)

var createGroupAssignmentTable = `
CREATE TABLE IF NOT EXISTS group_assignment
(
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  uuid BINARY(16) NOT NULL UNIQUE,
	role_id BIGINT NOT NULL,
	group_id VARCHAR(511) NOT NULL,
  role_id_group_hash VARCHAR(64) AS (SHA2(CONCAT(role_id, group_id), 256)) VIRTUAL UNIQUE
)
`

var createGroupAssignmentTableMariaDB = `
CREATE TABLE IF NOT EXISTS group_assignment
(
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  uuid BINARY(16) NOT NULL UNIQUE,
	role_id BIGINT NOT NULL,
	group_id VARCHAR(511) NOT NULL,
  role_id_group_hash VARCHAR(64) AS (SHA2(CONCAT(role_id, group_id), 256)) PERSISTENT UNIQUE
)
`

var addGroupAssignmentRoleIDForeignKey = `
ALTER TABLE
	group_assignment
ADD CONSTRAINT
	group_assignment_role_id_fkey
FOREIGN KEY(role_id) REFERENCES role(id)
ON DELETE CASCADE
`
var dropGroupAssignmentTable = `DROP TABLE IF EXISTS group_assignment`

func createGroupAssignmentTableUp(ctx context.Context, logger logx.Logger, tx *sqlx.Tx) error {
	logger = logger.WithName("create-group-assignment-table")
	logger.Debug(starting)
	defer logger.Debug(finished)
	var err error

	if tx.Flavor() == sqlx.DBFlavorMariaDB && strings.HasPrefix(tx.Version(), "10.1") {
		_, err = tx.ExecContext(ctx, createGroupAssignmentTableMariaDB)
	} else {
		_, err = tx.ExecContext(ctx, createGroupAssignmentTable)
	}
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, addGroupAssignmentRoleIDForeignKey)
	if err != nil {
		return err
	}

	return err
}

func createGroupAssignmentTableDown(ctx context.Context, logger logx.Logger, tx *sqlx.Tx) error {
	logger = logger.WithName("create-role-assignments-table")
	logger.Debug(starting)
	defer logger.Debug(finished)

	_, err := tx.ExecContext(ctx, dropGroupAssignmentTable)

	return err
}
