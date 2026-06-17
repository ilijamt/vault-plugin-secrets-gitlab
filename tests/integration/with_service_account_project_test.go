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

func TestWithServiceAccountProject(t *testing.T) {
	runServiceAccountTokenTest(t, serviceAccountCase{
		roleName:  "project-service-account",
		tokenType: token.TypeProjectServiceAccount,
		setupSA: func(t *testing.T, ctx context.Context, client glab.Client, gClient *g.Client) string {
			// Project bootstrapped by local-env/tf/_shared/service_accounts.tf
			projectId, err := client.GetProjectIdByPath(ctx, "service-accounts/project")
			require.NoError(t, err)
			pid := strconv.FormatInt(projectId, 10)

			sa, _, err := gClient.Projects.CreateProjectServiceAccount(pid, &g.CreateProjectServiceAccountOptions{})
			require.NoError(t, err)
			require.NotNil(t, sa)
			t.Cleanup(func() { _, _ = gClient.Projects.DeleteProjectServiceAccount(pid, sa.ID, nil) })
			return fmt.Sprintf("%s/%s", pid, sa.Username)
		},
	})
}
