//go:build unit

package gitlab_test

import (
	"cmp"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestPathRolesDeployTokens(t *testing.T) {
	var defaultConfig = map[string]any{
		"token":    "glpat-secret-random-token",
		"base_url": cmp.Or(os.Getenv("GITLAB_URL"), "http://localhost:8080/"),
		"type":     gitlab.TypeSelfManaged.String(),
	}

	var tests = []struct {
		tokenType   gitlab.TokenType
		accessLevel gitlab.AccessLevel
		scopes      []string
		ttl         string
		path        string
		name        string
	}{
		{
			tokenType: gitlab.TokenTypeProjectDeploy,
			path:      "example/example",
			scopes:    []string{gitlab.TokenScopeApi.String()},
		},
		{
			tokenType: gitlab.TokenTypeGroupDeploy,
			path:      "test/test1",
			scopes:    []string{gitlab.TokenScopeApi.String()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.tokenType.String(), func(t *testing.T) {
			t.Run("should create role successfully", func(t *testing.T) {
				ctx := getCtxGitlabClient(t, "unit")
				var b, l, err = getBackendWithConfig(ctx, defaultConfig)
				require.NoError(t, err)
				resp, err := b.HandleRequest(ctx, &logical.Request{
					Operation: logical.CreateOperation,
					Path:      fmt.Sprintf("%s/%d", gitlab.PathRoleStorage, time.Now().UnixNano()), Storage: l,
					Data: map[string]any{
						"path":         tt.path,
						"name":         tt.name,
						"access_level": cmp.Or(tt.accessLevel, gitlab.AccessLevelNoPermissions).String(),
						"token_type":   tt.tokenType.String(),
						"scopes":       tt.scopes,
						"ttl":          cmp.Or(tt.ttl, "1h"),
					},
				})
				require.Error(t, err)
				require.NotNil(t, resp)
				var errorMap = countErrByName(err.(*multierror.Error))
				assert.EqualValues(t, 2, errorMap[gitlab.ErrFieldInvalidValue.Error()])
			})
		})
	}
}