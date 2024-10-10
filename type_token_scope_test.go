//go:build !integration

package gitlab_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestTokenScope(t *testing.T) {
	var tests = []struct {
		expected gitlab.TokenScope
		input    string
		err      bool
	}{
		{
			expected: gitlab.TokenScopeApi,
			input:    gitlab.TokenScopeApi.String(),
		},
		{
			expected: gitlab.TokenScopeReadApi,
			input:    gitlab.TokenScopeReadApi.String(),
		},
		{
			expected: gitlab.TokenScopeReadRegistry,
			input:    gitlab.TokenScopeReadRegistry.String(),
		},
		{
			expected: gitlab.TokenScopeWriteRegistry,
			input:    gitlab.TokenScopeWriteRegistry.String(),
		},
		{
			expected: gitlab.TokenScopeReadRepository,
			input:    gitlab.TokenScopeReadRepository.String(),
		},
		{
			expected: gitlab.TokenScopeWriteRepository,
			input:    gitlab.TokenScopeWriteRepository.String(),
		},
		{
			expected: gitlab.TokenScopeCreateRunner,
			input:    gitlab.TokenScopeCreateRunner.String(),
		},
		{
			expected: gitlab.TokenScopeReadUser,
			input:    gitlab.TokenScopeReadUser.String(),
		},
		{
			expected: gitlab.TokenScopeSudo,
			input:    gitlab.TokenScopeSudo.String(),
		},
		{
			expected: gitlab.TokenScopeAdminMode,
			input:    gitlab.TokenScopeAdminMode.String(),
		},
		{
			expected: gitlab.TokenScopeUnknown,
			input:    "what",
			err:      true,
		},
		{
			expected: gitlab.TokenScopeUnknown,
			input:    "unknown",
			err:      true,
		},
	}

	for _, test := range tests {
		t.Logf("assert parse(%s) = %s (err: %v)", test.input, test.expected, test.err)
		val, err := gitlab.TokenScopeParse(test.input)
		assert.EqualValues(t, test.expected, val)
		assert.EqualValues(t, test.expected.Value(), test.expected.String())
		if test.err {
			assert.ErrorIs(t, err, gitlab.ErrUnknownTokenScope)
		} else {
			assert.NoError(t, err)
		}
	}
}
