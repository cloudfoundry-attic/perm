package db

import (
	"context"
	"database/sql"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/api/repos"
	"code.cloudfoundry.org/perm/pkg/perm"
	"code.cloudfoundry.org/perm/pkg/sqlx"
	"github.com/Masterminds/squirrel"
	"github.com/go-sql-driver/mysql"
	"github.com/satori/go.uuid"
)

type DataService struct {
	conn *sqlx.DB
}

func NewDataService(conn *sqlx.DB) *DataService {
	return &DataService{
		conn: conn,
	}
}

func createRoleAndAssignPermissions(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	roleName string,
	permissions ...*perm.Permission,
) (*role, error) {
	role, err := createRole(ctx, logger, conn, roleName)
	if err != nil {
		return nil, err
	}

	for _, permission := range permissions {
		_, err = createAction(ctx, logger, conn, permission.Action)
		if err != nil && err != errActionAlreadyExistsInDB {
			return nil, err
		}

		action, err := findAction(ctx, logger, conn, permission.Action)
		if err != nil {
			return nil, err
		}

		_, err = createPermission(ctx, logger, conn, action.ID, role.ID, permission.ResourcePattern, permission.Action)
		if err != nil {
			return nil, err
		}

	}

	return role, nil
}

func createRole(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	name string,
) (*role, error) {
	logger = logger.Session("create-role")
	u := uuid.NewV4().Bytes()

	result, err := squirrel.Insert("role").
		Columns("uuid", "name").
		Values(u, name).
		RunWith(conn).
		ExecContext(ctx)

	switch e := err.(type) {
	case nil:
		roleID, err2 := result.LastInsertId()
		if err2 != nil {
			logger.Error(failedToRetrieveID, err2)
			return nil, err2
		}

		role := &role{
			ID: roleID,
			Role: &perm.Role{
				Name: name,
			},
		}
		return role, nil
	case *mysql.MySQLError:
		if e.Number == MySQLErrorCodeDuplicateKey {
			logger.Debug(errRoleAlreadyExists)
			return nil, perm.ErrRoleAlreadyExists
		}

		logger.Error(failedToCreateRole, err)
		return nil, err
	default:
		logger.Error(failedToCreateRole, err)
		return nil, err
	}
}

func findRole(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	requestedRoleName string,
) (*role, error) {
	logger = logger.Session("find-role")

	var (
		roleID   int64
		roleName string
	)

	err := squirrel.Select("id", "name").
		From("role").
		Where(squirrel.Eq{
			"name": requestedRoleName,
		}).
		RunWith(conn).
		ScanContext(ctx, &roleID, &roleName)

	switch err {
	case nil:
		return &role{
			ID: roleID,
			Role: &perm.Role{
				Name: roleName,
			},
		}, nil
	case sql.ErrNoRows:
		logger.Debug(errRoleNotFound)
		return nil, perm.ErrRoleNotFound
	default:
		logger.Error(failedToFindRole, err)
		return nil, err
	}
}

func deleteRole(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	roleName string,
) error {
	logger = logger.Session("delete-role")
	result, err := squirrel.Delete("role").
		Where(squirrel.Eq{
			"name": roleName,
		}).
		RunWith(conn).
		ExecContext(ctx)

	switch err {
	case nil:
		n, err2 := result.RowsAffected()
		if err2 != nil {
			logger.Error(failedToCountRowsAffected, err2)
			return err2
		}

		if n == 0 {
			logger.Debug(errRoleNotFound)
			return perm.ErrRoleNotFound
		}

		return nil
	case sql.ErrNoRows:
		logger.Debug(errRoleNotFound)
		return perm.ErrRoleNotFound
	default:
		logger.Error(failedToDeleteRole, err)
		return err
	}
}

func assignRole(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	roleName string,
	actorID string,
	actorNamespace string,
) error {
	logger = logger.Session("assign-role")

	role, err := findRole(ctx, logger, conn, roleName)
	if err != nil {
		return err
	}

	return createRoleAssignment(ctx, logger, conn, role.ID, actorID, actorNamespace)
}

func createRoleAssignmentForGroup(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	roleID int64,
	groupID string,
) error {
	logger = logger.Session("create-role-assignment-for-group").WithData(lager.Data{
		"role.id":                   roleID,
		"group_assignment.group_id": groupID,
	})

	u := uuid.NewV4().Bytes()
	_, err := squirrel.Insert("group_assignment").
		Columns("uuid", "role_id", "group_id").
		Values(u, roleID, groupID).
		RunWith(conn).
		ExecContext(ctx)

	switch e := err.(type) {
	case nil:
		return nil
	case *mysql.MySQLError:
		if e.Number == MySQLErrorCodeDuplicateKey {
			logger.Debug(errRoleAssignmentAlreadyExists)
			return perm.ErrAssignmentAlreadyExists
		}

		logger.Error(failedToCreateRoleAssignment, err)
		return err
	default:
		logger.Error(failedToCreateRoleAssignment, err)
		return err
	}
}

