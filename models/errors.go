package models

import "code.cloudfoundry.org/perm/pkg/api/errdefs"

var (
	ErrRoleNotFound      = errdefs.NewErrNotFound("role")
	ErrRoleAlreadyExists = errdefs.NewErrAlreadyExists("role")

	ErrRoleAssignmentNotFound      = errdefs.NewErrNotFound("role-assignment")
	ErrRoleAssignmentAlreadyExists = errdefs.NewErrAlreadyExists("role-assignment")

	ErrActorNotFound      = errdefs.NewErrNotFound("actor")
	ErrActorAlreadyExists = errdefs.NewErrAlreadyExists("actor")

	ErrPermissionDefinitionNotFound      = errdefs.NewErrNotFound("permission-definition")
	ErrPermissionDefinitionAlreadyExists = errdefs.NewErrAlreadyExists("permission-definition")

	ErrPermissionNotFound      = errdefs.NewErrNotFound("permission")
	ErrPermissionAlreadyExists = errdefs.NewErrAlreadyExists("permission")
)
