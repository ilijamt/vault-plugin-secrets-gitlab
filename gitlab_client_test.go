package gitlab_test

import (
	"io"
	"net/http"
	"testing"
	"time"

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
		var client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
			Token:   "token",
			BaseURL: "https://example.com",
		}, &http.Client{}, nil)
		require.NoError(t, err)
		require.NotNil(t, client)
		require.ErrorIs(t, client.RevokeGroupServiceAccountAccessToken(""), gitlab.ErrNilValue)
		require.ErrorIs(t, client.RevokeUserServiceAccountAccessToken(""), gitlab.ErrNilValue)
	})
}

func TestGitlabClient_InvalidToken(t *testing.T) {
	var err error
	httpClient, url := getClient(t)
	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "super-secret-token",
		BaseURL: url,
	}, httpClient, nil)
	require.NoError(t, err)
	require.NotNil(t, client)

	require.True(t, client.Valid())

	token, err := client.CurrentTokenInfo()
	require.Error(t, err)
	require.Nil(t, token)

	newToken, oldToken, err := client.RotateCurrentToken()
	require.Error(t, err)
	require.Nil(t, newToken)
	require.Nil(t, oldToken)

	require.Error(t, client.RevokePersonalAccessToken(1))
	require.Error(t, client.RevokeGroupAccessToken(1, "group"))
	require.Error(t, client.RevokeProjectAccessToken(1, "project"))

	_, err = client.GetUserIdByUsername("username")
	require.Error(t, err)

	entryToken, err := client.CreateGroupAccessToken("groupId", "name", time.Now(), []string{"scope"}, gitlab.AccessLevelUnknown)
	require.Error(t, err)
	require.Nil(t, entryToken)

	entryToken, err = client.CreateProjectAccessToken("projectId", "name", time.Now(), []string{"scope"}, gitlab.AccessLevelUnknown)
	require.Error(t, err)
	require.Nil(t, entryToken)

	entryToken, err = client.CreatePersonalAccessToken("username", 0, "name", time.Now(), []string{"scope"})
	require.Error(t, err)
	require.Nil(t, entryToken)
}

func TestGitlabClient_RevokeToken_NotFound(t *testing.T) {
	var err error
	httpClient, url := getClient(t)
	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "glpat-secret-random-token",
		BaseURL: url,
	}, httpClient, nil)
	require.NoError(t, err)
	require.NotNil(t, client)

	require.True(t, client.Valid())

	require.ErrorIs(t, client.RevokePersonalAccessToken(999), gitlab.ErrAccessTokenNotFound)
	require.ErrorIs(t, client.RevokeGroupAccessToken(999, "group"), gitlab.ErrAccessTokenNotFound)
	require.ErrorIs(t, client.RevokeProjectAccessToken(999, "project"), gitlab.ErrAccessTokenNotFound)
}

func TestGitlabClient_GetUserIdByUsername(t *testing.T) {
	var err error
	httpClient, url := getClient(t)
	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "glpat-secret-random-token",
		BaseURL: url,
	}, httpClient, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid())

	userId, err := client.GetUserIdByUsername("root")
	require.NoError(t, err)
	require.EqualValues(t, 1, userId)
}

func TestGitlabClient_GetUserIdByUsernameDoesNotMatch(t *testing.T) {
	var err error
	httpClient, url := getClient(t)
	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "glpat-secret-random-token",
		BaseURL: url,
	}, httpClient, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid())

	userId, err := client.GetUserIdByUsername("ilijamt")
	require.ErrorIs(t, err, gitlab.ErrInvalidValue)
	require.NotEqualValues(t, 1, userId)

	userId, err = client.GetUserIdByUsername("demo")
	require.ErrorIs(t, err, gitlab.ErrInvalidValue)
	require.NotEqualValues(t, 1, userId)
}

func TestGitlabClient_Revoke_NonExistingTokens(t *testing.T) {
	var err error
	httpClient, url := getClient(t)
	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "glpat-secret-random-token",
		BaseURL: url,
	}, httpClient, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid())

	require.Error(t, client.RevokePersonalAccessToken(999))
	require.Error(t, client.RevokeGroupAccessToken(999, "example"))
	require.Error(t, client.RevokeProjectAccessToken(999, "example/example"))
}

func TestGitlabClient_CurrentTokenInfo(t *testing.T) {
	var err error
	httpClient, url := getClient(t)
	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "glpat-secret-random-token",
		BaseURL: url,
	}, httpClient, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid())

	token, err := client.CurrentTokenInfo()
	require.NoError(t, err)
	require.NotNil(t, token)
	assert.EqualValues(t, gitlab.TokenTypePersonal, token.TokenType)
}

func TestGitlabClient_CreateAccessToken_And_Revoke(t *testing.T) {
	var err error
	httpClient, url := getClient(t)
	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "glpat-secret-random-token",
		BaseURL: url,
	}, httpClient, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid())

	entryToken, err := client.CreateGroupAccessToken(
		"example",
		"name",
		time.Now(),
		[]string{gitlab.TokenScopeReadApi.String()},
		gitlab.AccessLevelGuestPermissions,
	)
	require.NoError(t, err)
	require.NotNil(t, entryToken)
	require.EqualValues(t, gitlab.TokenTypeGroup, entryToken.TokenType)
	require.NotEmpty(t, entryToken.Token)
	require.NoError(t, client.RevokeGroupAccessToken(entryToken.TokenID, "example"))

	entryToken, err = client.CreateProjectAccessToken(
		"example/example",
		"name",
		time.Now(),
		[]string{gitlab.TokenScopeReadApi.String()},
		gitlab.AccessLevelDeveloperPermissions,
	)
	require.NoError(t, err)
	require.NotNil(t, entryToken)
	require.EqualValues(t, gitlab.TokenTypeProject, entryToken.TokenType)
	require.NotEmpty(t, entryToken.Token)
	require.NoError(t, client.RevokeProjectAccessToken(entryToken.TokenID, "example/example"))

	entryToken, err = client.CreatePersonalAccessToken(
		"normal-user",
		1,
		"name",
		time.Now(),
		[]string{gitlab.TokenScopeReadApi.String()},
	)
	require.NoError(t, err)
	require.NotNil(t, entryToken)
	require.EqualValues(t, gitlab.TokenTypePersonal, entryToken.TokenType)
	require.NotEmpty(t, entryToken.Token)
	require.NoError(t, client.RevokePersonalAccessToken(entryToken.TokenID))
}

func TestGitlabClient_RotateCurrentToken(t *testing.T) {
	var err error
	httpClient, url := getClient(t)
	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "glpat-secret-admin-token-ar1",
		BaseURL: url,
	}, httpClient, logging.NewVaultLoggerWithWriter(io.Discard, log.Trace))

	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid())

	newToken, oldToken, err := client.RotateCurrentToken()
	require.NoError(t, err)
	require.NotNil(t, newToken)
	require.NotNil(t, oldToken)

	require.NotEqualValues(t, oldToken.Token, newToken.Token)
}
