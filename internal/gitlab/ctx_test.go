package gitlab_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
)

func TestEmptyGitlabClientFromContext(t *testing.T) {
	c, ok := gitlab.ClientFromContext(t.Context())
	require.False(t, ok)
	require.Nil(t, c)
}
