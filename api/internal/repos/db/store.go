package db

import (
	"context"
	"database/sql"

	"code.cloudfoundry.org/perm"
	"code.cloudfoundry.org/perm/api/internal/repos"
	"code.cloudfoundry.org/perm/internal/sqlx"
	"code.cloudfoundry.org/perm/logx"
	"github.com/Masterminds/squirrel"
	"github.com/go-sql-driver/mysql"
	"github.com/satori/go.uuid"
)

type Store struct {
	conn *sqlx.DB
}

func NewStore(conn *sqlx.DB) *Store {
	return &Store{
		conn: conn,
	}
}

func createRoleAndAssignPermissions(
	ctx context.Context,
	logger logx.Logger,
	conn squirrel.BaseRunner,
	roleName string,
	permissions ...perm.Permission,
) (role, error) {
	r, err := createRole(ctx, logger, conn, roleName)
	if err != nil {
		return role{}, err
	}

	for _, permission := range permissions {
		_, err = createAction(ctx, logger, conn, permission.Action)
		if err != nil && err != errActionAlreadyExistsInDB {
			return role{}, err
		}

		a, err := findAction(ctx, logger, conn, permission.Action)
		if err != nil {
			return role{}, err
		}

		_, err = createPermission(ctx, logger, conn, a.ID, r.ID, permission.ResourcePattern, permission.Action)
		if err != nil {
			return role{}, err
		}

	}

	return r, nil
}

func createRole(
	ctx context.Context,
	logger logx.Logger,
	conn squirrel.BaseRunner,
	name string,
) (role, error) {
	logger = logger.WithName("create-role")
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
			return role{}, err2
		}

		return role{
			ID: roleID,
			Role: perm.Role{
				Name: name,
			},
		}, nil
	case *mysql.MySQLError:
		if e.Number == MySQLErrorCodeDuplicateKey {
			logger.Debug(errRoleAlreadyExists)
			return role{}, perm.ErrRoleAlreadyExists
		}

		logger.Error(failedToCreateRole, err)
		return role{}, err
	default:
		logger.Error(failedToCreateRole, err)
		return role{}, err
	}
}

func findRole(
	ctx context.Context,
	logger logx.Logger,
	conn squirrel.BaseRunner,
	requestedRoleName string,
) (role, error) {
	logger = logger.WithName("find-role")

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
		return role{
			ID: roleID,
			Role: perm.Role{
				Name: roleName,
			},
		}, nil
	case sql.ErrNoRows:
		logger.Debug(errRoleNotFound)
		return role{}, perm.ErrRoleNotFound
	default:
		logger.Error(failedToFindRole, err)
		return role{}, err
	}
}

func deleteRole(
	ctx context.Context,
	logger logx.Logger,
	conn squirrel.BaseRunner,
	roleName string,
) error {
	logger = logger.WithName("delete-role")
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
	logger logx.Logger,
	conn squirrel.BaseRunner,
	roleName string,
	actorID string,
	actorNamespace string,
) error {
	logger = logger.WithName("assign-role")

	foundRole, err := findRole(ctx, logger, conn, roleName)
	if err != nil {
		return err
	}

	return createRoleAssignment(ctx, logger, conn, foundRole.ID, actorID, actorNamespace)
}

func createRoleAssignmentForGroup(
	ctx context.Context,
	logger logx.Logger,
	conn squirrel.BaseRunner,
	roleID int64,
	groupID string,
) error {
	logger = logger.WithName("create-role-assignment-for-group").WithData(
		logx.Data{Key: "role.id", Value: roleID},
		logx.Data{Key: "group_assignment.group_id", Value: groupID},
	)

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
	logger logx.Logger,
	conn squirrel.BaseRunner,
	roleName string,
	groupID string,
) error {
	logger = logger.WithName("assign-role-to-group")

	foundRole, err := findRole(ctx, logger, conn, roleName)
	if err != nil {
		return err
	}
	return createRoleAssignmentForGroup(ctx, logger, conn, foundRole.ID, groupID)
}

func createRoleAssignment(
	ctx context.Context,
	logger logx.Logger,
	conn squirrel.BaseRunner,
	roleID int64,
	actorID string,
	actorNamespace string,
) error {
	logger = logger.WithName("create-role-assignment").WithData(
		logx.Data{Key: "role.id", Value: roleID},
		logx.Data{Key: "assignment.actor_id", Value: actorID},
		logx.Data{Key: "assignment.actor_namespace", Value: actorNamespace},
	)

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
	logger logx.Logger,
	conn squirrel.BaseRunner,
	roleName string,
	actorID string,
	actorNamespace string,
) error {
	logger = logger.WithName("unassign-role")

	foundRole, err := findRole(ctx, logger, conn, roleName)
	if err != nil {
		return err
	}

	return deleteRoleAssignment(ctx, logger, conn, foundRole.ID, actorID, actorNamespace)
}

