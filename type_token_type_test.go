package gitlab_test

import (
	"testing"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	"github.com/stretchr/testify/assert"
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
			expected: gitlab.TokenTypeServiceAccount,
			input:    gitlab.TokenTypeServiceAccount.String(),
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
