package db

import (
	"context"
	"database/sql"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
	"code.cloudfoundry.org/perm/models"
	"github.com/Masterminds/squirrel"
	"github.com/go-sql-driver/mysql"
	"github.com/satori/go.uuid"
)

type DataService struct {
	conn *sql.DB
}

func NewDataService(conn *sql.DB) *DataService {
	return &DataService{
		conn: conn,
	}
}

func createRole(ctx context.Context, logger lager.Logger, conn squirrel.BaseRunner, name string) (*role, error) {
	logger = logger.Session("create-role")
	u := uuid.NewV4().Bytes()

	result, err := squirrel.Insert("role").
		Columns("uuid", "name").
		Values(u, name).
		RunWith(conn).
		ExecContext(ctx)

	switch e := err.(type) {
	case nil:
		id, err := result.LastInsertId()
		if err != nil {
			logger.Error(messages.FailedToRetrieveID, err)
			return nil, err
		}

		role := &role{
			ID: id,
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

func findRole(ctx context.Context, logger lager.Logger, conn squirrel.BaseRunner, query models.RoleQuery) (*role, error) {
	logger = logger.Session("find-role")

	var (
		id   int64
		name string
	)

	err := squirrel.Select("id", "name").
		From("role").
		Where(squirrel.Eq{
			"name": query.Name,
		}).
		RunWith(conn).
		ScanContext(ctx, &id, &name)

	switch err {
	case nil:
		return &role{
			ID: id,
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

func deleteRole(ctx context.Context, logger lager.Logger, conn squirrel.BaseRunner, query models.RoleQuery) error {
	logger = logger.Session("delete-role")
	result, err := squirrel.Delete("role").
		Where(squirrel.Eq{
			"name": query.Name,
		}).
		RunWith(conn).
		ExecContext(ctx)

	switch err {
	case nil:
		n, err := result.RowsAffected()
		if err != nil {
			logger.Error(messages.FailedToCountRowsAffected, err)
			return err
		}

		if n == 0 {
			logger.Debug(messages.ErrRoleNotFound)
			return models.ErrRoleNotFound
		}

		return nil
	case sql.ErrNoRows:
		return models.ErrRoleNotFound
	default:
		return err
	}
}

func createActor(ctx context.Context, logger lager.Logger, conn squirrel.BaseRunner, domainID, issuer string) (*actor, error) {
	logger = logger.Session("create-actor")

	u := uuid.NewV4().Bytes()

	result, err := squirrel.Insert("actor").
		Columns("uuid", "domain_id", "issuer").
		Values(u, domainID, issuer).
		RunWith(conn).
		ExecContext(ctx)

	switch e := err.(type) {
	case nil:
		id, err := result.LastInsertId()
		if err != nil {
			logger.Error(messages.FailedToRetrieveID, err)
			return nil, err
		}
		actor := &actor{
			ID: id,
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

func findActor(ctx context.Context, logger lager.Logger, conn squirrel.BaseRunner, query models.ActorQuery) (*actor, error) {
	logger = logger.Session("find-actor")

	sQuery := squirrel.Eq{}
	if query.DomainID != "" {
		sQuery["domain_id"] = query.DomainID
	}
	if query.Issuer != "" {
		sQuery["issuer"] = query.Issuer
	}

	var (
		id       int64
		domainID string
		issuer   string
	)
	err := squirrel.Select("id", "domain_id", "issuer").
		From("actor").
		Where(sQuery).
		RunWith(conn).
		ScanContext(ctx, &id, &domainID, &issuer)

	switch err {
	case nil:
		return &actor{
			ID: id,
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

func assignRole(ctx context.Context, logger lager.Logger, conn squirrel.BaseRunner, roleName, domainID, issuer string) error {
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

func createRoleAssignment(ctx context.Context, logger lager.Logger, conn squirrel.BaseRunner, roleID, actorID int64) error {
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

func unassignRole(ctx context.Context, logger lager.Logger, conn squirrel.BaseRunner, roleName, domainID, issuer string) error {
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

func deleteRoleAssignment(ctx context.Context, logger lager.Logger, conn squirrel.BaseRunner, roleID, actorID int64) error {
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
		n, err := result.RowsAffected()
		if err != nil {
			logger.Error(messages.FailedToDeleteRoleAssignment, err)
			return err
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

func hasRole(ctx context.Context, logger lager.Logger, conn squirrel.BaseRunner, query models.RoleAssignmentQuery) (bool, error) {
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

func findRoleAssignment(ctx context.Context, logger lager.Logger, conn squirrel.BaseRunner, roleID, actorID int64) (bool, error) {
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

func listActorRoles(ctx context.Context, logger lager.Logger, conn squirrel.BaseRunner, query models.ActorQuery) ([]*role, error) {
	logger = logger.Session("list-actor-roles")

	actor, err := findActor(ctx, logger, conn, query)
	if err != nil {
		return nil, err
	}

	return findActorRoleAssignments(ctx, logger, conn, actor.ID)
}

func findActorRoleAssignments(ctx context.Context, logger lager.Logger, conn squirrel.BaseRunner, actorID int64) ([]*role, error) {
	logger = logger.Session("find-actor-role-assignments").WithData(lager.Data{
		"actor.id": actorID,
	})

	rows, err := squirrel.Select("r.id", "r.name").From("role_assignment ra").Join("role r ON ra.role_id = r.id").
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
			id   int64
			name string
		)
		err := rows.Scan(&id, &name)
		if err != nil {
			logger.Error(messages.FailedToScanRow, err)
			return nil, err
		}

		roles = append(roles, &role{ID: id, Role: &models.Role{Name: name}})
	}

	err = rows.Err()
	if err != nil {
		logger.Error(messages.FailedToIterateOverRows, err)
		return nil, err
	}

	return roles, nil
}
