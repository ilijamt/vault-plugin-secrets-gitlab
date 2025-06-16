package gitlab

import (
	"crypto/sha1"
	"fmt"
	"maps"
	"strconv"
	"time"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

type IToken interface {
	Internal() map[string]any
	Data() map[string]any
	Event(map[string]string) map[string]string
	Type() token.TokenType
	SetConfigName(string)
	SetRoleName(string)
	SetGitlabRevokesToken(bool)
	SetExpiresAt(*time.Time)
	GetExpiresAt() time.Time
	GetCreatedAt() time.Time
	TTL() time.Duration
}

type Token struct {
	RoleName           string          `json:"role_name"`
	ConfigName         string          `json:"config_name"`
	GitlabRevokesToken bool            `json:"gitlab_revokes_token"`
	CreatedAt          *time.Time      `json:"created_at"`
	ExpiresAt          *time.Time      `json:"expires_at"`
	TokenType          token.TokenType `json:"type"`
	Token              string          `json:"token"`
	TokenID            int             `json:"token_id"`
	ParentID           string          `json:"parent_id"`
	Name               string          `json:"name"`
	Path               string          `json:"path"`
}

func (t *Token) TTL() time.Duration {
	return t.GetExpiresAt().Sub(t.GetCreatedAt())
}

func (t *Token) GetExpiresAt() (tm time.Time) {
	if t.ExpiresAt != nil {
		tm = *t.ExpiresAt
	}
	return tm
}
func (t *Token) GetCreatedAt() (tm time.Time) {
	if t.CreatedAt != nil {
		tm = *t.CreatedAt
	}
	return tm
}
func (t *Token) SetExpiresAt(expiresAt *time.Time) { t.ExpiresAt = expiresAt }
func (t *Token) SetConfigName(name string)         { t.ConfigName = name }
func (t *Token) SetRoleName(name string)           { t.RoleName = name }
func (t *Token) SetGitlabRevokesToken(b bool)      { t.GitlabRevokesToken = b }
func (t *Token) Type() token.TokenType             { return t.TokenType }

func (t *Token) Internal() map[string]any {
	return map[string]any{
		"name":                 t.Name,
		"path":                 t.Path,
		"token":                t.Token,
		"token_id":             t.TokenID,
		"parent_id":            t.ParentID,
		"role_name":            t.RoleName,
		"config_name":          t.ConfigName,
		"gitlab_revokes_token": t.GitlabRevokesToken,
		"created_at":           t.CreatedAt,
		"expires_at":           t.ExpiresAt,
		"token_type":           t.Type().String(),
	}
}

func (t *Token) Data() map[string]any {
	return map[string]any{
		"path":            t.Path,
		"name":            t.Name,
		"token":           t.Token,
		"token_sha1_hash": fmt.Sprintf("%x", sha1.Sum([]byte(t.Token))),
		"token_id":        t.TokenID,
		"token_type":      t.Type().String(),
		"parent_id":       t.ParentID,
		"role_name":       t.RoleName,
		"config_name":     t.ConfigName,
		"created_at":      t.CreatedAt,
		"expires_at":      t.ExpiresAt,
	}
}

func (t *Token) Event(m map[string]string) (d map[string]string) {
	d = map[string]string{
		"config_name": t.ConfigName,
		"role_name":   t.RoleName,
		"token_id":    strconv.Itoa(t.TokenID),
		"parent_id":   t.ParentID,
		"token_type":  t.Type().String(),
		"ttl":         t.TTL().String(),
		"name":        t.Name,
	}
	maps.Copy(d, m)
	return d
}

var _ IToken = (*Token)(nil)
