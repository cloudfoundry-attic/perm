package models

import "code.cloudfoundry.org/perm-go"

type PermissionDefinitionName string

type PermissionDefinition struct {
	Name PermissionDefinitionName
}

type PermissionName string
type PermissionResourcePattern string

type Permission struct {
	Name            PermissionName
	ResourcePattern PermissionResourcePattern
}

func (p *Permission) ToProto() *perm_go.Permission {
	return &perm_go.Permission{
		Name:            string(p.Name),
		ResourcePattern: string(p.ResourcePattern),
	}
}
