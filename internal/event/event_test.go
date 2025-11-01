package event_test

import (
	"context"
	"sync"
	"testing"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
)

type mockEventsSender struct {
	events []*logical.EventReceived
	mu     sync.Mutex
}

var _ logical.EventSender = (*mockEventsSender)(nil)

func (m *mockEventsSender) SendEvent(ctx context.Context, eventType logical.EventType, event *logical.EventData) error {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, &logical.EventReceived{
		EventType: string(eventType),
		Event:     event,
	})
	return nil
}

func TestEvent(t *testing.T) {
	t.Run("nil backend", func(t *testing.T) {
		require.ErrorIs(t,
			event.Event(
				t.Context(),
				nil,
				"test",
				map[string]string{"test": "test"},
			),
			errs.ErrNilValue,
		)
	})

	t.Run("no sender specified", func(t *testing.T) {
		b := &framework.Backend{}
		require.NoError(t, b.Setup(t.Context(), &logical.BackendConfig{}))
		require.ErrorIs(t,
			event.Event(
				t.Context(),
				&framework.Backend{},
				"test",
				map[string]string{"test": "test"},
			),
			framework.ErrNoEvents,
		)
	})

	t.Run("with event sender", func(t *testing.T) {
		b := &framework.Backend{}
		evt := &mockEventsSender{}
		require.NoError(t, b.Setup(t.Context(), &logical.BackendConfig{EventsSender: evt}))
		require.NoError(t,
			event.Event(
				t.Context(), b,
				"test",
				map[string]string{"test": "test"},
			),
		)
		require.Len(t, evt.events, 1)
	})
}
