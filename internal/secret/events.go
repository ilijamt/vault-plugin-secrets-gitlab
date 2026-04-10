package secret

import "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"

var eventRevoke = event.MustEventType("token-revoke")
