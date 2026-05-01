//go:build paths

package integration_test

import (
	"io"
	"net/http"
	"testing"

	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/helper/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	glab "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
	token2 "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestGitlabClient(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		client, err := glab.NewGitlabClient(nil, nil, nil)
		require.Nil(t, client)
		require.ErrorIs(t, err, errs.ErrNilValue)
	})

	t.Run("no token", func(t *testing.T) {
		var client, err = glab.NewGitlabClient(&config.EntryConfig{}, nil, nil)
		require.ErrorIs(t, err, errs.ErrInvalidValue)
		require.Nil(t, client)
	})

	t.Run("no base url", func(t *testing.T) {
		var client, err = glab.NewGitlabClient(&config.EntryConfig{}, nil, nil)
		require.ErrorIs(t, err, errs.ErrInvalidValue)
		require.Nil(t, client)
	})

	t.Run("with http client", func(t *testing.T) {
		var client, err = glab.NewGitlabClient(&config.EntryConfig{
			Token:   "token",
			BaseURL: "https://example.com",
		}, &http.Client{}, nil)
		require.NoError(t, err)
		require.NotNil(t, client)
	})

	t.Run("revoke service account token with empty token", func(t *testing.T) {
		var ctx = t.Context()
		var client, err = glab.NewGitlabClient(&config.EntryConfig{
			Token:   "token",
			BaseURL: "https://example.com",
		}, &http.Client{}, nil)
		require.NoError(t, err)
		require.NotNil(t, client)
		require.ErrorIs(t, client.RevokeGroupServiceAccountAccessToken(ctx, ""), errs.ErrNilValue)
		require.ErrorIs(t, client.RevokeUserServiceAccountAccessToken(ctx, ""), errs.ErrNilValue)
	})
}

func TestGitlabClient_InvalidToken(t *testing.T) {
	var tokenName = "super-secret-token"
	ctx, timeExpiresAt := ctxTestTime(t.Context(), t.Name(), tokenName)
	var err error
	httpClient, url := getClient(t, "paths")
	var client glab.Client
	client, err = glab.NewGitlabClient(&config.EntryConfig{
		Token:   tokenName,
		BaseURL: url,
	}, httpClient, nil)
	require.NoError(t, err)
	require.NotNil(t, client)

	require.True(t, client.Valid(ctx))

	token, err := client.CurrentTokenInfo(ctx)
	require.Error(t, err)
	require.Nil(t, token)

	newToken, oldToken, err := client.RotateCurrentToken(ctx)
	require.Error(t, err)
	require.Nil(t, newToken)
	require.Nil(t, oldToken)

	require.Error(t, client.RevokePersonalAccessToken(ctx, 1))
	require.Error(t, client.RevokeGroupAccessToken(ctx, 1, "group"))
	require.Error(t, client.RevokeProjectAccessToken(ctx, 1, "project"))

	_, err = client.GetUserIdByUsername(ctx, "username")
	require.Error(t, err)

	gatToken, err := client.CreateGroupAccessToken(ctx, "groupId", "name", timeExpiresAt, []string{"scope"}, token2.AccessLevelUnknown)
	require.Error(t, err)
	require.Nil(t, gatToken)

	prjAtToken, err := client.CreateProjectAccessToken(ctx, "projectId", "name", timeExpiresAt, []string{"scope"}, token2.AccessLevelUnknown)
	require.Error(t, err)
	require.Nil(t, prjAtToken)

	patToken, err := client.CreatePersonalAccessToken(ctx, "username", 0, "name", timeExpiresAt, []string{"scope"})
	require.Error(t, err)
	require.Nil(t, patToken)
}