func assignRoleToGroup(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	roleName string,
	groupID string,
) error {
	logger = logger.Session("assign-role-to-group")

	role, err := findRole(ctx, logger, conn, roleName)
	if err != nil {
		return err
	}
	return createRoleAssignmentForGroup(ctx, logger, conn, role.ID, groupID)
}

func createRoleAssignment(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	roleID int64,
	actorID string,
	actorNamespace string,
) error {
	logger = logger.Session("create-role-assignment").WithData(lager.Data{
		"role.id":                    roleID,
		"assignment.actor_id":        actorID,
		"assignment.actor_namespace": actorNamespace,
	})

	u := uuid.NewV4().Bytes()
	_, err := squirrel.Insert("assignment").
		Columns("uuid", "role_id", "actor_id", "actor_namespace").
		Values(u, roleID, actorID, actorNamespace).
		RunWith(conn).
		ExecContext(ctx)

	switch e := err.(type) {
	case nil:
		return nil
	case *mysql.MySQLError:
		if e.Number == MySQLErrorCodeDuplicateKey {
			logger.Debug(errRoleAssignmentAlreadyExists)
			return perm.ErrAssignmentAlreadyExists
		}

		logger.Error(failedToCreateRoleAssignment, err)
		return err
	default:
		logger.Error(failedToCreateRoleAssignment, err)
		return err
	}
}

func unassignRole(ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	roleName string,
	actorID string,
	actorNamespace string,
) error {
	logger = logger.Session("unassign-role")

	role, err := findRole(ctx, logger, conn, roleName)
	if err != nil {
		return err
	}

	return deleteRoleAssignment(ctx, logger, conn, role.ID, actorID, actorNamespace)
}

func deleteRoleAssignment(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	roleID int64,
	actorID string,
	actorNamespace string,
) error {
	logger = logger.Session("delete-role-assignment").WithData(lager.Data{
		"role.id":                    roleID,
		"assignment.actor_id":        actorID,
		"assignment.actor_namespace": actorNamespace,
	})

	result, err := squirrel.Delete("assignment").
		Where(squirrel.Eq{
			"role_id":         roleID,
			"actor_id":        actorID,
			"actor_namespace": actorNamespace,
		}).
		RunWith(conn).
		ExecContext(ctx)

	switch err {
	case nil:
		n, e := result.RowsAffected()
		if e != nil {
			logger.Error(failedToDeleteRoleAssignment, e)
			return e
		}

		if n == 0 {
			logger.Debug(errRoleAssignmentNotFound)
			return perm.ErrAssignmentNotFound
		}

		return nil
	case sql.ErrNoRows:
		logger.Debug(errRoleAssignmentNotFound)
		return perm.ErrAssignmentNotFound
	default:
		logger.Error(failedToDeleteRoleAssignment, err)
		return err
	}
}

func unassignRoleFromGroup(ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	roleName string,
	groupID string,
) error {
	logger = logger.Session("unassign-role-from-group")

	role, err := findRole(ctx, logger, conn, roleName)
	if err != nil {
		return err
	}

	return deleteGroupRoleAssignment(ctx, logger, conn, role.ID, groupID)
}

func deleteGroupRoleAssignment(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	roleID int64,
	groupID string,
) error {
	logger = logger.Session("delete-role-assignment").WithData(lager.Data{
		"role.id":                   roleID,
		"group_assignment.group_id": groupID,
	})

	result, err := squirrel.Delete("group_assignment").
		Where(squirrel.Eq{
			"role_id":  roleID,
			"group_id": groupID,
		}).
		RunWith(conn).
		ExecContext(ctx)

	switch err {
	case nil:
		n, e := result.RowsAffected()
		if e != nil {
			logger.Error(failedToDeleteRoleAssignment, e)
			return e
		}

		if n == 0 {
			logger.Debug(errRoleAssignmentNotFound)
			return perm.ErrAssignmentNotFound
		}

		return nil
	case sql.ErrNoRows:
		logger.Debug(errRoleAssignmentNotFound)
		return perm.ErrAssignmentNotFound
	default:
		logger.Error(failedToDeleteRoleAssignment, err)
		return err
	}
}

func hasRole(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	query repos.HasRoleQuery,
) (bool, error) {
	logger = logger.Session("has-role")
	role, err := findRole(ctx, logger, conn, query.RoleName)
	if err != nil {
		return false, err
	}

	return findRoleAssignment(ctx, logger, conn, role.ID, query.Actor.ID, query.Actor.Namespace)
}

