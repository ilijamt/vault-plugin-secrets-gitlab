package token_test

import (
	"context"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/helper/locksutil"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	modelRole "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/role"
	mt "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/token"
	tk "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

type mockTokenBackend struct {
	role      *modelRole.Role
	roleErr   error
	client    gitlab.Client
	clientErr error
	sendEvent func(ctx context.Context, eventType event.EventType, metadata map[string]string) error
}

func (m *mockTokenBackend) Logger() hclog.Logger { return hclog.NewNullLogger() }
func (m *mockTokenBackend) RoleLockForKey(_ string) *locksutil.LockEntry {
	return locksutil.CreateLocks()[0]
}
func (m *mockTokenBackend) GetRole(_ context.Context, _ logical.Storage, _ string) (*modelRole.Role, error) {
	return m.role, m.roleErr
}
func (m *mockTokenBackend) GetClientByName(_ context.Context, _ logical.Storage, _ string) (gitlab.Client, error) {
	return m.client, m.clientErr
}
func (m *mockTokenBackend) SendEvent(ctx context.Context, eventType event.EventType, metadata map[string]string) error {
	if m.sendEvent != nil {
		return m.sendEvent(ctx, eventType, metadata)
	}
	return nil
}

type mockGitlabClient struct {
	gitlab.Client
	token     tk.Token
	lookupErr error
	createErr error
}

func (m *mockGitlabClient) GetUserIdByUsername(_ context.Context, _ string) (int64, error) {
	return 1, m.lookupErr
}
func (m *mockGitlabClient) GetProjectIdByPath(_ context.Context, _ string) (int64, error) {
	return 1, m.lookupErr
}
func (m *mockGitlabClient) GetGroupIdByPath(_ context.Context, _ string) (int64, error) {
	return 1, m.lookupErr
}

func (m *mockGitlabClient) CreateProjectAccessToken(_ context.Context, _ string, _ string, _ time.Time, _ []string, _ tk.AccessLevel) (*mt.TokenProject, error) {
	if m.createErr != nil || m.token == nil {
		return nil, m.createErr
	}
	return m.token.(*mt.TokenProject), nil
}
func (m *mockGitlabClient) CreateGroupAccessToken(_ context.Context, _ string, _ string, _ time.Time, _ []string, _ tk.AccessLevel) (*mt.TokenGroup, error) {
	if m.createErr != nil || m.token == nil {
		return nil, m.createErr
	}
	return m.token.(*mt.TokenGroup), nil
}
func (m *mockGitlabClient) CreatePersonalAccessToken(_ context.Context, _ string, _ int64, _ string, _ time.Time, _ []string) (*mt.TokenPersonal, error) {
	if m.createErr != nil || m.token == nil {
		return nil, m.createErr
	}
	return m.token.(*mt.TokenPersonal), nil
}
func (m *mockGitlabClient) CreateUserServiceAccountAccessToken(_ context.Context, _ string, _ int64, _ string, _ time.Time, _ []string) (*mt.TokenUserServiceAccount, error) {
	if m.createErr != nil || m.token == nil {
		return nil, m.createErr
	}
	return m.token.(*mt.TokenUserServiceAccount), nil
}
func (m *mockGitlabClient) CreateGroupServiceAccountAccessToken(_ context.Context, _ string, _ string, _ int64, _ string, _ time.Time, _ []string) (*mt.TokenGroupServiceAccount, error) {
	if m.createErr != nil || m.token == nil {
		return nil, m.createErr
	}
	return m.token.(*mt.TokenGroupServiceAccount), nil
}
func (m *mockGitlabClient) CreateProjectDeployToken(_ context.Context, _ string, _ int64, _ string, _ *time.Time, _ []string) (*mt.TokenProjectDeploy, error) {
	if m.createErr != nil || m.token == nil {
		return nil, m.createErr
	}
	return m.token.(*mt.TokenProjectDeploy), nil
}
func (m *mockGitlabClient) CreateGroupDeployToken(_ context.Context, _ string, _ int64, _ string, _ *time.Time, _ []string) (*mt.TokenGroupDeploy, error) {
	if m.createErr != nil || m.token == nil {
		return nil, m.createErr
	}
	return m.token.(*mt.TokenGroupDeploy), nil
}
func (m *mockGitlabClient) CreatePipelineProjectTriggerAccessToken(_ context.Context, _ string, _ string, _ int64, _ string, _ *time.Time) (*mt.TokenPipelineProjectTrigger, error) {
	if m.createErr != nil || m.token == nil {
		return nil, m.createErr
	}
	return m.token.(*mt.TokenPipelineProjectTrigger), nil
}

func newToken(tokenType tk.Type, now, expiresAt time.Time) tk.Token {
	base := mt.Token{
		TokenID: 1, Token: "glpat-test", Name: "t", Path: "p",
		TokenType: tokenType, CreatedAt: &now, ExpiresAt: &expiresAt,
	}
	scopes := mt.TokenWithScopes{Token: base, Scopes: []string{"api"}}
	scopesAL := mt.TokenWithScopesAndAccessLevel{Token: base, Scopes: []string{"api"}, AccessLevel: tk.AccessLevelDeveloperPermissions}

	switch tokenType {
	case tk.TypeProject:
		return &mt.TokenProject{TokenWithScopesAndAccessLevel: scopesAL}
	case tk.TypeGroup:
		return &mt.TokenGroup{TokenWithScopesAndAccessLevel: scopesAL}
	case tk.TypePersonal:
		return &mt.TokenPersonal{TokenWithScopes: scopes, UserID: 1}
	case tk.TypeUserServiceAccount:
		return &mt.TokenUserServiceAccount{TokenWithScopes: scopes}
	case tk.TypeGroupServiceAccount:
		return &mt.TokenGroupServiceAccount{TokenWithScopes: scopes, UserID: 1}
	case tk.TypeProjectDeploy:
		return &mt.TokenProjectDeploy{TokenWithScopes: scopes}
	case tk.TypeGroupDeploy:
		return &mt.TokenGroupDeploy{TokenWithScopes: scopes}
	case tk.TypePipelineProjectTrigger:
		return &mt.TokenPipelineProjectTrigger{Token: base}
	}
	return nil
}