func TestGitlabClient_RevokeToken_NotFound(t *testing.T) {
	var ctx = t.Context()
	var err error
	httpClient, url := getClient(t, "paths")
	var client glab.Client
	client, err = glab.NewGitlabClient(&config.EntryConfig{
		Token:   getGitlabToken("admin_user_root").Token,
		BaseURL: url,
	}, httpClient, nil)
	require.NoError(t, err)
	require.NotNil(t, client)

	require.True(t, client.Valid(ctx))

	require.ErrorIs(t, client.RevokePersonalAccessToken(ctx, 999), errs.ErrAccessTokenNotFound)
	require.ErrorIs(t, client.RevokeGroupAccessToken(ctx, 999, "group"), errs.ErrAccessTokenNotFound)
	require.ErrorIs(t, client.RevokeProjectAccessToken(ctx, 999, "project"), errs.ErrAccessTokenNotFound)
}

func TestGitlabClient_GetGroupIdByPath(t *testing.T) {
	var tokenName = "admin_user_root"
	httpClient, url := getClient(t, "paths")
	client, err := glab.NewGitlabClient(&config.EntryConfig{
		Token:   getGitlabToken(tokenName).Token,
		BaseURL: url,
	}, httpClient, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid(t.Context()))

	tests := []struct {
		name        string
		path        string
		expectedId  int64
		expectedErr error
	}{
		{name: "top-level group", path: "test", expectedId: 2},
		{name: "nested group", path: "level-1/level-2/level-3"},
		{name: "first level nested group", path: "level-1/level-2"},
		{name: "nonexistent group", path: "nonexistent", expectedErr: errs.ErrInvalidValue},
		{name: "nonexistent second level nested group", path: "level-1/level-2/nonexistent", expectedErr: errs.ErrInvalidValue},
		{name: "nonexistent first level nested group", path: "level-1/nonexistent/level-3", expectedErr: errs.ErrInvalidValue},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			groupId, err := client.GetGroupIdByPath(t.Context(), tc.path)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				if tc.expectedId > 0 {
					require.EqualValues(t, tc.expectedId, groupId)
				} else {
					require.Greater(t, groupId, int64(0))
				}
			}
		})
	}
}

func TestGitlabClient_GetUserIdByUsername(t *testing.T) {
	var ctx = t.Context()
	var err error
	httpClient, url := getClient(t, "paths")
	var client glab.Client
	client, err = glab.NewGitlabClient(&config.EntryConfig{
		Token:   getGitlabToken("admin_user_root").Token,
		BaseURL: url,
	}, httpClient, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid(ctx))

	userId, err := client.GetUserIdByUsername(ctx, "root")
	require.NoError(t, err)
	require.EqualValues(t, 1, userId)
}

func TestGitlabClient_GetUserIdByUsernameDoesNotMatch(t *testing.T) {
	var err error
	var ctx = t.Context()
	httpClient, url := getClient(t, "paths")
	var client glab.Client
	client, err = glab.NewGitlabClient(&config.EntryConfig{
		Token:   getGitlabToken("admin_user_root").Token,
		BaseURL: url,
	}, httpClient, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid(ctx))

	userId, err := client.GetUserIdByUsername(ctx, "ilijamt")
	require.ErrorIs(t, err, errs.ErrInvalidValue)
	require.NotEqualValues(t, 1, userId)

	userId, err = client.GetUserIdByUsername(ctx, "demo")
	require.ErrorIs(t, err, errs.ErrInvalidValue)
	require.NotEqualValues(t, 1, userId)
}

func TestGitlabClient_Revoke_NonExistingTokens(t *testing.T) {
	var ctx = t.Context()
	var err error
	httpClient, url := getClient(t, "paths")
	var client glab.Client
	client, err = glab.NewGitlabClient(&config.EntryConfig{
		Token:   getGitlabToken("admin_user_root").Token,
		BaseURL: url,
	}, httpClient, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid(ctx))

	require.Error(t, client.RevokePersonalAccessToken(ctx, 999))
	require.Error(t, client.RevokeGroupAccessToken(ctx, 999, "example"))
	require.Error(t, client.RevokeProjectAccessToken(ctx, 999, "example/example"))
}

