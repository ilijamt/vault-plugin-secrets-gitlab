package gitlab_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode"

	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault/sdk/helper/logging"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
	g "github.com/xanzy/go-gitlab"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func countErrByName(err *multierror.Error) map[string]int {
	var data = make(map[string]int)

	for _, e := range err.Errors {
		name := errors.Unwrap(e).Error()
		if _, ok := data[name]; !ok {
			data[name] = 0
		}
		data[name]++
	}

	return data
}

type expectedEvent struct {
	eventType string
}

type mockEventsSender struct {
	eventsProcessed []*logical.EventReceived
}

func (m *mockEventsSender) SendEvent(ctx context.Context, eventType logical.EventType, event *logical.EventData) error {
	if m == nil {
		return nil
	}
	m.eventsProcessed = append(m.eventsProcessed, &logical.EventReceived{
		EventType: string(eventType),
		Event:     event,
	})
	return nil
}

func (m *mockEventsSender) expectEvents(t *testing.T, expectedEvents []expectedEvent) {
	t.Helper()
	require.EqualValuesf(t, len(m.eventsProcessed), len(expectedEvents), "Expected events: %v\nEvents processed: %v", expectedEvents, m.eventsProcessed)
	for i, expected := range expectedEvents {
		actual := m.eventsProcessed[i]
		require.EqualValuesf(t, expected.eventType, actual.EventType, "Mismatched event type at index %d. Expected %s, got %s\n%v", i, expected.eventType, actual.EventType, m.eventsProcessed)
	}
}

func getBackendWithEvents(ctx context.Context) (*gitlab.Backend, logical.Storage, *mockEventsSender, error) {
	events := &mockEventsSender{}
	config := &logical.BackendConfig{
		Logger:       logging.NewVaultLoggerWithWriter(io.Discard, log.NoLevel),
		System:       &logical.StaticSystemView{},
		StorageView:  &logical.InmemStorage{},
		BackendUUID:  "test",
		EventsSender: events,
	}

	b, err := gitlab.Factory(ctx, config)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to create Backend: %w", err)
	}

	return b.(*gitlab.Backend), config.StorageView, events, nil
}

func writeBackendConfig(ctx context.Context, b *gitlab.Backend, l logical.Storage, config map[string]any) error {
	var _, err = b.HandleRequest(ctx, &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      gitlab.PathConfigStorage, Storage: l,
		Data: config,
	})
	return err
}

func getBackendWithEventsAndConfig(ctx context.Context, config map[string]any) (*gitlab.Backend, logical.Storage, *mockEventsSender, error) {
	var b, storage, events, _ = getBackendWithEvents(ctx)
	return b, storage, events, writeBackendConfig(ctx, b, storage, config)
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
		accessTokens: make(map[string]gitlab.EntryToken),

		mainTokenInfo:   gitlab.EntryToken{CreatedAt: g.Ptr(time.Now()), ExpiresAt: g.Ptr(time.Now())},
		rotateMainToken: gitlab.EntryToken{CreatedAt: g.Ptr(time.Now()), ExpiresAt: g.Ptr(time.Now())},
	}
}

type inMemoryClient struct {
	internalCounter int
	users           []string
	muLock          sync.Mutex
	valid           bool

	personalAccessTokenRevokeError bool
	groupAccessTokenRevokeError    bool
	projectAccessTokenRevokeError  bool
	personalAccessTokenCreateError bool
	groupAccessTokenCreateError    bool
	projectAccessTokenCreateError  bool

	calledMainToken       int
	calledRotateMainToken int
	calledValid           int

	mainTokenInfo   gitlab.EntryToken
	rotateMainToken gitlab.EntryToken

	accessTokens map[string]gitlab.EntryToken
}

func (i *inMemoryClient) CurrentTokenInfo() (*gitlab.EntryToken, error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	i.calledMainToken++
	return &i.mainTokenInfo, nil
}

func (i *inMemoryClient) RotateCurrentToken() (*gitlab.EntryToken, *gitlab.EntryToken, error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	i.calledRotateMainToken++
	return &i.rotateMainToken, &i.mainTokenInfo, nil
}

func (i *inMemoryClient) Valid() bool {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	i.calledValid++
	return i.valid
}

