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
			name:     "project-service-account",
			expected: token.TypeProjectServiceAccount,
			input:    token.TypeProjectServiceAccount.String(),
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
			if test.err {
				assert.ErrorIs(t, err, errs.ErrUnknownTokenType)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsTokenTypeAllowed(t *testing.T) {
	var tests = []struct {
		name     string
		input    token.Type
		version  string
		expected bool
	}{
		{name: "project service account on minimum", input: token.TypeProjectServiceAccount, version: "18.11.0", expected: true},
		{name: "project service account on newer", input: token.TypeProjectServiceAccount, version: "19.1.1-ee", expected: true},
		{name: "project service account on first ee release", input: token.TypeProjectServiceAccount, version: "18.11.0-ee", expected: true},
		{name: "project service account on older", input: token.TypeProjectServiceAccount, version: "18.10.9", expected: false},
		{name: "group service account on 17.11", input: token.TypeGroupServiceAccount, version: "17.11.7", expected: true},
		{name: "personal has no minimum", input: token.TypePersonal, version: "12.0.0", expected: true},
		{name: "empty version is lenient", input: token.TypeProjectServiceAccount, version: "", expected: true},
		{name: "unparseable version is lenient", input: token.TypeProjectServiceAccount, version: "version", expected: true},
		{name: "unknown type", input: token.TypeUnknown, version: "19.1.1", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, token.IsTokenTypeAllowed(tt.input, tt.version))
		})
	}
}

func TestMinVersionForType(t *testing.T) {
	minVersion, ok := token.MinVersionForType(token.TypeProjectServiceAccount)
	assert.True(t, ok)
	assert.Equal(t, "18.11", minVersion)

	_, ok = token.MinVersionForType(token.TypeUnknown)
	assert.False(t, ok)

	for _, tokenType := range token.ValidTokenTypes {
		_, ok = token.MinVersionForType(token.Type(tokenType))
		assert.True(t, ok, tokenType)
	}
}
