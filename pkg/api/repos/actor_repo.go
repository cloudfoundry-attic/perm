package repos

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/perm"
)

type ActorRepo interface {
	CreateActor(
		ctx context.Context,
		logger lager.Logger,
		domainID,
		issuer string,
	) (*perm.Actor, error)
}
