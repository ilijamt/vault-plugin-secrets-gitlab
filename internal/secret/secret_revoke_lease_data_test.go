package secret_test

import (
	"context"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/secret"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func leaseData(tokenType token.Type) map[string]any {
	return map[string]any{
		"token_id":             int64(42),
		"user_id":              int64(7),
		"token_type":           tokenType.String(),
		"gitlab_revokes_token": false,
		"parent_id":            "",
		"path":                 "path",
		"name":                 "name",
		"config_name":          "default",
	}
}

func revoke(t *testing.T, data map[string]any, client *stubClient, events *[]map[string]string) error {
	t.Helper()
	mb := &mockSecretBackend{
		getClientByName: func(context.Context, logical.Storage, string) (gitlab.Client, error) { return client, nil },
		sendEvent: func(_ context.Context, _ event.EventType, metadata map[string]string) error {
			*events = append(*events, metadata)
			return nil
		},
	}
	_, err := secret.NewSecret(mb).HandleRevoke(t.Context(), &logical.Request{
		Storage: &logical.InmemStorage{},
		Secret:  &logical.Secret{InternalData: data},
	})
	return err
}

func TestRevokeAccessToken_InvalidLeaseData(t *testing.T) {
	var tests = []struct {
		name      string
		tokenType token.Type
		change    func(map[string]any)
		key       string
	}{
		{name: "missing gitlab_revokes_token", tokenType: token.TypePersonal, change: func(d map[string]any) { delete(d, "gitlab_revokes_token") }, key: "gitlab_revokes_token"},
		{name: "wrong type gitlab_revokes_token", tokenType: token.TypePersonal, change: func(d map[string]any) { d["gitlab_revokes_token"] = "false" }, key: "gitlab_revokes_token"},
		{name: "missing parent_id", tokenType: token.TypePersonal, change: func(d map[string]any) { delete(d, "parent_id") }, key: "parent_id"},
		{name: "missing token_type", tokenType: token.TypePersonal, change: func(d map[string]any) { delete(d, "token_type") }, key: "token_type"},
		{name: "unknown token_type", tokenType: token.TypePersonal, change: func(d map[string]any) { d["token_type"] = "nope" }, key: "token_type"},
		{name: "missing config_name", tokenType: token.TypePersonal, change: func(d map[string]any) { delete(d, "config_name") }, key: "config_name"},
		{name: "wrong type config_name", tokenType: token.TypePersonal, change: func(d map[string]any) { d["config_name"] = 1 }, key: "config_name"},
		{name: "group service account without user_id", tokenType: token.TypeGroupServiceAccount, change: func(d map[string]any) { delete(d, "user_id") }, key: "user_id"},
		{name: "project service account without user_id", tokenType: token.TypeProjectServiceAccount, change: func(d map[string]any) { delete(d, "user_id") }, key: "user_id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := leaseData(tt.tokenType)
			tt.change(data)

			var events []map[string]string
			err := revoke(t, data, &stubClient{}, &events)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.key)
			assert.Empty(t, events)
		})
	}
}

func TestRevokeAccessToken_InvalidLeaseDataErrors(t *testing.T) {
	data := leaseData(token.TypePersonal)
	delete(data, "parent_id")

	var events []map[string]string
	require.ErrorIs(t, revoke(t, data, &stubClient{}, &events), errs.ErrInvalidValue)

	data = leaseData(token.TypePersonal)
	data["token_type"] = "nope"
	require.ErrorIs(t, revoke(t, data, &stubClient{}, &events), errs.ErrUnknownTokenType)
}

func TestRevokeAccessToken_MissingEventDataStillRevokes(t *testing.T) {
	var revoked bool
	client := &stubClient{revokePersonalAccessToken: func(context.Context, int64) error {
		revoked = true
		return nil
	}}
	data := leaseData(token.TypePersonal)
	delete(data, "path")
	delete(data, "name")

	var events []map[string]string
	require.NoError(t, revoke(t, data, client, &events))

	assert.True(t, revoked)
	require.Len(t, events, 1)
	assert.Empty(t, events[0]["path"])
}

func TestRevokeAccessToken_GitlabRevokesIgnoresTokenFields(t *testing.T) {
	data := leaseData(token.TypeUserServiceAccount)
	data["gitlab_revokes_token"] = true
	data["token_type"] = "nope"
	delete(data, "user_id")

	var events []map[string]string
	require.NoError(t, revoke(t, data, &stubClient{}, &events))
	assert.Len(t, events, 1)
}

func TestRevokeAccessToken_AlreadyRevokedInGitlab(t *testing.T) {
	notFound := errs.ErrAccessTokenNotFound
	client := &stubClient{
		revokePersonalAccessToken:               func(context.Context, int64) error { return notFound },
		revokeProjectAccessToken:                func(context.Context, int64, string) error { return notFound },
		revokeGroupAccessToken:                  func(context.Context, int64, string) error { return notFound },
		revokeUserServiceAccountAccessToken:     func(context.Context, int64) error { return notFound },
		revokeGroupServiceAccountAccessToken:    func(context.Context, int64, string, int64) error { return notFound },
		revokeProjectServiceAccountAccessToken:  func(context.Context, int64, string, int64) error { return notFound },
		revokePipelineProjectTriggerAccessToken: func(context.Context, int64, string) error { return notFound },
		revokeGroupDeployToken:                  func(context.Context, int64, string) error { return notFound },
		revokeProjectDeployToken:                func(context.Context, int64, string) error { return notFound },
	}

	for _, tokenType := range token.ValidTokenTypes {
		t.Run(tokenType, func(t *testing.T) {
			var events []map[string]string
			require.NoError(t, revoke(t, leaseData(token.Type(tokenType)), client, &events))
			assert.Len(t, events, 1)
		})
	}
}
