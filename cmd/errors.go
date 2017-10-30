package cmd

import "errors"

var (
	ErrMigrationsOutOfSync = errors.New("migrations out of sync: not all migrations applied")
)
