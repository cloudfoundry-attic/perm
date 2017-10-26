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
	conn *sql.DB
}

func NewDataService(conn *sql.DB) *DataService {
	return &DataService{
		conn: conn,
	}
}

func (s *DataService) CreateRole(ctx context.Context, logger lager.Logger, name string) (*models.Role, error) {
	role, err := createRole(ctx, s.conn, name)
	if err != nil {
		return nil, err
	}

	return role.Role, nil
}

func (s *DataService) FindRole(ctx context.Context, logger lager.Logger, query models.RoleQuery) (*models.Role, error) {
	role, err := findRole(ctx, s.conn, query)
	if err != nil {
		return nil, err
	}
	return role.Role, nil
}

func (s *DataService) DeleteRole(ctx context.Context, logger lager.Logger, query models.RoleQuery) error {
	return deleteRole(ctx, s.conn, query)
}

func (s *DataService) AssignRole(ctx context.Context, logger lager.Logger, roleName string, domainID string, issuer string) (err error) {
	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(messages.FailedToStartTransaction, err)
		return
	}

	defer func() {
		if err != nil {
			// TODO log stuff
			return
		}
		err = sqlx.Commit(logger, tx, err)
	}()

	err = assignRole(ctx, tx, roleName, domainID, issuer)

	return
}

func (s *DataService) UnassignRole(ctx context.Context, logger lager.Logger, roleName string, domainID string, issuer string) (err error) {
	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(messages.FailedToStartTransaction, err)
		return
	}

	defer func() {
		if err != nil {
			// TODO log stuff
			return
		}
		err = sqlx.Commit(logger, tx, err)
	}()

	err = unassignRole(ctx, tx, roleName, domainID, issuer)

	return
}

func (s *DataService) HasRole(ctx context.Context, logger lager.Logger, query models.RoleAssignmentQuery) (bool, error) {
	return hasRole(ctx, s.conn, query)
}

func (s *DataService) ListActorRoles(ctx context.Context, logger lager.Logger, query models.ActorQuery) ([]*models.Role, error) {
	actor, err := findActor(ctx, s.conn, query)
	if err != nil {
		return nil, err
	}

	rows, err := squirrel.Select("r.name").From("role_assignment ra").Join("role r ON ra.role_id = r.id").
		Where(squirrel.Eq{"actor_id": actor.ID}).
		RunWith(s.conn).
		QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			return nil, err
		}

		roles = append(roles, &models.Role{Name: name})
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (s *DataService) CreateActor(ctx context.Context, logger lager.Logger, domainID, issuer string) (*models.Actor, error) {
	actor, err := createActor(ctx, s.conn, domainID, issuer)
	if err != nil {
		return nil, err
	}

	return actor.Actor, nil
}

func (s *DataService) FindActor(ctx context.Context, logger lager.Logger, query models.ActorQuery) (*models.Actor, error) {
	actor, err := findActor(ctx, s.conn, query)
	if err != nil {
		return nil, err
	}

	return actor.Actor, nil
}

func createRole(ctx context.Context, conn squirrel.BaseRunner, name string) (*role, error) {
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
			return nil, models.ErrRoleAlreadyExists
		}
		return nil, err
	default:
		return nil, err
	}
}

func findRole(ctx context.Context, conn squirrel.BaseRunner, query models.RoleQuery) (*role, error) {
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
		return nil, models.ErrRoleNotFound
	default:
		return nil, err
	}
}

func deleteRole(ctx context.Context, conn squirrel.BaseRunner, query models.RoleQuery) error {
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
			return err
		}

		if n == 0 {
			return models.ErrRoleNotFound
		}

		return nil
	case sql.ErrNoRows:
		return models.ErrRoleNotFound
	default:
		return err
	}
}

func createActor(ctx context.Context, conn squirrel.BaseRunner, domainID, issuer string) (*actor, error) {
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
			return nil, models.ErrActorAlreadyExists
		}
		return nil, err
	default:
		return nil, err
	}
}

func findActor(ctx context.Context, conn squirrel.BaseRunner, query models.ActorQuery) (*actor, error) {
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
		return nil, models.ErrActorNotFound
	default:
		return nil, err
	}
}

func assignRole(ctx context.Context, conn squirrel.BaseRunner, roleName, domainID, issuer string) error {
	role, err := findRole(ctx, conn, models.RoleQuery{Name: roleName})
	if err != nil {
		return err
	}

	_, err = createActor(ctx, conn, domainID, issuer)
	if err != nil && err != models.ErrActorAlreadyExists {
		return err
	}

	actor, err := findActor(ctx, conn, models.ActorQuery{DomainID: domainID, Issuer: issuer})
	if err != nil {
		return err
	}

	_, err = squirrel.Insert("role_assignment").
		Columns("role_id", "actor_id").
		Values(role.ID, actor.ID).
		RunWith(conn).
		ExecContext(ctx)

	if e, ok := err.(*mysql.MySQLError); ok && e.Number == MySQLErrorCodeDuplicateKey {
		return models.ErrRoleAssignmentAlreadyExists
	}

	return nil
}

func unassignRole(ctx context.Context, conn squirrel.BaseRunner, roleName, domainID, issuer string) error {
	role, err := findRole(ctx, conn, models.RoleQuery{Name: roleName})
	if err != nil {
		return err
	}

	actor, err := findActor(ctx, conn, models.ActorQuery{DomainID: domainID, Issuer: issuer})
	if err != nil {
		return err
	}

	result, err := squirrel.Delete("role_assignment").
		Where(squirrel.Eq{
			"role_id":  role.ID,
			"actor_id": actor.ID,
		}).
		RunWith(conn).
		ExecContext(ctx)

	switch err {
	case nil:
		n, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if n == 0 {
			return models.ErrRoleAssignmentNotFound
		}

		return nil
	case sql.ErrNoRows:
		return models.ErrRoleAssignmentNotFound
	default:
		return err
	}
}

func hasRole(ctx context.Context, conn squirrel.BaseRunner, query models.RoleAssignmentQuery) (bool, error) {
	actor, err := findActor(ctx, conn, query.ActorQuery)
	if err != nil {
		return false, models.ErrActorNotFound
	}

	role, err := findRole(ctx, conn, query.RoleQuery)
	if err != nil {
		return false, models.ErrRoleNotFound
	}

	var actorID int64
	err = squirrel.Select("actor_id").
		From("role_assignment").
		Where(squirrel.Eq{"actor_id": actor.ID, "role_id": role.ID}).
		RunWith(conn).
		ScanContext(ctx, &actorID)

	switch err {
	case nil:
		return true, nil
	case sql.ErrNoRows:
		return false, nil
	default:
		return false, err
	}
}
