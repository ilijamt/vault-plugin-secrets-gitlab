//go:build paths || saas || selfhosted || e2e

package integration_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	g "gitlab.com/gitlab-org/api/client-go/v2"

	glab "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/token"
	t "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

var _ glab.Client = new(inMemoryClient)

func newInMemoryClient(valid bool) *inMemoryClient {
	return &inMemoryClient{
		users:        make([]string, 0),
		valid:        valid,
		accessTokens: make(map[string]t.Token),

		mainTokenInfo: token.TokenConfig{
			TokenWithScopes: token.TokenWithScopes{
				Token: token.Token{
					CreatedAt: g.Ptr(time.Now()),
					ExpiresAt: g.Ptr(time.Now()),
				},
			},
		},
		rotateMainToken: token.TokenConfig{
			TokenWithScopes: token.TokenWithScopes{
				Token: token.Token{
					CreatedAt: g.Ptr(time.Now()),
					ExpiresAt: g.Ptr(time.Now()),
				},
			},
		},
	}
}

type inMemoryClient struct {
	internalCounter int64
	users           []string
	groups          []string
	muLock          sync.Mutex
	valid           bool

	personalAccessTokenRevokeError                    bool
	groupAccessTokenRevokeError                       bool
	projectAccessTokenRevokeError                     bool
	personalAccessTokenCreateError                    bool
	groupAccessTokenCreateError                       bool
	projectAccessTokenCreateError                     bool
	revokeUserServiceAccountPersonalAccessTokenError  bool
	revokeGroupServiceAccountPersonalAccessTokenError bool
	createUserServiceAccountAccessTokenError          bool
	createGroupServiceAccountAccessTokenError         bool
	createPipelineProjectTriggerAccessTokenError      bool
	revokePipelineProjectTriggerAccessTokenError      bool
	metadataError                                     bool
	revokeProjectDeployTokenError                     bool
	revokeGroupDeployTokenError                       bool
	createProjectDeployTokenError                     bool
	createGroupDeployTokenError                       bool
	getProjectIdByPathError                           bool

	calledMainToken       int64
	calledRotateMainToken int64
	calledValid           int64

	mainTokenInfo   token.TokenConfig
	rotateMainToken token.TokenConfig

	accessTokens map[string]t.Token

	valueGetProjectIdByPath int64
}

// LiveTokens returns a snapshot of access tokens currently held by the
// in-memory client (those created but not yet revoked). Pair with
// requireNoDanglingTokens to assert that a test cleaned up after itself.
func (i *inMemoryClient) LiveTokens() map[string]t.Token {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	out := make(map[string]t.Token, len(i.accessTokens))
	for k, v := range i.accessTokens {
		out[k] = v
	}
	return out
}

func (i *inMemoryClient) GetProjectIdByPath(ctx context.Context, path string) (int64, error) {
	if i.getProjectIdByPathError {
		return -1, fmt.Errorf("unable to get project id by path")
	}
	return i.valueGetProjectIdByPath, nil
}

func (i *inMemoryClient) CreateProjectDeployToken(ctx context.Context, path string, projectId int64, name string, expiresAt *time.Time, scopes []string) (et *token.TokenProjectDeploy, err error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.createProjectDeployTokenError {
		return nil, fmt.Errorf("unable to create project deploy token")
	}
	i.internalCounter++
	var tokenId = i.internalCounter
	key := fmt.Sprintf("%s_%v_%v", t.TypeProjectDeploy.String(), projectId, tokenId)
	var entryToken = &token.TokenProjectDeploy{
		TokenWithScopes: token.TokenWithScopes{
			Token: token.Token{
				TokenID:   tokenId,
				ParentID:  strconv.FormatInt(projectId, 10),
				Path:      path,
				Name:      name,
				Token:     fmt.Sprintf("glpat-%s", uuid.New().String()),
				TokenType: t.TypeProjectDeploy,
				ExpiresAt: expiresAt,
				CreatedAt: g.Ptr(time.Now())},
			Scopes: scopes,
		},
		Username: uuid.New().String(),
	}
	i.accessTokens[key] = entryToken
	return entryToken, nil
}

