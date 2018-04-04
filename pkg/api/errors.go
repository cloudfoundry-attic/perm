package api

import "errors"

var (
	ErrServerStopped       = errors.New("perm: the server has been stopped")
	ErrServerFailedToStart = errors.New("perm: the server failed to start")
)
