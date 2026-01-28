package token_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestIsValidPath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		tokenType token.Type
		valid     bool
	}{
		// Test cases
		{"personal access token - dynamic path", "admin-user", token.TypePersonal, true},
		{"project access token - dynamic path", "example/example", token.TypeProject, true},
		{"group access token - dynamic path", "example", token.TypeGroup, true},

		// TypePersonal and TypeUserServiceAccount: single segment
		{"single valid - letters", "userone", token.TypePersonal, true},
		{"single valid - underscore", "user_one", token.TypeUserServiceAccount, true},
		{"single valid - hyphen+dot", "user.one-two", token.TypePersonal, true},
		{"single valid - digits", "user2024", token.TypePersonal, true},
		{"starts with invalid prefix '-'", "-user", token.TypePersonal, false},
		{"starts with invalid prefix '_'", "_user", token.TypeUserServiceAccount, false},
		{"starts with invalid prefix '.'", ".user", token.TypePersonal, false},
		{"ends with invalid suffix '-'", "user-", token.TypePersonal, false},
		{"ends with invalid suffix '_'", "user_", token.TypeUserServiceAccount, false},
		{"ends with invalid suffix '.'", "user.", token.TypePersonal, false},
		{"ends with invalid suffix '.git'", "user.git", token.TypeUserServiceAccount, false},
		{"ends with invalid suffix '.atom'", "user.atom", token.TypePersonal, false},
		{"ends with invalid suffix (mixed case)", "user.Atom", token.TypePersonal, true}, // Only lower ".atom" is invalid
		{"too many segments", "user/one", token.TypePersonal, false},
		{"empty path", "", token.TypePersonal, false},
		{"whitespace path", "   ", token.TypeUserServiceAccount, false},

		// TypeGroupServiceAccount: two segments
		{"group SA valid", "group1/account2", token.TypeGroupServiceAccount, true},
		{"group SA valid underscore", "group1/_account", token.TypeGroupServiceAccount, false},
		{"group SA valid, dot middle", "team.service/acct-2", token.TypeGroupServiceAccount, true},
		{"group SA too few segments", "group1", token.TypeGroupServiceAccount, false},
		{"group SA too many segments", "g/a/too/many", token.TypeGroupServiceAccount, false},
		{"group SA segment starts with invalid", "-group/acct", token.TypeGroupServiceAccount, false},
		{"group SA segment ends with invalid", "group/acct-", token.TypeGroupServiceAccount, false},

		// TypeProject, TypeGroup, TypeProjectDeploy, TypeGroupDeploy, TypePipelineProjectTrigger types
		{"one segment", "myproj", token.TypeProject, true},
		{"two segments", "group/proj", token.TypeGroup, true},
		{"segments invalid", "grp-1/pro.j2/_b", token.TypeProjectDeploy, false},
		{"segments valid", "grp1/pro.j2/b_c", token.TypeGroup, true},
		{"forbidden prefix", "-group/project", token.TypeGroupDeploy, false},
		{"forbidden suffix", "g1/proj.git", token.TypeProject, false},
		{"trailing slash", "g1/", token.TypeProject, false},
		{"leading slash", "/g1", token.TypeProjectDeploy, false},
		{"double slash (empty segment)", "g1//p2", token.TypeProjectDeploy, false},
		{"ends with forbidden segment edge", "g1/g2.", token.TypeProject, false},
		{"dots and hyphens", "g1.part-2/project_3.one", token.TypeGroup, true},

		// Empty segment from double slash
		{"double slash", "foo//bar", token.TypeProject, false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.valid, token.IsValidPath(tt.path, tt.tokenType))
	}
}
