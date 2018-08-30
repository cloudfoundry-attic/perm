package db

import "code.cloudfoundry.org/perm"

var (
	errActorNotFoundDB           = perm.NewErrNotFound("actor")
	errActionNotFoundDB          = perm.NewErrNotFound("permission-definition")
	errActionAlreadyExistsInDB   = perm.NewErrAlreadyExists("permission-definition")
	errPermissionAlreadyExistsDB = perm.NewErrAlreadyExists("permission")
)
