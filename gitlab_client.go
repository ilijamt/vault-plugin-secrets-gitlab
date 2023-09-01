package gitlab

import (
	"errors"
	"fmt"
	g "github.com/xanzy/go-gitlab"
	"golang.org/x/time/rate"
	"net/http"
	"time"
)

var (
	ErrAccessTokenNotFound = errors.New("access token not found")
	ErrRoleNotFound        = errors.New("role not found")
)

type Client interface {
	Valid() bool

	MainTokenInfo() (*EntryToken, error)
	RotateMainToken(revokeOldToken bool) (*EntryToken, error)
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
	config *entryConfig
}

func (gc *gitlabClient) MainTokenInfo() (*EntryToken, error) {
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

func (gc *gitlabClient) RotateMainToken(revokeOldToken bool) (*EntryToken, error) {
	if gc.client == nil {
		return nil, fmt.Errorf("client: %w", ErrNilValue)
	}

	var currentEntryToken, err = gc.MainTokenInfo()
	if err != nil {
		return nil, err
	}
	var expiresAt = *currentEntryToken.ExpiresAt
	var durationTTL = currentEntryToken.CreatedAt.Sub(expiresAt)

	var usr *g.User
	usr, _, err = gc.client.Users.GetUser(currentEntryToken.UserID, g.GetUsersOptions{})
	if err != nil {
		return nil, err
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
		return nil, err
	}

	gc.config.Token = token.Token
	if token.ExpiresAt != nil {
		gc.config.TokenExpiresAt = *token.ExpiresAt
	}

	if revokeOldToken {
		_, err = gc.client.PersonalAccessTokens.RevokePersonalAccessToken(currentEntryToken.TokenID)
	}

	return token, err
}

func (gc *gitlabClient) GetUserIdByUsername(username string) (int, error) {
	l := &g.ListUsersOptions{
		Username: g.String(username),
	}

	u, _, err := gc.client.Users.ListUsers(l)
	if err != nil {
		return fmt.Printf("%v", err)
	}
	if username != u[0].Username {
		return fmt.Printf("%v", username)
	}

	return u[0].ID, nil
}

func (gc *gitlabClient) CreatePersonalAccessToken(username string, userId int, name string, expiresAt time.Time, scopes []string) (*EntryToken, error) {
	at, _, err := gc.client.Users.CreatePersonalAccessToken(userId, &g.CreatePersonalAccessTokenOptions{
		Name:      g.String(name),
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
		Name:        g.String(name),
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
		Name:        g.String(name),
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

func NewGitlabClient(config *entryConfig) (client Client, err error) {
	if config == nil {
		return nil, fmt.Errorf("configure the backend first, config: %w", ErrNilValue)
	}

	var gc *g.Client
	if gc, err = g.NewClient(config.Token,
		g.WithBaseURL(fmt.Sprintf("%s/api/v4", config.BaseURL)),
		g.WithCustomLimiter(rate.NewLimiter(rate.Inf, 0)),
	); err != nil {
		return nil, err
	}

	return &gitlabClient{client: gc, config: config}, nil
}
