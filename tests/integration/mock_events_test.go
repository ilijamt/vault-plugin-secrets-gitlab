//go:build paths || saas || selfhosted || e2e

package integration_test

import (
	"context"
	"sync"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
)

type expectedEvent struct {
	eventType string
}

type mockEventsSender struct {
	eventsProcessed []*logical.EventReceived
	mu              sync.Mutex
}

func (m *mockEventsSender) resetEvents(t *testing.T) {
	t.Helper()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventsProcessed = make([]*logical.EventReceived, 0)
}

func (m *mockEventsSender) SendEvent(ctx context.Context, eventType logical.EventType, event *logical.EventData) error {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventsProcessed = append(m.eventsProcessed, &logical.EventReceived{
		EventType: string(eventType),
		Event:     event,
	})
	return nil
}

func (m *mockEventsSender) expectEvents(t *testing.T, expectedEvents []expectedEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t.Helper()
	require.EqualValuesf(t, len(m.eventsProcessed), len(expectedEvents), "Expected events: %v\nEvents processed: %v", expectedEvents, m.eventsProcessed)
	for i, expected := range expectedEvents {
		actual := m.eventsProcessed[i]
		require.EqualValuesf(t, expected.eventType, actual.EventType, "Mismatched event type at index %d. Expected %s, got %s\n%v", i, expected.eventType, actual.EventType, m.eventsProcessed)
	}
}
