package gitlab

import (
	"maps"
	"strings"
)

type TokenWithScopes struct {
	Token `json:",inline"`

	Scopes []string `json:"scopes"`
}

func (t *TokenWithScopes) Internal() (d map[string]any) {
	d = map[string]any{"scopes": t.Scopes}
	maps.Copy(d, t.Token.Internal())
	return d
}

func (t *TokenWithScopes) Data() (d map[string]any) {
	d = map[string]any{"scopes": t.Scopes}
	maps.Copy(d, t.Token.Data())
	return d
}

func (t *TokenWithScopes) Event(m map[string]string) (d map[string]string) {
	d = map[string]string{"scopes": strings.Join(t.Scopes, ",")}
	maps.Copy(d, t.Token.Event(m))
	return d
}

var _ IToken = (*TokenWithScopes)(nil)
