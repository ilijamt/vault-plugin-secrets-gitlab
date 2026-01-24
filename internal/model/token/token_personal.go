package token

import (
	"maps"
	"strconv"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

type TokenPersonal struct {
	TokenWithScopes `json:",inline"`

	UserID int64 `json:"user_id"`
}

func (t *TokenPersonal) Internal() (d map[string]any) {
	d = map[string]any{"user_id": t.UserID}
	maps.Copy(d, t.TokenWithScopes.Internal())
	return d
}

func (t *TokenPersonal) Data() (d map[string]any) {
	d = map[string]any{"user_id": t.UserID}
	maps.Copy(d, t.TokenWithScopes.Data())
	return d
}

func (t *TokenPersonal) Event(m map[string]string) (d map[string]string) {
	d = map[string]string{"user_id": strconv.FormatInt(t.UserID, 10)}
	maps.Copy(d, t.Token.Event(m))
	return d
}

var _ token.Token = (*TokenPersonal)(nil)
