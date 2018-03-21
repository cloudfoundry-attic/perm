package db

import (
	"context"
	"database/sql"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/models"
	"code.cloudfoundry.org/perm/pkg/sqlx"
	"code.cloudfoundry.org/perm/repos"
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
	roleName models.RoleName,
	permissions ...*models.Permission,
) (*role, error) {
	role, err := createRole(ctx, logger, conn, roleName)
	if err != nil {
		return nil, err
	}

	for _, permission := range permissions {
		permissionName := models.PermissionName(permission.Name)
		_, err = createPermissionDefinition(ctx, logger, conn, permissionName)
		if err != nil && err != models.ErrPermissionDefinitionAlreadyExists {
			return nil, err
		}

		var p *permissionDefinition
		p, err = findPermissionDefinition(ctx, logger, conn, permissionName)
		if err != nil {
			return nil, err
		}

		_, err = createPermission(ctx, logger, conn, p.ID, role.ID, permission.ResourcePattern, permissionName)
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
	name models.RoleName,
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
			ID: id(roleID),
			Role: &models.Role{
				Name: name,
			},
		}
		return role, nil
	case *mysql.MySQLError:
		if e.Number == MySQLErrorCodeDuplicateKey {
			logger.Debug(errRoleAlreadyExists)
			return nil, models.ErrRoleAlreadyExists
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
		roleID   id
		roleName models.RoleName
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
			Role: &models.Role{
				Name: roleName,
			},
		}, nil
	case sql.ErrNoRows:
		logger.Debug(errRoleNotFound)
		return nil, models.ErrRoleNotFound
	default:
		logger.Error(failedToFindRole, err)
		return nil, err
	}
}

func deleteRole(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	roleName models.RoleName,
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
			return models.ErrRoleNotFound
		}

		return nil
	case sql.ErrNoRows:
		logger.Debug(errRoleNotFound)
		return models.ErrRoleNotFound
	default:
		logger.Error(failedToDeleteRole, err)
		return err
	}
}

func createActor(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	domainID models.ActorDomainID,
	issuer models.ActorIssuer,
) (*actor, error) {
	logger = logger.Session("create-actor")

	u := uuid.NewV4().Bytes()

	result, err := squirrel.Insert("actor").
		Columns("uuid", "domain_id", "issuer").
		Values(u, domainID, issuer).
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
			ID: id(actorID),
			Actor: &models.Actor{
				DomainID: domainID,
				Issuer:   issuer,
			},
		}
		return actor, nil
	case *mysql.MySQLError:
		if e.Number == MySQLErrorCodeDuplicateKey {
			logger.Debug(errActorAlreadyExists)
			return nil, models.ErrActorAlreadyExists
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
	domainID models.ActorDomainID,
	issuer models.ActorIssuer,
) (id, error) {
	logger = logger.Session("find-actor")

	sQuery := squirrel.Eq{}
	if domainID != "" {
		sQuery["domain_id"] = domainID
	}
	if issuer != "" {
		sQuery["issuer"] = issuer
	}

	var (
		actorID id
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
		return actorID, models.ErrActorNotFound
	default:
		logger.Error(failedToFindActor, err)
		return actorID, err
	}
}

func assignRole(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	roleName models.RoleName,
	domainID models.ActorDomainID,
	issuer models.ActorIssuer,
) error {
	logger = logger.Session("assign-role")

	role, err := findRole(ctx, logger, conn, repos.FindRoleQuery{RoleName: roleName})
	if err != nil {
		return err
	}

	_, err = createActor(ctx, logger, conn, domainID, issuer)
	if err != nil && err != models.ErrActorAlreadyExists {
		return err
	}

	actorID, err := findActorID(ctx, logger, conn, domainID, issuer)
	if err != nil {
		return err
	}

	return createRoleAssignment(ctx, logger, conn, role.ID, actorID)
}

