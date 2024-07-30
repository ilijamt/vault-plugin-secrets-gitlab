package gitlab

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/helper/logging"
	g "github.com/xanzy/go-gitlab"
	"golang.org/x/time/rate"
)

var (
	ErrAccessTokenNotFound = errors.New("access token not found")
	ErrRoleNotFound        = errors.New("role not found")
)

type Client interface {
	Valid() bool
	CurrentTokenInfo() (*EntryToken, error)
	RotateCurrentToken() (newToken *EntryToken, oldToken *EntryToken, err error)
	CreatePersonalAccessToken(username string, userId int, name string, expiresAt time.Time, scopes []string) (*EntryToken, error)
	CreateServiceAccountPersonalAccessToken(path string, groupId string, userId int, name string, expiresAt time.Time, scopes []string) (*EntryToken, error)
	CreateGroupAccessToken(groupId string, name string, expiresAt time.Time, scopes []string, accessLevel AccessLevel) (*EntryToken, error)
	CreateProjectAccessToken(projectId string, name string, expiresAt time.Time, scopes []string, accessLevel AccessLevel) (*EntryToken, error)
	RevokePersonalAccessToken(tokenId int) error
	RevokeProjectAccessToken(tokenId int, projectId string) error
	RevokeGroupAccessToken(tokenId int, groupId string) error
	GetUserIdByUsername(username string) (int, error)
	GetRolePathParts(path string) (interface{}, interface{}, error)
}

type gitlabClient struct {
	client *g.Client
	config *EntryConfig
	logger hclog.Logger
}

func (gc *gitlabClient) CurrentTokenInfo() (et *EntryToken, err error) {
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

func (gc *gitlabClient) RotateCurrentToken() (token *EntryToken, currentEntryToken *EntryToken, err error) {
	var expiresAt time.Time
	defer func() {
		gc.logger.Debug("Rotate current token", "token", token, "currentEntryToken", currentEntryToken, "expiresAt", expiresAt, "error", err)
	}()

	currentEntryToken, err = gc.CurrentTokenInfo()
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
	_, expiresAt, _ = calculateGitlabTTL(durationTTL, time.Now())
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

func (gc *gitlabClient) GetUserIdByUsername(username string) (userId int, err error) {
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

func (gc *gitlabClient) CreatePersonalAccessToken(username string, userId int, name string, expiresAt time.Time, scopes []string) (et *EntryToken, err error) {
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

func (gc *gitlabClient) GetRolePathParts(path string) (interface{}, interface{}, error) {
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		return nil, nil, errors.New("Too many arguments for service account path - eg: 1234/my-service-account")
	}
	groupId := parts[0]
	username := parts[1]

	return groupId, username, nil
}

func (gc *gitlabClient) CreateServiceAccountPersonalAccessToken(path string, groupId string, userId int, name string, expiresAt time.Time, scopes []string) (et *EntryToken, err error) {
	var at *g.PersonalAccessToken
	defer func() {
		gc.logger.Debug("Create service account personal access token", "pat", at, "et", et, "groupId", groupId, "serviceAccountId", userId, "name", name, "expiresAt", expiresAt, "scopes", scopes, "error", err)
	}()
	at, _, err = gc.client.Groups.CreateServiceAccountPersonalAccessToken(groupId, userId, &g.CreateServiceAccountPersonalAccessTokenOptions{
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
		ParentID:    groupId,
		Path:        path,
		Name:        name,
		Token:       at.Token,
		TokenType:   TokenTypeServiceAccount,
		CreatedAt:   at.CreatedAt,
		ExpiresAt:   (*time.Time)(at.ExpiresAt),
		Scopes:      scopes,
		AccessLevel: AccessLevelUnknown,
	}
	return et, nil
}

func (gc *gitlabClient) CreateGroupAccessToken(groupId string, name string, expiresAt time.Time, scopes []string, accessLevel AccessLevel) (et *EntryToken, err error) {
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

func (gc *gitlabClient) CreateProjectAccessToken(projectId string, name string, expiresAt time.Time, scopes []string, accessLevel AccessLevel) (*EntryToken, error) {
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

func (gc *gitlabClient) RevokePersonalAccessToken(tokenId int) (err error) {
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

func (gc *gitlabClient) RevokeProjectAccessToken(tokenId int, projectId string) (err error) {
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

func (gc *gitlabClient) RevokeGroupAccessToken(tokenId int, groupId string) (err error) {
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

func (gc *gitlabClient) Valid() bool {
	return gc.client != nil && gc.config != nil
}

var _ Client = new(gitlabClient)

func NewGitlabClient(config *EntryConfig, httpClient *http.Client, logger hclog.Logger) (client Client, err error) {
	if config == nil {
		return nil, fmt.Errorf("configure the backend first, config: %w", ErrNilValue)
	}

	if logger == nil {
		logger = logging.NewVaultLoggerWithWriter(io.Discard, hclog.NoLevel)
	}

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

	var gc *g.Client
	gc, err = g.NewClient(config.Token, opts...)

	return &gitlabClient{client: gc, config: config, logger: logger}, err
}
