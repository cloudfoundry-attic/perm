package models

import "code.cloudfoundry.org/perm-go"

type RoleName string

type Role struct {
	Name RoleName
}

func (r *Role) ToProto() *perm_go.Role {
	return &perm_go.Role{
		Name: string(r.Name),
	}
}
