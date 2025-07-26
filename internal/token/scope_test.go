//go:build unit

package token_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestTokenScope(t *testing.T) {
	var tests = []struct {
		expected token.Scope
		input    string
		err      bool
	}{
		{
			expected: token.ScopeApi,
			input:    token.ScopeApi.String(),
		},
		{
			expected: token.ScopeReadApi,
			input:    token.ScopeReadApi.String(),
		},
		{
			expected: token.ScopeReadRegistry,
			input:    token.ScopeReadRegistry.String(),
		},
		{
			expected: token.ScopeWriteRegistry,
			input:    token.ScopeWriteRegistry.String(),
		},
		{
			expected: token.ScopeReadRepository,
			input:    token.ScopeReadRepository.String(),
		},
		{
			expected: token.ScopeWriteRepository,
			input:    token.ScopeWriteRepository.String(),
		},
		{
			expected: token.ScopeCreateRunner,
			input:    token.ScopeCreateRunner.String(),
		},
		{
			expected: token.ScopeReadUser,
			input:    token.ScopeReadUser.String(),
		},
		{
			expected: token.ScopeSudo,
			input:    token.ScopeSudo.String(),
		},
		{
			expected: token.ScopeAdminMode,
			input:    token.ScopeAdminMode.String(),
		},
		{
			expected: token.ScopeReadPackageRegistry,
			input:    token.ScopeReadPackageRegistry.String(),
		},
		{
			expected: token.ScopeWritePackageRegistry,
			input:    token.ScopeWritePackageRegistry.String(),
		},
		{
			expected: token.ScopeUnknown,
			input:    "what",
			err:      true,
		},
		{
			expected: token.ScopeUnknown,
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
			assert.ErrorIs(t, err, errs.ErrUnknownTokenScope)
		} else {
			assert.NoError(t, err)
		}
	}
}
