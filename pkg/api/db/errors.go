package db

import "code.cloudfoundry.org/perm/pkg/perm"

var (
	errActorNotFoundDB                     = perm.NewErrNotFound("actor")
	errPermissionDefinitionNotFoundDB      = perm.NewErrNotFound("permission-definition")
	errPermissionDefinitionAlreadyExistsDB = perm.NewErrAlreadyExists("permission-definition")
	errPermissionAlreadyExistsDB           = perm.NewErrAlreadyExists("permission")
)