func (i *inMemoryClient) CreatePersonalAccessToken(username string, userId int, name string, expiresAt time.Time, scopes []string) (*gitlab.EntryToken, error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.personalAccessTokenCreateError {
		return nil, fmt.Errorf("CreatePersonalAccessToken")
	}
	i.internalCounter++
	var tokenId = i.internalCounter
	var entryToken = gitlab.EntryToken{
		TokenID:   tokenId,
		UserID:    userId,
		ParentID:  "",
		Path:      username,
		Name:      name,
		Token:     "",
		TokenType: gitlab.TokenTypePersonal,
		CreatedAt: g.Ptr(time.Now()),
		ExpiresAt: &expiresAt,
		Scopes:    scopes,
	}
	i.accessTokens[fmt.Sprintf("%s_%v", gitlab.TokenTypePersonal.String(), tokenId)] = entryToken
	return &entryToken, nil
}

func (i *inMemoryClient) CreateGroupAccessToken(groupId string, name string, expiresAt time.Time, scopes []string, accessLevel gitlab.AccessLevel) (*gitlab.EntryToken, error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.groupAccessTokenCreateError {
		return nil, fmt.Errorf("CreateGroupAccessToken")
	}
	i.internalCounter++
	var tokenId = i.internalCounter
	var entryToken = gitlab.EntryToken{
		TokenID:     tokenId,
		UserID:      0,
		ParentID:    groupId,
		Path:        groupId,
		Name:        name,
		Token:       "",
		TokenType:   gitlab.TokenTypeGroup,
		CreatedAt:   g.Ptr(time.Now()),
		ExpiresAt:   &expiresAt,
		Scopes:      scopes,
		AccessLevel: accessLevel,
	}
	i.accessTokens[fmt.Sprintf("%s_%v", gitlab.TokenTypeGroup.String(), tokenId)] = entryToken
	return &entryToken, nil
}

func (i *inMemoryClient) CreateProjectAccessToken(projectId string, name string, expiresAt time.Time, scopes []string, accessLevel gitlab.AccessLevel) (*gitlab.EntryToken, error) {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.projectAccessTokenCreateError {
		return nil, fmt.Errorf("CreateProjectAccessToken")
	}
	i.internalCounter++
	var tokenId = i.internalCounter
	var entryToken = gitlab.EntryToken{
		TokenID:     tokenId,
		UserID:      0,
		ParentID:    projectId,
		Path:        projectId,
		Name:        name,
		Token:       "",
		TokenType:   gitlab.TokenTypeProject,
		CreatedAt:   g.Ptr(time.Now()),
		ExpiresAt:   &expiresAt,
		Scopes:      scopes,
		AccessLevel: accessLevel,
	}
	i.accessTokens[fmt.Sprintf("%s_%v", gitlab.TokenTypeProject.String(), tokenId)] = entryToken
	return &entryToken, nil
}

func (i *inMemoryClient) RevokePersonalAccessToken(tokenId int) error {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.personalAccessTokenRevokeError {
		return fmt.Errorf("RevokePersonalAccessToken")
	}
	delete(i.accessTokens, fmt.Sprintf("%s_%v", gitlab.TokenTypePersonal.String(), tokenId))
	return nil
}

func (i *inMemoryClient) RevokeProjectAccessToken(tokenId int, projectId string) error {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.projectAccessTokenRevokeError {
		return fmt.Errorf("RevokeProjectAccessToken")
	}
	delete(i.accessTokens, fmt.Sprintf("%s_%v", gitlab.TokenTypeProject.String(), tokenId))
	return nil
}

func (i *inMemoryClient) RevokeGroupAccessToken(tokenId int, groupId string) error {
	i.muLock.Lock()
	defer i.muLock.Unlock()
	if i.groupAccessTokenRevokeError {
		return fmt.Errorf("RevokeGroupAccessToken")
	}
	delete(i.accessTokens, fmt.Sprintf("%s_%v", gitlab.TokenTypeGroup.String(), tokenId))
	return nil
}

func (i *inMemoryClient) GetUserIdByUsername(username string) (int, error) {
	idx := slices.Index(i.users, username)
	if idx == -1 {
		i.users = append(i.users, username)
		idx = slices.Index(i.users, username)
	}
	return idx, nil
}

var _ gitlab.Client = new(inMemoryClient)

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

func getCtxGitlabClient(t *testing.T) context.Context {
	httpClient, _ := getClient(t)
	return gitlab.HttpClientNewContext(context.Background(), httpClient)
}

func getCtxGitlabClientWithUrl(t *testing.T) (context.Context, string) {
	httpClient, url := getClient(t)
	return gitlab.HttpClientNewContext(context.Background(), httpClient), url
}
