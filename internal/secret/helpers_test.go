package secret_test

import (
	"context"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/mocks"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

type mockSecretBackend struct {
	*mocks.MockClientProvider
	*mocks.MockEventSender
}

func (m *mockSecretBackend) GetClient(name string) gitlab.Client {
	return m.MockClientProvider.GetClient(name)
}

func (m *mockSecretBackend) SetClient(client gitlab.Client, name string) {
	m.MockClientProvider.SetClient(client, name)
}

func (m *mockSecretBackend) DeleteClient(name string) {
	m.MockClientProvider.DeleteClient(name)
}

func (m *mockSecretBackend) GetClientByName(ctx context.Context, s logical.Storage, name string) (gitlab.Client, error) {
	return m.MockClientProvider.GetClientByName(ctx, s, name)
}

func (m *mockSecretBackend) SendEvent(ctx context.Context, eventType event.EventType, metadata map[string]string) error {
	return m.MockEventSender.SendEvent(ctx, eventType, metadata)
}

func newMockSecretBackend(t *testing.T) *mockSecretBackend {
	t.Helper()
	return &mockSecretBackend{
		MockClientProvider: mocks.NewMockClientProvider(t),
		MockEventSender:    mocks.NewMockEventSender(t),
	}
}

func newRevokeSecret(tokenType token.Type, parentId string, extra map[string]any) *logical.Secret {
	data := map[string]any{
		"token_id":             int64(42),
		"gitlab_revokes_token": false,
		"parent_id":            parentId,
		"token_type":           tokenType.String(),
		"path":                 "some/path",
		"name":                 "test-token",
		"config_name":          "default",
	}
	for k, v := range extra {
		data[k] = v
	}
	return &logical.Secret{InternalData: data}
}
