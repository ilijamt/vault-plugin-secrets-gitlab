package secret_test

import (
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/mocks"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/secret"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestRevokeAccessToken_GitlabRevokesToken(t *testing.T) {
	mb := newMockSecretBackend(t)
	mb.MockEventSender.EXPECT().SendEvent(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

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
}

func TestRevokeAccessToken_VaultRevokes(t *testing.T) {
	tests := []struct {
		name      string
		tokenType token.Type
		parentId  string
		extra     map[string]any
		setupMock func(*mocks.MockClient)
	}{
		{
			name:      "personal",
			tokenType: token.TypePersonal,
			parentId:  "user1",
			setupMock: func(c *mocks.MockClient) {
				c.EXPECT().RevokePersonalAccessToken(mock.Anything, int64(42)).Return(nil).Once()
			},
		},
		{
			name:      "project",
			tokenType: token.TypeProject,
			parentId:  "proj1",
			setupMock: func(c *mocks.MockClient) {
				c.EXPECT().RevokeProjectAccessToken(mock.Anything, int64(42), "proj1").Return(nil).Once()
			},
		},
		{
			name:      "group",
			tokenType: token.TypeGroup,
			parentId:  "grp1",
			setupMock: func(c *mocks.MockClient) {
				c.EXPECT().RevokeGroupAccessToken(mock.Anything, int64(42), "grp1").Return(nil).Once()
			},
		},
		{
			name:      "user service account",
			tokenType: token.TypeUserServiceAccount,
			parentId:  "user1",
			extra:     map[string]any{"token": "glpat-secret"},
			setupMock: func(c *mocks.MockClient) {
				c.EXPECT().RevokeUserServiceAccountAccessToken(mock.Anything, "glpat-secret").Return(nil).Once()
			},
		},
		{
			name:      "group service account",
			tokenType: token.TypeGroupServiceAccount,
			parentId:  "grp1",
			extra:     map[string]any{"token": "glpat-secret"},
			setupMock: func(c *mocks.MockClient) {
				c.EXPECT().RevokeGroupServiceAccountAccessToken(mock.Anything, "glpat-secret").Return(nil).Once()
			},
		},
		{
			name:      "pipeline project trigger",
			tokenType: token.TypePipelineProjectTrigger,
			parentId:  "100",
			setupMock: func(c *mocks.MockClient) {
				c.EXPECT().RevokePipelineProjectTriggerAccessToken(mock.Anything, int64(100), int64(42)).Return(nil).Once()
			},
		},
		{
			name:      "group deploy",
			tokenType: token.TypeGroupDeploy,
			parentId:  "200",
			setupMock: func(c *mocks.MockClient) {
				c.EXPECT().RevokeGroupDeployToken(mock.Anything, int64(200), int64(42)).Return(nil).Once()
			},
		},
		{
			name:      "project deploy",
			tokenType: token.TypeProjectDeploy,
			parentId:  "300",
			setupMock: func(c *mocks.MockClient) {
				c.EXPECT().RevokeProjectDeployToken(mock.Anything, int64(300), int64(42)).Return(nil).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := newMockSecretBackend(t)
			client := mocks.NewMockClient(t)
			tt.setupMock(client)
			mb.MockClientProvider.EXPECT().GetClientByName(mock.Anything, mock.Anything, "default").Return(client, nil).Once()
			mb.MockEventSender.EXPECT().SendEvent(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

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
	mb := newMockSecretBackend(t)
	client := mocks.NewMockClient(t)
	client.EXPECT().RevokePersonalAccessToken(mock.Anything, int64(42)).Return(errs.ErrAccessTokenNotFound).Once()
	mb.MockClientProvider.EXPECT().GetClientByName(mock.Anything, mock.Anything, "default").Return(client, nil).Once()
	mb.MockEventSender.EXPECT().SendEvent(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	s := secret.NewSecret(mb, "default")

	resp, err := s.HandleRevoke(t.Context(), &logical.Request{
		Storage: &logical.InmemStorage{},
		Secret:  newRevokeSecret(token.TypePersonal, "user1", nil),
	})
	require.NoError(t, err)
	require.Nil(t, resp)
}

func TestRevokeAccessToken_UsesConfigNameFromSecret(t *testing.T) {
	mb := newMockSecretBackend(t)
	client := mocks.NewMockClient(t)
	client.EXPECT().RevokePersonalAccessToken(mock.Anything, int64(42)).Return(nil).Once()
	mb.MockClientProvider.EXPECT().GetClientByName(mock.Anything, mock.Anything, "custom-config").Return(client, nil).Once()
	mb.MockEventSender.EXPECT().SendEvent(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

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