func hasRoleForGroup(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	query repos.HasRoleForGroupQuery,
) (bool, error) {
	logger = logger.Session("has-role-for-group")
	role, err := findRole(ctx, logger, conn, query.RoleName)
	if err != nil {
		return false, err
	}

	return findRoleAssignmentForGroup(ctx, logger, conn, role.ID, query.Group.ID)
}

func findRoleAssignment(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	roleID int64,
	actorID string,
	actorNamespace string,
) (bool, error) {
	logger = logger.Session("find-role-assignment").WithData(lager.Data{
		"role.id":                    roleID,
		"assignment.actor_id":        actorID,
		"assignment.actor_namespace": actorNamespace,
	})

	var count int

	err := squirrel.Select("count(role_id)").
		From("assignment").
		Where(squirrel.Eq{
			"role_id":         roleID,
			"actor_id":        actorID,
			"actor_namespace": actorNamespace,
		}).
		RunWith(conn).
		ScanContext(ctx, &count)
	if err != nil {
		logger.Error(failedToFindRoleAssignment, err)
		return false, err
	}

	return count > 0, nil
}

func findRoleAssignmentForGroup(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	roleID int64,
	groupID string,
) (bool, error) {
	logger = logger.Session("find-role-assignment-for-group").WithData(lager.Data{
		"role.id":                   roleID,
		"group_assignment.group_id": groupID,
	})

	var count int

	err := squirrel.Select("count(role_id)").
		From("group_assignment").
		Where(squirrel.Eq{
			"role_id":  roleID,
			"group_id": groupID,
		}).
		RunWith(conn).
		ScanContext(ctx, &count)
	if err != nil {
		logger.Error(failedToFindRoleAssignment, err)
		return false, err
	}

	return count > 0, nil
}

func listRolePermissions(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	query repos.ListRolePermissionsQuery,
) ([]*permission, error) {
	logger = logger.Session("list-role-permissions")

	role, err := findRole(ctx, logger, conn, query.RoleName)
	if err != nil {
		return nil, err
	}

	return findRolePermissions(ctx, logger, conn, role.ID)
}

func createAction(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	name string,
) (*action, error) {
	logger = logger.Session("create-permission-definition")
	u := uuid.NewV4().Bytes()

	result, err := squirrel.Insert("action").
		Columns("uuid", "name").
		Values(u, name).
		RunWith(conn).
		ExecContext(ctx)

	switch e := err.(type) {
	case nil:
		actionID, err2 := result.LastInsertId()
		if err2 != nil {
			logger.Error(failedToRetrieveID, err2)
			return nil, err2
		}

		return &action{
			ID: actionID,
			Action: &perm.Action{
				Name: name,
			},
		}, nil
	case *mysql.MySQLError:
		if e.Number == MySQLErrorCodeDuplicateKey {
			logger.Debug(errActionAlreadyExists)
			return nil, errActionAlreadyExistsInDB
		}

		logger.Error(failedToCreateAction, err)
		return nil, err
	default:
		logger.Error(failedToCreateAction, err)
		return nil, err
	}
}

func findAction(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	actionName string,
) (*action, error) {
	logger = logger.Session("find-permission-definition")

	var (
		actionID int64
		name     string
	)

	err := squirrel.Select("id", "name").
		From("action").
		Where(squirrel.Eq{
			"name": actionName,
		}).
		RunWith(conn).
		ScanContext(ctx, &actionID, &name)

	switch err {
	case nil:
		return &action{
			ID: actionID,
			Action: &perm.Action{
				Name: name,
			},
		}, nil
	case sql.ErrNoRows:
		logger.Debug(errActionNotFound)
		return nil, errActionNotFoundDB
	default:
		logger.Error(failedToFindAction, err)
		return nil, err
	}
}

func findRolePermissions(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	roleID int64,
) ([]*permission, error) {
	logger = logger.Session("find-role-permissions").WithData(lager.Data{
		"role.id": roleID,
	})

	rows, err := squirrel.Select("permission.id", "action.name", "permission.resource_pattern").
		From("permission").
		JoinClause("INNER JOIN role ON permission.role_id = role.id").
		JoinClause("INNER JOIN action action ON permission.action_id = action.id").
		Where(squirrel.Eq{"role_id": roleID}).
		RunWith(conn).
		QueryContext(ctx)
	if err != nil {
		logger.Error(failedToFindPermissions, err)
		return nil, err
	}
	defer rows.Close()

	var permissions []*permission
	for rows.Next() {
		var (
			permissionID    int64
			action          string
			resourcePattern string
		)
		e := rows.Scan(&permissionID, &action, &resourcePattern)
		if e != nil {
			logger.Error(failedToScanRow, e)
			return nil, e
		}

		p := permission{
			ID: permissionID,
			Permission: &perm.Permission{
				Action:          action,
				ResourcePattern: resourcePattern,
			},
		}
		permissions = append(permissions, &p)
	}

	err = rows.Err()
	if err != nil {
		logger.Error(failedToIterateOverRows, err)
		return nil, err
	}

	return permissions, nil
}

