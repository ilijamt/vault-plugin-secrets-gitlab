package flags_test

import (
	"context"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
)

// mockFlagsBackend is a hand-written mock satisfying the flagsBackend interface.
type mockFlagsBackend struct {
	flags       flags.Flags
	updateFlags func(fn func(*flags.Flags))
	sendEvent   func(ctx context.Context, eventType event.EventType, metadata map[string]string) error
}

func (m *mockFlagsBackend) Flags() flags.Flags {
	return m.flags
}

func (m *mockFlagsBackend) UpdateFlags(fn func(*flags.Flags)) {
	if m.updateFlags != nil {
		m.updateFlags(fn)
		return
	}
	fn(&m.flags)
}

func (m *mockFlagsBackend) SendEvent(ctx context.Context, eventType event.EventType, metadata map[string]string) error {
	if m.sendEvent != nil {
		return m.sendEvent(ctx, eventType, metadata)
	}
	return nil
}
