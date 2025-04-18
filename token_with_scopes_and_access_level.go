package gitlab

import (
	"maps"
	"strings"
)

type TokenWithScopesAndAccessLevel struct {
	Token `json:",inline"`

	Scopes      []string    `json:"scopes"`
	AccessLevel AccessLevel `json:"access_level"`
}

func (t *TokenWithScopesAndAccessLevel) Internal() (d map[string]any) {
	d = map[string]any{
		"scopes":       t.Scopes,
		"access_level": t.AccessLevel.String(),
	}
	maps.Copy(d, t.Token.Internal())
	return d
}

func (t *TokenWithScopesAndAccessLevel) Data() (d map[string]any) {
	d = map[string]any{
		"scopes":       t.Scopes,
		"access_level": t.AccessLevel.String(),
	}
	maps.Copy(d, t.Token.Data())
	return d
}

func (t *TokenWithScopesAndAccessLevel) Event(m map[string]string) (d map[string]string) {
	d = map[string]string{
		"scopes":       strings.Join(t.Scopes, ","),
		"access_level": t.AccessLevel.String(),
	}
	maps.Copy(d, t.Token.Event(m))
	return d
}

var _ IToken = (*TokenWithScopesAndAccessLevel)(nil)
