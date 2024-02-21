package gitlab

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

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
	RotateCurrentToken(revokeOldToken bool) (newToken *EntryToken, oldToken *EntryToken, err error)
	CreatePersonalAccessToken(username string, userId int, name string, expiresAt time.Time, scopes []string) (*EntryToken, error)
	CreateGroupAccessToken(groupId string, name string, expiresAt time.Time, scopes []string, accessLevel AccessLevel) (*EntryToken, error)
	CreateProjectAccessToken(projectId string, name string, expiresAt time.Time, scopes []string, accessLevel AccessLevel) (*EntryToken, error)
	RevokePersonalAccessToken(tokenId int) error
	RevokeProjectAccessToken(tokenId int, projectId string) error
	RevokeGroupAccessToken(tokenId int, groupId string) error
	GetUserIdByUsername(username string) (int, error)
}

type gitlabClient struct {
	client *g.Client
	config *EntryConfig
}

func (gc *gitlabClient) CurrentTokenInfo() (*EntryToken, error) {
	var pat, _, err = gc.client.PersonalAccessTokens.GetSinglePersonalAccessToken()
	if err != nil {
		return nil, err
	}
	return &EntryToken{
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
	}, nil
}

func (gc *gitlabClient) RotateCurrentToken(revokeOldToken bool) (*EntryToken, *EntryToken, error) {
	var currentEntryToken, err = gc.CurrentTokenInfo()
	if err != nil {
		return nil, nil, err
	}
	var expiresAt = *currentEntryToken.ExpiresAt
	var durationTTL = expiresAt.Sub(*currentEntryToken.CreatedAt)

	var usr *g.User
	usr, _, err = gc.client.Users.GetUser(currentEntryToken.UserID, g.GetUsersOptions{})
	if err != nil {
		return nil, nil, err
	}

	var token *EntryToken
	token, err = gc.CreatePersonalAccessToken(
		usr.Username,
		currentEntryToken.UserID,
		fmt.Sprintf("%s-%d", currentEntryToken.Name, time.Now().Unix()),
		time.Now().Add(durationTTL),
		currentEntryToken.Scopes,
	)
	if err != nil {
		return nil, nil, err
	}

	gc.config.Token = token.Token
	if token.ExpiresAt != nil {
		gc.config.TokenExpiresAt = *token.ExpiresAt
	}

	if revokeOldToken {
		_, err = gc.client.PersonalAccessTokens.RevokePersonalAccessToken(currentEntryToken.TokenID)
	}

	gc.client = nil
	return token, currentEntryToken, err
}

func (gc *gitlabClient) GetUserIdByUsername(username string) (int, error) {
	l := &g.ListUsersOptions{
		Username: g.Ptr(username),
	}

	u, _, err := gc.client.Users.ListUsers(l)
	if err != nil {
		return 0, fmt.Errorf("%v", err)
	}
	if username != u[0].Username {
		return 0, fmt.Errorf("%v does not match with %s: %w", u[0].Username, username, ErrInvalidValue)
	}

	return u[0].ID, nil
}

func (gc *gitlabClient) CreatePersonalAccessToken(username string, userId int, name string, expiresAt time.Time, scopes []string) (*EntryToken, error) {
	at, _, err := gc.client.Users.CreatePersonalAccessToken(userId, &g.CreatePersonalAccessTokenOptions{
		Name:      g.Ptr(name),
		ExpiresAt: (*g.ISOTime)(&expiresAt),
		Scopes:    &scopes,
	})
	if err != nil {
		return nil, err
	}
	return &EntryToken{
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
	}, nil
}

func (gc *gitlabClient) CreateGroupAccessToken(groupId string, name string, expiresAt time.Time, scopes []string, accessLevel AccessLevel) (*EntryToken, error) {
	var al = new(g.AccessLevelValue)
	*al = g.AccessLevelValue(accessLevel.Value())
	at, _, err := gc.client.GroupAccessTokens.CreateGroupAccessToken(groupId, &g.CreateGroupAccessTokenOptions{
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
		ParentID:    groupId,
		Path:        groupId,
		Name:        name,
		Token:       at.Token,
		TokenType:   TokenTypeGroup,
		CreatedAt:   at.CreatedAt,
		ExpiresAt:   (*time.Time)(at.ExpiresAt),
		Scopes:      scopes,
		AccessLevel: accessLevel,
	}, nil
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

func (gc *gitlabClient) RevokePersonalAccessToken(tokenId int) error {
	var resp, err = gc.client.PersonalAccessTokens.RevokePersonalAccessToken(tokenId)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("personal: %w", ErrAccessTokenNotFound)
	}
	if err != nil {
		return err
	}
	return nil
}

func (gc *gitlabClient) RevokeProjectAccessToken(tokenId int, projectId string) error {
	var resp, err = gc.client.ProjectAccessTokens.RevokeProjectAccessToken(projectId, tokenId)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("project: %w", ErrAccessTokenNotFound)
	}
	if err != nil {
		return err
	}
	return nil
}

func (gc *gitlabClient) RevokeGroupAccessToken(tokenId int, groupId string) error {
	var resp, err = gc.client.GroupAccessTokens.RevokeGroupAccessToken(groupId, tokenId)
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

func NewGitlabClient(config *EntryConfig, httpClient *http.Client) (client Client, err error) {
	if config == nil {
		return nil, fmt.Errorf("configure the backend first, config: %w", ErrNilValue)
	}

	if "" == strings.TrimSpace(config.BaseURL) || "" == strings.TrimSpace(config.Token) {
		return nil, fmt.Errorf("base url or token is empty: %w", ErrInvalidValue)
	}

	var opts = []g.ClientOptionFunc{
		g.WithBaseURL(fmt.Sprintf("%s/api/v4", config.BaseURL)),
		g.WithCustomLimiter(rate.NewLimiter(rate.Inf, 0)),
	}

	if httpClient != nil {
		opts = append(opts, g.WithHTTPClient(httpClient))
	}

	var gc *g.Client
	gc, err = g.NewClient(config.Token, opts...)

	return &gitlabClient{client: gc, config: config}, err
}
