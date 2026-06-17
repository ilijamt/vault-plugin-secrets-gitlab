//go:build serviceaccount

package integration_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	g "gitlab.com/gitlab-org/api/client-go/v2"

	glab "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestWithServiceAccountUser(t *testing.T) {
	runServiceAccountTokenTest(t, serviceAccountCase{
		roleName:  "user-service-account",
		tokenType: token.TypeUserServiceAccount,
		setupSA: func(t *testing.T, ctx context.Context, client glab.Client, gClient *g.Client) string {
			usr, _, err := gClient.Users.CreateServiceAccountUser(&g.CreateServiceAccountUserOptions{})
			require.NoError(t, err)
			require.NotNil(t, usr)
			t.Cleanup(func() { _, _ = gClient.Users.DeleteUser(usr.ID) })
			return usr.Username
		},
	})
}
