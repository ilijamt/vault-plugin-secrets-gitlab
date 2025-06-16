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
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	gitlab2 "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestPathRolesPipelineProjectTrigger(t *testing.T) {
	var defaultConfig = map[string]any{
		"token":    getGitlabToken("admin_user_root").Token,
		"base_url": cmp.Or(os.Getenv("GITLAB_URL"), "http://localhost:8080/"),
		"type":     gitlab2.TypeSelfManaged.String(),
	}

	t.Run("should fail if have defined scopes or access level", func(t *testing.T) {
		ctx := getCtxGitlabClient(t, "unit")
		var b, l, err = getBackendWithConfig(ctx, defaultConfig)
		require.NoError(t, err)
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.CreateOperation,
			Path:      fmt.Sprintf("%s/%d", gitlab.PathRoleStorage, time.Now().UnixNano()), Storage: l,
			Data: map[string]any{
				"path":         "user",
				"name":         "Example user personal token",
				"access_level": gitlab.AccessLevelNoPermissions.String(),
				"token_type":   token.TypePipelineProjectTrigger.String(),
				"scopes":       []string{token.ScopeApi.String()},
				"ttl":          "1h",
			},
		})
		require.Error(t, err)
		require.NotNil(t, resp)
		var errorMap = countErrByName(err.(*multierror.Error))
		assert.EqualValues(t, 2, errorMap[errs.ErrFieldInvalidValue.Error()])
	})

	t.Run("ttl is set", func(t *testing.T) {
		ctx := getCtxGitlabClient(t, "unit")
		var b, l, err = getBackendWithConfig(ctx, defaultConfig)
		require.NoError(t, err)
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.CreateOperation,
			Path:      fmt.Sprintf("%s/%d", gitlab.PathRoleStorage, time.Now().UnixNano()), Storage: l,
			Data: map[string]any{
				"path":         "user",
				"name":         "Example user personal token",
				"access_level": gitlab.AccessLevelUnknown.String(),
				"token_type":   token.TypePipelineProjectTrigger.String(),
				"scopes":       []string{},
				"ttl":          "1h",
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.EqualValues(t, 3600, resp.Data["ttl"])
	})

	t.Run("ttl is optional", func(t *testing.T) {
		ctx := getCtxGitlabClient(t, "unit")
		var b, l, err = getBackendWithConfig(ctx, defaultConfig)
		require.NoError(t, err)
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.CreateOperation,
			Path:      fmt.Sprintf("%s/%d", gitlab.PathRoleStorage, time.Now().UnixNano()), Storage: l,
			Data: map[string]any{
				"path":         "user",
				"name":         "Example user personal token",
				"access_level": gitlab.AccessLevelUnknown.String(),
				"token_type":   token.TypePipelineProjectTrigger.String(),
				"scopes":       []string{},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.EqualValues(t, 0, resp.Data["ttl"])
	})
}
