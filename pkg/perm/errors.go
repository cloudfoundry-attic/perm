package perm

import (
	"errors"
)

var (
	ErrUnknown = errors.New("perm: unknown error")

	ErrFailedToConnect     = errors.New("perm: failed to connect")
	ErrNoTransportSecurity = errors.New("perm: no transport security set (use perm.WithTransportCredentials() to set)")
	ErrClientConnClosing   = errors.New("perm: the client connection is already closing or closed")

	ErrRoleNotFound      = NewErrNotFound("role")
	ErrRoleAlreadyExists = NewErrAlreadyExists("role")

	ErrAssignmentNotFound      = NewErrNotFound("assignment")
	ErrAssignmentAlreadyExists = NewErrAlreadyExists("assignment")
	ErrActorAlreadyExists      = NewErrAlreadyExists("actor")

	ErrActorNamespaceEmpty = NewErrCannotBeEmpty("actor namespace")
)
