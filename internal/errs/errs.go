package errs

import "errors"

var (
	// ErrNilValue represents an error indicating a nil value was encountered where it is not allowed.
	ErrNilValue = errors.New("nil value")

	// ErrInvalidValue indicates that an operation encountered a value that is considered invalid or inappropriate.
	ErrInvalidValue = errors.New("invalid value")

	// ErrFieldRequired represents an error when a required field is missing
	ErrFieldRequired = errors.New("required field")

	// ErrFieldInvalidValue represents an error when a field contains an invalid value
	ErrFieldInvalidValue = errors.New("invalid value for field")

	// ErrBackendNotConfigured represents an error when trying to use a backend that hasn't been properly configured
	ErrBackendNotConfigured = errors.New("backend not configured")
)
