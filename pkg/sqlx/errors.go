package sqlx

import "errors"

var (
	ErrUnsupportedSQLDriver        = errors.New("unsupported sql driver")
	ErrFailedToEstablishConnection = errors.New("failed to establish connection")
)
