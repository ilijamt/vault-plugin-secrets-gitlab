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
			if test.err {
				assert.ErrorIs(t, err, errs.ErrUnknownTokenScope)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidScopesFor_Applicability(t *testing.T) {
	t.Run("pipeline trigger not applicable", func(t *testing.T) {
		scopes, applicable := token.ValidScopesFor(token.TypePipelineProjectTrigger, "18.0")
		assert.False(t, applicable)
		assert.Empty(t, scopes)
	})
	t.Run("personal applicable", func(t *testing.T) {
		_, applicable := token.ValidScopesFor(token.TypePersonal, "17.0")
		assert.True(t, applicable)
	})
}

func TestIsScopeAllowed_VersionGating(t *testing.T) {
	tests := []struct {
		name      string
		tokenType token.Type
		scope     token.Scope
		version   string
		want      bool
	}{
		// k8s_proxy GA at 16.4 → always allowed in supported window (17.0+)
		{"k8s_proxy on 17.0 personal", token.TypePersonal, token.ScopeK8SProxy, "17.0", true},
		// manage_runner since 17.1
		{"manage_runner gated below 17.1", token.TypePersonal, token.ScopeManageRunner, "17.0", false},
		{"manage_runner allowed at 17.1", token.TypePersonal, token.ScopeManageRunner, "17.1", true},
		// self_rotate since 17.9
		{"self_rotate gated on 17.0 group", token.TypeGroup, token.ScopeSelfRotate, "17.0", false},
		{"self_rotate allowed on 17.9 group", token.TypeGroup, token.ScopeSelfRotate, "17.9", true},
		// virtual_registry since 18.0 — present on group AT, not on project AT
		{"read_virtual_registry gated below 18.0 on group", token.TypeGroup, token.ScopeReadVirtualRegistry, "17.11", false},
		{"read_virtual_registry allowed at 18.0 on group", token.TypeGroup, token.ScopeReadVirtualRegistry, "18.0", true},
		{"read_virtual_registry not allowed on project at all", token.TypeProject, token.ScopeReadVirtualRegistry, "18.5", false},
		// read_service_ping is PAT-only
		{"read_service_ping personal at 17.1", token.TypePersonal, token.ScopeReadServicePing, "17.1", true},
		{"read_service_ping not on project AT", token.TypeProject, token.ScopeReadServicePing, "18.0", false},
		// deploy tokens have a different scope set
		{"read_package_registry on project deploy", token.TypeProjectDeploy, token.ScopeReadPackageRegistry, "17.0", true},
		{"api not on project deploy", token.TypeProjectDeploy, token.ScopeApi, "18.0", false},
		// pipeline trigger never has scopes
		{"pipeline trigger rejects api", token.TypePipelineProjectTrigger, token.ScopeApi, "18.0", false},
		// empty version = lenient
		{"self_rotate lenient on empty version", token.TypeGroup, token.ScopeSelfRotate, "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, token.IsScopeAllowed(tc.tokenType, tc.scope, tc.version))
		})
	}
}

func TestAllValidScopes(t *testing.T) {
	all := token.AllValidScopes()
	assert.Contains(t, all, token.ScopeApi.String())
	assert.Contains(t, all, token.ScopeSelfRotate.String())
	assert.Contains(t, all, token.ScopeReadVirtualRegistry.String())
	assert.NotContains(t, all, token.ScopeUnknown.String())
}
