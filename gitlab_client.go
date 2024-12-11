package gitlab

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/helper/logging"
	g "gitlab.com/gitlab-org/api/client-go"
	"golang.org/x/time/rate"
)

var (
	ErrAccessTokenNotFound = errors.New("access token not found")
	ErrRoleNotFound        = errors.New("role not found")
)

type Client interface {
	GitlabClient(ctx context.Context) *g.Client
	Valid(ctx context.Context) bool
	CurrentTokenInfo(ctx context.Context) (*EntryToken, error)
	RotateCurrentToken(ctx context.Context) (newToken *EntryToken, oldToken *EntryToken, err error)
	CreatePersonalAccessToken(ctx context.Context, username string, userId int, name string, expiresAt time.Time, scopes []string) (*EntryToken, error)
	CreateGroupAccessToken(ctx context.Context, groupId string, name string, expiresAt time.Time, scopes []string, accessLevel AccessLevel) (*EntryToken, error)
	CreateProjectAccessToken(ctx context.Context, projectId string, name string, expiresAt time.Time, scopes []string, accessLevel AccessLevel) (*EntryToken, error)
	RevokePersonalAccessToken(ctx context.Context, tokenId int) error
	RevokeProjectAccessToken(ctx context.Context, tokenId int, projectId string) error
	RevokeGroupAccessToken(ctx context.Context, tokenId int, groupId string) error
	GetUserIdByUsername(ctx context.Context, username string) (int, error)
	GetGroupIdByPath(ctx context.Context, path string) (int, error)
	CreateGroupServiceAccountAccessToken(ctx context.Context, group string, groupId string, userId int, name string, expiresAt time.Time, scopes []string) (*EntryToken, error)
	CreateUserServiceAccountAccessToken(ctx context.Context, username string, userId int, name string, expiresAt time.Time, scopes []string) (*EntryToken, error)
	RevokeUserServiceAccountAccessToken(ctx context.Context, token string) error
	RevokeGroupServiceAccountAccessToken(ctx context.Context, token string) error
	CreatePipelineProjectTriggerAccessToken(ctx context.Context, projectId int, description string) error
	RevokePipelineProjectTriggerAccessToken(ctx context.Context, projectId int, tokenId int) error
}

type gitlabClient struct {
	client     *g.Client
	httpClient *http.Client
	config     *EntryConfig
	logger     hclog.Logger
}

func (gc *gitlabClient) CreatePipelineProjectTriggerAccessToken(ctx context.Context, projectId int, description string) (err error) {
	defer func() {
		gc.logger.Debug("Created a pipeline project trigger access token", "projectId", description, "description", "error", err)
	}()

	return err
}

func (gc *gitlabClient) RevokePipelineProjectTriggerAccessToken(ctx context.Context, projectId int, tokenId int) (err error) {
	defer func() {
		gc.logger.Debug("Revoked pipeline project trigger access token", "projectId", projectId, "tokenId", tokenId, "error", err)
	}()

	return err
}

func (gc *gitlabClient) GetGroupIdByPath(ctx context.Context, path string) (groupId int, err error) {
	defer func() {
		gc.logger.Debug("Get group id by path", "path", path, "groupId", groupId, "error", err)
	}()

	l := &g.ListGroupsOptions{
		Search: g.Ptr(path),
	}

	g, _, err := gc.client.Groups.ListGroups(l)
	if err != nil {
		return 0, fmt.Errorf("%v", err)
	}
	if len(g) == 0 {
		return 0, fmt.Errorf("path '%s' not found: %w", path, ErrInvalidValue)
	}
	groupId = g[0].ID
	return groupId, nil

}

func (gc *gitlabClient) GitlabClient(ctx context.Context) *g.Client {
	return gc.client
}

