package repos

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/api/models"
)

type ActorRepo interface {
	CreateActor(
		ctx context.Context,
		logger lager.Logger,
		domainID models.ActorDomainID,
		issuer models.ActorIssuer,
	) (*models.Actor, error)
}
