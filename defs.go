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
	DefaultConfigFieldAccessTokenMaxTTL = 7 * 24 * time.Hour
	DefaultConfigFieldAccessTokenRotate = DefaultAutoRotateBeforeMinTTL
	DefaultRoleFieldAccessTokenMaxTTL   = 24 * time.Hour
	DefaultAccessTokenMinTTL            = 24 * time.Hour
	DefaultAccessTokenMaxPossibleTTL    = 365 * 24 * time.Hour
	DefaultAutoRotateBeforeMinTTL       = 24 * time.Hour
	DefaultAutoRotateBeforeMaxTTL       = 730 * time.Hour
)
