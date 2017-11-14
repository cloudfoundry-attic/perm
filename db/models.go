package db

import (
	"code.cloudfoundry.org/perm/models"
)

type actor struct {
	ID int64
	*models.Actor
}

type role struct {
	ID int64
	*models.Role
}

type permission struct {
	ID int64
	*models.Permission
}

type roleAssignment struct {
	Actor actor
	Role  role
}
