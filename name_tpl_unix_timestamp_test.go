//go:build unit

package gitlab_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	g "github.com/ilijamt/vault-plugin-secrets-gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestTokenNameGenerator_UnixTimeStamp(t *testing.T) {
	now := time.Now().UTC().Unix()
	val, err := g.TokenName(
		&g.EntryRole{
			RoleName:            "test",
			TTL:                 time.Hour,
			Path:                "/path",
			Name:                "{{ .unix_timestamp_utc }}",
			Scopes:              []string{g.TokenScopeApi.String()},
			AccessLevel:         g.AccessLevelNoPermissions,
			TokenType:           token.TypePersonal,
			GitlabRevokesTokens: false,
		},
	)
	require.NoError(t, err)
	require.NotEmpty(t, val)
	i, err := strconv.ParseInt(val, 10, 64)
	require.NoError(t, err)
	require.GreaterOrEqual(t, i, now)
}
