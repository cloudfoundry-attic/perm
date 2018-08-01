package recording

type FailedToObserveDurationError struct {
	Err error
}

func (e FailedToObserveDurationError) Error() string {
	return e.Err.Error()
}
