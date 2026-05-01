//go:build paths

package integration_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
)

func TestBackend(t *testing.T) {
	var err error
	var b *gitlab.Backend
	ctx := getCtxGitlabClient(t, "paths")
	b, _, err = getBackend(ctx)
	require.NoError(t, err)
	require.NotNil(t, b)
	require.Nil(t, b.GetClient(backend.DefaultConfigName))
	b.SetClient(newInMemoryClient(true), backend.DefaultConfigName)
	require.NotNil(t, b.GetClient(backend.DefaultConfigName))
	b.Invalidate(ctx, fmt.Sprintf("%s/%s", backend.PathConfigStorage, backend.DefaultConfigName))
	require.Nil(t, b.GetClient(backend.DefaultConfigName))
	b.SetClient(newInMemoryClient(true), backend.DefaultConfigName)
	require.NotNil(t, b.GetClient(backend.DefaultConfigName))
	require.EqualValues(t, gitlab.Version, b.PluginVersion().Version)
}
