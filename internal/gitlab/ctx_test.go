package gitlab_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
)

type stubClient struct {
	gitlab.Client
}

func TestEmptyGitlabClientFromContext(t *testing.T) {
	ctx := gitlab.ClientNewContext(t.Context(), nil)
	c, ok := gitlab.ClientFromContext(ctx)
	require.False(t, ok)
	require.Nil(t, c)
}

func TestGitlabClientFromContext(t *testing.T) {
	want := &stubClient{}
	ctx := gitlab.ClientNewContext(t.Context(), want)
	got, ok := gitlab.ClientFromContext(ctx)
	require.True(t, ok)
	require.Same(t, want, got)
}
