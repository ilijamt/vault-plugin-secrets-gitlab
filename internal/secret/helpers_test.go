package secret_test

import (
	"context"

	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

// mockSecretBackend is a hand-written mock satisfying the secretBackend interface.
type mockSecretBackend struct {
	getClientByName func(ctx context.Context, s logical.Storage, name string) (gitlab.Client, error)
	sendEvent       func(ctx context.Context, eventType event.EventType, metadata map[string]string) error
}

func (m *mockSecretBackend) GetClientByName(ctx context.Context, s logical.Storage, name string) (gitlab.Client, error) {
	return m.getClientByName(ctx, s, name)
}

func (m *mockSecretBackend) SendEvent(ctx context.Context, eventType event.EventType, metadata map[string]string) error {
	if m.sendEvent != nil {
		return m.sendEvent(ctx, eventType, metadata)
	}
	return nil
}

// stubClient embeds gitlab.Client (nil, panics on unimplemented methods)
// and provides function fields for the revoke methods used in tests.
type stubClient struct {
	gitlab.Client
	revokePersonalAccessToken               func(ctx context.Context, tokenId int64) error
	revokeProjectAccessToken                func(ctx context.Context, tokenId int64, projectId string) error
	revokeGroupAccessToken                  func(ctx context.Context, tokenId int64, groupId string) error
	revokeUserServiceAccountAccessToken     func(ctx context.Context, token string) error
	revokeGroupServiceAccountAccessToken    func(ctx context.Context, token string) error
	revokePipelineProjectTriggerAccessToken func(ctx context.Context, projectId int64, tokenId int64) error
	revokeGroupDeployToken                  func(ctx context.Context, groupId, deployTokenId int64) error
	revokeProjectDeployToken                func(ctx context.Context, projectId, deployTokenId int64) error
}

func (s *stubClient) RevokePersonalAccessToken(ctx context.Context, tokenId int64) error {
	return s.revokePersonalAccessToken(ctx, tokenId)
}

func (s *stubClient) RevokeProjectAccessToken(ctx context.Context, tokenId int64, projectId string) error {
	return s.revokeProjectAccessToken(ctx, tokenId, projectId)
}

func (s *stubClient) RevokeGroupAccessToken(ctx context.Context, tokenId int64, groupId string) error {
	return s.revokeGroupAccessToken(ctx, tokenId, groupId)
}

func (s *stubClient) RevokeUserServiceAccountAccessToken(ctx context.Context, token string) error {
	return s.revokeUserServiceAccountAccessToken(ctx, token)
}

func (s *stubClient) RevokeGroupServiceAccountAccessToken(ctx context.Context, token string) error {
	return s.revokeGroupServiceAccountAccessToken(ctx, token)
}

func (s *stubClient) RevokePipelineProjectTriggerAccessToken(ctx context.Context, projectId int64, tokenId int64) error {
	return s.revokePipelineProjectTriggerAccessToken(ctx, projectId, tokenId)
}

func (s *stubClient) RevokeGroupDeployToken(ctx context.Context, groupId, deployTokenId int64) error {
	return s.revokeGroupDeployToken(ctx, groupId, deployTokenId)
}

func (s *stubClient) RevokeProjectDeployToken(ctx context.Context, projectId, deployTokenId int64) error {
	return s.revokeProjectDeployToken(ctx, projectId, deployTokenId)
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
