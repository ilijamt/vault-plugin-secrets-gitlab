package token_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestTokenScope(t *testing.T) {
	var tests = []struct {
		name     string
		expected token.Scope
		input    string
		err      bool
	}{
		{
			name:     "api",
			expected: token.ScopeApi,
			input:    token.ScopeApi.String(),
		},
		{
			name:     "read_api",
			expected: token.ScopeReadApi,
			input:    token.ScopeReadApi.String(),
		},
		{
			name:     "read_registry",
			expected: token.ScopeReadRegistry,
			input:    token.ScopeReadRegistry.String(),
		},
		{
			name:     "write_registry",
			expected: token.ScopeWriteRegistry,
			input:    token.ScopeWriteRegistry.String(),
		},
		{
			name:     "read_repository",
			expected: token.ScopeReadRepository,
			input:    token.ScopeReadRepository.String(),
		},
		{
			name:     "write_repository",
			expected: token.ScopeWriteRepository,
			input:    token.ScopeWriteRepository.String(),
		},
		{
			name:     "read_package_registry",
			expected: token.ScopeReadPackageRegistry,
			input:    token.ScopeReadPackageRegistry.String(),
		},
		{
			name:     "write_package_registry",
			expected: token.ScopeWritePackageRegistry,
			input:    token.ScopeWritePackageRegistry.String(),
		},
		{
			name:     "create_runner",
			expected: token.ScopeCreateRunner,
			input:    token.ScopeCreateRunner.String(),
		},
		{
			name:     "manage_runner",
			expected: token.ScopeManageRunner,
			input:    token.ScopeManageRunner.String(),
		},
		{
			name:     "read_user",
			expected: token.ScopeReadUser,
			input:    token.ScopeReadUser.String(),
		},
		{
			name:     "sudo",
			expected: token.ScopeSudo,
			input:    token.ScopeSudo.String(),
		},
		{
			name:     "admin_mode",
			expected: token.ScopeAdminMode,
			input:    token.ScopeAdminMode.String(),
		},
		{
			name:     "ai_features",
			expected: token.ScopeAiFeatures,
			input:    token.ScopeAiFeatures.String(),
		},
		{
			name:     "k8s_proxy",
			expected: token.ScopeK8SProxy,
			input:    token.ScopeK8SProxy.String(),
		},
		{
			name:     "read_service_ping",
			expected: token.ScopeReadServicePing,
			input:    token.ScopeReadServicePing.String(),
		},
		{
			name:     "self_rotate",
			expected: token.ScopeSelfRotate,
			input:    token.ScopeSelfRotate.String(),
		},
		{
			name:     "read_virtual_registry",
			expected: token.ScopeReadVirtualRegistry,
			input:    token.ScopeReadVirtualRegistry.String(),
		},
		{
			name:     "write_virtual_registry",
			expected: token.ScopeWriteVirtualRegistry,
			input:    token.ScopeWriteVirtualRegistry.String(),
		},
		{
			name:     "unknown",
			expected: token.ScopeUnknown,
			input:    "unknown",
			err:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			val, err := token.ParseScope(test.input)
			assert.EqualValues(t, test.expected, val)
			assert.EqualValues(t, test.expected.Value(), test.expected.String())
			if test.err {
				assert.ErrorIs(t, err, errs.ErrUnknownTokenScope)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
