package secret_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/secret"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestRevokeAccessToken_NilStorage(t *testing.T) {
	mb := &mockSecretBackend{}
	s := secret.NewSecret(mb, "default")

	resp, err := s.HandleRevoke(t.Context(), &logical.Request{})
	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorIs(t, err, errs.ErrNilValue)
}

func TestRevokeAccessToken_NilSecret(t *testing.T) {
	mb := &mockSecretBackend{}
	s := secret.NewSecret(mb, "default")

	resp, err := s.HandleRevoke(t.Context(), &logical.Request{
		Storage: &logical.InmemStorage{},
	})
	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorIs(t, err, errs.ErrNilValue)
}

func TestRevokeAccessToken_InvalidTokenId(t *testing.T) {
	mb := &mockSecretBackend{}
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
	mb := &mockSecretBackend{
		getClientByName: func(_ context.Context, _ logical.Storage, name string) (gitlab.Client, error) {
			require.Equal(t, "default", name)
			return nil, errors.New("client error")
		},
	}

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
	client := &stubClient{
		revokePersonalAccessToken: func(_ context.Context, tokenId int64) error {
			require.Equal(t, int64(42), tokenId)
			return errors.New("revoke failed")
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
	require.Error(t, err)
	require.NotNil(t, resp)
}
