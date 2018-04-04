package perm

import (
	"errors"

	"code.cloudfoundry.org/perm/pkg/api/errdefs"
)

var (
	ErrUnknown = errors.New("perm: unknown error")

	ErrFailedToConnect     = errors.New("perm: failed to connect")
	ErrNoTransportSecurity = errors.New("perm: no transport security set (use perm.WithTransportCredentials() to set)")
	ErrClientConnClosing   = errors.New("perm: the client connection is already closing or closed")

	ErrRoleNotFound      = errdefs.NewErrNotFound("role")
	ErrRoleAlreadyExists = errdefs.NewErrAlreadyExists("role")

	ErrAssignmentNotFound      = errdefs.NewErrNotFound("assignment")
	ErrAssignmentAlreadyExists = errdefs.NewErrAlreadyExists("assignment")
)