func (gc *gitlabClient) CreateGroupServiceAccountAccessToken(ctx context.Context, path string, groupId string, userId int, name string, expiresAt time.Time, scopes []string) (et *EntryToken, err error) {
	var at *g.PersonalAccessToken
	defer func() {
		gc.logger.Debug("Created group service access token", "pat", at, "et", et, "path", path, "groupId", groupId, "userId", userId, "name", name, "expiresAt", expiresAt, "scopes", scopes, "error", err)
	}()
	at, _, err = gc.client.Groups.CreateServiceAccountPersonalAccessToken(groupId, userId, &g.CreateServiceAccountPersonalAccessTokenOptions{
		Name:      g.Ptr(name),
		ExpiresAt: (*g.ISOTime)(&expiresAt),
		Scopes:    &scopes,
	})
	if err == nil {
		et = &EntryToken{
			TokenID:     at.ID,
			UserID:      userId,
			ParentID:    groupId,
			Path:        path,
			Name:        name,
			Token:       at.Token,
			TokenType:   TokenTypeGroupServiceAccount,
			CreatedAt:   at.CreatedAt,
			ExpiresAt:   (*time.Time)(at.ExpiresAt),
			Scopes:      scopes,
			AccessLevel: AccessLevelUnknown,
		}
	}
	return et, err
}

func (gc *gitlabClient) CreateUserServiceAccountAccessToken(ctx context.Context, username string, userId int, name string, expiresAt time.Time, scopes []string) (et *EntryToken, err error) {
	defer func() {
		gc.logger.Debug("Created user service access token", "et", et, "username", username, "userId", userId, "name", name, "expiresAt", expiresAt, "scopes", scopes, "error", err)
	}()
	et, err = gc.CreatePersonalAccessToken(ctx, username, userId, name, expiresAt, scopes)
	if err == nil && et != nil {
		et.TokenType = TokenTypeUserServiceAccount
	}
	return et, err
}

func (gc *gitlabClient) RevokeUserServiceAccountAccessToken(ctx context.Context, token string) (err error) {
	defer func() { gc.logger.Debug("Revoke user service account token", "token", token, "error", err) }()
	if token == "" {
		err = fmt.Errorf("%w: empty token", ErrNilValue)
		return err
	}

	var c *g.Client
	if c, err = newGitlabClient(&EntryConfig{
		BaseURL: gc.config.BaseURL,
		Token:   token,
	}, gc.httpClient); err == nil {
		_, err = c.PersonalAccessTokens.RevokePersonalAccessTokenSelf()
	}

	return err
}

func (gc *gitlabClient) RevokeGroupServiceAccountAccessToken(ctx context.Context, token string) (err error) {
	defer func() { gc.logger.Debug("Revoke group service account token", "token", token, "error", err) }()
	if token == "" {
		err = fmt.Errorf("%w: empty token", ErrNilValue)
		return err
	}

	var c *g.Client
	if c, err = newGitlabClient(&EntryConfig{
		BaseURL: gc.config.BaseURL,
		Token:   token,
	}, gc.httpClient); err == nil {
		_, err = c.PersonalAccessTokens.RevokePersonalAccessTokenSelf()
	}

	return err
}

func (gc *gitlabClient) CurrentTokenInfo(ctx context.Context) (et *EntryToken, err error) {
	var pat *g.PersonalAccessToken
	defer func() { gc.logger.Debug("Current token info", "token", et, "error", err) }()
	pat, _, err = gc.client.PersonalAccessTokens.GetSinglePersonalAccessToken()
	if err != nil {
		return nil, err
	}
	et = &EntryToken{
		TokenID:     pat.ID,
		UserID:      pat.UserID,
		ParentID:    "",
		Path:        "",
		Name:        pat.Name,
		Token:       pat.Token,
		TokenType:   TokenTypePersonal,
		CreatedAt:   pat.CreatedAt,
		ExpiresAt:   (*time.Time)(pat.ExpiresAt),
		Scopes:      pat.Scopes,
		AccessLevel: "",
	}
	return et, nil
}

