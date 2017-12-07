package cmd

import (
	"errors"
	"fmt"
)

var (
	ErrMigrationsOutOfSync = errors.New("migrations out of sync: not all migrations applied")

	ErrFailedToAppendCertsFromPem = errors.New("failed to append certs from pem")
	ErrUnsupportedSQLDriver       = errors.New("unsupported sql driver")
)

type AttemptError int

func NewAttemptError(attempts int) AttemptError {
	return AttemptError(attempts)
}

func (i AttemptError) Error() string {
	return fmt.Sprintf("failed to talk to database within %d attempts", i)
}
