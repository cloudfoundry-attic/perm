package models

import "code.cloudfoundry.org/perm-go"

type ActorDomainID string

type ActorIssuer string

type Actor struct {
	DomainID ActorDomainID
	Issuer   ActorIssuer
}

func (a *Actor) ToProto() *protos.Actor {
	return &protos.Actor{
		ID:     string(a.DomainID),
		Issuer: string(a.Issuer),
	}
}
