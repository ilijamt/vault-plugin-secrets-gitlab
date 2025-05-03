//go:build unit

package token_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestScope(t *testing.T) {
	var tests = []struct {
		expected token.Scope
		input    string
		err      bool
	}{
		{
			expected: token.TokenScopeApi,
			input:    token.TokenScopeApi.String(),
		},
		{
			expected: token.TokenScopeReadApi,
			input:    token.TokenScopeReadApi.String(),
		},
		{
			expected: token.TokenScopeReadRegistry,
			input:    token.TokenScopeReadRegistry.String(),
		},
		{
			expected: token.TokenScopeWriteRegistry,
			input:    token.TokenScopeWriteRegistry.String(),
		},
		{
			expected: token.TokenScopeReadRepository,
			input:    token.TokenScopeReadRepository.String(),
		},
		{
			expected: token.TokenScopeWriteRepository,
			input:    token.TokenScopeWriteRepository.String(),
		},
		{
			expected: token.TokenScopeCreateRunner,
			input:    token.TokenScopeCreateRunner.String(),
		},
		{
			expected: token.TokenScopeReadUser,
			input:    token.TokenScopeReadUser.String(),
		},
		{
			expected: token.TokenScopeSudo,
			input:    token.TokenScopeSudo.String(),
		},
		{
			expected: token.TokenScopeAdminMode,
			input:    token.TokenScopeAdminMode.String(),
		},
		{
			expected: token.TokenScopeReadPackageRegistry,
			input:    token.TokenScopeReadPackageRegistry.String(),
		},
		{
			expected: token.TokenScopeWritePackageRegistry,
			input:    token.TokenScopeWritePackageRegistry.String(),
		},
		{
			expected: token.TokenScopeUnknown,
			input:    "what",
			err:      true,
		},
		{
			expected: token.TokenScopeUnknown,
			input:    "unknown",
			err:      true,
		},
	}

	for _, test := range tests {
		t.Logf("assert parse(%s) = %s (err: %v)", test.input, test.expected, test.err)
		val, err := token.ParseScope(test.input)
		assert.EqualValues(t, test.expected, val)
		assert.EqualValues(t, test.expected.Value(), test.expected.String())
		if test.err {
			assert.ErrorIs(t, err, token.ErrUnknownTokenScope)
		} else {
			assert.NoError(t, err)
		}
	}
}
