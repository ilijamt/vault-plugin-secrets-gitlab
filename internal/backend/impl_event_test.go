package backend_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
)

func TestSendEvent(t *testing.T) {
	b := newTestBackend(t)
	err := b.SendEvent(t.Context(), event.MustEventType("test"), nil)
	assert.ErrorContains(t, err, "no event sender")
}
