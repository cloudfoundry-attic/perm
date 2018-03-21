package db

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/api/models"
)

func (s *DataService) CreateActor(
	ctx context.Context,
	logger lager.Logger,
	domainID models.ActorDomainID,
	issuer models.ActorIssuer,
) (*models.Actor, error) {
	actor, err := createActor(ctx, logger.Session("data-service"), s.conn, domainID, issuer)
	if err != nil {
		return nil, err
	}

	return actor.Actor, nil
}
