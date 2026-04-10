package flags_test

import (
	"context"
	"testing"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/mocks"
)

type mockFlagsBackend struct {
	*mocks.MockFlagsProvider
	*mocks.MockEventSender
}

func (m *mockFlagsBackend) Flags() flags.Flags {
	return m.MockFlagsProvider.Flags()
}

func (m *mockFlagsBackend) UpdateFlags(fn func(*flags.Flags)) {
	m.MockFlagsProvider.UpdateFlags(fn)
}

func (m *mockFlagsBackend) SendEvent(ctx context.Context, eventType event.EventType, metadata map[string]string) error {
	return m.MockEventSender.SendEvent(ctx, eventType, metadata)
}

func newMockFlagsBackend(t *testing.T) *mockFlagsBackend {
	t.Helper()
	return &mockFlagsBackend{
		MockFlagsProvider: mocks.NewMockFlagsProvider(t),
		MockEventSender:   mocks.NewMockEventSender(t),
	}
}
