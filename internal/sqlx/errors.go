package sqlx

import "errors"

var (
	ErrUnsupportedSQLDriver        = errors.New("unsupported sql driver")
	ErrFailedToEstablishConnection = errors.New("failed to establish connection")

	ErrMigrationsOutOfSync = errors.New("migrations out of sync: not all migrations applied")
)
