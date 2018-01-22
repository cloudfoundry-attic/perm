package models

import "code.cloudfoundry.org/perm/protos"

type RoleName string

type Role struct {
	Name RoleName
}

func (r *Role) ToProto() *protos.Role {
	return &protos.Role{
		Name: string(r.Name),
	}
}
