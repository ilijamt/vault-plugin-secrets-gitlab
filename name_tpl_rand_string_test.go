//go:build unit

package gitlab_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	g "github.com/ilijamt/vault-plugin-secrets-gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestTokenNameGenerator_RandString(t *testing.T) {
	val, err := g.TokenName(
		&g.EntryRole{
			RoleName:            "test",
			TTL:                 time.Hour,
			Path:                "/path",
			Name:                "{{ randHexString 8 }}",
			Scopes:              []string{g.TokenScopeApi.String()},
			AccessLevel:         g.AccessLevelNoPermissions,
			TokenType:           token.TokenTypePersonal,
			GitlabRevokesTokens: false,
		},
	)
	require.NoError(t, err)
	require.NotEmpty(t, val)
	require.Len(t, val, 16)
}
