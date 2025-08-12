package gitlab

import (
	"time"
)

const (
	DefaultConfigFieldAccessTokenMaxTTL = 7 * 24 * time.Hour
	DefaultConfigFieldAccessTokenRotate = DefaultAutoRotateBeforeMinTTL
	DefaultRoleFieldAccessTokenMaxTTL   = 24 * time.Hour
	DefaultAccessTokenMinTTL            = 24 * time.Hour
	DefaultAccessTokenMaxPossibleTTL    = 365 * 24 * time.Hour
	DefaultAutoRotateBeforeMinTTL       = 24 * time.Hour
	DefaultAutoRotateBeforeMaxTTL       = 730 * time.Hour
	DefaultConfigName                   = "default"
)