func (i *inMemoryClient) CreateGroupDeployToken(ctx context.Context, path string, groupId int64, name string, expiresAt *time.Time, scopes []string) (et *token.TokenGroupDeploy, err error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.createGroupDeployTokenError {
		return nil, fmt.Errorf("unable to create project deploy token")
	}
	i.internalCounter++
	var tokenId = i.internalCounter
	key := fmt.Sprintf("%s_%v_%v", t.TypeGroupDeploy.String(), groupId, tokenId)
	var entryToken = &token.TokenGroupDeploy{
		TokenWithScopes: token.TokenWithScopes{
			Token: token.Token{
				TokenID:   tokenId,
				ParentID:  strconv.FormatInt(groupId, 10),
				Path:      path,
				Name:      name,
				Token:     fmt.Sprintf("glpat-%s", uuid.New().String()),
				TokenType: t.TypeGroupDeploy,
				ExpiresAt: expiresAt,
				CreatedAt: g.Ptr(time.Now()),
			},
			Scopes: scopes,
		},
		Username: uuid.New().String(),
	}
	i.accessTokens[key] = entryToken
	return entryToken, nil
}

func (i *inMemoryClient) RevokeProjectDeployToken(ctx context.Context, projectId, deployTokenId int64) (err error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.revokeProjectDeployTokenError {
		return errors.New("revoke project deploy token error")
	}
	key := fmt.Sprintf("%s_%v_%v", t.TypeProjectDeploy.String(), projectId, deployTokenId)
	delete(i.accessTokens, key)
	return nil
}

func (i *inMemoryClient) RevokeGroupDeployToken(ctx context.Context, groupId, deployTokenId int64) (err error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.revokeGroupDeployTokenError {
		return errors.New("revoke group deploy token error")
	}
	key := fmt.Sprintf("%s_%v_%v", t.TypeGroupDeploy.String(), groupId, deployTokenId)
	delete(i.accessTokens, key)
	return nil
}

func (i *inMemoryClient) Metadata(ctx context.Context) (*g.Metadata, error) {
	if i.metadataError {
		return nil, errors.New("metadata error")
	}
	return &g.Metadata{
		Version:    "version",
		Revision:   "revision",
		Enterprise: false,
	}, nil
}

func (i *inMemoryClient) CreatePipelineProjectTriggerAccessToken(ctx context.Context, path, name string, projectId int64, description string, expiresAt *time.Time) (et *token.TokenPipelineProjectTrigger, err error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.createPipelineProjectTriggerAccessTokenError {
		return nil, fmt.Errorf("CreatePipelineProjectTriggerAccessToken")
	}
	i.internalCounter++
	var tokenId = i.internalCounter
	key := fmt.Sprintf("%s_%v_%v", t.TypePipelineProjectTrigger.String(), projectId, tokenId)
	var entryToken = &token.TokenPipelineProjectTrigger{
		Token: token.Token{
			TokenID:   tokenId,
			ParentID:  strconv.FormatInt(projectId, 10),
			Path:      strconv.FormatInt(projectId, 10),
			Name:      name,
			Token:     fmt.Sprintf("glptt-%s", uuid.New().String()),
			TokenType: t.TypePipelineProjectTrigger,
			ExpiresAt: expiresAt,
			CreatedAt: g.Ptr(time.Now()),
		},
	}
	i.accessTokens[key] = entryToken
	return entryToken, nil
}

func (i *inMemoryClient) RevokePipelineProjectTriggerAccessToken(ctx context.Context, projectId int64, tokenId int64) error {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.revokePipelineProjectTriggerAccessTokenError {
		return fmt.Errorf("RevokePipelineProjectTriggerAccessToken")
	}
	key := fmt.Sprintf("%s_%v_%v", t.TypePipelineProjectTrigger.String(), projectId, tokenId)
	delete(i.accessTokens, key)
	return nil
}

func (i *inMemoryClient) GetGroupIdByPath(ctx context.Context, path string) (int64, error) {
	idx := slices.Index(i.groups, path)
	if idx == -1 {
		i.groups = append(i.groups, path)
		idx = slices.Index(i.groups, path)
	}
	return int64(idx), nil
}

