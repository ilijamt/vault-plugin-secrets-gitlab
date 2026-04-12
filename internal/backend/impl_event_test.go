package backend_test

import (
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
)

func TestSendEvent(t *testing.T) {
	b := backend.New(flags.Flags{})
	mockSender := &logical.MockEventSender{}
	require.NoError(t, b.Init(t.Context(), &logical.BackendConfig{
		System:       &logical.StaticSystemView{},
		EventsSender: mockSender,
	}))

	err := b.SendEvent(t.Context(), event.MustEventType("test"), nil)
	assert.NoError(t, err)
	assert.Len(t, mockSender.Events, 1)
}
