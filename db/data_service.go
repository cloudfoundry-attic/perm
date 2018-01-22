package db

import (
	"context"
	"database/sql"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
	"code.cloudfoundry.org/perm/models"
	"code.cloudfoundry.org/perm/sqlx"
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
		permissionDefinitionName := models.PermissionDefinitionName(permission.Name)
		_, err = createPermissionDefinition(ctx, logger, conn, permissionDefinitionName)
		if err != nil && err != models.ErrPermissionDefinitionAlreadyExists {
			return nil, err
		}

		var p *permissionDefinition
		query := models.PermissionDefinitionQuery{
			Name: permissionDefinitionName,
		}
		p, err = findPermissionDefinition(ctx, logger, conn, &query)
		if err != nil {
			return nil, err
		}

		_, err = createPermission(ctx, logger, conn, p.ID, role.ID, permission.ResourcePattern, permission.Name)
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
			logger.Error(messages.FailedToRetrieveID, err2)
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
			logger.Debug(messages.ErrRoleAlreadyExists)
			return nil, models.ErrRoleAlreadyExists
		}

		logger.Error(messages.FailedToCreateRole, err)
		return nil, err
	default:
		logger.Error(messages.FailedToCreateRole, err)
		return nil, err
	}
}

func findRole(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	query models.RoleQuery,
) (*role, error) {
	logger = logger.Session("find-role")

	var (
		roleID id
		name   models.RoleName
	)

	err := squirrel.Select("id", "name").
		From("role").
		Where(squirrel.Eq{
			"name": query.Name,
		}).
		RunWith(conn).
		ScanContext(ctx, &roleID, &name)

	switch err {
	case nil:
		return &role{
			ID: roleID,
			Role: &models.Role{

				Name: name,
			},
		}, nil
	case sql.ErrNoRows:
		logger.Debug(messages.ErrRoleNotFound)
		return nil, models.ErrRoleNotFound
	default:
		logger.Error(messages.FailedToFindRole, err)
		return nil, err
	}
}

func deleteRole(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	query models.RoleQuery,
) error {
	logger = logger.Session("delete-role")
	result, err := squirrel.Delete("role").
		Where(squirrel.Eq{
			"name": query.Name,
		}).
		RunWith(conn).
		ExecContext(ctx)

	switch err {
	case nil:
		n, err2 := result.RowsAffected()
		if err2 != nil {
			logger.Error(messages.FailedToCountRowsAffected, err2)
			return err2
		}

		if n == 0 {
			logger.Debug(messages.ErrRoleNotFound)
			return models.ErrRoleNotFound
		}

		return nil
	case sql.ErrNoRows:
		logger.Debug(messages.ErrRoleNotFound)
		return models.ErrRoleNotFound
	default:
		logger.Error(messages.FailedToDeleteRole, err)
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
			logger.Error(messages.FailedToRetrieveID, err2)
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
			logger.Debug(messages.ErrActorAlreadyExists)
			return nil, models.ErrActorAlreadyExists
		}
		logger.Error(messages.FailedToCreateActor, err)
		return nil, err
	default:
		logger.Error(messages.FailedToCreateActor, err)
		return nil, err
	}
}

func findActor(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	query models.ActorQuery,
) (*actor, error) {
	logger = logger.Session("find-actor")

	sQuery := squirrel.Eq{}
	if query.DomainID != "" {
		sQuery["domain_id"] = query.DomainID
	}
	if query.Issuer != "" {
		sQuery["issuer"] = query.Issuer
	}

	var (
		actorID  id
		domainID models.ActorDomainID
		issuer   models.ActorIssuer
	)
	err := squirrel.Select("id", "domain_id", "issuer").
		From("actor").
		Where(sQuery).
		RunWith(conn).
		ScanContext(ctx, &actorID, &domainID, &issuer)

	switch err {
	case nil:
		return &actor{
			ID: actorID,
			Actor: &models.Actor{
				DomainID: domainID,
				Issuer:   issuer,
			},
		}, nil
	case sql.ErrNoRows:
		logger.Debug(messages.ErrActorNotFound)
		return nil, models.ErrActorNotFound
	default:
		logger.Error(messages.FailedToFindActor, err)
		return nil, err
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

	role, err := findRole(ctx, logger, conn, models.RoleQuery{Name: roleName})
	if err != nil {
		return err
	}

	_, err = createActor(ctx, logger, conn, domainID, issuer)
	if err != nil && err != models.ErrActorAlreadyExists {
		return err
	}

	actor, err := findActor(ctx, logger, conn, models.ActorQuery{DomainID: domainID, Issuer: issuer})
	if err != nil {
		return err
	}

	return createRoleAssignment(ctx, logger, conn, role.ID, actor.ID)
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
			logger.Debug(messages.ErrRoleAssignmentAlreadyExists)
			return models.ErrRoleAssignmentAlreadyExists
		}

		logger.Error(messages.FailedToCreateRoleAssignment, err)
		return err
	default:
		logger.Error(messages.FailedToCreateRoleAssignment, err)
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

	role, err := findRole(ctx, logger, conn, models.RoleQuery{Name: roleName})
	if err != nil {
		return err
	}

	actor, err := findActor(ctx, logger, conn, models.ActorQuery{DomainID: domainID, Issuer: issuer})
	if err != nil {
		return err
	}

	return deleteRoleAssignment(ctx, logger, conn, role.ID, actor.ID)
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
			logger.Error(messages.FailedToDeleteRoleAssignment, e)
			return e
		}

		if n == 0 {
			logger.Debug(messages.ErrRoleAssignmentNotFound)
			return models.ErrRoleAssignmentNotFound
		}

		return nil
	case sql.ErrNoRows:
		logger.Debug(messages.ErrRoleAssignmentNotFound)
		return models.ErrRoleAssignmentNotFound
	default:
		logger.Error(messages.FailedToDeleteRoleAssignment, err)
		return err
	}
}

