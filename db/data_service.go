package db

import (
	"context"
	"database/sql"

	"code.cloudfoundry.org/lager"
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

func (s *DataService) CreateRole(ctx context.Context, logger lager.Logger, name string) (*models.Role, error) {
	return createRole(ctx, s.conn, name)
}

func (s *DataService) FindRole(ctx context.Context, logger lager.Logger, query models.RoleQuery) (*models.Role, error) {
	return findRole(ctx, s.conn, query)
}

func (s *DataService) DeleteRole(ctx context.Context, logger lager.Logger, query models.RoleQuery) error {
	return deleteRole(ctx, s.conn, query)
}

func createRole(ctx context.Context, conn *sql.DB, name string) (*models.Role, error) {
	u := uuid.NewV4().Bytes()
	role := &models.Role{
		Name: name,
	}

	_, err := squirrel.Insert("role").
		Columns("uuid", "name").
		Values(u, name).
		RunWith(conn).
		ExecContext(ctx)

	switch e := err.(type) {
	case nil:
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

func findRole(ctx context.Context, conn *sql.DB, query models.RoleQuery) (*models.Role, error) {
	var name string

	err := squirrel.Select("name").
		From("role").
		Where(squirrel.Eq{
			"name": query.Name,
		}).
		RunWith(conn).
		ScanContext(ctx, &name)

	switch err {
	case nil:
		return &models.Role{Name: name}, nil
	case sql.ErrNoRows:
		return nil, models.ErrRoleNotFound
	default:
		return nil, err
	}
}

func deleteRole(ctx context.Context, conn *sql.DB, query models.RoleQuery) error {
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
