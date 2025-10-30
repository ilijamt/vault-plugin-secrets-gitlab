package role

import (
	"strings"
	"time"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

var _ model.Named = (*Role)(nil)
var _ model.IsNil = (*Role)(nil)
var _ model.LogicalResponseData = (*Role)(nil)

type Role struct {
	RoleName            string            `json:"role_name" structs:"role_name" mapstructure:"role_name"`
	TTL                 time.Duration     `json:"ttl" structs:"ttl" mapstructure:"ttl"`
	Path                string            `json:"path" structs:"path" mapstructure:"path"`
	Name                string            `json:"name" structs:"name" mapstructure:"name"`
	Scopes              []string          `json:"scopes" structs:"scopes" mapstructure:"scopes"`
	AccessLevel         token.AccessLevel `json:"access_level" structs:"access_level" mapstructure:"access_level,omitempty"`
	TokenType           token.Type        `json:"token_type" structs:"token_type" mapstructure:"token_type"`
	GitlabRevokesTokens bool              `json:"gitlab_revokes_token" structs:"gitlab_revokes_token" mapstructure:"gitlab_revokes_token"`
	ConfigName          string            `json:"config_name" structs:"config_name" mapstructure:"config_name"`
}

func (e Role) IsNil() bool { return false }

func (e Role) GetName() string {
	return e.Name
}

func (e Role) LogicalResponseData() map[string]any {
	return map[string]any{
		"role_name":            e.RoleName,
		"path":                 e.Path,
		"name":                 e.Name,
		"scopes":               strings.Join(e.Scopes, ", "),
		"access_level":         e.AccessLevel.String(),
		"ttl":                  int64(e.TTL / time.Second),
		"token_type":           e.TokenType.String(),
		"gitlab_revokes_token": e.GitlabRevokesTokens,
		"config_name":          e.ConfigName,
	}
}
