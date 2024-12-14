//go:build unit

package gitlab_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	g "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestTokenNameGenerator(t *testing.T) {
	var tests = []struct {
		in     *g.EntryRole
		outVal string
		outErr bool
	}{
		{nil, "", true},

		// invalid template
		{
			&g.EntryRole{
				RoleName:            "test",
				TTL:                 time.Hour,
				Path:                "/path",
				Name:                "{{ .role_name",
				Scopes:              []string{g.TokenScopeApi.String()},
				AccessLevel:         g.AccessLevelNoPermissions,
				TokenType:           g.TokenTypePersonal,
				GitlabRevokesTokens: true,
			},
			"",
			true,
		},

		// combination template
		{
			&g.EntryRole{
				RoleName:            "test",
				TTL:                 time.Hour,
				Path:                "/path",
				Name:                "{{ .role_name }}-{{ .token_type }}-access-token-{{ yesNoBool .gitlab_revokes_token }}",
				Scopes:              []string{g.TokenScopeApi.String()},
				AccessLevel:         g.AccessLevelNoPermissions,
				TokenType:           g.TokenTypePersonal,
				GitlabRevokesTokens: true,
			},
			"test-personal-access-token-yes",
			false,
		},

		// with stringsJoin
		{
			&g.EntryRole{
				RoleName:            "test",
				TTL:                 time.Hour,
				Path:                "/path",
				Name:                "{{ .role_name }}-{{ .token_type }}-{{ stringsJoin .scopes \"-\" }}-{{ yesNoBool .gitlab_revokes_token }}",
				Scopes:              []string{g.TokenScopeApi.String(), g.TokenScopeSudo.String()},
				AccessLevel:         g.AccessLevelNoPermissions,
				TokenType:           g.TokenTypePersonal,
				GitlabRevokesTokens: false,
			},
			"test-personal-api-sudo-no",
			false,
		},

		// with timeNowFormat
		{
			&g.EntryRole{
				RoleName:            "test",
				TTL:                 time.Hour,
				Path:                "/path",
				Name:                "{{ .role_name }}-{{ .token_type }}-{{ timeNowFormat \"2006-01\" }}",
				Scopes:              []string{g.TokenScopeApi.String(), g.TokenScopeSudo.String()},
				AccessLevel:         g.AccessLevelNoPermissions,
				TokenType:           g.TokenTypePersonal,
				GitlabRevokesTokens: false,
			},
			fmt.Sprintf("test-personal-%d-%02d", time.Now().UTC().Year(), time.Now().UTC().Month()),
			false,
		},
	}

	for _, tst := range tests {
		t.Logf("TokenName(%v)", tst.in)
		val, err := g.TokenName(tst.in)
		assert.Equal(t, tst.outVal, val)
		if tst.outErr {
			assert.Error(t, err, tst.outErr)
		} else {
			assert.NoError(t, err)
		}
	}
}
