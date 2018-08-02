package monitor

type ExceededMaxLatencyError struct{}

func (e ExceededMaxLatencyError) Error() string {
	return "probe: an API call timed out"
}

type HasAssignedPermissionError struct{}

func (e HasAssignedPermissionError) Error() string {
	return "probe: incorrect result, HasPermission should have returned true"
}

type HasUnassignedPermissionError struct{}

func (e HasUnassignedPermissionError) Error() string {
	return "probe: incorrect result, HasPermission should have returned false"
}
