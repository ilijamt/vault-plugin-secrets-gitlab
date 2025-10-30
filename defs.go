package gitlab

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
)
