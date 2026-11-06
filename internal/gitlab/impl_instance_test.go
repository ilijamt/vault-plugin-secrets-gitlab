package gitlab_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	modelConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
)

const settingsSecret = "must-not-be-stored"

func newInstanceClient(t *testing.T, userStatus, settingsStatus int) gitlab.Client {
	t.Helper()
	return newInstanceClientWith(t, userStatus, settingsStatus, map[string]any{
		"enforce_granular_tokens":                   true,
		"granular_tokens_enforced_after":            "2026-12-01",
		"max_personal_access_token_lifetime":        30,
		"require_personal_access_token_expiry":      true,
		"service_access_tokens_expiration_enforced": true,
		"disable_personal_access_tokens":            true,
		"slack_app_verification_token":              settingsSecret,
	})
}

func newInstanceClientWith(t *testing.T, userStatus, settingsStatus int, settings map[string]any) gitlab.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v4/user":
			w.WriteHeader(userStatus)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "is_admin": true})
		case "/api/v4/application/settings":
			w.WriteHeader(settingsStatus)
			_ = json.NewEncoder(w).Encode(settings)
		}
	}))
	t.Cleanup(srv.Close)
	client, err := gitlab.NewGitlabClient(&modelConfig.EntryConfig{BaseURL: srv.URL, Token: "config-token"}, srv.Client(), nil)
	require.NoError(t, err)
	return client
}

func TestInstanceInfo(t *testing.T) {
	disabled := true
	t.Run("reads only the allowed settings", func(t *testing.T) {
		info, err := newInstanceClient(t, http.StatusOK, http.StatusOK).InstanceInfo(t.Context())
		require.NoError(t, err)
		assert.Equal(t, &modelConfig.InstanceInfo{
			TokenOwnerIsAdmin:           true,
			SettingsKnown:               true,
			EnforceGranularTokens:       true,
			GranularTokensEnforcedAfter: "2026-12-01",
			MaxTokenLifetimeDays:        30,
			RequireTokenExpiry:          true,
			ServiceTokenExpiryEnforced:  true,
			PersonalTokensDisabled:      &disabled,
		}, info)

		stored, err := json.Marshal(info)
		require.NoError(t, err)
		assert.NotContains(t, string(stored), settingsSecret)
	})

	t.Run("personal tokens setting stays unknown when gitlab omits it", func(t *testing.T) {
		info, err := newInstanceClientWith(t, http.StatusOK, http.StatusOK, map[string]any{}).InstanceInfo(t.Context())
		require.NoError(t, err)
		assert.True(t, info.SettingsKnown)
		assert.Nil(t, info.PersonalTokensDisabled)

		info, err = newInstanceClientWith(t, http.StatusOK, http.StatusOK, map[string]any{
			"disable_personal_access_tokens": false,
		}).InstanceInfo(t.Context())
		require.NoError(t, err)
		require.NotNil(t, info.PersonalTokensDisabled)
		assert.False(t, *info.PersonalTokensDisabled)
	})

	t.Run("settings are unknown without admin access", func(t *testing.T) {
		info, err := newInstanceClient(t, http.StatusOK, http.StatusForbidden).InstanceInfo(t.Context())
		require.NoError(t, err)
		assert.False(t, info.SettingsKnown)
		assert.True(t, info.TokenOwnerIsAdmin)
	})

	t.Run("other failures are errors", func(t *testing.T) {
		_, err := newInstanceClient(t, http.StatusOK, http.StatusBadRequest).InstanceInfo(t.Context())
		require.Error(t, err)

		_, err = newInstanceClient(t, http.StatusUnauthorized, http.StatusOK).InstanceInfo(t.Context())
		require.Error(t, err)
	})
}
