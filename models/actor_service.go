package models

import (
	"context"

	"code.cloudfoundry.org/lager"
)

type ActorQuery struct {
	DomainID ActorDomainID
	Issuer   ActorIssuer
}

type ActorService interface {
	CreateActor(
		ctx context.Context,
		logger lager.Logger,
		domainID ActorDomainID,
		issuer ActorIssuer,
	) (*Actor, error)

	FindActor(
		ctx context.Context,
		logger lager.Logger,
		query ActorQuery,
	) (*Actor, error)
}
