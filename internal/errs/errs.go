package errs

import "errors"

var (
	ErrNilValue             = errors.New("nil value")
	ErrInvalidValue         = errors.New("invalid value")
	ErrFieldRequired        = errors.New("required field")
	ErrFieldInvalidValue    = errors.New("invalid value for field")
	ErrBackendNotConfigured = errors.New("backend not configured")
)
