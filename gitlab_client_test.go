package gitlab_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/helper/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestGitlabClient(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		client, err := gitlab.NewGitlabClient(nil, nil, nil)
		require.Nil(t, client)
		require.ErrorIs(t, err, gitlab.ErrNilValue)
	})

	t.Run("no token", func(t *testing.T) {
		var client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{}, nil, nil)
		require.ErrorIs(t, err, gitlab.ErrInvalidValue)
		require.Nil(t, client)
	})

	t.Run("no base url", func(t *testing.T) {
		var client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{}, nil, nil)
		require.ErrorIs(t, err, gitlab.ErrInvalidValue)
		require.Nil(t, client)
	})

	t.Run("with http client", func(t *testing.T) {
		var client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
			Token:   "token",
			BaseURL: "https://example.com",
		}, &http.Client{}, nil)
		require.NoError(t, err)
		require.NotNil(t, client)
	})

	t.Run("revoke service account token with empty token", func(t *testing.T) {
		var ctx = context.Background()
		var client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
			Token:   "token",
			BaseURL: "https://example.com",
		}, &http.Client{}, nil)
		require.NoError(t, err)
		require.NotNil(t, client)
		require.ErrorIs(t, client.RevokeGroupServiceAccountAccessToken(ctx, ""), gitlab.ErrNilValue)
		require.ErrorIs(t, client.RevokeUserServiceAccountAccessToken(ctx, ""), gitlab.ErrNilValue)
	})
}

func TestGitlabClient_InvalidToken(t *testing.T) {
	ctx, timeExpiresAt := ctxTestTime(context.Background(), t.Name())
	var err error
	httpClient, url := getClient(t)
	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "super-secret-token",
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

	entryToken, err := client.CreateGroupAccessToken(ctx, "groupId", "name", timeExpiresAt, []string{"scope"}, gitlab.AccessLevelUnknown)
	require.Error(t, err)
	require.Nil(t, entryToken)

	entryToken, err = client.CreateProjectAccessToken(ctx, "projectId", "name", timeExpiresAt, []string{"scope"}, gitlab.AccessLevelUnknown)
	require.Error(t, err)
	require.Nil(t, entryToken)

	entryToken, err = client.CreatePersonalAccessToken(ctx, "username", 0, "name", timeExpiresAt, []string{"scope"})
	require.Error(t, err)
	require.Nil(t, entryToken)
}

func TestGitlabClient_RevokeToken_NotFound(t *testing.T) {
	var ctx = context.Background()
	var err error
	httpClient, url := getClient(t)
	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "glpat-secret-random-token",
		BaseURL: url,
	}, httpClient, nil)
	require.NoError(t, err)
	require.NotNil(t, client)

	require.True(t, client.Valid(ctx))

	require.ErrorIs(t, client.RevokePersonalAccessToken(ctx, 999), gitlab.ErrAccessTokenNotFound)
	require.ErrorIs(t, client.RevokeGroupAccessToken(ctx, 999, "group"), gitlab.ErrAccessTokenNotFound)
	require.ErrorIs(t, client.RevokeProjectAccessToken(ctx, 999, "project"), gitlab.ErrAccessTokenNotFound)
}

func TestGitlabClient_GetGroupIdByPath(t *testing.T) {
	var ctx = context.Background()
	var err error
	httpClient, url := getClient(t)
	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "glpat-secret-random-token",
		BaseURL: url,
	}, httpClient, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid(ctx))

	groupId, err := client.GetGroupIdByPath(ctx, "test")
	require.NoError(t, err)
	require.EqualValues(t, 37, groupId)

	_, err = client.GetGroupIdByPath(ctx, "nonexistent")
	require.ErrorIs(t, err, gitlab.ErrInvalidValue)
}

