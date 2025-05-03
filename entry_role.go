package gitlab

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/access"
)

type EntryRole struct {
	RoleName            string             `json:"role_name" structs:"role_name" mapstructure:"role_name"`
	TTL                 time.Duration      `json:"ttl" structs:"ttl" mapstructure:"ttl"`
	Path                string             `json:"path" structs:"path" mapstructure:"path"`
	Name                string             `json:"name" structs:"name" mapstructure:"name"`
	Scopes              []string           `json:"scopes" structs:"scopes" mapstructure:"scopes"`
	AccessLevel         access.AccessLevel `json:"access_level" structs:"access_level" mapstructure:"access_level,omitempty"`
	TokenType           TokenType          `json:"token_type" structs:"token_type" mapstructure:"token_type"`
	GitlabRevokesTokens bool               `json:"gitlab_revokes_token" structs:"gitlab_revokes_token" mapstructure:"gitlab_revokes_token"`
	ConfigName          string             `json:"config_name" structs:"config_name" mapstructure:"config_name"`
}

func (e EntryRole) LogicalResponseData() map[string]any {
	return map[string]any{
		"role_name":            e.RoleName,
		"path":                 e.Path,
		"name":                 e.Name,
		"scopes":               e.Scopes,
		"access_level":         e.AccessLevel.String(),
		"ttl":                  int64(e.TTL / time.Second),
		"token_type":           e.TokenType.String(),
		"gitlab_revokes_token": e.GitlabRevokesTokens,
		"config_name":          e.ConfigName,
	}
}

func getRole(ctx context.Context, name string, s logical.Storage) (role *EntryRole, err error) {
	var entry *logical.StorageEntry
	if entry, err = s.Get(ctx, fmt.Sprintf("%s/%s", PathRoleStorage, name)); err == nil {
		if entry == nil {
			return nil, nil
		}
		role = new(EntryRole)
		_ = entry.DecodeJSON(role)
	}
	return role, err

}
