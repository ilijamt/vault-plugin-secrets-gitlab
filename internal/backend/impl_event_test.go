package backend_test

import (
	"testing"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
)

func TestSendEvent(t *testing.T) {
	b := newTestBackend(t)
	_ = b.SendEvent(t.Context(), event.MustEventType("test"), nil)
}