func createRoleAssignment(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	roleID, actorID id,
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
			return models.ErrRoleAssignmentAlreadyExists
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
	roleName models.RoleName,
	domainID models.ActorDomainID,
	issuer models.ActorIssuer,
) error {
	logger = logger.Session("unassign-role")

	role, err := findRole(ctx, logger, conn, repos.FindRoleQuery{RoleName: roleName})
	if err != nil {
		return err
	}

	actorID, err := findActorID(ctx, logger, conn, domainID, issuer)
	if err == models.ErrActorNotFound {
		return models.ErrRoleAssignmentNotFound
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
	actorID id,
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
			return models.ErrRoleAssignmentNotFound
		}

		return nil
	case sql.ErrNoRows:
		logger.Debug(errRoleAssignmentNotFound)
		return models.ErrRoleAssignmentNotFound
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

	actorID, err := findActorID(ctx, logger, conn, query.Actor.DomainID, query.Actor.Issuer)
	if err == models.ErrActorNotFound {
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
	roleID id,
	actorID id,
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

	actorID, err := findActorID(ctx, logger, conn, query.Actor.DomainID, query.Actor.Issuer)
	if err == models.ErrActorNotFound {
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
	actorID id,
) ([]*role, error) {
	logger = logger.Session("find-actor-role-assignments").WithData(lager.Data{
		"actor.id": actorID,
	})

	rows, err := squirrel.Select("r.id", "r.name").
		From("role_assignment ra").
		JoinClause("INNER JOIN role r ON ra.role_id = r.id").
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
			roleID id
			name   models.RoleName
		)
		e := rows.Scan(&roleID, &name)
		if e != nil {
			logger.Error(failedToScanRow, e)
			return nil, e
		}

		roles = append(roles, &role{ID: roleID, Role: &models.Role{Name: name}})
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

func createPermissionDefinition(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	name models.PermissionName,
) (*permissionDefinition, error) {
	logger = logger.Session("create-permission-definition")
	u := uuid.NewV4().Bytes()

	result, err := squirrel.Insert("permission_definition").
		Columns("uuid", "name").
		Values(u, name).
		RunWith(conn).
		ExecContext(ctx)

	switch e := err.(type) {
	case nil:
		permissionDefinitionID, err2 := result.LastInsertId()
		if err2 != nil {
			logger.Error(failedToRetrieveID, err2)
			return nil, err2
		}

		permissionDefinition := &permissionDefinition{
			ID: id(permissionDefinitionID),
			PermissionDefinition: &models.PermissionDefinition{
				Name: name,
			},
		}
		return permissionDefinition, nil
	case *mysql.MySQLError:
		if e.Number == MySQLErrorCodeDuplicateKey {
			logger.Debug(errPermissionDefinitionAlreadyExists)
			return nil, models.ErrPermissionDefinitionAlreadyExists
		}

		logger.Error(failedToCreatePermissionDefinition, err)
		return nil, err
	default:
		logger.Error(failedToCreatePermissionDefinition, err)
		return nil, err
	}
}

func findPermissionDefinition(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	permissionName models.PermissionName,
) (*permissionDefinition, error) {
	logger = logger.Session("find-permission-definition")

	var (
		permissionDefinitionID id
		name                   models.PermissionName
	)

	err := squirrel.Select("id", "name").
		From("permission_definition").
		Where(squirrel.Eq{
			"name": permissionName,
		}).
		RunWith(conn).
		ScanContext(ctx, &permissionDefinitionID, &name)

	switch err {
	case nil:
		return &permissionDefinition{
			ID: permissionDefinitionID,
			PermissionDefinition: &models.PermissionDefinition{
				Name: name,
			},
		}, nil
	case sql.ErrNoRows:
		logger.Debug(errPermissionDefinitionNotFound)
		return nil, models.ErrPermissionDefinitionNotFound
	default:
		logger.Error(failedToFindPermissionDefinition, err)
		return nil, err
	}
}

func findRolePermissions(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	roleID id,
) ([]*permission, error) {
	logger = logger.Session("find-role-permissions").WithData(lager.Data{
		"role.id": roleID,
	})

	rows, err := squirrel.Select("p.id", "pd.name", "p.resource_pattern").
		From("permission p").
		JoinClause("INNER JOIN role r ON p.role_id = r.id").
		JoinClause("INNER JOIN permission_definition pd ON p.permission_definition_id = pd.id").
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
			permissionID    id
			name            models.PermissionName
			resourcePattern models.PermissionResourcePattern
		)
		e := rows.Scan(&permissionID, &name, &resourcePattern)
		if e != nil {
			logger.Error(failedToScanRow, e)
			return nil, e
		}

		p := permission{
			ID: permissionID,
			Permission: &models.Permission{
				Name:            name,
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
		"actor.issuer":               query.Actor.Issuer,
		"actor.domainID":             query.Actor.DomainID,
		"permission.name":            query.PermissionName,
		"permission.resourcePattern": query.ResourcePattern,
	})

	var count int

	err := squirrel.Select("count(ra.role_id)").
		From("role_assignment ra").
		JoinClause("INNER JOIN actor a ON a.id = ra.actor_id").
		JoinClause("INNER JOIN permission p ON ra.role_id = p.role_id").
		JoinClause("INNER JOIN permission_definition pd ON p.permission_definition_id = pd.id").
		Where(squirrel.Eq{
			"a.issuer":           query.Actor.Issuer,
			"a.domain_id":        query.Actor.DomainID,
			"pd.name":            query.PermissionName,
			"p.resource_pattern": query.ResourcePattern,
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
	permissionDefinitionID id,
	roleID id,
	resourcePattern models.PermissionResourcePattern,
	permissionName models.PermissionName,
) (*permission, error) {
	logger = logger.Session("create-permission-definition")
	u := uuid.NewV4().Bytes()

	result, err := squirrel.Insert("permission").
		Columns("uuid", "permission_definition_id", "role_id", "resource_pattern").
		Values(u, permissionDefinitionID, roleID, resourcePattern).
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
			ID: id(permissionID),
			Permission: &models.Permission{
				Name:            permissionName,
				ResourcePattern: resourcePattern,
			},
		}
		return permission, nil
	case *mysql.MySQLError:
		if e.Number == MySQLErrorCodeDuplicateKey {
			logger.Debug(errPermissionAlreadyExists)
			return nil, models.ErrPermissionAlreadyExists
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
) ([]models.PermissionResourcePattern, error) {
	permissionName := query.PermissionName
	issuer := query.Actor.Issuer
	domainID := query.Actor.DomainID

	logger = logger.Session("list-resource-patterns").
		WithData(lager.Data{
			"actor.issuer":    issuer,
			"actor.domainID":  domainID,
			"permission.name": permissionName,
		})

	rows, err := squirrel.Select("permission.resource_pattern").
		Distinct().
		From("role").
		Join("role_assignment ON role.id = role_assignment.role_id").
		Join("actor ON actor.id = role_assignment.actor_id").
		Join("permission ON permission.role_id = role.id").
		Join("permission_definition ON permission.permission_definition_id = permission_definition.id").
		Where(squirrel.Eq{
			"permission_definition.name": permissionName,
			"actor.domain_id":            domainID,
			"actor.issuer":               issuer,
		}).
		RunWith(conn).
		QueryContext(ctx)
	if err != nil {
		logger.Error(failedToListResourcePatterns, err)
		return nil, err
	}
	defer rows.Close()

	var resourcePatterns []models.PermissionResourcePattern
	for rows.Next() {
		var resourcePattern models.PermissionResourcePattern

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
