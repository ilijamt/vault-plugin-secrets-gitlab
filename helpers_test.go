//go:build unit || saas || selfhosted || local

package gitlab_test

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode"

	"github.com/google/uuid"
	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/helper/logging"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
	g "gitlab.com/gitlab-org/api/client-go"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/token"
	t "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

var _ gitlab.Client = new(inMemoryClient)

var (
	gitlabComPersonalAccessToken = cmp.Or(os.Getenv("GITLAB_COM_TOKEN"), "glpat-invalid-value")
	gitlabComUrl                 = cmp.Or(os.Getenv("GITLAB_COM_URL"), "https://gitlab.com")
	gitlabServiceAccountUrl      = cmp.Or(os.Getenv("GITLAB_SERVICE_ACCOUNT_URL"), "http://localhost:8080")
	gitlabServiceAccountToken    = cmp.Or(os.Getenv("GITLAB_SERVICE_ACCOUNT_TOKEN"), "REPLACED-TOKEN")
)

type expectedEvent struct {
	eventType string
}

type mockEventsSender struct {
	eventsProcessed []*logical.EventReceived
	mu              sync.Mutex
}

func (m *mockEventsSender) resetEvents(t *testing.T) {
	t.Helper()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventsProcessed = make([]*logical.EventReceived, 0)
}

func (m *mockEventsSender) SendEvent(ctx context.Context, eventType logical.EventType, event *logical.EventData) error {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventsProcessed = append(m.eventsProcessed, &logical.EventReceived{
		EventType: string(eventType),
		Event:     event,
	})
	return nil
}

func (m *mockEventsSender) expectEvents(t *testing.T, expectedEvents []expectedEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t.Helper()
	require.EqualValuesf(t, len(m.eventsProcessed), len(expectedEvents), "Expected events: %v\nEvents processed: %v", expectedEvents, m.eventsProcessed)
	for i, expected := range expectedEvents {
		actual := m.eventsProcessed[i]
		require.EqualValuesf(t, expected.eventType, actual.EventType, "Mismatched event type at index %d. Expected %s, got %s\n%v", i, expected.eventType, actual.EventType, m.eventsProcessed)
	}
}

func getBackendWithEvents(ctx context.Context) (*gitlab.Backend, logical.Storage, *mockEventsSender, error) {
	return getBackendWithFlagsWithEvents(ctx, flags.Flags{})
}

func getBackendWithFlagsWithEvents(ctx context.Context, flags flags.Flags) (*gitlab.Backend, logical.Storage, *mockEventsSender, error) {
	events := &mockEventsSender{}
	config := &logical.BackendConfig{
		Logger:       logging.NewVaultLoggerWithWriter(io.Discard, log.NoLevel),
		System:       &logical.StaticSystemView{},
		StorageView:  &logical.InmemStorage{},
		BackendUUID:  "test",
		EventsSender: events,
	}

	b, err := gitlab.Factory(flags)(ctx, config)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to create Backend: %w", err)
	}

	return b.(*gitlab.Backend), config.StorageView, events, nil
}

func writeBackendConfigWithName(ctx context.Context, b *gitlab.Backend, l logical.Storage, config map[string]any, name string) error {
	var _, err = b.HandleRequest(ctx, &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, cmp.Or(name, gitlab.DefaultConfigName)), Storage: l,
		Data: config,
	})
	return err
}

func writeBackendConfig(ctx context.Context, b *gitlab.Backend, l logical.Storage, config map[string]any) error {
	return writeBackendConfigWithName(ctx, b, l, config, gitlab.DefaultConfigName)
}

func getBackendWithEventsAndConfig(ctx context.Context, config map[string]any) (*gitlab.Backend, logical.Storage, *mockEventsSender, error) {
	var b, storage, events, _ = getBackendWithEvents(ctx)
	return b, storage, events, writeBackendConfig(ctx, b, storage, config)
}

func getBackendWithEventsAndConfigName(ctx context.Context, config map[string]any, name string) (*gitlab.Backend, logical.Storage, *mockEventsSender, error) {
	var b, storage, events, _ = getBackendWithEvents(ctx)
	return b, storage, events, writeBackendConfigWithName(ctx, b, storage, config, name)
}

func getBackendWithConfig(ctx context.Context, config map[string]any) (*gitlab.Backend, logical.Storage, error) {
	var b, storage, _, _ = getBackendWithEvents(ctx)
	return b, storage, writeBackendConfig(ctx, b, storage, config)
}

