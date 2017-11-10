package errdefs // import "code.cloudfoundry.org/perm/errdefs"

import "fmt"

type ErrNotFound struct {
	model string
}

func NewErrNotFound(model string) ErrNotFound {
	return ErrNotFound{
		model: model,
	}
}

func (err ErrNotFound) Error() string {
	return fmt.Sprintf("%s not found", err.model)
}

type ErrAlreadyExists struct {
	model string
}

func NewErrAlreadyExists(model string) ErrAlreadyExists {
	return ErrAlreadyExists{
		model: model,
	}
}

func (err ErrAlreadyExists) Error() string {
	return fmt.Sprintf("%s already exists", err.model)
}
