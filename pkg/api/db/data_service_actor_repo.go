package db

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/perm"
)

func (s *DataService) CreateActor(
	ctx context.Context,
	logger lager.Logger,
	domainID string,
	issuer string,
) (*perm.Actor, error) {
	actor, err := createActor(ctx, logger.Session("data-service"), s.conn, domainID, issuer)
	if err != nil {
		return nil, err
	}

	return actor.Actor, nil
}
