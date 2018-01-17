package db

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/models"
)

func (s *DataService) CreateActor(
	ctx context.Context,
	logger lager.Logger,
	domainID,
	issuer string,
) (*models.Actor, error) {
	actor, err := createActor(ctx, logger.Session("data-service"), s.conn, domainID, issuer)
	if err != nil {
		return nil, err
	}

	return actor.Actor, nil
}

func (s *DataService) FindActor(
	ctx context.Context,
	logger lager.Logger,
	query models.ActorQuery,
) (*models.Actor, error) {
	actor, err := findActor(ctx, logger.Session("data-service"), s.conn, query)
	if err != nil {
		return nil, err
	}

	return actor.Actor, nil
}
