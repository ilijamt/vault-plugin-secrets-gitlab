package gitlab_test

import (
	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGitlabClient(t *testing.T) {
	r, err := getVcr("fixtures/gitlab-client")
	require.NoError(t, err)
	require.NotNil(t, r)
	defer func() {
		require.NoError(t, r.Stop())
	}()

	t.Run("nil config", func(t *testing.T) {
		client, err := gitlab.NewGitlabClient(nil, nil)
		require.Nil(t, client)
		require.ErrorIs(t, err, gitlab.ErrNilValue)
	})
}
