package backend

import (
	"time"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
)

const (
	DefaultConfigFieldAccessTokenMaxTTL = 7 * 24 * time.Hour
	DefaultConfigFieldAccessTokenRotate = config.DefaultAutoRotateBeforeMinTTL
	DefaultRoleFieldAccessTokenMaxTTL   = 24 * time.Hour
	DefaultAccessTokenMinTTL            = 24 * time.Hour
	DefaultAccessTokenMaxPossibleTTL    = 365 * 24 * time.Hour
	DefaultConfigName                   = "default"

	// PathConfigStorage is the storage key prefix for config entries.
	PathConfigStorage = "config"

	// PathRoleStorage is the storage key prefix for role entries.
	PathRoleStorage = "roles"
)
