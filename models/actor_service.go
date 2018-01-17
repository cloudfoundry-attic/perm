package models

import (
	"context"

	"code.cloudfoundry.org/lager"
)

type ActorQuery struct {
	DomainID string
	Issuer   string
}

type ActorService interface {
	CreateActor(
		ctx context.Context,
		logger lager.Logger,
		domainID,
		issuer string,
	) (*Actor, error)

	FindActor(
		ctx context.Context,
		logger lager.Logger,
		query ActorQuery,
	) (*Actor, error)
}
