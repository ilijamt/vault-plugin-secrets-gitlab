package gitlab_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	modelConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
)

func TestRevoke(t *testing.T) {
	var tests = []struct {
		name   string
		path   string
		revoke func(ctx context.Context, c gitlab.Client) error
	}{
		{"personal", "/api/v4/personal_access_tokens/1", func(ctx context.Context, c gitlab.Client) error {
			return c.RevokePersonalAccessToken(ctx, 1)
		}},
		{"project", "/api/v4/projects/2/access_tokens/1", func(ctx context.Context, c gitlab.Client) error {
			return c.RevokeProjectAccessToken(ctx, 1, "2")
		}},
		{"group", "/api/v4/groups/2/access_tokens/1", func(ctx context.Context, c gitlab.Client) error {
			return c.RevokeGroupAccessToken(ctx, 1, "2")
		}},
		{"user service account", "/api/v4/personal_access_tokens/1", func(ctx context.Context, c gitlab.Client) error {
			return c.RevokeUserServiceAccountAccessToken(ctx, 1)
		}},
		{"group service account", "/api/v4/groups/2/service_accounts/3/personal_access_tokens/1", func(ctx context.Context, c gitlab.Client) error {
			return c.RevokeGroupServiceAccountAccessToken(ctx, 1, "2", 3)
		}},
		{"project service account", "/api/v4/projects/2/service_accounts/3/personal_access_tokens/1", func(ctx context.Context, c gitlab.Client) error {
			return c.RevokeProjectServiceAccountAccessToken(ctx, 1, "2", 3)
		}},
		{"pipeline trigger", "/api/v4/projects/2/triggers/1", func(ctx context.Context, c gitlab.Client) error {
			return c.RevokePipelineProjectTriggerAccessToken(ctx, 1, "2")
		}},
		{"project deploy", "/api/v4/projects/2/deploy_tokens/1", func(ctx context.Context, c gitlab.Client) error {
			return c.RevokeProjectDeployToken(ctx, 1, "2")
		}},
		{"group deploy", "/api/v4/groups/2/deploy_tokens/1", func(ctx context.Context, c gitlab.Client) error {
			return c.RevokeGroupDeployToken(ctx, 1, "2")
		}},
	}

	newClient := func(t *testing.T, status int, path *string) gitlab.Client {
		t.Helper()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*path = r.Method + " " + r.URL.Path
			w.WriteHeader(status)
		}))
		t.Cleanup(srv.Close)
		client, err := gitlab.NewGitlabClient(&modelConfig.EntryConfig{BaseURL: srv.URL, Token: "config-token"}, srv.Client(), nil)
		require.NoError(t, err)
		return client
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			require.NoError(t, tt.revoke(t.Context(), newClient(t, http.StatusNoContent, &path)))
			assert.Equal(t, "DELETE "+tt.path, path)

			require.ErrorIs(t, tt.revoke(t.Context(), newClient(t, http.StatusNotFound, &path)), errs.ErrAccessTokenNotFound)

			err := tt.revoke(t.Context(), newClient(t, http.StatusForbidden, &path))
			require.Error(t, err)
			require.NotErrorIs(t, err, errs.ErrAccessTokenNotFound)
		})
	}
}