func hasPermission(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	query repos.HasPermissionQuery,
) (bool, error) {
	logger = logger.Session("has-permission").WithData(lager.Data{
		"actor.issuer":               query.Actor.Namespace,
		"assignment.actor_id":        query.Actor.ID,
		"permission.action":          query.Action,
		"permission.resourcePattern": query.ResourcePattern,
		"group_assignment.groups":    query.Groups,
	})

	var count int

	// Actor-based access grant.
	err := squirrel.Select("count(assignment.role_id)").
		From("assignment").
		JoinClause("INNER JOIN permission permission ON assignment.role_id = permission.role_id").
		JoinClause("INNER JOIN action ON permission.action_id = action.id").
		Where(squirrel.Eq{
			"assignment.actor_id":         query.Actor.ID,
			"assignment.actor_namespace":  query.Actor.Namespace,
			"action.name":                 query.Action,
			"permission.resource_pattern": query.ResourcePattern,
		}).
		RunWith(conn).
		ScanContext(ctx, &count)

	if err != nil {
		logger.Error(failedToFindPermissions, err)
		return false, err
	}

	if count > 0 {
		return true, nil
	}
	// Group-based access grant.
	for _, group := range query.Groups {
		err := squirrel.Select("count(group_assignment.role_id)").
			From("group_assignment").
			JoinClause("INNER JOIN permission permission ON group_assignment.role_id = permission.role_id").
			JoinClause("INNER JOIN action ON permission.action_id = action.id").
			Where(squirrel.Eq{
				"group_assignment.group_id":   group.ID,
				"action.name":                 query.Action,
				"permission.resource_pattern": query.ResourcePattern,
			}).
			RunWith(conn).
			ScanContext(ctx, &count)

		if err != nil {
			logger.Error(failedToFindPermissions, err)
			return false, err
		}

		if count > 0 {
			return true, nil
		}
	}
	return false, nil
}

func createPermission(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	actionID int64,
	roleID int64,
	resourcePattern string,
	action string,
) (*permission, error) {
	logger = logger.Session("create-permission-definition")
	u := uuid.NewV4().Bytes()

	result, err := squirrel.Insert("permission").
		Columns("uuid", "action_id", "role_id", "resource_pattern").
		Values(u, actionID, roleID, resourcePattern).
		RunWith(conn).
		ExecContext(ctx)

	switch e := err.(type) {
	case nil:
		permissionID, err2 := result.LastInsertId()
		if err2 != nil {
			logger.Error(failedToRetrieveID, err2)
			return nil, err2
		}

		permission := &permission{
			ID: permissionID,
			Permission: &perm.Permission{
				Action:          action,
				ResourcePattern: resourcePattern,
			},
		}
		return permission, nil
	case *mysql.MySQLError:
		if e.Number == MySQLErrorCodeDuplicateKey {
			logger.Debug(errPermissionAlreadyExists)
			return nil, errPermissionAlreadyExistsDB
		}

		logger.Error(failedToCreatePermission, err)
		return nil, err
	default:
		logger.Error(failedToCreatePermission, err)
		return nil, err
	}
}

func listResourcePatterns(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	query repos.ListResourcePatternsQuery,
) ([]string, error) {
	action := query.Action
	namespace := query.Actor.Namespace
	id := query.Actor.ID

	logger = logger.Session("list-resource-patterns").
		WithData(lager.Data{
			"assignment.actor_namespace": namespace,
			"assignment.actor_id":        id,
			"permission.action":          action,
		})

	rows, err := squirrel.Select("permission.resource_pattern").
		Distinct().
		From("role").
		Join("assignment ON role.id = assignment.role_id").
		Join("permission ON permission.role_id = role.id").
		Join("action ON permission.action_id = action.id").
		Where(squirrel.Eq{
			"action.name":                action,
			"assignment.actor_id":        id,
			"assignment.actor_namespace": namespace,
		}).
		RunWith(conn).
		QueryContext(ctx)
	if err != nil {
		logger.Error(failedToListResourcePatterns, err)
		return nil, err
	}
	defer rows.Close()

	var resourcePatterns []string
	for rows.Next() {
		var resourcePattern string

		err = rows.Scan(&resourcePattern)
		if err != nil {
			logger.Error(failedToScanRow, err)
			return nil, err
		}

		resourcePatterns = append(resourcePatterns, resourcePattern)
	}

	err = rows.Err()
	if err != nil {
		logger.Error(failedToIterateOverRows, err)
		return nil, err
	}

	return resourcePatterns, nil
}
