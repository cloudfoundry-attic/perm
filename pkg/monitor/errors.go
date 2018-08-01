package monitor

import "errors"

var (
	ErrExceededMaxLatency      = errors.New("probe: request took too long")
	ErrHasAssignedPermission   = errors.New("probe: incorrect result, HasPermission should have returned true")
	ErrHasUnassignedPermission = errors.New("probe: incorrect result, HasPermission should have returned false")
)
