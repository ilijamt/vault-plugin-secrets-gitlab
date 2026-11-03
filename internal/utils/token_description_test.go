package utils_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	modelRole "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/role"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

func TestTokenDescription(t *testing.T) {
	const fallback = "Issued by Vault: mount=gitlab/ role=ci config=default"

	var tests = []struct {
		name     string
		template string
		expected string
	}{
		{name: "not set", expected: fallback},
		{name: "renders empty", template: `{{ "" }}`, expected: fallback},
		{name: "renders spaces", template: `{{ "  " }}`, expected: fallback},
		{name: "template", template: `{{ .mount }}{{ .role_name }} {{ .token_type }}`, expected: "gitlab/ci personal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := modelRole.Role{RoleName: "ci", ConfigName: "default", TokenType: token.TypePersonal, Description: tt.template}
			description, err := utils.TokenDescription(role, "gitlab/")
			require.NoError(t, err)
			assert.Equal(t, tt.expected, description)
		})
	}
}

func TestTokenDescription_Truncated(t *testing.T) {
	role := modelRole.Role{Description: strings.Repeat("é", 300)}
	description, err := utils.TokenDescription(role, "gitlab/")
	require.NoError(t, err)
	assert.Equal(t, 255, utf8.RuneCountInString(description))
	assert.True(t, utf8.ValidString(description))
}

func TestTokenDescription_InvalidTemplate(t *testing.T) {
	role := modelRole.Role{Description: "{{"}
	require.Error(t, utils.ValidateTokenDescription(role))
	_, err := utils.TokenDescription(role, "gitlab/")
	require.Error(t, err)
}
