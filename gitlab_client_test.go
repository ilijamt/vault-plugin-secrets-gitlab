package gitlab_test

import (
	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	"github.com/stretchr/testify/require"
	"net/http"
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
}

func TestGitlabClient_InvalidTokenAndUrl(t *testing.T) {
	var err error
	var httpClient *http.Client

	//var r *recorder.Recorder
	//r, err = getVcr("fixtures/gitlab-client-invalid-credentials")
	//require.NoError(t, err)
	//require.NotNil(t, r)
	//defer func() {
	//	require.NoError(t, r.Stop())
	//}()
	//httpClient = r.GetDefaultClient()

	var client gitlab.Client
	client, err = gitlab.NewGitlabClient(&gitlab.EntryConfig{
		Token:   "super-secret-token",
		BaseURL: "https://git.example.com",
	}, httpClient)
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
