package models

import "code.cloudfoundry.org/perm/errdefs"

var (
	ErrRoleNotFound      = errdefs.NewErrNotFound("role")
	ErrRoleAlreadyExists = errdefs.NewErrAlreadyExists("role")

	ErrRoleAssignmentNotFound      = errdefs.NewErrNotFound("role-assignment")
	ErrRoleAssignmentAlreadyExists = errdefs.NewErrAlreadyExists("role-assignment")

	ErrActorNotFound = errdefs.NewErrNotFound("actor")
	ErrActorAlreadyExists = errdefs.NewErrAlreadyExists("actor")
)
