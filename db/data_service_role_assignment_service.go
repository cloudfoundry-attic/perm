package db

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
	"code.cloudfoundry.org/perm/models"
	"code.cloudfoundry.org/perm/sqlx"
	"github.com/Masterminds/squirrel"
)

func (s *DataService) AssignRole(ctx context.Context, logger lager.Logger, roleName string, domainID string, issuer string) (err error) {
	logger = logger.Session("data-service")

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

	err = assignRole(ctx, logger, tx, roleName, domainID, issuer)

	return
}

func (s *DataService) UnassignRole(ctx context.Context, logger lager.Logger, roleName string, domainID string, issuer string) (err error) {
	logger = logger.Session("data-service")

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

	err = unassignRole(ctx, logger, tx, roleName, domainID, issuer)

	return
}

func (s *DataService) HasRole(ctx context.Context, logger lager.Logger, query models.RoleAssignmentQuery) (bool, error) {
	logger = logger.Session("data-service")

	return hasRole(ctx, logger, s.conn, query)
}

func (s *DataService) ListActorRoles(ctx context.Context, logger lager.Logger, query models.ActorQuery) ([]*models.Role, error) {
	logger = logger.Session("data-service")

	actor, err := findActor(ctx, logger, s.conn, query)
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
