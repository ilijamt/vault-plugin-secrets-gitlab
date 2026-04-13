package token_test

import (
	"crypto/sha1"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	modelToken "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestToken(t *testing.T) {
	t.Run("zero value gives 0 ttl", func(t *testing.T) {
		require.EqualValues(t, 0, (&modelToken.Token{}).TTL())
	})

	t.Run("ttl from created and expires", func(t *testing.T) {
		cat := time.Date(2025, 1, 1, 1, 0, 0, 0, time.UTC)
		eat := time.Date(2025, 1, 1, 2, 0, 0, 0, time.UTC)
		data := modelToken.Token{CreatedAt: &cat}
		data.SetExpiresAt(&eat)
		require.EqualValues(t, time.Hour, data.TTL())
	})

	t.Run("setters", func(t *testing.T) {
		data := modelToken.Token{}
		data.SetRoleName("role-name")
		data.SetConfigName("config-name")
		data.SetGitlabRevokesToken(true)

		require.EqualValues(t, "role-name", data.RoleName)
		require.EqualValues(t, "config-name", data.ConfigName)
		require.True(t, data.GitlabRevokesToken)
	})

	t.Run("data contains sha1 hash", func(t *testing.T) {
		d := (&modelToken.Token{Token: "secret-token"}).Data()
		assert.Equal(t, fmt.Sprintf("%x", sha1.Sum([]byte("secret-token"))), d["token_sha1_hash"])
	})

	t.Run("type returns token type", func(t *testing.T) {
		assert.Equal(t, token.TypePersonal, (&modelToken.Token{TokenType: token.TypePersonal}).Type())
	})
}

func TestTokenWithScopes(t *testing.T) {
	data := &modelToken.TokenWithScopes{Scopes: []string{"scope1", "scope2"}}
	assert.EqualValues(t, []string{"scope1", "scope2"}, data.Data()["scopes"])
	assert.EqualValues(t, []string{"scope1", "scope2"}, data.Internal()["scopes"])
	assert.EqualValues(t, "scope1,scope2", data.Event(nil)["scopes"])
}

func TestTokenWithScopesAndAccessLevel(t *testing.T) {
	data := &modelToken.TokenWithScopesAndAccessLevel{
		Scopes:      []string{"scope1", "scope2"},
		AccessLevel: token.AccessLevelNoPermissions,
	}
	assert.EqualValues(t, []string{"scope1", "scope2"}, data.Data()["scopes"])
	assert.EqualValues(t, []string{"scope1", "scope2"}, data.Internal()["scopes"])
	assert.EqualValues(t, "scope1,scope2", data.Event(nil)["scopes"])
	assert.EqualValues(t, token.AccessLevelNoPermissions, data.Data()["access_level"])
	assert.EqualValues(t, token.AccessLevelNoPermissions, data.Internal()["access_level"])
	assert.EqualValues(t, token.AccessLevelNoPermissions.String(), data.Event(nil)["access_level"])
}

func TestTokenVariants(t *testing.T) {
	tests := []struct {
		name      string
		tok       token.Token
		wantKey   string
		wantData  any
		wantEvent string
		scopes    string
	}{
		{
			name:      "TokenConfig",
			tok:       &modelToken.TokenConfig{TokenWithScopes: modelToken.TokenWithScopes{Scopes: []string{"api", "read_user"}}, UserID: 1},
			wantKey:   "user_id",
			wantData:  int64(1),
			wantEvent: "1",
			scopes:    "api,read_user",
		},
		{
			name:      "TokenPersonal",
			tok:       &modelToken.TokenPersonal{TokenWithScopes: modelToken.TokenWithScopes{Scopes: []string{"api", "read_user"}}, UserID: 1},
			wantKey:   "user_id",
			wantData:  int64(1),
			wantEvent: "1",
			scopes:    "api,read_user",
		},
		{
			name:      "TokenGroupServiceAccount",
			tok:       &modelToken.TokenGroupServiceAccount{TokenWithScopes: modelToken.TokenWithScopes{Scopes: []string{"api"}}, UserID: 1},
			wantKey:   "user_id",
			wantData:  int64(1),
			wantEvent: "1",
			scopes:    "api",
		},
		{
			name:      "TokenProjectDeploy",
			tok:       &modelToken.TokenProjectDeploy{TokenWithScopes: modelToken.TokenWithScopes{Scopes: []string{"read_repository"}}, Username: "deploy-bot"},
			wantKey:   "username",
			wantData:  "deploy-bot",
			wantEvent: "deploy-bot",
			scopes:    "read_repository",
		},
		{
			name:      "TokenGroupDeploy",
			tok:       &modelToken.TokenGroupDeploy{TokenWithScopes: modelToken.TokenWithScopes{Scopes: []string{"read_repository"}}, Username: "deploy-bot"},
			wantKey:   "username",
			wantData:  "deploy-bot",
			wantEvent: "deploy-bot",
			scopes:    "read_repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualValues(t, tt.wantData, tt.tok.Data()[tt.wantKey])
			assert.EqualValues(t, tt.wantData, tt.tok.Internal()[tt.wantKey])
			assert.EqualValues(t, tt.wantEvent, tt.tok.Event(nil)[tt.wantKey])
			assert.EqualValues(t, tt.scopes, tt.tok.Event(nil)["scopes"])
		})
	}
}
