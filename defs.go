package gitlab

import (
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	configPaths "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/config"
	flagsPaths "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/flags"
	rolePaths "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/role"
	tokenPaths "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/token"
)

const (
	DefaultConfigFieldAccessTokenMaxTTL = backend.DefaultConfigFieldAccessTokenMaxTTL
	DefaultConfigFieldAccessTokenRotate = backend.DefaultConfigFieldAccessTokenRotate
	DefaultRoleFieldAccessTokenMaxTTL   = backend.DefaultRoleFieldAccessTokenMaxTTL
	DefaultAccessTokenMinTTL            = backend.DefaultAccessTokenMinTTL
	DefaultAccessTokenMaxPossibleTTL    = backend.DefaultAccessTokenMaxPossibleTTL
	DefaultConfigName                   = backend.DefaultConfigName
	PathConfigStorage                   = backend.PathConfigStorage
	PathRoleStorage                     = backend.PathRoleStorage
	PathConfigFlags                     = flagsPaths.PathConfigFlags
	PathTokenRoleStorage                = tokenPaths.PathTokenRoleStorage
	TypeConfigDefault                   = DefaultConfigName
)

var (
	FieldSchemaConfig    = configPaths.FieldSchemaConfig
	FieldSchemaRoles     = rolePaths.FieldSchemaRoles
	FieldSchemaFlags     = flagsPaths.FieldSchemaFlags
	FieldSchemaTokenRole = tokenPaths.FieldSchemaTokenRole
)
