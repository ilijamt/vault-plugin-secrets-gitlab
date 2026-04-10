package config

import "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"

var (
	eventWrite       = event.MustEventType("config-write")
	eventDelete      = event.MustEventType("config-delete")
	eventPatch       = event.MustEventType("config-patch")
	eventTokenRotate = event.MustEventType("config-token-rotate")
)
