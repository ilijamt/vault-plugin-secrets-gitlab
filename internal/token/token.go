package token

import (
	"time"
)

type Token interface {
	Internal() map[string]any
	Data() map[string]any
	Event(map[string]string) map[string]string
	Type() Type
	SetConfigName(string)
	SetRoleName(string)
	SetGitlabRevokesToken(bool)
	SetExpiresAt(*time.Time)
	GetExpiresAt() time.Time
	GetCreatedAt() time.Time
	TTL() time.Duration
}
