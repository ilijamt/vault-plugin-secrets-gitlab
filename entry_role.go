package gitlab

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
)

type entryRole struct {
	RoleName            string        `json:"role_name" structs:"role_name" mapstructure:"role_name"`
	TTL                 time.Duration `json:"ttl" structs:"ttl" mapstructure:"ttl"`
	MaxTTL              time.Duration `json:"max_ttl" structs:"max_ttl" mapstructure:"max_ttl"`
	Path                string        `json:"path" structs:"path" mapstructure:"path"`
	Name                string        `json:"name" structs:"name" mapstructure:"name"`
	Scopes              []string      `json:"scopes" structs:"scopes" mapstructure:"scopes"`
	AccessLevel         AccessLevel   `json:"access_level" structs:"access_level" mapstructure:"access_level,omitempty"`
	TokenType           TokenType     `json:"token_type" structs:"token_type" mapstructure:"token_type"`
	GitlabRevokesTokens bool          `json:"gitlab_revokes_token" structs:"gitlab_revokes_token" mapstructure:"gitlab_revokes_token"`
}

func (e entryRole) LogicalResponseData() map[string]any {
	return map[string]any{
		"role_name":            e.RoleName,
		"path":                 e.Path,
		"name":                 e.Name,
		"scopes":               e.Scopes,
		"access_level":         e.AccessLevel.String(),
		"ttl":                  int64(e.TTL / time.Second),
		"max_ttl":              int64(e.MaxTTL / time.Second),
		"token_type":           e.TokenType.String(),
		"gitlab_revokes_token": e.GitlabRevokesTokens,
	}
}

func getRole(ctx context.Context, name string, s logical.Storage) (*entryRole, error) {
	entry, err := s.Get(ctx, fmt.Sprintf("%s/%s", PathRoleStorage, name))
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	role := new(entryRole)
	if err := entry.DecodeJSON(role); err != nil {
		return nil, err
	}
	return role, nil
}