func TestGitlabClient_CurrentTokenInfo(t *testing.T) {
	var err error
	var ctx = t.Context()
	httpClient, url := getClient(t, "paths")
	var client glab.Client
	client, err = glab.NewGitlabClient(&config.EntryConfig{
		Token:   getGitlabToken("admin_user_root").Token,
		BaseURL: url,
	}, httpClient, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid(ctx))

	token, err := client.CurrentTokenInfo(ctx)
	require.NoError(t, err)
	require.NotNil(t, token)
	assert.EqualValues(t, token2.TypePersonal, token.TokenType)
}

func TestGitlabClient_Metadata(t *testing.T) {
	var err error
	var ctx = t.Context()
	httpClient, url := getClient(t, "paths")
	var client glab.Client
	client, err = glab.NewGitlabClient(&config.EntryConfig{
		Token:   getGitlabToken("admin_user_root").Token,
		BaseURL: url,
	}, httpClient, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid(ctx))

	metadata, err := client.Metadata(ctx)
	require.NoError(t, err)
	require.NotNil(t, metadata)
}

func TestGitlabClient_CreateAccessToken_And_Revoke(t *testing.T) {
	var err error
	var tokenName = "admin_user_root"
	ctx, timeExpiresAt := ctxTestTime(t.Context(), t.Name(), tokenName)
	httpClient, url := getClient(t, "paths")
	var client glab.Client
	client, err = glab.NewGitlabClient(&config.EntryConfig{
		Token:   getGitlabToken(tokenName).Token,
		BaseURL: url,
	}, httpClient, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid(ctx))

	gatToken, err := client.CreateGroupAccessToken(
		ctx,
		"example",
		"name",
		timeExpiresAt,
		[]string{token2.ScopeReadApi.String()},
		token2.AccessLevelGuestPermissions,
	)
	require.NoError(t, err)
	require.NotNil(t, gatToken)
	require.EqualValues(t, token2.TypeGroup, gatToken.TokenType)
	require.NotEmpty(t, gatToken.Token)
	require.NoError(t, client.RevokeGroupAccessToken(ctx, gatToken.TokenID, "example"))

	prjatToken, err := client.CreateProjectAccessToken(
		ctx,
		"example/example",
		"name",
		timeExpiresAt,
		[]string{token2.ScopeReadApi.String()},
		token2.AccessLevelDeveloperPermissions,
	)
	require.NoError(t, err)
	require.NotNil(t, prjatToken)
	require.EqualValues(t, token2.TypeProject, prjatToken.TokenType)
	require.NotEmpty(t, prjatToken.Token)
	require.NoError(t, client.RevokeProjectAccessToken(ctx, prjatToken.TokenID, "example/example"))

	patToken, err := client.CreatePersonalAccessToken(
		ctx,
		"normal-user",
		1,
		"name",
		timeExpiresAt,
		[]string{token2.ScopeReadApi.String()},
	)
	require.NoError(t, err)
	require.NotNil(t, patToken)
	require.EqualValues(t, token2.TypePersonal, patToken.TokenType)
	require.NotEmpty(t, patToken.Token)
	require.NoError(t, client.RevokePersonalAccessToken(ctx, patToken.TokenID))
}

func TestGitlabClient_RotateCurrentToken(t *testing.T) {
	var err error
	var ctx = t.Context()
	httpClient, url := getClient(t, "paths")
	var client glab.Client
	var tokenName = "admin_user_auto_rotate_token_1"
	client, err = glab.NewGitlabClient(&config.EntryConfig{
		Token:   getGitlabToken(tokenName).Token,
		BaseURL: url,
	}, httpClient, logging.NewVaultLoggerWithWriter(io.Discard, log.Trace))

	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid(ctx))
	ctx, _ = ctxTestTime(ctx, t.Name(), tokenName)
	newToken, oldToken, err := client.RotateCurrentToken(ctx)
	require.NoError(t, err)
	require.NotNil(t, newToken)
	require.NotNil(t, oldToken)

	require.NotEqualValues(t, oldToken.Token, newToken.Token)
}
