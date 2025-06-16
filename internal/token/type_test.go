//go:build unit

package token_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestTokenType(t *testing.T) {
	var tests = []struct {
		expected gitlab.TokenType
		input    string
		err      bool
	}{
		{
			expected: gitlab.TokenTypePersonal,
			input:    gitlab.TokenTypePersonal.String(),
		},
		{
			expected: gitlab.TokenTypeGroup,
			input:    gitlab.TokenTypeGroup.String(),
		},
		{
			expected: gitlab.TokenTypeProject,
			input:    gitlab.TokenTypeProject.String(),
		},
		{
			expected: gitlab.TokenTypeUserServiceAccount,
			input:    gitlab.TokenTypeUserServiceAccount.String(),
		},
		{
			expected: gitlab.TokenTypeGroupServiceAccount,
			input:    gitlab.TokenTypeGroupServiceAccount.String(),
		},
		{
			expected: gitlab.TokenTypePipelineProjectTrigger,
			input:    gitlab.TokenTypePipelineProjectTrigger.String(),
		},
		{
			expected: gitlab.TokenTypeProjectDeploy,
			input:    gitlab.TokenTypeProjectDeploy.String(),
		},
		{
			expected: gitlab.TokenTypeGroupDeploy,
			input:    gitlab.TokenTypeGroupDeploy.String(),
		},
		{
			expected: gitlab.TokenTypeUnknown,
			input:    "unknown",
			err:      true,
		},
		{
			expected: gitlab.TokenTypeUnknown,
			input:    "unknown",
			err:      true,
		},
	}

	for _, test := range tests {
		t.Logf("assert parse(%s) = %s (err: %v)", test.input, test.expected, test.err)
		val, err := gitlab.TokenTypeParse(test.input)
		assert.EqualValues(t, test.expected, val)
		assert.EqualValues(t, test.expected.Value(), test.expected.String())
		if test.err {
			assert.ErrorIs(t, err, gitlab.ErrUnknownTokenType)
		} else {
			assert.NoError(t, err)
		}
	}
}
