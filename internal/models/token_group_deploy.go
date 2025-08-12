package models

import (
	"maps"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

type TokenGroupDeploy struct {
	TokenWithScopes `json:",inline"`

	Username string `json:"username"`
}

func (t *TokenGroupDeploy) Internal() (d map[string]any) {
	d = map[string]any{"username": t.Username}
	maps.Copy(d, t.TokenWithScopes.Internal())
	return d
}

func (t *TokenGroupDeploy) Data() (d map[string]any) {
	d = map[string]any{"username": t.Username}
	maps.Copy(d, t.TokenWithScopes.Data())
	return d
}

func (t *TokenGroupDeploy) Event(m map[string]string) (d map[string]string) {
	d = map[string]string{"username": t.Username}
	maps.Copy(d, t.Token.Event(m))
	return d
}

var _ token.Token = (*TokenGroupDeploy)(nil)
