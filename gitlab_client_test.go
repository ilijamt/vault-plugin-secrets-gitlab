package gitlab_test

import (
	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGitlabClient(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		client, err := gitlab.NewGitlabClient(nil)
		require.Nil(t, client)
		require.ErrorIs(t, err, gitlab.ErrNilValue)
	})
}
