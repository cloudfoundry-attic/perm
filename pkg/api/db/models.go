package db

import (
	"code.cloudfoundry.org/perm/models"
)

type id int64

type actor struct {
	ID id
	*models.Actor
}

type role struct {
	ID id
	*models.Role
}

type permission struct {
	ID id
	*models.Permission
}

type permissionDefinition struct {
	ID id
	*models.PermissionDefinition
}
