package repos

import "code.cloudfoundry.org/perm/models"

type ActorQuery struct {
	DomainID models.ActorDomainID
	Issuer   models.ActorIssuer
}
