package token

import "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"

var eventWrite = event.MustEventType("token-write")
