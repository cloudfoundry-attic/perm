package probe

import "errors"

var (
	ErrExceededMaxLatency     = errors.New("probe: request took too long")
	ErrIncorrectHasPermission = errors.New("probe: incorrect HasPermission result")
)
