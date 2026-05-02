//go:build paths || saas || selfhosted || e2e

package integration_test

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	g "gitlab.com/gitlab-org/api/client-go/v2"

	glab "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/token"
	t "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

var _ glab.Client = new(inMemoryClient)

func newInMemoryClient(tt *testing.T, valid bool) *inMemoryClient {
	tt.Helper()
	c := &inMemoryClient{
		users:           make([]string, 0),
		groups:          make([]string, 0),
		valid:           valid,
		accessTokens:    make(map[string]t.Token),
		injectedErrors:  make(map[string]bool),
		mainTokenInfo:   newSeededTokenConfig(),
		rotateMainToken: newSeededTokenConfig(),
	}
	tt.Cleanup(func() {
		assert.Empty(tt, c.LiveTokens(), "test left unrevoked tokens in inMemoryClient")
	})
	return c
}

func (i *inMemoryClient) ForgetToken(key string) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	delete(i.accessTokens, key)
}

func (i *inMemoryClient) ForgetAllTokens() {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	clear(i.accessTokens)
}

func newSeededTokenConfig() token.TokenConfig {
	now := time.Now()
	return token.TokenConfig{
		TokenWithScopes: token.TokenWithScopes{
			Token: token.Token{
				CreatedAt: g.Ptr(now),
				ExpiresAt: g.Ptr(now),
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

	injectedErrors map[string]bool

	calledMainToken       int64
	calledRotateMainToken int64
	calledValid           int64

	mainTokenInfo   token.TokenConfig
	rotateMainToken token.TokenConfig

	accessTokens map[string]t.Token

	valueGetProjectIdByPath int64
}

func (i *inMemoryClient) InjectError(method string) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	i.injectedErrors[method] = true
}

func (i *inMemoryClient) injectedErrLocked(method string) error {
	if i.injectedErrors[method] {
		return fmt.Errorf("%s", method)
	}
	return nil
}

func tokenKey(typ t.Type, parts ...any) string {
	segments := make([]string, 0, len(parts)+1)
	segments = append(segments, typ.String())
	for _, p := range parts {
		segments = append(segments, fmt.Sprintf("%v", p))
	}
	return strings.Join(segments, "_")
}

func (i *inMemoryClient) nextID() int64 {
	i.internalCounter++
	return i.internalCounter
}

func newTokenBase(id int64, parentID, path, name, prefix string, typ t.Type, expiresAt *time.Time) token.Token {
	return token.Token{
		TokenID:   id,
		ParentID:  parentID,
		Path:      path,
		Name:      name,
		Token:     fmt.Sprintf("%s-%s", prefix, uuid.New().String()),
		TokenType: typ,
		CreatedAt: g.Ptr(time.Now()),
		ExpiresAt: expiresAt,
	}
}

func indexOrAppend(s *[]string, val string) int {
	if idx := slices.Index(*s, val); idx >= 0 {
		return idx
	}
	*s = append(*s, val)
	return len(*s) - 1
}

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
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if err := i.injectedErrLocked("GetProjectIdByPath"); err != nil {
		return -1, err
	}
	return i.valueGetProjectIdByPath, nil
}

func (i *inMemoryClient) CreateProjectDeployToken(ctx context.Context, path string, projectId int64, name string, expiresAt *time.Time, scopes []string) (et *token.TokenProjectDeploy, err error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if err := i.injectedErrLocked("CreateProjectDeployToken"); err != nil {
		return nil, err
	}
	id := i.nextID()
	entryToken := &token.TokenProjectDeploy{
		TokenWithScopes: token.TokenWithScopes{
			Token:  newTokenBase(id, strconv.FormatInt(projectId, 10), path, name, "glpat", t.TypeProjectDeploy, expiresAt),
			Scopes: scopes,
		},
		Username: uuid.New().String(),
	}
	i.accessTokens[tokenKey(t.TypeProjectDeploy, projectId, id)] = entryToken
	return entryToken, nil
}

func (i *inMemoryClient) CreateGroupDeployToken(ctx context.Context, path string, groupId int64, name string, expiresAt *time.Time, scopes []string) (et *token.TokenGroupDeploy, err error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if err := i.injectedErrLocked("CreateGroupDeployToken"); err != nil {
		return nil, err
	}
	id := i.nextID()
	entryToken := &token.TokenGroupDeploy{
		TokenWithScopes: token.TokenWithScopes{
			Token:  newTokenBase(id, strconv.FormatInt(groupId, 10), path, name, "glpat", t.TypeGroupDeploy, expiresAt),
			Scopes: scopes,
		},
		Username: uuid.New().String(),
	}
	i.accessTokens[tokenKey(t.TypeGroupDeploy, groupId, id)] = entryToken
	return entryToken, nil
}

func (i *inMemoryClient) RevokeProjectDeployToken(ctx context.Context, projectId, deployTokenId int64) (err error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if err := i.injectedErrLocked("RevokeProjectDeployToken"); err != nil {
		return err
	}
	delete(i.accessTokens, tokenKey(t.TypeProjectDeploy, projectId, deployTokenId))
	return nil
}

func (i *inMemoryClient) RevokeGroupDeployToken(ctx context.Context, groupId, deployTokenId int64) (err error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if err := i.injectedErrLocked("RevokeGroupDeployToken"); err != nil {
		return err
	}
	delete(i.accessTokens, tokenKey(t.TypeGroupDeploy, groupId, deployTokenId))
	return nil
}

func (i *inMemoryClient) Metadata(ctx context.Context) (*g.Metadata, error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if err := i.injectedErrLocked("Metadata"); err != nil {
		return nil, err
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
	if err := i.injectedErrLocked("CreatePipelineProjectTriggerAccessToken"); err != nil {
		return nil, err
	}
	id := i.nextID()
	pid := strconv.FormatInt(projectId, 10)
	entryToken := &token.TokenPipelineProjectTrigger{
		Token: newTokenBase(id, pid, pid, name, "glptt", t.TypePipelineProjectTrigger, expiresAt),
	}
	i.accessTokens[tokenKey(t.TypePipelineProjectTrigger, projectId, id)] = entryToken
	return entryToken, nil
}

func (i *inMemoryClient) RevokePipelineProjectTriggerAccessToken(ctx context.Context, projectId int64, tokenId int64) error {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if err := i.injectedErrLocked("RevokePipelineProjectTriggerAccessToken"); err != nil {
		return err
	}
	delete(i.accessTokens, tokenKey(t.TypePipelineProjectTrigger, projectId, tokenId))
	return nil
}

func (i *inMemoryClient) GetGroupIdByPath(ctx context.Context, path string) (int64, error) {
	return int64(indexOrAppend(&i.groups, path)), nil
}

func (i *inMemoryClient) GitlabClient(ctx context.Context) *g.Client {
	return nil
}

func (i *inMemoryClient) CreateGroupServiceAccountAccessToken(ctx context.Context, path string, groupId string, userId int64, name string, expiresAt time.Time, scopes []string) (*token.TokenGroupServiceAccount, error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if err := i.injectedErrLocked("CreateGroupServiceAccountAccessToken"); err != nil {
		return nil, err
	}
	return nil, nil
}

func (i *inMemoryClient) CreateUserServiceAccountAccessToken(ctx context.Context, username string, userId int64, name string, expiresAt time.Time, scopes []string) (*token.TokenUserServiceAccount, error) {
	i.muLock.Lock()
	if err := i.injectedErrLocked("CreateUserServiceAccountAccessToken"); err != nil {
		i.muLock.Unlock()
		return nil, err
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

func (i *inMemoryClient) RevokeUserServiceAccountAccessToken(ctx context.Context, tok string) error {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if err := i.injectedErrLocked("RevokeUserServiceAccountAccessToken"); err != nil {
		return err
	}
	delete(i.accessTokens, tokenKey(t.TypeUserServiceAccount, tok))
	return nil
}

func (i *inMemoryClient) RevokeGroupServiceAccountAccessToken(ctx context.Context, tok string) error {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if err := i.injectedErrLocked("RevokeGroupServiceAccountAccessToken"); err != nil {
		return err
	}
	delete(i.accessTokens, tokenKey(t.TypeGroupServiceAccount, tok))
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
	if err := i.injectedErrLocked("CreatePersonalAccessToken"); err != nil {
		return nil, err
	}
	id := i.nextID()
	entryToken := &token.TokenPersonal{
		TokenWithScopes: token.TokenWithScopes{
			Token:  newTokenBase(id, "", username, name, "glpat", t.TypePersonal, &expiresAt),
			Scopes: scopes,
		},
		UserID: userId,
	}
	i.accessTokens[tokenKey(t.TypePersonal, id)] = entryToken
	return entryToken, nil
}

func (i *inMemoryClient) CreateGroupAccessToken(ctx context.Context, groupId string, name string, expiresAt time.Time, scopes []string, accessLevel t.AccessLevel) (*token.TokenGroup, error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if err := i.injectedErrLocked("CreateGroupAccessToken"); err != nil {
		return nil, err
	}
	id := i.nextID()
	entryToken := &token.TokenGroup{
		TokenWithScopesAndAccessLevel: token.TokenWithScopesAndAccessLevel{
			Token:       newTokenBase(id, groupId, groupId, name, "glgat", t.TypeGroup, &expiresAt),
			Scopes:      scopes,
			AccessLevel: accessLevel,
		},
	}
	i.accessTokens[tokenKey(t.TypeGroup, id)] = entryToken
	return entryToken, nil
}

func (i *inMemoryClient) CreateProjectAccessToken(ctx context.Context, projectId string, name string, expiresAt time.Time, scopes []string, accessLevel t.AccessLevel) (*token.TokenProject, error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if err := i.injectedErrLocked("CreateProjectAccessToken"); err != nil {
		return nil, err
	}
	id := i.nextID()
	entryToken := &token.TokenProject{
		TokenWithScopesAndAccessLevel: token.TokenWithScopesAndAccessLevel{
			Token:       newTokenBase(id, projectId, projectId, name, "glpat", t.TypeProject, &expiresAt),
			Scopes:      scopes,
			AccessLevel: accessLevel,
		},
	}
	i.accessTokens[tokenKey(t.TypeProject, id)] = entryToken
	return entryToken, nil
}

func (i *inMemoryClient) RevokePersonalAccessToken(ctx context.Context, tokenId int64) error {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if err := i.injectedErrLocked("RevokePersonalAccessToken"); err != nil {
		return err
	}
	delete(i.accessTokens, tokenKey(t.TypePersonal, tokenId))
	return nil
}

func (i *inMemoryClient) RevokeProjectAccessToken(ctx context.Context, tokenId int64, projectId string) error {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if err := i.injectedErrLocked("RevokeProjectAccessToken"); err != nil {
		return err
	}
	delete(i.accessTokens, tokenKey(t.TypeProject, tokenId))
	return nil
}

func (i *inMemoryClient) RevokeGroupAccessToken(ctx context.Context, tokenId int64, groupId string) error {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if err := i.injectedErrLocked("RevokeGroupAccessToken"); err != nil {
		return err
	}
	delete(i.accessTokens, tokenKey(t.TypeGroup, tokenId))
	return nil
}

func (i *inMemoryClient) GetUserIdByUsername(ctx context.Context, username string) (int64, error) {
	return int64(indexOrAppend(&i.users, username)), nil
}