func (gc *gitlabClient) RotateCurrentToken(ctx context.Context) (token *EntryToken, currentEntryToken *EntryToken, err error) {
	var expiresAt time.Time
	defer func() {
		gc.logger.Debug("Rotate current token", "token", token, "currentEntryToken", currentEntryToken, "expiresAt", expiresAt, "error", err)
	}()

	currentEntryToken, err = gc.CurrentTokenInfo(ctx)
	if err != nil {
		return nil, nil, err
	}

	var usr *g.User
	usr, _, err = gc.client.Users.GetUser(currentEntryToken.UserID, g.GetUsersOptions{})
	if err != nil {
		return nil, nil, err
	}

	var pat *g.PersonalAccessToken
	var durationTTL = currentEntryToken.ExpiresAt.Sub(*currentEntryToken.CreatedAt)
	_, expiresAt, _ = calculateGitlabTTL(durationTTL, TimeFromContext(ctx))
	pat, _, err = gc.client.PersonalAccessTokens.RotatePersonalAccessToken(
		currentEntryToken.TokenID,
		&g.RotatePersonalAccessTokenOptions{ExpiresAt: (*g.ISOTime)(&expiresAt)},
	)

	if err != nil {
		return nil, nil, err
	}

	token = &EntryToken{
		TokenID:     pat.ID,
		UserID:      pat.UserID,
		ParentID:    "",
		Path:        usr.Username,
		Name:        pat.Name,
		Token:       pat.Token,
		TokenType:   TokenTypePersonal,
		CreatedAt:   pat.CreatedAt,
		ExpiresAt:   (*time.Time)(pat.ExpiresAt),
		Scopes:      pat.Scopes,
		AccessLevel: AccessLevelUnknown,
	}

	gc.config.Token = token.Token
	gc.config.TokenId = token.TokenID
	gc.config.Scopes = token.Scopes
	if token.CreatedAt != nil {
		gc.config.TokenCreatedAt = *token.CreatedAt
	}
	if token.ExpiresAt != nil {
		gc.config.TokenExpiresAt = *token.ExpiresAt
	}

	gc.client = nil
	return token, currentEntryToken, err
}

func (gc *gitlabClient) GetUserIdByUsername(ctx context.Context, username string) (userId int, err error) {
	defer func() {
		gc.logger.Debug("Get user id by username", "username", username, "userId", userId, "error", err)
	}()

	l := &g.ListUsersOptions{
		Username: g.Ptr(username),
	}

	u, _, err := gc.client.Users.ListUsers(l)
	if err != nil {
		return 0, fmt.Errorf("%v", err)
	}
	if len(u) == 0 {
		return 0, fmt.Errorf("username '%s' not found: %w", username, ErrInvalidValue)
	}
	userId = u[0].ID
	return userId, nil
}

func (gc *gitlabClient) CreatePersonalAccessToken(ctx context.Context, username string, userId int, name string, expiresAt time.Time, scopes []string) (et *EntryToken, err error) {
	var at *g.PersonalAccessToken
	defer func() {
		gc.logger.Debug("Create personal access token", "pat", at, "et", et, "username", username, "userId", userId, "name", name, "expiresAt", expiresAt, "scopes", scopes, "error", err)
	}()
	at, _, err = gc.client.Users.CreatePersonalAccessToken(userId, &g.CreatePersonalAccessTokenOptions{
		Name:      g.Ptr(name),
		ExpiresAt: (*g.ISOTime)(&expiresAt),
		Scopes:    &scopes,
	})
	if err != nil {
		return nil, err
	}
	et = &EntryToken{
		TokenID:     at.ID,
		UserID:      userId,
		ParentID:    "",
		Path:        username,
		Name:        name,
		Token:       at.Token,
		TokenType:   TokenTypePersonal,
		CreatedAt:   at.CreatedAt,
		ExpiresAt:   (*time.Time)(at.ExpiresAt),
		Scopes:      scopes,
		AccessLevel: AccessLevelUnknown,
	}
	return et, nil
}

func (gc *gitlabClient) CreateGroupAccessToken(ctx context.Context, groupId string, name string, expiresAt time.Time, scopes []string, accessLevel AccessLevel) (et *EntryToken, err error) {
	var at *g.GroupAccessToken
	defer func() {
		gc.logger.Debug("Create group access token", "gat", at, "et", et, "groupId", groupId, "name", name, "expiresAt", expiresAt, "scopes", scopes, "accessLevel", accessLevel, "error", err)
	}()
	var al = new(g.AccessLevelValue)
	*al = g.AccessLevelValue(accessLevel.Value())
	at, _, err = gc.client.GroupAccessTokens.CreateGroupAccessToken(groupId, &g.CreateGroupAccessTokenOptions{
		Name:        g.Ptr(name),
		Scopes:      &scopes,
		ExpiresAt:   (*g.ISOTime)(&expiresAt),
		AccessLevel: al,
	})
	if err != nil {
		return nil, err
	}
	et = &EntryToken{
		TokenID:     at.ID,
		UserID:      0,
		ParentID:    groupId,
		Path:        groupId,
		Name:        name,
		Token:       at.Token,
		TokenType:   TokenTypeGroup,
		CreatedAt:   at.CreatedAt,
		ExpiresAt:   (*time.Time)(at.ExpiresAt),
		Scopes:      scopes,
		AccessLevel: accessLevel,
	}
	return et, nil
}