func TestGitlabClient_GetUserIdByUsername(t *testing.T) {
	var ctx = context.Background()
	var err error
	httpClient, url := getClient(t)
	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "glpat-secret-random-token",
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
	var ctx = context.Background()
	httpClient, url := getClient(t)
	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "glpat-secret-random-token",
		BaseURL: url,
	}, httpClient, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid(ctx))

	userId, err := client.GetUserIdByUsername(ctx, "ilijamt")
	require.ErrorIs(t, err, gitlab.ErrInvalidValue)
	require.NotEqualValues(t, 1, userId)

	userId, err = client.GetUserIdByUsername(ctx, "demo")
	require.ErrorIs(t, err, gitlab.ErrInvalidValue)
	require.NotEqualValues(t, 1, userId)
}

func TestGitlabClient_Revoke_NonExistingTokens(t *testing.T) {
	var ctx = context.Background()
	var err error
	httpClient, url := getClient(t)
	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "glpat-secret-random-token",
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
	var ctx = context.Background()
	httpClient, url := getClient(t)
	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "glpat-secret-random-token",
		BaseURL: url,
	}, httpClient, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid(ctx))

	token, err := client.CurrentTokenInfo(ctx)
	require.NoError(t, err)
	require.NotNil(t, token)
	assert.EqualValues(t, gitlab.TokenTypePersonal, token.TokenType)
}

func TestGitlabClient_CreateAccessToken_And_Revoke(t *testing.T) {
	var err error
	ctx, timeExpiresAt := ctxTestTime(context.Background(), t.Name())
	httpClient, url := getClient(t)
	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "glpat-secret-random-token",
		BaseURL: url,
	}, httpClient, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid(ctx))

	entryToken, err := client.CreateGroupAccessToken(
		ctx,
		"example",
		"name",
		timeExpiresAt,
		[]string{gitlab.TokenScopeReadApi.String()},
		gitlab.AccessLevelGuestPermissions,
	)
	require.NoError(t, err)
	require.NotNil(t, entryToken)
	require.EqualValues(t, gitlab.TokenTypeGroup, entryToken.TokenType)
	require.NotEmpty(t, entryToken.Token)
	require.NoError(t, client.RevokeGroupAccessToken(ctx, entryToken.TokenID, "example"))

	entryToken, err = client.CreateProjectAccessToken(
		ctx,
		"example/example",
		"name",
		timeExpiresAt,
		[]string{gitlab.TokenScopeReadApi.String()},
		gitlab.AccessLevelDeveloperPermissions,
	)
	require.NoError(t, err)
	require.NotNil(t, entryToken)
	require.EqualValues(t, gitlab.TokenTypeProject, entryToken.TokenType)
	require.NotEmpty(t, entryToken.Token)
	require.NoError(t, client.RevokeProjectAccessToken(ctx, entryToken.TokenID, "example/example"))

	entryToken, err = client.CreatePersonalAccessToken(
		ctx,
		"normal-user",
		1,
		"name",
		timeExpiresAt,
		[]string{gitlab.TokenScopeReadApi.String()},
	)
	require.NoError(t, err)
	require.NotNil(t, entryToken)
	require.EqualValues(t, gitlab.TokenTypePersonal, entryToken.TokenType)
	require.NotEmpty(t, entryToken.Token)
	require.NoError(t, client.RevokePersonalAccessToken(ctx, entryToken.TokenID))
}

func TestGitlabClient_RotateCurrentToken(t *testing.T) {
	var err error
	var ctx = context.Background()
	httpClient, url := getClient(t)
	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "glpat-secret-admin-token-ar1",
		BaseURL: url,
	}, httpClient, logging.NewVaultLoggerWithWriter(io.Discard, log.Trace))

	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid(ctx))
	ctx, _ = ctxTestTime(ctx, t.Name())
	newToken, oldToken, err := client.RotateCurrentToken(ctx)
	require.NoError(t, err)
	require.NotNil(t, newToken)
	require.NotNil(t, oldToken)

	require.NotEqualValues(t, oldToken.Token, newToken.Token)
}
