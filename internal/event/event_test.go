package event_test

import (
	"testing"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
)

func TestEvent(t *testing.T) {
	t.Run("nil backend", func(t *testing.T) {
		require.ErrorIs(t,
			event.Event(
				t.Context(),
				nil,
				event.MustEventType("test"),
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
				event.MustEventType("test"),
				map[string]string{"test": "test"},
			),
			framework.ErrNoEvents,
		)
	})

	t.Run("with event sender", func(t *testing.T) {
		b := &framework.Backend{}
		evt := &logical.MockEventSender{}
		require.NoError(t, b.Setup(t.Context(), &logical.BackendConfig{EventsSender: evt}))
		require.NoError(t,
			event.Event(
				t.Context(), b,
				event.MustEventType("test"),
				map[string]string{"test": "test"},
			),
		)
		require.Len(t, evt.Events, 1)
	})
}
