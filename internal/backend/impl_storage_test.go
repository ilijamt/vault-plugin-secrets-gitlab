package backend_test

import (
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/role"
)

func TestConfigAndRoleStore(t *testing.T) {
	b := newTestBackend(t, flags.Flags{})
	s := &logical.InmemStorage{}
	ctx := t.Context()

	got, err := b.GetConfig(ctx, s, "nope")
	require.NoError(t, err)
	assert.Nil(t, got)

	require.NoError(t, b.SaveConfig(ctx, &config.EntryConfig{Name: "cfg1", BaseURL: "https://gl.io"}, s))
	got, err = b.GetConfig(ctx, s, "cfg1")
	require.NoError(t, err)
	assert.Equal(t, "cfg1", got.Name)

	r, err := b.GetRole(ctx, "nope", s)
	require.NoError(t, err)
	assert.Nil(t, r)

	entry, err := logical.StorageEntryJSON("roles/myrole", &role.Role{RoleName: "myrole", Path: "grp/proj"})
	require.NoError(t, err)
	require.NoError(t, s.Put(ctx, entry))
	r, err = b.GetRole(ctx, "myrole", s)
	require.NoError(t, err)
	assert.Equal(t, "myrole", r.RoleName)
}