func (i *inMemoryClient) GitlabClient(ctx context.Context) *g.Client {
	return nil
}

func (i *inMemoryClient) CreateGroupServiceAccountAccessToken(ctx context.Context, path string, groupId string, userId int64, name string, expiresAt time.Time, scopes []string) (*token.TokenGroupServiceAccount, error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.createGroupServiceAccountAccessTokenError {
		return nil, fmt.Errorf("CreateGroupServiceAccountAccessToken")
	}
	return nil, nil
}

func (i *inMemoryClient) CreateUserServiceAccountAccessToken(ctx context.Context, username string, userId int64, name string, expiresAt time.Time, scopes []string) (*token.TokenUserServiceAccount, error) {
	i.muLock.Lock()
	if i.createUserServiceAccountAccessTokenError {
		i.muLock.Unlock()
		return nil, fmt.Errorf("CreateUserServiceAccountAccessToken")
	}
	i.muLock.Unlock()
	var tok *token.TokenUserServiceAccount
	var err error
	var cpat *token.TokenPersonal
	if cpat, err = i.CreatePersonalAccessToken(ctx, username, userId, name, expiresAt, scopes); err == nil && cpat != nil {
		tok = &token.TokenUserServiceAccount{
			TokenWithScopes: token.TokenWithScopes{
				Token: token.Token{
					CreatedAt: cpat.CreatedAt,
					ExpiresAt: cpat.ExpiresAt,
					TokenType: t.TypeUserServiceAccount,
					Token:     cpat.Token.Token,
					TokenID:   cpat.TokenID,
					ParentID:  cpat.ParentID,
					Name:      cpat.Name,
					Path:      cpat.Path,
				},
				Scopes: cpat.Scopes,
			},
		}

	}
	return tok, err
}

func (i *inMemoryClient) RevokeUserServiceAccountAccessToken(ctx context.Context, token string) error {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.revokeUserServiceAccountPersonalAccessTokenError {
		return errors.New("RevokeServiceAccountPersonalAccessToken")
	}
	delete(i.accessTokens, fmt.Sprintf("%s_%v", t.TypeUserServiceAccount.String(), token))
	return nil
}

func (i *inMemoryClient) RevokeGroupServiceAccountAccessToken(ctx context.Context, token string) error {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.revokeGroupServiceAccountPersonalAccessTokenError {
		return errors.New("RevokeServiceAccountPersonalAccessToken")
	}
	delete(i.accessTokens, fmt.Sprintf("%s_%v", t.TypeGroupServiceAccount.String(), token))
	return nil
}

func (i *inMemoryClient) CurrentTokenInfo(ctx context.Context) (*token.TokenConfig, error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	i.calledMainToken++
	return &i.mainTokenInfo, nil
}

func (i *inMemoryClient) RotateCurrentToken(ctx context.Context) (*token.TokenConfig, *token.TokenConfig, error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	i.calledRotateMainToken++
	return &i.rotateMainToken, &i.mainTokenInfo, nil
}

func (i *inMemoryClient) Valid(ctx context.Context) bool {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	i.calledValid++
	return i.valid
}

func (i *inMemoryClient) CreatePersonalAccessToken(ctx context.Context, username string, userId int64, name string, expiresAt time.Time, scopes []string) (*token.TokenPersonal, error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.personalAccessTokenCreateError {
		return nil, fmt.Errorf("CreatePersonalAccessToken")
	}
	i.internalCounter++
	var tokenId = i.internalCounter
	var entryToken = &token.TokenPersonal{
		TokenWithScopes: token.TokenWithScopes{
			Token: token.Token{
				TokenID:   tokenId,
				ParentID:  "",
				Path:      username,
				Name:      name,
				Token:     fmt.Sprintf("glpat-%s", uuid.New().String()),
				TokenType: t.TypePersonal,
				CreatedAt: g.Ptr(time.Now()),
				ExpiresAt: &expiresAt,
			},
			Scopes: scopes,
		},
		UserID: userId,
	}
	i.accessTokens[fmt.Sprintf("%s_%v", t.TypePersonal.String(), tokenId)] = entryToken
	return entryToken, nil
}

