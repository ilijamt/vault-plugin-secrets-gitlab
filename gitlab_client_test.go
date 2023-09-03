package gitlab_test

import (
	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestGitlabClient(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		client, err := gitlab.NewGitlabClient(nil, nil)
		require.Nil(t, client)
		require.ErrorIs(t, err, gitlab.ErrNilValue)
	})

	t.Run("no token", func(t *testing.T) {
		var client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{}, nil)
		require.ErrorIs(t, err, gitlab.ErrInvalidValue)
		require.Nil(t, client)
	})

	t.Run("no base url", func(t *testing.T) {
		var client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{}, nil)
		require.ErrorIs(t, err, gitlab.ErrInvalidValue)
		require.Nil(t, client)
	})

	t.Run("with http client", func(t *testing.T) {
		var client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
			Token:   "token",
			BaseURL: "https://example.com",
		}, &http.Client{})
		require.NoError(t, err)
		require.NotNil(t, client)
	})
}

func TestGitlabClient_InvalidToken(t *testing.T) {
	var err error

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "super-secret-token",
		BaseURL: server.URL,
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, client)

	require.True(t, client.Valid())

	token, err := client.CurrentTokenInfo()
	require.Error(t, err)
	require.Nil(t, token)

	newToken, oldToken, err := client.RotateCurrentToken(true)
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "super-secret-token",
		BaseURL: server.URL,
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, client)

	require.True(t, client.Valid())

	require.ErrorIs(t, client.RevokePersonalAccessToken(1), gitlab.ErrAccessTokenNotFound)
	require.ErrorIs(t, client.RevokeGroupAccessToken(1, "group"), gitlab.ErrAccessTokenNotFound)
	require.ErrorIs(t, client.RevokeProjectAccessToken(1, "project"), gitlab.ErrAccessTokenNotFound)
}

func TestGitlabClient_GetUserIdByUsername(t *testing.T) {
	var err error

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/api/v4/users?username=ilijamt" {
			w.WriteHeader(http.StatusOK)
			data, _ := os.ReadFile("testdata/list_users.json")
			_, _ = w.Write(data)

		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "super-secret-token",
		BaseURL: server.URL,
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid())

	userId, err := client.GetUserIdByUsername("ilijamt")
	require.NoError(t, err)
	require.EqualValues(t, 1, userId)
}

func TestGitlabClient_GetUserIdByUsernameDoesNotMatch(t *testing.T) {
	var err error

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/api/v4/users?username=ilijamt" {
			w.WriteHeader(http.StatusOK)
			data, _ := os.ReadFile("testdata/list_users_username_not_matched.json")
			_, _ = w.Write(data)

		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "super-secret-token",
		BaseURL: server.URL,
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid())

	userId, err := client.GetUserIdByUsername("ilijamt")
	require.ErrorIs(t, err, gitlab.ErrInvalidValue)
	require.NotEqualValues(t, 1, userId)
}

func TestGitlabClient_Revoke_Success(t *testing.T) {
	var err error

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "super-secret-token",
		BaseURL: server.URL,
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid())

	require.NoError(t, client.RevokePersonalAccessToken(1))
	require.NoError(t, client.RevokeGroupAccessToken(1, "group"))
	require.NoError(t, client.RevokeProjectAccessToken(1, "project"))
}

func TestGitlabClient_Revoke_BadRequest(t *testing.T) {
	var err error

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "super-secret-token",
		BaseURL: server.URL,
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid())

	require.Error(t, client.RevokePersonalAccessToken(1))
	require.Error(t, client.RevokeGroupAccessToken(1, "group"))
	require.Error(t, client.RevokeProjectAccessToken(1, "project"))
}

func TestGitlabClient_CurrentTokenInfo(t *testing.T) {
	var err error

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/api/v4/personal_access_tokens/self" {
			w.WriteHeader(http.StatusOK)
			data, _ := os.ReadFile("testdata/personal_access_tokens_self.json")
			_, _ = w.Write(data)

		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "super-secret-token",
		BaseURL: server.URL,
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid())

	token, err := client.CurrentTokenInfo()
	require.NoError(t, err)
	require.NotNil(t, token)
	assert.EqualValues(t, gitlab.TokenTypePersonal, token.TokenType)
}

func TestGitlabClient_CreateAccessToken(t *testing.T) {
	var err error

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/api/v4/personal_access_tokens/self" {
			w.WriteHeader(http.StatusOK)
			data, _ := os.ReadFile("testdata/personal_access_tokens_self.json")
			_, _ = w.Write(data)

		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "super-secret-token",
		BaseURL: server.URL,
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, client.Valid())

	token, err := client.CurrentTokenInfo()
	require.NoError(t, err)
	require.NotNil(t, token)
	assert.EqualValues(t, gitlab.TokenTypePersonal, token.TokenType)
}
