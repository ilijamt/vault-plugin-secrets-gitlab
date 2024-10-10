//go:build !integration

package gitlab_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestEmptyGitlabClientFromContext(t *testing.T) {
	c, ok := gitlab.GitlabClientFromContext(context.Background())
	require.False(t, ok)
	require.Nil(t, c)
}

func TestEmptyHttpClientFromContext(t *testing.T) {
	c, ok := gitlab.HttpClientFromContext(context.Background())
	require.False(t, ok)
	require.Nil(t, c)
}
