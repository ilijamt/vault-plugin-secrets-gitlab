package gitlab

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
)

type entryRole struct {
	RoleName    string        `json:"role_name" structs:"role_name" mapstructure:"role_name"`
	TokenTTL    time.Duration `json:"token_ttl" structs:"token_ttl" mapstructure:"token_ttl"`
	Path        string        `json:"path" structs:"path" mapstructure:"path"`
	Name        string        `json:"name" structs:"name" mapstructure:"name"`
	Scopes      []string      `json:"scopes" structs:"scopes" mapstructure:"scopes"`
	AccessLevel AccessLevel   `json:"access_level" structs:"access_level" mapstructure:"access_level,omitempty"`
	TokenType   TokenType     `json:"token_type" structs:"token_type" mapstructure:"token_type"`
}

func (e entryRole) LogicalResponseData() map[string]any {
	return map[string]any{
		"role_name":    e.RoleName,
		"path":         e.Path,
		"name":         e.Name,
		"scopes":       e.Scopes,
		"access_level": e.AccessLevel.String(),
		"token_ttl":    int64(e.TokenTTL / time.Second),
		"token_type":   e.TokenType.String(),
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
