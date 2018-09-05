package monitor

type errExceededMaxLatency struct{}

func (e errExceededMaxLatency) Error() string {
	return "probe: exceeded max latency"
}

type errIncorrectPermission struct {
	expected bool
	actual   bool
}

func (e errIncorrectPermission) Error() string {
	return "probe: incorrect permission"
}
