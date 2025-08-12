//go:build unit

package gitlab_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

func TestEmptyGitlabClientFromContext(t *testing.T) {
	c, ok := gitlab.ClientFromContext(t.Context())
	require.False(t, ok)
	require.Nil(t, c)
}

func TestEmptyHttpClientFromContext(t *testing.T) {
	c, ok := utils.HttpClientFromContext(t.Context())
	require.False(t, ok)
	require.Nil(t, c)
}
