package gitlab

import (
	"maps"
)

type TokenWithAccessLevel struct {
	*Token `json:",inline"`

	AccessLevel AccessLevel `json:"access_level"`
}

func (t *TokenWithAccessLevel) Internal() (d map[string]any) {
	d = map[string]any{"access_level": t.AccessLevel.String()}
	maps.Copy(d, t.Token.Internal())
	return d
}

func (t *TokenWithAccessLevel) Data() (d map[string]any) {
	d = map[string]any{"access_level": t.AccessLevel.String()}
	maps.Copy(d, t.Token.Data())
	return d
}

func (t *TokenWithAccessLevel) Event(m map[string]string) (d map[string]string) {
	d = map[string]string{"access_level": t.AccessLevel.String()}
	maps.Copy(d, t.Token.Event(m))
	return d
}

var _ IToken = (*TokenWithAccessLevel)(nil)
