package secret_test

import (
	"errors"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/mocks"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/secret"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestRevokeAccessToken_NilStorage(t *testing.T) {
	mb := newMockSecretBackend(t)
	s := secret.NewSecret(mb, "default")

	resp, err := s.HandleRevoke(t.Context(), &logical.Request{})
	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorIs(t, err, errs.ErrNilValue)
}

func TestRevokeAccessToken_NilSecret(t *testing.T) {
	mb := newMockSecretBackend(t)
	s := secret.NewSecret(mb, "default")

	resp, err := s.HandleRevoke(t.Context(), &logical.Request{
		Storage: &logical.InmemStorage{},
	})
	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorIs(t, err, errs.ErrNilValue)
}

func TestRevokeAccessToken_InvalidTokenId(t *testing.T) {
	mb := newMockSecretBackend(t)
	s := secret.NewSecret(mb, "default")

	resp, err := s.HandleRevoke(t.Context(), &logical.Request{
		Storage: &logical.InmemStorage{},
		Secret: &logical.Secret{
			InternalData: map[string]any{
				"token_id": "not-a-number",
			},
		},
	})
	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorIs(t, err, errs.ErrInvalidValue)
}

func TestRevokeAccessToken_ClientError(t *testing.T) {
	mb := newMockSecretBackend(t)
	mb.MockClientProvider.EXPECT().GetClientByName(mock.Anything, mock.Anything, "default").Return(nil, errors.New("client error")).Once()

	s := secret.NewSecret(mb, "default")

	resp, err := s.HandleRevoke(t.Context(), &logical.Request{
		Storage: &logical.InmemStorage{},
		Secret: &logical.Secret{
			InternalData: map[string]any{
				"token_id":             int64(1),
				"gitlab_revokes_token": false,
				"parent_id":            "user1",
				"token_type":           token.TypePersonal.String(),
				"path":                 "user1",
				"name":                 "test-token",
				"config_name":          "default",
			},
		},
	})
	require.Error(t, err)
	require.Nil(t, resp)
}

func TestRevokeAccessToken_RevokeError(t *testing.T) {
	mb := newMockSecretBackend(t)
	client := mocks.NewMockClient(t)
	client.EXPECT().RevokePersonalAccessToken(mock.Anything, int64(42)).Return(errors.New("revoke failed")).Once()
	mb.MockClientProvider.EXPECT().GetClientByName(mock.Anything, mock.Anything, "default").Return(client, nil).Once()

	s := secret.NewSecret(mb, "default")

	resp, err := s.HandleRevoke(t.Context(), &logical.Request{
		Storage: &logical.InmemStorage{},
		Secret:  newRevokeSecret(token.TypePersonal, "user1", nil),
	})
	require.Error(t, err)
	require.NotNil(t, resp)
}
