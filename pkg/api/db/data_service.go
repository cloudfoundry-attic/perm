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
	query repos.FindRoleQuery,
) (*role, error) {
	logger = logger.Session("find-role")

	var (
		roleID   int64
		roleName string
	)

	err := squirrel.Select("id", "name").
		From("role").
		Where(squirrel.Eq{
			"name": query.RoleName,
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

func createActor(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	id string,
	namespace string,
) (*actor, error) {
	logger = logger.Session("create-actor")

	u := uuid.NewV4().Bytes()

	result, err := squirrel.Insert("actor").
		Columns("uuid", "domain_id", "issuer").
		Values(u, id, namespace).
		RunWith(conn).
		ExecContext(ctx)

	switch e := err.(type) {
	case nil:
		actorID, err2 := result.LastInsertId()
		if err2 != nil {
			logger.Error(failedToRetrieveID, err2)
			return nil, err2
		}
		actor := &actor{
			ID: actorID,
			Actor: &perm.Actor{
				ID:        id,
				Namespace: namespace,
			},
		}
		return actor, nil
	case *mysql.MySQLError:
		if e.Number == MySQLErrorCodeDuplicateKey {
			logger.Debug(errActorAlreadyExists)
			return nil, perm.ErrActorAlreadyExists
		}
		logger.Error(failedToCreateActor, err)
		return nil, err
	default:
		logger.Error(failedToCreateActor, err)
		return nil, err
	}
}

func findActorID(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	id string,
	namespace string,
) (int64, error) {
	logger = logger.Session("find-actor")

	sQuery := squirrel.Eq{}
	if id != "" {
		sQuery["domain_id"] = id
	}
	if namespace != "" {
		sQuery["issuer"] = namespace
	}

	var (
		actorID int64
	)
	err := squirrel.Select("id").
		From("actor").
		Where(sQuery).
		RunWith(conn).
		ScanContext(ctx, &actorID)

	switch err {
	case nil:
		return actorID, nil
	case sql.ErrNoRows:
		logger.Debug(errActorNotFound)
		return actorID, errActorNotFoundDB
	default:
		logger.Error(failedToFindActor, err)
		return actorID, err
	}
}

func assignRole(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	roleName string,
	id string,
	namespace string,
) error {
	logger = logger.Session("assign-role")

	role, err := findRole(ctx, logger, conn, repos.FindRoleQuery{RoleName: roleName})
	if err != nil {
		return err
	}

	_, err = createActor(ctx, logger, conn, id, namespace)
	if err != nil && err != perm.ErrActorAlreadyExists {
		return err
	}

	actorID, err := findActorID(ctx, logger, conn, id, namespace)
	if err != nil {
		return err
	}

	return createRoleAssignment(ctx, logger, conn, role.ID, actorID)
}

func createRoleAssignment(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	roleID, actorID int64,
) error {
	logger = logger.Session("create-role-assignment").WithData(lager.Data{
		"role.id":  roleID,
		"actor.id": actorID,
	})

	_, err := squirrel.Insert("role_assignment").
		Columns("role_id", "actor_id").
		Values(roleID, actorID).
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
	id string,
	namespace string,
) error {
	logger = logger.Session("unassign-role")

	role, err := findRole(ctx, logger, conn, repos.FindRoleQuery{RoleName: roleName})
	if err != nil {
		return err
	}

	actorID, err := findActorID(ctx, logger, conn, id, namespace)
	if err == errActorNotFoundDB {
		return perm.ErrAssignmentNotFound
	} else if err != nil {
		return err
	}

	return deleteRoleAssignment(ctx, logger, conn, role.ID, actorID)
}

func deleteRoleAssignment(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	roleID,
	actorID int64,
) error {
	logger = logger.Session("delete-role-assignment").WithData(lager.Data{
		"role.id":  roleID,
		"actor.id": actorID,
	})

	result, err := squirrel.Delete("role_assignment").
		Where(squirrel.Eq{
			"role_id":  roleID,
			"actor_id": actorID,
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

	actorID, err := findActorID(ctx, logger, conn, query.Actor.ID, query.Actor.Namespace)
	if err == errActorNotFoundDB {
		return false, nil
	} else if err != nil {
		return false, err
	}

	findRoleQuery := repos.FindRoleQuery{
		RoleName: query.RoleName,
	}
	role, err := findRole(ctx, logger, conn, findRoleQuery)
	if err != nil {
		return false, err
	}

	return findRoleAssignment(ctx, logger, conn, role.ID, actorID)
}

func findRoleAssignment(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	roleID int64,
	actorID int64,
) (bool, error) {
	logger = logger.Session("find-role-assignment").WithData(lager.Data{
		"role.id":  roleID,
		"actor.id": actorID,
	})

	err := squirrel.Select("actor_id").
		From("role_assignment").
		Where(squirrel.Eq{"actor_id": actorID, "role_id": roleID}).
		RunWith(conn).
		ScanContext(ctx, &actorID)

	switch err {
	case nil:
		return true, nil
	case sql.ErrNoRows:
		return false, nil
	default:
		logger.Error(failedToFindRoleAssignment, err)
		return false, err
	}
}

func listActorRoles(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	query repos.ListActorRolesQuery,
) ([]*role, error) {
	logger = logger.Session("list-actor-roles")

	actorID, err := findActorID(ctx, logger, conn, query.Actor.ID, query.Actor.Namespace)
	if err == errActorNotFoundDB {
		return []*role{}, nil
	} else if err != nil {
		return nil, err
	}

	return findActorRoleAssignments(ctx, logger, conn, actorID)
}

func findActorRoleAssignments(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	actorID int64,
) ([]*role, error) {
	logger = logger.Session("find-actor-role-assignments").WithData(lager.Data{
		"actor.id": actorID,
	})

	rows, err := squirrel.Select("role.id", "role.name").
		From("role_assignment").
		JoinClause("INNER JOIN role ON role_assignment.role_id = role.id").
		Where(squirrel.Eq{"actor_id": actorID}).
		RunWith(conn).
		QueryContext(ctx)
	if err != nil {
		logger.Error(failedToFindRoleAssignments, err)
		return nil, err
	}
	defer rows.Close()

	var roles []*role
	for rows.Next() {
		var (
			roleID int64
			action string
		)
		e := rows.Scan(&roleID, &action)
		if e != nil {
			logger.Error(failedToScanRow, e)
			return nil, e
		}

		roles = append(roles, &role{ID: roleID, Role: &perm.Role{Name: action}})
	}

	err = rows.Err()
	if err != nil {
		logger.Error(failedToIterateOverRows, err)
		return nil, err
	}

	return roles, nil
}

func listRolePermissions(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	query repos.ListRolePermissionsQuery,
) ([]*permission, error) {
	logger = logger.Session("list-role-permissions")

	role, err := findRole(ctx, logger, conn, repos.FindRoleQuery{RoleName: query.RoleName})
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
		"actor.id":                   query.Actor.ID,
		"permission.action":          query.Action,
		"permission.resourcePattern": query.ResourcePattern,
	})

	var count int

	err := squirrel.Select("count(role_assignment.role_id)").
		From("role_assignment").
		JoinClause("INNER JOIN actor ON actor.id = role_assignment.actor_id").
		JoinClause("INNER JOIN permission permission ON role_assignment.role_id = permission.role_id").
		JoinClause("INNER JOIN action ON permission.action_id = action.id").
		Where(squirrel.Eq{
			"actor.issuer":                query.Actor.Namespace,
			"actor.domain_id":             query.Actor.ID,
			"action.name":                 query.Action,
			"permission.resource_pattern": query.ResourcePattern,
		}).
		RunWith(conn).
		ScanContext(ctx, &count)

	if err != nil {
		logger.Error(failedToFindPermissions, err)
		return false, err
	}

	if count == 0 {
		return false, nil
	}

	return true, nil
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
			"actor.issuer":      namespace,
			"actor.id":          id,
			"permission.action": action,
		})

	rows, err := squirrel.Select("permission.resource_pattern").
		Distinct().
		From("role").
		Join("role_assignment ON role.id = role_assignment.role_id").
		Join("actor ON actor.id = role_assignment.actor_id").
		Join("permission ON permission.role_id = role.id").
		Join("action ON permission.action_id = action.id").
		Where(squirrel.Eq{
			"action.name":     action,
			"actor.domain_id": id,
			"actor.issuer":    namespace,
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
