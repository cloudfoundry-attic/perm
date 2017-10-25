package models

import "code.cloudfoundry.org/perm/errdefs"

var (
	ErrRoleNotFound = errdefs.NewErrNotFound("role")
	ErrRoleAlreadyExists = errdefs.NewErrAlreadyExists("role")
)
