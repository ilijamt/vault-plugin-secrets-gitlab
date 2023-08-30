package gitlab

import "time"

type EntryToken struct {
	TokenID     int         `json:"token_id"`
	UserID      int         `json:"user_id"`
	ParentID    string      `json:"parent_id"`
	Path        string      `json:"path"`
	Name        string      `json:"name"`
	Token       string      `json:"token"`
	TokenType   TokenType   `json:"token_type"`
	CreatedAt   *time.Time  `json:"created_at"`
	ExpiresAt   *time.Time  `json:"expires_at"`
	Scopes      []string    `json:"scopes"`
	AccessLevel AccessLevel `json:"access_level"` // not used for personal access tokens
}

func (e EntryToken) SecretResponse() (map[string]interface{}, map[string]interface{}) {
	return map[string]interface{}{
			"name":         e.Name,
			"token":        e.Token,
			"path":         e.Path,
			"scopes":       e.Scopes,
			"access_level": e.AccessLevel.String(),
			"created_at":   e.CreatedAt,
			"expires_at":   e.ExpiresAt,
		},
		map[string]interface{}{
			"path":         e.Path,
			"name":         e.Name,
			"user_id":      e.UserID,
			"parent_id":    e.ParentID,
			"token_id":     e.TokenID,
			"token_type":   e.TokenType.String(),
			"scopes":       e.Scopes,
			"access_level": e.AccessLevel.String(),
		}
}
