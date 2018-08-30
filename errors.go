package perm

import (
	"errors"
)

var (
	ErrFailedToConnect     = errors.New("perm: failed to connect")
	ErrUnauthenticated     = errors.New("perm: unauthenticated")
	ErrNoTransportSecurity = errors.New("perm: no transport security set (use perm.WithTLSConfig() to set)")
	ErrClientConnClosing   = errors.New("perm: the client connection is already closing or closed")

	ErrRoleNotFound      = NewErrNotFound("role")
	ErrRoleAlreadyExists = NewErrAlreadyExists("role")

	ErrAssignmentNotFound      = NewErrNotFound("assignment")
	ErrAssignmentAlreadyExists = NewErrAlreadyExists("assignment")
	ErrActorAlreadyExists      = NewErrAlreadyExists("actor")

	ErrActorNamespaceEmpty = NewErrCannotBeEmpty("actor namespace")
)
