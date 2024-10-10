//go:build !integration

package gitlab_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestBackend(t *testing.T) {
	var err error
	var b *gitlab.Backend
	ctx := getCtxGitlabClient(t)
	b, _, err = getBackend(ctx)
	require.NoError(t, err)
	require.NotNil(t, b)
	fv := reflect.ValueOf(b).Elem().FieldByName("client")
	require.True(t, fv.IsNil())
	b.SetClient(newInMemoryClient(true))
	require.False(t, fv.IsNil())
	b.Invalidate(ctx, gitlab.PathConfigStorage)
	require.True(t, fv.IsNil())
	b.SetClient(newInMemoryClient(true))
	require.False(t, fv.IsNil())
	require.EqualValues(t, gitlab.Version, b.PluginVersion().Version)
}