func (gc *gitlabClient) CreateProjectAccessToken(ctx context.Context, projectId string, name string, expiresAt time.Time, scopes []string, accessLevel AccessLevel) (*EntryToken, error) {
	var al = new(g.AccessLevelValue)
	*al = g.AccessLevelValue(accessLevel.Value())
	at, _, err := gc.client.ProjectAccessTokens.CreateProjectAccessToken(projectId, &g.CreateProjectAccessTokenOptions{
		Name:        g.Ptr(name),
		Scopes:      &scopes,
		ExpiresAt:   (*g.ISOTime)(&expiresAt),
		AccessLevel: al,
	})
	if err != nil {
		return nil, err
	}
	return &EntryToken{
		TokenID:     at.ID,
		UserID:      0,
		ParentID:    projectId,
		Path:        projectId,
		Name:        name,
		Token:       at.Token,
		TokenType:   TokenTypeProject,
		CreatedAt:   at.CreatedAt,
		ExpiresAt:   (*time.Time)(at.ExpiresAt),
		Scopes:      scopes,
		AccessLevel: accessLevel,
	}, nil
}

func (gc *gitlabClient) RevokePersonalAccessToken(ctx context.Context, tokenId int) (err error) {
	defer func() {
		gc.logger.Debug("Revoke personal access token", "tokenId", tokenId, "error", err)
	}()
	var resp *g.Response
	resp, err = gc.client.PersonalAccessTokens.RevokePersonalAccessToken(tokenId)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("personal: %w", ErrAccessTokenNotFound)
	}
	if err != nil {
		return err
	}
	return nil
}

func (gc *gitlabClient) RevokeProjectAccessToken(ctx context.Context, tokenId int, projectId string) (err error) {
	defer func() {
		gc.logger.Debug("Revoke project access token", "tokenId", tokenId, "error", err)
	}()
	var resp *g.Response
	resp, err = gc.client.ProjectAccessTokens.RevokeProjectAccessToken(projectId, tokenId)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("project: %w", ErrAccessTokenNotFound)
	}
	if err != nil {
		return err
	}
	return nil
}

func (gc *gitlabClient) RevokeGroupAccessToken(ctx context.Context, tokenId int, groupId string) (err error) {
	defer func() {
		gc.logger.Debug("Revoke group access token", "tokenId", tokenId, "error", err)
	}()
	var resp *g.Response
	resp, err = gc.client.GroupAccessTokens.RevokeGroupAccessToken(groupId, tokenId)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("group: %w", ErrAccessTokenNotFound)
	}
	if err != nil {
		return err
	}
	return nil
}

func (gc *gitlabClient) Valid(ctx context.Context) bool {
	return gc.client != nil && gc.config != nil
}

var _ Client = new(gitlabClient)

func newGitlabClient(config *EntryConfig, httpClient *http.Client) (gc *g.Client, err error) {
	if "" == strings.TrimSpace(config.BaseURL) {
		err = errors.Join(err, fmt.Errorf("gitlab base url: %w", ErrInvalidValue))
	}

	if "" == strings.TrimSpace(config.Token) {
		err = errors.Join(err, fmt.Errorf("gitlab token: %w", ErrInvalidValue))
	}

	if err != nil {
		return nil, err
	}

	var opts = []g.ClientOptionFunc{
		g.WithBaseURL(fmt.Sprintf("%s/api/v4", strings.TrimSuffix(config.BaseURL, "/"))),
		g.WithCustomLimiter(rate.NewLimiter(rate.Inf, 0)),
	}

	if httpClient != nil {
		opts = append(opts, g.WithHTTPClient(httpClient))
	}

	return g.NewClient(config.Token, opts...)
}

func NewGitlabClient(config *EntryConfig, httpClient *http.Client, logger hclog.Logger) (client Client, err error) {
	if config == nil {
		return nil, fmt.Errorf("configure the backend first, config: %w", ErrNilValue)
	}

	if logger == nil {
		logger = logging.NewVaultLoggerWithWriter(io.Discard, hclog.NoLevel)
	}

	var gc *g.Client
	if gc, err = newGitlabClient(config, httpClient); err != nil {
		return nil, err
	}

	return &gitlabClient{client: gc, config: config, logger: logger, httpClient: httpClient}, err
}
