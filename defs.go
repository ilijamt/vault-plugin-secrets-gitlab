package gitlab

import (
	"errors"
	"time"
)

var (
	ErrNilValue             = errors.New("nil value")
	ErrInvalidValue         = errors.New("invalid value")
	ErrFieldRequired        = errors.New("required field")
	ErrFieldInvalidValue    = errors.New("invalid value for field")
	ErrBackendNotConfigured = errors.New("backend not configured")
)

const (
	DefaultConfigFieldAccessTokenMaxTTL = time.Duration(0)
	DefaultRoleFieldAccessTokenMaxTTL   = 24 * time.Hour
	DefaultAccessTokenMinTTL            = 24 * time.Hour
	DefaultAccessTokenMaxPossibleTTL    = 365 * 24 * time.Hour
)
