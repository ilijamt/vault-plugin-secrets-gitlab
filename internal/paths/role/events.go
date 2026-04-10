package role

import "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"

var (
	eventWrite  = event.MustEventType("role-write")
	eventDelete = event.MustEventType("role-delete")
)
