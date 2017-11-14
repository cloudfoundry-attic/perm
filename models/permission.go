package models

import "code.cloudfoundry.org/perm/protos"

type Permission struct {
	Name            string
	ResourcePattern string
}

func (p *Permission) ToProto() *protos.Permission {
	return &protos.Permission{
		Name:            p.Name,
		ResourcePattern: p.ResourcePattern,
	}
}
