package gitlab

import (
	"maps"
	"strconv"
)

type TokenGroupServiceAccount struct {
	TokenWithScopes `json:",inline"`

	UserID int `json:"user_id"`
}

func (t *TokenGroupServiceAccount) Internal() (d map[string]any) {
	d = map[string]any{"user_id": t.UserID}
	maps.Copy(d, t.TokenWithScopes.Internal())
	return d
}

func (t *TokenGroupServiceAccount) Data() (d map[string]any) {
	d = map[string]any{"user_id": t.UserID}
	maps.Copy(d, t.TokenWithScopes.Internal())
	return d
}

func (t *TokenGroupServiceAccount) Event(m map[string]string) (d map[string]string) {
	d = map[string]string{"user_id": strconv.Itoa(t.UserID)}
	maps.Copy(d, t.Token.Event(m))
	return d
}