func hasRole(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	query models.RoleAssignmentQuery,
) (bool, error) {
	logger = logger.Session("has-role")

	actor, err := findActor(ctx, logger, conn, query.ActorQuery)
	if err != nil {
		return false, err
	}

	role, err := findRole(ctx, logger, conn, query.RoleQuery)
	if err != nil {
		return false, err
	}

	return findRoleAssignment(ctx, logger, conn, role.ID, actor.ID)
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
		logger.Error(messages.FailedToFindRoleAssignment, err)
		return false, err
	}
}

func listActorRoles(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	query models.ActorQuery,
) ([]*role, error) {
	logger = logger.Session("list-actor-roles")

	actor, err := findActor(ctx, logger, conn, query)
	if err != nil {
		return nil, err
	}

	return findActorRoleAssignments(ctx, logger, conn, actor.ID)
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
		logger.Error(messages.FailedToFindRoleAssignments, err)
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
			logger.Error(messages.FailedToScanRow, e)
			return nil, e
		}

		roles = append(roles, &role{ID: roleID, Role: &models.Role{Name: name}})
	}

	err = rows.Err()
	if err != nil {
		logger.Error(messages.FailedToIterateOverRows, err)
		return nil, err
	}

	return roles, nil
}

func listRolePermissions(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	query models.RoleQuery,
) ([]*permission, error) {
	logger = logger.Session("list-role-permissions")

	role, err := findRole(ctx, logger, conn, query)
	if err != nil {
		return nil, err
	}

	return findRolePermissions(ctx, logger, conn, role.ID)
}

func createPermissionDefinition(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	name models.PermissionDefinitionName,
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
			logger.Error(messages.FailedToRetrieveID, err2)
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
			logger.Debug(messages.ErrPermissionDefinitionAlreadyExists)
			return nil, models.ErrPermissionDefinitionAlreadyExists
		}

		logger.Error(messages.FailedToCreatePermissionDefinition, err)
		return nil, err
	default:
		logger.Error(messages.FailedToCreatePermissionDefinition, err)
		return nil, err
	}
}

func findPermissionDefinition(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	query *models.PermissionDefinitionQuery,
) (*permissionDefinition, error) {
	logger = logger.Session("find-permission-definition")

	var (
		permissionDefinitionID id
		name                   models.PermissionDefinitionName
	)

	err := squirrel.Select("id", "name").
		From("permission_definition").
		Where(squirrel.Eq{
			"name": query.Name,
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
		logger.Debug(messages.ErrPermissionDefinitionNotFound)
		return nil, models.ErrPermissionDefinitionNotFound
	default:
		logger.Error(messages.FailedToFindPermissionDefinition, err)
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
		logger.Error(messages.FailedToFindPermissions, err)
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
			logger.Error(messages.FailedToScanRow, e)
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
		logger.Error(messages.FailedToIterateOverRows, err)
		return nil, err
	}

	return permissions, nil
}

func hasPermission(
	ctx context.Context,
	logger lager.Logger,
	conn squirrel.BaseRunner,
	query models.HasPermissionQuery,
) (bool, error) {
	logger = logger.Session("has-permission").WithData(lager.Data{
		"actor.issuer":          query.ActorQuery.Issuer,
		"actor.domainID":        query.ActorQuery.DomainID,
		"permission.name":       query.PermissionQuery.PermissionName,
		"permission.resourceID": query.PermissionQuery.ResourceID,
	})

	var count int

	err := squirrel.Select("count(ra.role_id)").
		From("role_assignment ra").
		JoinClause("INNER JOIN actor a ON a.id = ra.actor_id").
		JoinClause("INNER JOIN permission p ON ra.role_id = p.role_id").
		JoinClause("INNER JOIN permission_definition pd ON p.permission_definition_id = pd.id").
		Where(squirrel.Eq{
			"a.issuer":           query.ActorQuery.Issuer,
			"a.domain_id":        query.ActorQuery.DomainID,
			"pd.name":            query.PermissionQuery.PermissionName,
			"p.resource_pattern": query.PermissionQuery.ResourceID,
		}).
		RunWith(conn).
		ScanContext(ctx, &count)

	if err != nil {
		logger.Error(messages.FailedToFindPermissions, err)
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
			logger.Error(messages.FailedToRetrieveID, err2)
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
			logger.Debug(messages.ErrPermissionAlreadyExists)
			return nil, models.ErrPermissionAlreadyExists
		}

		logger.Error(messages.FailedToCreatePermission, err)
		return nil, err
	default:
		logger.Error(messages.FailedToCreatePermission, err)
		return nil, err
	}
}
