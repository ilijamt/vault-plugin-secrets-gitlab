package config_test

import (
	"bytes"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	gitlabTypes "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab/types"
	pathConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/config"
)

func TestConfigTokenNotLoggedWhenShownInResponse(t *testing.T) {
	configPath := pathConfig.New(&mockConfigBackend{}).Paths()[0]
	writeRaw := map[string]interface{}{
		"token":    "glpat-write-token-value",
		"base_url": "https://gitlab.example.com",
		"type":     gitlabTypes.TypeSelfManaged.String(),
	}

	var tests = []struct {
		name      string
		operation logical.Operation
		raw       map[string]interface{}
		token     string
	}{
		{name: "read", operation: logical.ReadOperation, raw: map[string]interface{}{"config_name": "default"}, token: testConfig().Token},
		{name: "write", operation: logical.UpdateOperation, raw: writeRaw, token: writeRaw["token"].(string)},
		{name: "patch", operation: logical.PatchOperation, raw: map[string]interface{}{"config_name": "default", "type": gitlabTypes.TypeSaaS.String()}, token: testConfig().Token},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			mb := &mockConfigBackend{
				config:   testConfig(),
				flagsVal: flags.Flags{ShowConfigToken: true},
				logger:   hclog.New(&hclog.LoggerOptions{Output: &logs, Level: hclog.Debug}),
			}
			op := pathConfig.New(mb).Paths()[0].Operations[tt.operation].Handler()
			ctx := gitlab.ClientNewContext(t.Context(), &mockGitlabClient{tokenInfo: testTokenInfo(), metadata: testMetadata()})

			resp, err := op(ctx, &logical.Request{Storage: &logical.InmemStorage{}}, &framework.FieldData{Raw: tt.raw, Schema: configPath.Fields})
			require.NoError(t, err)
			require.False(t, resp.IsError())

			assert.Equal(t, tt.token, resp.Data["token"])
			assert.NotEmpty(t, logs.String())
			assert.NotContains(t, logs.String(), tt.token)
		})
	}
}