func deleteRoleAssignment(
	ctx context.Context,
	logger logx.Logger,
	conn squirrel.BaseRunner,
	roleID int64,
	actorID string,
	actorNamespace string,
) error {
	logger = logger.WithName("delete-role-assignment").WithData(
		logx.Data{Key: "role.id", Value: roleID},
		logx.Data{Key: "assignment.actor_id", Value: actorID},
		logx.Data{Key: "assignment.actor_namespace", Value: actorNamespace},
	)

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
	logger logx.Logger,
	conn squirrel.BaseRunner,
	roleName string,
	groupID string,
) error {
	logger = logger.WithName("unassign-role-from-group")

	foundRole, err := findRole(ctx, logger, conn, roleName)
	if err != nil {
		return err
	}

	return deleteGroupRoleAssignment(ctx, logger, conn, foundRole.ID, groupID)
}

func deleteGroupRoleAssignment(
	ctx context.Context,
	logger logx.Logger,
	conn squirrel.BaseRunner,
	roleID int64,
	groupID string,
) error {
	logger = logger.WithName("delete-role-assignment").WithData(
		logx.Data{Key: "role.id", Value: roleID},
		logx.Data{Key: "group_assignment.group_id", Value: groupID},
	)

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
	logger logx.Logger,
	conn squirrel.BaseRunner,
	query repos.HasRoleQuery,
) (bool, error) {
	logger = logger.WithName("has-role")
	foundRole, err := findRole(ctx, logger, conn, query.RoleName)
	if err != nil {
		return false, err
	}

	return findRoleAssignment(ctx, logger, conn, foundRole.ID, query.Actor.ID, query.Actor.Namespace)
}

func hasRoleForGroup(
	ctx context.Context,
	logger logx.Logger,
	conn squirrel.BaseRunner,
	query repos.HasRoleForGroupQuery,
) (bool, error) {
	logger = logger.WithName("has-role-for-group")
	foundRole, err := findRole(ctx, logger, conn, query.RoleName)
	if err != nil {
		return false, err
	}

	return findRoleAssignmentForGroup(ctx, logger, conn, foundRole.ID, query.Group.ID)
}

func findRoleAssignment(
	ctx context.Context,
	logger logx.Logger,
	conn squirrel.BaseRunner,
	roleID int64,
	actorID string,
	actorNamespace string,
) (bool, error) {
	logger = logger.WithName("find-role-assignment").WithData(
		logx.Data{Key: "role.id", Value: roleID},
		logx.Data{Key: "assignment.actor_id", Value: actorID},
		logx.Data{Key: "assignment.actor_namespace", Value: actorNamespace},
	)

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
	logger logx.Logger,
	conn squirrel.BaseRunner,
	roleID int64,
	groupID string,
) (bool, error) {
	logger = logger.WithName("find-role-assignment-for-group").WithData(
		logx.Data{Key: "role.id", Value: roleID},
		logx.Data{Key: "group_assignment.group_id", Value: groupID},
	)

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
	logger logx.Logger,
	conn squirrel.BaseRunner,
	query repos.ListRolePermissionsQuery,
) ([]permission, error) {
	logger = logger.WithName("list-role-permissions")

	foundRole, err := findRole(ctx, logger, conn, query.RoleName)
	if err != nil {
		return nil, err
	}

	return findRolePermissions(ctx, logger, conn, foundRole.ID)
}

func createAction(
	ctx context.Context,
	logger logx.Logger,
	conn squirrel.BaseRunner,
	name string,
) (action, error) {
	logger = logger.WithName("create-permission-definition")
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
			return action{}, err2
		}

		return action{
			ID: actionID,
			Action: perm.Action{
				Name: name,
			},
		}, nil
	case *mysql.MySQLError:
		if e.Number == MySQLErrorCodeDuplicateKey {
			logger.Debug(errActionAlreadyExists)
			return action{}, errActionAlreadyExistsInDB
		}

		logger.Error(failedToCreateAction, err)
		return action{}, err
	default:
		logger.Error(failedToCreateAction, err)
		return action{}, err
	}
}

