package gitlab

import (
	"strconv"
	"time"
)

type EntryToken struct {
	TokenID            int         `json:"token_id"`
	UserID             int         `json:"user_id"`
	ParentID           string      `json:"parent_id"`
	Path               string      `json:"path"`
	Name               string      `json:"name"`
	Token              string      `json:"token"`
	TokenType          TokenType   `json:"token_type"`
	CreatedAt          *time.Time  `json:"created_at"`
	ExpiresAt          *time.Time  `json:"expires_at"`
	Scopes             []string    `json:"scopes"`
	AccessLevel        AccessLevel `json:"access_level"` // not used for personal access tokens
	RoleName           string      `json:"role_name"`
	ConfigName         string      `json:"config_name"`
	GitlabRevokesToken bool        `json:"gitlab_revokes_token"`
}

func (e EntryToken) SecretResponse() (map[string]any, map[string]any) {
	return map[string]any{
			"name":         e.Name,
			"token":        e.Token,
			"path":         e.Path,
			"scopes":       e.Scopes,
			"role_name":    e.RoleName,
			"access_level": e.AccessLevel.String(),
			"created_at":   e.CreatedAt,
			"expires_at":   e.ExpiresAt,
		},
		map[string]any{
			"path":                 e.Path,
			"name":                 e.Name,
			"token":                e.Token,
			"user_id":              e.UserID,
			"parent_id":            e.ParentID,
			"token_id":             e.TokenID,
			"token_type":           e.TokenType.String(),
			"scopes":               e.Scopes,
			"access_level":         e.AccessLevel.String(),
			"role_name":            e.RoleName,
			"config_name":          e.ConfigName,
			"gitlab_revokes_token": strconv.FormatBool(e.GitlabRevokesToken),
		}
}
