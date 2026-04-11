package secret_test

import (
	"context"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/secret"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestRevokeAccessToken_GitlabRevokesToken(t *testing.T) {
	var eventSent bool
	mb := &mockSecretBackend{
		sendEvent: func(_ context.Context, _ event.EventType, _ map[string]string) error {
			eventSent = true
			return nil
		},
	}

	s := secret.NewSecret(mb, "default")

	resp, err := s.HandleRevoke(t.Context(), &logical.Request{
		Storage: &logical.InmemStorage{},
		Secret: &logical.Secret{
			InternalData: map[string]any{
				"token_id":             int64(123),
				"gitlab_revokes_token": true,
				"parent_id":            "group1",
				"token_type":           token.TypePersonal.String(),
				"path":                 "user1",
				"name":                 "test-token",
				"config_name":          "default",
			},
		},
	})
	require.NoError(t, err)
	require.Nil(t, resp)
	require.True(t, eventSent)
}

func TestRevokeAccessToken_VaultRevokes(t *testing.T) {
	tests := []struct {
		name      string
		tokenType token.Type
		parentId  string
		extra     map[string]any
		setupStub func(*stubClient)
	}{
		{
			name:      "personal",
			tokenType: token.TypePersonal,
			parentId:  "user1",
			setupStub: func(c *stubClient) {
				c.revokePersonalAccessToken = func(_ context.Context, tokenId int64) error {
					require.Equal(t, int64(42), tokenId)
					return nil
				}
			},
		},
		{
			name:      "project",
			tokenType: token.TypeProject,
			parentId:  "proj1",
			setupStub: func(c *stubClient) {
				c.revokeProjectAccessToken = func(_ context.Context, tokenId int64, projectId string) error {
					require.Equal(t, int64(42), tokenId)
					require.Equal(t, "proj1", projectId)
					return nil
				}
			},
		},
		{
			name:      "group",
			tokenType: token.TypeGroup,
			parentId:  "grp1",
			setupStub: func(c *stubClient) {
				c.revokeGroupAccessToken = func(_ context.Context, tokenId int64, groupId string) error {
					require.Equal(t, int64(42), tokenId)
					require.Equal(t, "grp1", groupId)
					return nil
				}
			},
		},
		{
			name:      "user service account",
			tokenType: token.TypeUserServiceAccount,
			parentId:  "user1",
			extra:     map[string]any{"token": "glpat-secret"},
			setupStub: func(c *stubClient) {
				c.revokeUserServiceAccountAccessToken = func(_ context.Context, tok string) error {
					require.Equal(t, "glpat-secret", tok)
					return nil
				}
			},
		},
		{
			name:      "group service account",
			tokenType: token.TypeGroupServiceAccount,
			parentId:  "grp1",
			extra:     map[string]any{"token": "glpat-secret"},
			setupStub: func(c *stubClient) {
				c.revokeGroupServiceAccountAccessToken = func(_ context.Context, tok string) error {
					require.Equal(t, "glpat-secret", tok)
					return nil
				}
			},
		},
		{
			name:      "pipeline project trigger",
			tokenType: token.TypePipelineProjectTrigger,
			parentId:  "100",
			setupStub: func(c *stubClient) {
				c.revokePipelineProjectTriggerAccessToken = func(_ context.Context, projectId int64, tokenId int64) error {
					require.Equal(t, int64(100), projectId)
					require.Equal(t, int64(42), tokenId)
					return nil
				}
			},
		},
		{
			name:      "group deploy",
			tokenType: token.TypeGroupDeploy,
			parentId:  "200",
			setupStub: func(c *stubClient) {
				c.revokeGroupDeployToken = func(_ context.Context, groupId, deployTokenId int64) error {
					require.Equal(t, int64(200), groupId)
					require.Equal(t, int64(42), deployTokenId)
					return nil
				}
			},
		},
		{
			name:      "project deploy",
			tokenType: token.TypeProjectDeploy,
			parentId:  "300",
			setupStub: func(c *stubClient) {
				c.revokeProjectDeployToken = func(_ context.Context, projectId, deployTokenId int64) error {
					require.Equal(t, int64(300), projectId)
					require.Equal(t, int64(42), deployTokenId)
					return nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &stubClient{}
			tt.setupStub(client)

			mb := &mockSecretBackend{
				getClientByName: func(_ context.Context, _ logical.Storage, name string) (gitlab.Client, error) {
					require.Equal(t, "default", name)
					return client, nil
				},
			}

			s := secret.NewSecret(mb, "default")

			resp, err := s.HandleRevoke(t.Context(), &logical.Request{
				Storage: &logical.InmemStorage{},
				Secret:  newRevokeSecret(tt.tokenType, tt.parentId, tt.extra),
			})
			require.NoError(t, err)
			require.Nil(t, resp)
		})
	}
}

func TestRevokeAccessToken_TokenNotFound(t *testing.T) {
	client := &stubClient{
		revokePersonalAccessToken: func(_ context.Context, tokenId int64) error {
			require.Equal(t, int64(42), tokenId)
			return errs.ErrAccessTokenNotFound
		},
	}
	mb := &mockSecretBackend{
		getClientByName: func(_ context.Context, _ logical.Storage, name string) (gitlab.Client, error) {
			require.Equal(t, "default", name)
			return client, nil
		},
	}

	s := secret.NewSecret(mb, "default")

	resp, err := s.HandleRevoke(t.Context(), &logical.Request{
		Storage: &logical.InmemStorage{},
		Secret:  newRevokeSecret(token.TypePersonal, "user1", nil),
	})
	require.NoError(t, err)
	require.Nil(t, resp)
}

func TestRevokeAccessToken_UsesConfigNameFromSecret(t *testing.T) {
	client := &stubClient{
		revokePersonalAccessToken: func(_ context.Context, tokenId int64) error {
			require.Equal(t, int64(42), tokenId)
			return nil
		},
	}
	mb := &mockSecretBackend{
		getClientByName: func(_ context.Context, _ logical.Storage, name string) (gitlab.Client, error) {
			require.Equal(t, "custom-config", name)
			return client, nil
		},
	}

	s := secret.NewSecret(mb, "default")

	sec := newRevokeSecret(token.TypePersonal, "user1", nil)
	sec.InternalData["config_name"] = "custom-config"

	resp, err := s.HandleRevoke(t.Context(), &logical.Request{
		Storage: &logical.InmemStorage{},
		Secret:  sec,
	})
	require.NoError(t, err)
	require.Nil(t, resp)
}
