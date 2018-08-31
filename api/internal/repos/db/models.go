package db

import "code.cloudfoundry.org/perm"

type actor struct {
	ID int64
	perm.Actor
}

type role struct {
	ID int64
	perm.Role
}

type permission struct {
	ID int64
	perm.Permission
}

type action struct {
	ID int64
	perm.Action
}
