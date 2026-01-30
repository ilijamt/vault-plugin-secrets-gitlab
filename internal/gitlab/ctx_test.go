package gitlab_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
)

func TestEmptyGitlabClientFromContext(t *testing.T) {
	ctx := gitlab.ClientNewContext(t.Context(), nil)
	c, ok := gitlab.ClientFromContext(ctx)
	require.False(t, ok)
	require.Nil(t, c)
}
