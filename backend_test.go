//go:build unit

package gitlab_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestBackend(t *testing.T) {
	var err error
	var b *gitlab.Backend
	ctx := getCtxGitlabClient(t, "unit")
	b, _, err = getBackend(ctx)
	require.NoError(t, err)
	require.NotNil(t, b)
	require.Nil(t, b.GetClient(gitlab.DefaultConfigName))
	b.SetClient(newInMemoryClient(true), gitlab.DefaultConfigName)
	require.NotNil(t, b.GetClient(gitlab.DefaultConfigName))
	b.Invalidate(ctx, fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName))
	require.Nil(t, b.GetClient(gitlab.DefaultConfigName))
	b.SetClient(newInMemoryClient(true), gitlab.DefaultConfigName)
	require.NotNil(t, b.GetClient(gitlab.DefaultConfigName))
	require.EqualValues(t, gitlab.Version, b.PluginVersion().Version)
}