func (i *inMemoryClient) CreateGroupAccessToken(ctx context.Context, groupId string, name string, expiresAt time.Time, scopes []string, accessLevel t.AccessLevel) (*token.TokenGroup, error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.groupAccessTokenCreateError {
		return nil, fmt.Errorf("CreateGroupAccessToken")
	}
	i.internalCounter++
	var tokenId = i.internalCounter
	var entryToken = &token.TokenGroup{
		TokenWithScopesAndAccessLevel: token.TokenWithScopesAndAccessLevel{
			Token: token.Token{
				TokenID:   tokenId,
				ParentID:  groupId,
				Path:      groupId,
				Name:      name,
				Token:     fmt.Sprintf("glgat-%s", uuid.New().String()),
				TokenType: t.TypeGroup,
				CreatedAt: g.Ptr(time.Now()),
				ExpiresAt: &expiresAt,
			},
			Scopes:      scopes,
			AccessLevel: accessLevel,
		},
	}
	i.accessTokens[fmt.Sprintf("%s_%v", t.TypeGroup.String(), tokenId)] = entryToken
	return entryToken, nil
}

func (i *inMemoryClient) CreateProjectAccessToken(ctx context.Context, projectId string, name string, expiresAt time.Time, scopes []string, accessLevel t.AccessLevel) (*token.TokenProject, error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.projectAccessTokenCreateError {
		return nil, fmt.Errorf("CreateProjectAccessToken")
	}
	i.internalCounter++
	var tokenId = i.internalCounter
	var entryToken = &token.TokenProject{
		TokenWithScopesAndAccessLevel: token.TokenWithScopesAndAccessLevel{
			Token: token.Token{
				Token:     fmt.Sprintf("glpat-%s", uuid.New().String()),
				TokenType: t.TypeProject,
				CreatedAt: g.Ptr(time.Now()),
				ExpiresAt: &expiresAt,
				TokenID:   tokenId,
				ParentID:  projectId,
				Name:      name,
				Path:      projectId,
			},
			Scopes:      scopes,
			AccessLevel: accessLevel,
		},
	}
	i.accessTokens[fmt.Sprintf("%s_%v", t.TypeProject.String(), tokenId)] = entryToken
	return entryToken, nil
}

func (i *inMemoryClient) RevokePersonalAccessToken(ctx context.Context, tokenId int64) error {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.personalAccessTokenRevokeError {
		return fmt.Errorf("RevokePersonalAccessToken")
	}
	delete(i.accessTokens, fmt.Sprintf("%s_%v", t.TypePersonal.String(), tokenId))
	return nil
}

func (i *inMemoryClient) RevokeProjectAccessToken(ctx context.Context, tokenId int64, projectId string) error {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.projectAccessTokenRevokeError {
		return fmt.Errorf("RevokeProjectAccessToken")
	}
	delete(i.accessTokens, fmt.Sprintf("%s_%v", t.TypeProject.String(), tokenId))
	return nil
}

func (i *inMemoryClient) RevokeGroupAccessToken(ctx context.Context, tokenId int64, groupId string) error {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.groupAccessTokenRevokeError {
		return fmt.Errorf("RevokeGroupAccessToken")
	}
	delete(i.accessTokens, fmt.Sprintf("%s_%v", t.TypeGroup.String(), tokenId))
	return nil
}

func (i *inMemoryClient) GetUserIdByUsername(ctx context.Context, username string) (int64, error) {
	idx := slices.Index(i.users, username)
	if idx == -1 {
		i.users = append(i.users, username)
		idx = slices.Index(i.users, username)
	}
	return int64(idx), nil
}

// requireNoDanglingTokens fails the test if the in-memory client still holds
// any unrevoked access tokens. Register via t.Cleanup in tests whose contract
// is that every token created should also be revoked.
func requireNoDanglingTokens(t *testing.T, c *inMemoryClient) {
	t.Helper()
	if live := c.LiveTokens(); len(live) > 0 {
		t.Errorf("test left %d unrevoked tokens in inMemoryClient: %v", len(live), live)
	}
}
