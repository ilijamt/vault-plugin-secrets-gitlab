package token_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestTokenType(t *testing.T) {
	var tests = []struct {
		name     string
		expected token.Type
		input    string
		err      bool
	}{
		{
			name:     "personal",
			expected: token.TypePersonal,
			input:    token.TypePersonal.String(),
		},
		{
			name:     "group",
			expected: token.TypeGroup,
			input:    token.TypeGroup.String(),
		},
		{
			name:     "project",
			expected: token.TypeProject,
			input:    token.TypeProject.String(),
		},
		{
			name:     "user-service-account",
			expected: token.TypeUserServiceAccount,
			input:    token.TypeUserServiceAccount.String(),
		},
		{
			name:     "group-service-account",
			expected: token.TypeGroupServiceAccount,
			input:    token.TypeGroupServiceAccount.String(),
		},
		{
			name:     "pipeline-project-trigger",
			expected: token.TypePipelineProjectTrigger,
			input:    token.TypePipelineProjectTrigger.String(),
		},
		{
			name:     "project-deploy",
			expected: token.TypeProjectDeploy,
			input:    token.TypeProjectDeploy.String(),
		},
		{
			name:     "group-deploy",
			expected: token.TypeGroupDeploy,
			input:    token.TypeGroupDeploy.String(),
		},
		{
			name:     "unknown",
			expected: token.TypeUnknown,
			input:    "unknown",
			err:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			val, err := token.ParseType(test.input)
			assert.EqualValues(t, test.expected, val)
			assert.EqualValues(t, test.expected.Value(), test.expected.String())
			if test.err {
				assert.ErrorIs(t, err, errs.ErrUnknownTokenType)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
