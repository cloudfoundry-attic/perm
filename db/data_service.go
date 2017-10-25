package db

import (
	"context"
	"database/sql"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/models"
	"code.cloudfoundry.org/perm/sqlx"
	"github.com/Masterminds/squirrel"
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

func (s *DataService) CreateRole(ctx context.Context, logger lager.Logger, name string) (role *models.Role, err error) {
	var tx *sql.Tx
	tx, err = s.conn.Begin()
	if err != nil {
		return nil, err
	}
	u := uuid.NewV4().Bytes()
	role = &models.Role{
		Name: name,
	}

	defer func() {
		if err != nil {
			logger.Error("failed-to-insert-role", err)
		}
		err = sqlx.Commit(logger, tx, err)
	}()

	_, err = squirrel.Insert("role").
		Columns("uuid", "name").
		Values(u, name).
		RunWith(tx).
		ExecContext(ctx)

	return
}
func (s *DataService) FindRole(context.Context, lager.Logger, models.RoleQuery) (*models.Role, error) {
	return nil, nil
}
func (s *DataService) DeleteRole(context.Context, lager.Logger, models.RoleQuery) error {
	return nil
}
