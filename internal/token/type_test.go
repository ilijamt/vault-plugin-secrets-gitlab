//go:build unit

package token_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestTokenType(t *testing.T) {
	var tests = []struct {
		expected token.Type
		input    string
		err      bool
	}{
		{
			expected: token.TypePersonal,
			input:    token.TypePersonal.String(),
		},
		{
			expected: token.TypeGroup,
			input:    token.TypeGroup.String(),
		},
		{
			expected: token.TypeProject,
			input:    token.TypeProject.String(),
		},
		{
			expected: token.TypeUserServiceAccount,
			input:    token.TypeUserServiceAccount.String(),
		},
		{
			expected: token.TypeGroupServiceAccount,
			input:    token.TypeGroupServiceAccount.String(),
		},
		{
			expected: token.TypePipelineProjectTrigger,
			input:    token.TypePipelineProjectTrigger.String(),
		},
		{
			expected: token.TypeProjectDeploy,
			input:    token.TypeProjectDeploy.String(),
		},
		{
			expected: token.TypeGroupDeploy,
			input:    token.TypeGroupDeploy.String(),
		},
		{
			expected: token.TypeUnknown,
			input:    "unknown",
			err:      true,
		},
		{
			expected: token.TypeUnknown,
			input:    "unknown",
			err:      true,
		},
	}

	for _, test := range tests {
		t.Logf("assert parse(%s) = %s (err: %v)", test.input, test.expected, test.err)
		val, err := token.TypeParse(test.input)
		assert.EqualValues(t, test.expected, val)
		assert.EqualValues(t, test.expected.Value(), test.expected.String())
		if test.err {
			assert.ErrorIs(t, err, errs.ErrUnknownTokenType)
		} else {
			assert.NoError(t, err)
		}
	}
}
