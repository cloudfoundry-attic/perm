package perm

import "errors"

var (
	ErrFailedToConnect     = errors.New("perm: failed to connect")
	ErrNoTransportSecurity = errors.New("perm: no transport security set (use perm.WithTransportCredentials() to set)")
	ErrClientConnClosing   = errors.New("perm: the client connection is already closing or closed")
)