func getBackend(ctx context.Context) (*gitlab.Backend, logical.Storage, error) {
	b, storage, _, err := getBackendWithEvents(ctx)
	return b, storage, err
}

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
	internalCounter int
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

	calledMainToken       int
	calledRotateMainToken int
	calledValid           int

	mainTokenInfo   token.TokenConfig
	rotateMainToken token.TokenConfig

	accessTokens map[string]t.Token

	valueGetProjectIdByPath int
}

func (i *inMemoryClient) GetProjectIdByPath(ctx context.Context, path string) (int, error) {
	if i.getProjectIdByPathError {
		return -1, fmt.Errorf("unable to get project id by path")
	}
	return i.valueGetProjectIdByPath, nil
}

func (i *inMemoryClient) CreateProjectDeployToken(ctx context.Context, path string, projectId int, name string, expiresAt *time.Time, scopes []string) (et *token.TokenProjectDeploy, err error) {
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
				ParentID:  strconv.Itoa(projectId),
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

func (i *inMemoryClient) CreateGroupDeployToken(ctx context.Context, path string, groupId int, name string, expiresAt *time.Time, scopes []string) (et *token.TokenGroupDeploy, err error) {
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
				ParentID:  strconv.Itoa(groupId),
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

func (i *inMemoryClient) RevokeProjectDeployToken(ctx context.Context, projectId, deployTokenId int) (err error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.revokeProjectDeployTokenError {
		return errors.New("revoke project deploy token error")
	}
	key := fmt.Sprintf("%s_%v_%v", t.TypeProjectDeploy.String(), projectId, deployTokenId)
	delete(i.accessTokens, key)
	return nil
}

func (i *inMemoryClient) RevokeGroupDeployToken(ctx context.Context, groupId, deployTokenId int) (err error) {
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

func (i *inMemoryClient) CreatePipelineProjectTriggerAccessToken(ctx context.Context, path, name string, projectId int, description string, expiresAt *time.Time) (et *token.TokenPipelineProjectTrigger, err error) {
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
			ParentID:  strconv.Itoa(projectId),
			Path:      strconv.Itoa(projectId),
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

func (i *inMemoryClient) RevokePipelineProjectTriggerAccessToken(ctx context.Context, projectId int, tokenId int) error {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.revokePipelineProjectTriggerAccessTokenError {
		return fmt.Errorf("RevokePipelineProjectTriggerAccessToken")
	}
	key := fmt.Sprintf("%s_%v_%v", t.TypePipelineProjectTrigger.String(), projectId, tokenId)
	delete(i.accessTokens, key)
	return nil
}

func (i *inMemoryClient) GetGroupIdByPath(ctx context.Context, path string) (int, error) {
	idx := slices.Index(i.groups, path)
	if idx == -1 {
		i.users = append(i.groups, path)
		idx = slices.Index(i.groups, path)
	}
	return idx, nil
}

func (i *inMemoryClient) GitlabClient(ctx context.Context) *g.Client {
	return nil
}

func (i *inMemoryClient) CreateGroupServiceAccountAccessToken(ctx context.Context, path string, groupId string, userId int, name string, expiresAt time.Time, scopes []string) (*token.TokenGroupServiceAccount, error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.createGroupServiceAccountAccessTokenError {
		return nil, fmt.Errorf("CreateGroupServiceAccountAccessToken")
	}
	return nil, nil
}

func (i *inMemoryClient) CreateUserServiceAccountAccessToken(ctx context.Context, username string, userId int, name string, expiresAt time.Time, scopes []string) (*token.TokenUserServiceAccount, error) {
	i.muLock.Lock()
	if i.createUserServiceAccountAccessTokenError {
		i.muLock.Unlock()
		return nil, fmt.Errorf("CreateUserServiceAccountAccessToken")
	}
	i.muLock.Unlock()
	var tok *token.TokenUserServiceAccount
	var err error
	var cpat *token.TokenPersonal
	if cpat, err = i.CreatePersonalAccessToken(ctx, username, userId, name, expiresAt, scopes); err != nil && cpat != nil {
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

func (i *inMemoryClient) CreatePersonalAccessToken(ctx context.Context, username string, userId int, name string, expiresAt time.Time, scopes []string) (*token.TokenPersonal, error) {
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

func (i *inMemoryClient) RevokePersonalAccessToken(ctx context.Context, tokenId int) error {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.personalAccessTokenRevokeError {
		return fmt.Errorf("RevokePersonalAccessToken")
	}
	delete(i.accessTokens, fmt.Sprintf("%s_%v", t.TypePersonal.String(), tokenId))
	return nil
}

func (i *inMemoryClient) RevokeProjectAccessToken(ctx context.Context, tokenId int, projectId string) error {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.projectAccessTokenRevokeError {
		return fmt.Errorf("RevokeProjectAccessToken")
	}
	delete(i.accessTokens, fmt.Sprintf("%s_%v", t.TypeProject.String(), tokenId))
	return nil
}

func (i *inMemoryClient) RevokeGroupAccessToken(ctx context.Context, tokenId int, groupId string) error {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.groupAccessTokenRevokeError {
		return fmt.Errorf("RevokeGroupAccessToken")
	}
	delete(i.accessTokens, fmt.Sprintf("%s_%v", t.TypeGroup.String(), tokenId))
	return nil
}

func (i *inMemoryClient) GetUserIdByUsername(ctx context.Context, username string) (int, error) {
	idx := slices.Index(i.users, username)
	if idx == -1 {
		i.users = append(i.users, username)
		idx = slices.Index(i.users, username)
	}
	return idx, nil
}

func sanitizePath(path string) string {
	var builder strings.Builder

	for _, r := range path {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			builder.WriteRune(r)
		} else {
			builder.WriteRune('_')
		}
	}

	return strings.ReplaceAll(builder.String(), "__", "_")
}

func getCtxGitlabClient(t *testing.T, target string) context.Context {
	httpClient, _ := getClient(t, target)
	return utils.HttpClientNewContext(t.Context(), httpClient)
}

func getCtxGitlabClientWithUrl(t *testing.T, target string) (context.Context, string) {
	httpClient, url := getClient(t, target)
	return utils.HttpClientNewContext(t.Context(), httpClient), url
}

func parseTimeFromFile(name string) (t time.Time, err error) {
	var buff []byte
	if buff, err = os.ReadFile(fmt.Sprintf("./testdata/%s", name)); err != nil {
		return t, err
	}
	return time.Parse(time.RFC3339, string(buff))
}

func ctxTestTime(ctx context.Context, testName string, tokenName string) (_ context.Context, t time.Time) {
	var token = getGitlabToken(tokenName)
	if token.Empty() {
		var err error
		switch testName {
		case "TestGitlabClient_InvalidToken":
			// no token for this test
		case "TestWithGitlabUser_RotateToken":
			if t, err = parseTimeFromFile("gitlab-com"); err != nil {
				panic(err)
			}
		case "TestWithServiceAccountUser",
			"TestWithServiceAccountGroup",
			"TestWithServiceAccountUserFail_dedicated",
			"TestWithServiceAccountUserFail_saas":
			if t, err = parseTimeFromFile("gitlab-selfhosted"); err != nil {
				panic(err)
			}
		default:
			panic(fmt.Errorf("unknown test name %s", testName))
		}
	} else {
		t = token.CreatedAtTime()
	}
	return utils.WithStaticTime(ctx, t), t
}

func filterSlice[T any, Slice ~[]T](collection Slice, predicate func(item T, index int) bool) Slice {
	result := make(Slice, 0, len(collection))

	for i := range collection {
		if predicate(collection[i], i) {
			result = append(result, collection[i])
		}
	}

	return result
}

type generatedToken struct {
	ID        string `json:"id"`
	Token     string `json:"token"`
	CreatedAt string `json:"created_at"`
}

func (g generatedToken) Empty() bool {
	return generatedToken{} == g
}

const (
	gitlabTimeLayout = "2006-01-02 15:04:05.000 -0700 MST"
)

func (g generatedToken) CreatedAtTime() (t time.Time) {
	t, _ = time.Parse(gitlabTimeLayout, g.CreatedAt)
	return t
}

type generatedTokens map[string]generatedToken

var loadTokens = sync.OnceValues(func() (t generatedTokens, err error) {
	var payload []byte
	if payload, err = os.ReadFile("./testdata/tokens.json"); err != nil {
		return t, err
	}

	err = json.Unmarshal(payload, &t)
	return t, err
})

func getGitlabToken(name string) generatedToken {
	var tokens, _ = loadTokens()
	if token, ok := tokens[name]; ok {
		return token
	}
	return generatedToken{}
}
