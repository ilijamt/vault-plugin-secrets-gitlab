//go:build unit

package gitlab_test

import (
	"cmp"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
	gitlab2 "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestPathRolesWithDynamicPath(t *testing.T) {
	var defaultConfig = map[string]any{
		"token":    getGitlabToken("admin_user_root").Token,
		"base_url": cmp.Or(os.Getenv("GITLAB_URL"), "http://localhost:8080/"),
		"type":     gitlab2.TypeSelfManaged.String(),
	}

	ctx := getCtxGitlabClient(t, "unit")
	var bFlags = flags.Flags{}
	b, l, _, _ := getBackendWithFlagsWithEvents(ctx, bFlags)
	require.NoError(t, writeBackendConfigWithName(ctx, b, l, defaultConfig, gitlab.DefaultConfigName))

	tests := []struct {
		name        string
		operation   logical.Operation
		path        string
		tokenType   token.Type
		accessLevel token.AccessLevel
		dynamicPath bool
		valid       bool
	}{
		{
			"valid regexp",
			logical.CreateOperation,
			"test-.*123$",
			token.TypePersonal,
			token.AccessLevelUnknown,
			true,
			true,
		},
		{
			"valid regexp without dynamic path",
			logical.CreateOperation,
			"test-.*123$",
			token.TypePersonal,
			token.AccessLevelUnknown,
			false,
			false,
		},
		{
			"valid regexp all values accepted",
			logical.CreateOperation,
			".*",
			token.TypePersonal,
			token.AccessLevelUnknown,
			true,
			true,
		},
		{
			"invalid regexp",
			logical.CreateOperation,
			`[a-z`,
			token.TypePersonal,
			token.AccessLevelUnknown,
			true,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: tt.operation,
				Path:      fmt.Sprintf("%s/%d", gitlab.PathRoleStorage, time.Now().UnixNano()), Storage: l,
				Data: map[string]any{
					"path":         tt.path,
					"name":         tt.name,
					"access_level": tt.accessLevel.String(),
					"token_type":   tt.tokenType.String(),
					"dynamic_path": tt.dynamicPath,
					"scopes":       []string{},
					"ttl":          "1h",
				},
			})
			require.NotNil(t, resp)
			if tt.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