func findAction(
	ctx context.Context,
	logger logx.Logger,
	conn squirrel.BaseRunner,
	actionName string,
) (action, error) {
	logger = logger.WithName("find-permission-definition")

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
		return action{
			ID: actionID,
			Action: perm.Action{
				Name: name,
			},
		}, nil
	case sql.ErrNoRows:
		logger.Debug(errActionNotFound)
		return action{}, errActionNotFoundDB
	default:
		logger.Error(failedToFindAction, err)
		return action{}, err
	}
}

func findRolePermissions(
	ctx context.Context,
	logger logx.Logger,
	conn squirrel.BaseRunner,
	roleID int64,
) ([]permission, error) {
	logger = logger.WithName("find-role-permissions").WithData(logx.Data{Key: "role.id", Value: roleID})

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

	var permissions []permission
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
			Permission: perm.Permission{
				Action:          action,
				ResourcePattern: resourcePattern,
			},
		}
		permissions = append(permissions, p)
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
	logger logx.Logger,
	conn squirrel.BaseRunner,
	query repos.HasPermissionQuery,
) (bool, error) {
	logger = logger.WithName("has-permission").WithData(
		logx.Data{Key: "actor.issuer", Value: query.Actor.Namespace},
		logx.Data{Key: "assignment.actor_id", Value: query.Actor.ID},
		logx.Data{Key: "permission.action", Value: query.Action},
		logx.Data{Key: "permission.resourcePattern", Value: query.ResourcePattern},
		logx.Data{Key: "group_assignment.groups", Value: query.Actor.Groups},
	)

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
	for _, group := range query.Actor.Groups {
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
	logger logx.Logger,
	conn squirrel.BaseRunner,
	actionID int64,
	roleID int64,
	resourcePattern string,
	action string,
) (permission, error) {
	logger = logger.WithName("create-permission-definition")
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
			return permission{}, err2
		}

		return permission{
			ID: permissionID,
			Permission: perm.Permission{
				Action:          action,
				ResourcePattern: resourcePattern,
			},
		}, nil
	case *mysql.MySQLError:
		if e.Number == MySQLErrorCodeDuplicateKey {
			logger.Debug(errPermissionAlreadyExists)
			return permission{}, errPermissionAlreadyExistsDB
		}

		logger.Error(failedToCreatePermission, err)
		return permission{}, err
	default:
		logger.Error(failedToCreatePermission, err)
		return permission{}, err
	}
}

func listResourcePatterns(
	ctx context.Context,
	logger logx.Logger,
	conn squirrel.BaseRunner,
	query repos.ListResourcePatternsQuery,
) ([]string, error) {
	action := query.Action
	actorNamespace := query.Actor.Namespace
	actorID := query.Actor.ID

	logger = logger.WithName("list-resource-patterns").WithData(
		logx.Data{Key: "assignment.actor_namespace", Value: actorNamespace},
		logx.Data{Key: "assignment.actor_id", Value: actorID},
		logx.Data{Key: "permission.action", Value: action},
	)

	rows, err := squirrel.Select("permission.resource_pattern").
		Distinct().
		From("role").
		Join("assignment ON role.id = assignment.role_id").
		Join("permission ON permission.role_id = role.id").
		Join("action ON permission.action_id = action.id").
		Where(squirrel.Eq{
			"action.name":                action,
			"assignment.actor_id":        actorID,
			"assignment.actor_namespace": actorNamespace,
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

	var groupIDs []string
	for _, group := range query.Actor.Groups {
		groupIDs = append(groupIDs, group.ID)
	}

	gRows, err := squirrel.Select("permission.resource_pattern").
		Distinct().
		From("role").
		Join("group_assignment ON role.id = group_assignment.role_id").
		Join("permission ON permission.role_id = role.id").
		Join("action ON permission.action_id = action.id").
		Where(squirrel.Eq{
			"action.name":               action,
			"group_assignment.group_id": groupIDs,
		}).
		RunWith(conn).
		QueryContext(ctx)

	if err != nil {
		logger.Error(failedToListResourcePatterns, err)
		return nil, err
	}
	defer gRows.Close()

	for gRows.Next() {
		var resourcePattern string

		err = gRows.Scan(&resourcePattern)
		if err != nil {
			logger.Error(failedToScanRow, err)
			return nil, err
		}

		resourcePatterns = append(resourcePatterns, resourcePattern)
	}

	err = gRows.Err()
	if err != nil {
		logger.Error(failedToIterateOverRows, err)
		return nil, err
	}

	return resourcePatterns, nil
}
