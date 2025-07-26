package models_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/models"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestToken(t *testing.T) {
	t.Run("expires is not set so we get a 0 ttl", func(t *testing.T) {
		data := models.Token{}
		require.EqualValues(t, 0, data.TTL())
	})

	t.Run("ttl has a value if both created and expires are set", func(t *testing.T) {
		cat := time.Date(2025, 1, 1, 1, 0, 0, 0, time.UTC)
		data := models.Token{CreatedAt: &cat}
		eat := time.Date(2025, 1, 1, 2, 0, 0, 0, time.UTC)
		data.SetExpiresAt(&eat)
		require.EqualValues(t, time.Hour, data.TTL())
	})

	t.Run("setters", func(t *testing.T) {
		data := models.Token{}
		data.SetRoleName("role-name")
		data.SetConfigName("config-name")
		data.SetGitlabRevokesToken(true)

		require.EqualValues(t, "role-name", data.RoleName)
		require.EqualValues(t, "config-name", data.ConfigName)
		require.EqualValues(t, true, data.GitlabRevokesToken)

	})
}

func TestTokenWithScopes(t *testing.T) {
	data := models.TokenWithScopes{Scopes: []string{"scope1", "scope2"}}
	assert.Contains(t, data.Data(), "scopes")
	assert.Contains(t, data.Event(nil), "scopes")
	assert.Contains(t, data.Internal(), "scopes")
	assert.EqualValues(t, []string{"scope1", "scope2"}, data.Data()["scopes"])
	assert.EqualValues(t, "scope1,scope2", data.Event(nil)["scopes"])
	assert.EqualValues(t, []string{"scope1", "scope2"}, data.Internal()["scopes"])
}

func TestTokenWithScopesAndAccessLevel(t *testing.T) {
	data := models.TokenWithScopesAndAccessLevel{
		Scopes:      []string{"scope1", "scope2"},
		AccessLevel: token.AccessLevelNoPermissions,
	}
	assert.Contains(t, data.Data(), "scopes")
	assert.Contains(t, data.Event(nil), "scopes")
	assert.Contains(t, data.Internal(), "scopes")
	assert.EqualValues(t, []string{"scope1", "scope2"}, data.Data()["scopes"])
	assert.EqualValues(t, "scope1,scope2", data.Event(nil)["scopes"])
	assert.EqualValues(t, []string{"scope1", "scope2"}, data.Internal()["scopes"])
	assert.Contains(t, data.Data(), "access_level")
	assert.Contains(t, data.Event(nil), "access_level")
	assert.Contains(t, data.Internal(), "access_level")
	assert.EqualValues(t, token.AccessLevelNoPermissions, data.Data()["access_level"])
	assert.EqualValues(t, token.AccessLevelNoPermissions, data.Event(nil)["access_level"])
	assert.EqualValues(t, token.AccessLevelNoPermissions, data.Internal()["access_level"])
}
