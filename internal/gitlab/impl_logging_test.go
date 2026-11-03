package gitlab_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/logging"
	modelConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

const canaryToken = "glpat-canary-must-not-be-logged"

func newLoggingClient(t *testing.T, status int) (gitlab.Client, *bytes.Buffer) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case status != http.StatusOK:
			w.WriteHeader(status)
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": 1, "name": "name", "token": canaryToken, "user_id": 1, "username": "user",
				"scopes": []string{"api"}, "access_level": 40,
				"created_at": "2026-01-01T00:00:00Z", "expires_at": "2027-01-01",
			})
		}
	}))
	t.Cleanup(srv.Close)

	var out bytes.Buffer
	logger := hclog.New(&hclog.LoggerOptions{Output: &out, Level: hclog.Debug})
	client, err := gitlab.NewGitlabClient(&modelConfig.EntryConfig{BaseURL: srv.URL, Token: "config-token"}, srv.Client(), logger)
	require.NoError(t, err)
	return client, &out
}

func TestClientNeverLogsTokenValues(t *testing.T) {
	var expires = time.Now().Add(24 * time.Hour)
	var scopes = []string{"api"}

	var calls = map[string]func(ctx context.Context, c gitlab.Client) error{
		"personal": func(ctx context.Context, c gitlab.Client) error {
			_, err := c.CreatePersonalAccessToken(ctx, "user", 1, "name", "description", expires, scopes)
			return err
		},
		"group": func(ctx context.Context, c gitlab.Client) error {
			_, err := c.CreateGroupAccessToken(ctx, "1", "name", "description", expires, scopes, token.AccessLevelOwnerPermissions)
			return err
		},
		"project": func(ctx context.Context, c gitlab.Client) error {
			_, err := c.CreateProjectAccessToken(ctx, "1", "name", "description", expires, scopes, token.AccessLevelOwnerPermissions)
			return err
		},
		"user service account": func(ctx context.Context, c gitlab.Client) error {
			_, err := c.CreateUserServiceAccountAccessToken(ctx, "user", 1, "name", "description", expires, scopes)
			return err
		},
		"group service account": func(ctx context.Context, c gitlab.Client) error {
			_, err := c.CreateGroupServiceAccountAccessToken(ctx, "1/sa", "1", 1, "name", "description", expires, scopes)
			return err
		},
		"project service account": func(ctx context.Context, c gitlab.Client) error {
			_, err := c.CreateProjectServiceAccountAccessToken(ctx, "1/sa", "1", 1, "name", "description", expires, scopes)
			return err
		},
		"current token info": func(ctx context.Context, c gitlab.Client) error {
			_, err := c.CurrentTokenInfo(ctx)
			return err
		},
		"rotate current token": func(ctx context.Context, c gitlab.Client) error {
			_, _, err := c.RotateCurrentToken(ctx)
			return err
		},
	}

	for name, call := range calls {
		t.Run(name+" succeeds", func(t *testing.T) {
			client, logs := newLoggingClient(t, http.StatusOK)
			require.NoError(t, call(t.Context(), client))
			assert.NotContains(t, logs.String(), canaryToken)
			assert.Contains(t, logs.String(), "token_sha1_hash="+logging.Hash(canaryToken))
		})

		t.Run(name+" fails", func(t *testing.T) {
			client, logs := newLoggingClient(t, http.StatusForbidden)
			require.Error(t, call(t.Context(), client))
			assert.NotContains(t, logs.String(), canaryToken)
		})
	}
}
