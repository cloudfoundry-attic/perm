package models

import "code.cloudfoundry.org/perm/protos"

type Actor struct {
	DomainID string
	Issuer   string
}

func (a *Actor) ToProto() *protos.Actor {
	return &protos.Actor{
		ID:     a.DomainID,
		Issuer: a.Issuer,
	}
}
