package gitlab_test

import (
	"context"
	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestBackend(t *testing.T) {
	var err error
	var b *gitlab.Backend
	b, _, err = getBackend()
	require.NoError(t, err)
	require.NotNil(t, b)
	fv := reflect.ValueOf(b).Elem().FieldByName("client")
	require.True(t, fv.IsNil())
	b.SetClient(newInMemoryClient(true))
	require.False(t, fv.IsNil())
	b.Invalidate(context.Background(), gitlab.PathConfigStorage)
	require.True(t, fv.IsNil())
	b.SetClient(newInMemoryClient(true))
	require.False(t, fv.IsNil())
	require.EqualValues(t, gitlab.Version, b.PluginVersion().Version)
}
