//go:build serviceaccount

package integration_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	g "gitlab.com/gitlab-org/api/client-go/v2"

	glab "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestWithServiceAccountGroup(t *testing.T) {
	runServiceAccountTokenTest(t, serviceAccountCase{
		roleName:  "group-service-account",
		tokenType: token.TypeGroupServiceAccount,
		setupSA: func(t *testing.T, ctx context.Context, client glab.Client, gClient *g.Client) string {
			// Group bootstrapped by local-env/tf/_shared/service_accounts.tf
			groupId, err := client.GetGroupIdByPath(ctx, "service-accounts")
			require.NoError(t, err)
			gid := strconv.FormatInt(groupId, 10)

			sa, _, err := gClient.Groups.CreateServiceAccount(gid, &g.CreateServiceAccountOptions{})
			require.NoError(t, err)
			require.NotNil(t, sa)
			t.Cleanup(func() { _, _ = gClient.Users.DeleteUser(sa.ID) })
			return fmt.Sprintf("%s/%s", gid, sa.UserName)
		},
	})
}
