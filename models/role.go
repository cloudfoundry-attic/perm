package models

import "code.cloudfoundry.org/perm/protos"

type Role struct {
	Name string
}

func (r *Role) ToProto() *protos.Role {
	return &protos.Role{
		Name: r.Name,
	}
}
