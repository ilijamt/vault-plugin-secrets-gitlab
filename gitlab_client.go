package gitlab

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
	Metadata(ctx context.Context) (*g.Metadata, error)
	CurrentTokenInfo(ctx context.Context) (*TokenConfig, error)
	RotateCurrentToken(ctx context.Context) (newToken *TokenConfig, oldToken *TokenConfig, err error)
	CreatePersonalAccessToken(ctx context.Context, username string, userId int, name string, expiresAt time.Time, scopes []string) (*TokenPersonal, error)
	CreateGroupAccessToken(ctx context.Context, groupId string, name string, expiresAt time.Time, scopes []string, accessLevel AccessLevel) (*TokenGroup, error)
	CreateProjectAccessToken(ctx context.Context, projectId string, name string, expiresAt time.Time, scopes []string, accessLevel AccessLevel) (*TokenProject, error)
	RevokePersonalAccessToken(ctx context.Context, tokenId int) error
	RevokeProjectAccessToken(ctx context.Context, tokenId int, projectId string) error
	RevokeGroupAccessToken(ctx context.Context, tokenId int, groupId string) error
	GetUserIdByUsername(ctx context.Context, username string) (int, error)
	GetGroupIdByPath(ctx context.Context, path string) (int, error)
	GetProjectIdByPath(ctx context.Context, path string) (int, error)
	CreateGroupServiceAccountAccessToken(ctx context.Context, group string, groupId string, userId int, name string, expiresAt time.Time, scopes []string) (*TokenGroupServiceAccount, error)
	CreateUserServiceAccountAccessToken(ctx context.Context, username string, userId int, name string, expiresAt time.Time, scopes []string) (*TokenUserServiceAccount, error)
	RevokeUserServiceAccountAccessToken(ctx context.Context, token string) error
	RevokeGroupServiceAccountAccessToken(ctx context.Context, token string) error
	CreatePipelineProjectTriggerAccessToken(ctx context.Context, path, name string, projectId int, description string, expiresAt *time.Time) (*TokenPipelineProjectTrigger, error)
	RevokePipelineProjectTriggerAccessToken(ctx context.Context, projectId int, tokenId int) error
	CreateProjectDeployToken(ctx context.Context, path string, projectId int, name string, expiresAt *time.Time, scopes []string) (et *TokenProjectDeploy, err error)
	RevokeProjectDeployToken(ctx context.Context, projectId, deployTokenId int) (err error)
	CreateGroupDeployToken(ctx context.Context, path string, groupId int, name string, expiresAt *time.Time, scopes []string) (et *TokenGroupDeploy, err error)
	RevokeGroupDeployToken(ctx context.Context, groupId, deployTokenId int) (err error)
}

type gitlabClient struct {
	client     *g.Client
	httpClient *http.Client
	config     *EntryConfig
	logger     hclog.Logger
}

func (gc *gitlabClient) GetProjectIdByPath(ctx context.Context, path string) (projectId int, err error) {
	defer func() {
		gc.logger.Debug("Get project id by path", "path", path, "projectId", projectId, "error", err)
	}()

	projectId = -1
	var project *g.Project
	if project, _, err = gc.client.Projects.GetProject(path, &g.GetProjectOptions{}, g.WithContext(ctx)); err == nil {
		projectId = project.ID
	}

	return projectId, err
}

func (gc *gitlabClient) CreateGroupDeployToken(ctx context.Context, path string, groupId int, name string, expiresAt *time.Time, scopes []string) (et *TokenGroupDeploy, err error) {
	var dt *g.DeployToken
	defer func() {
		gc.logger.Debug("Create group deploy token", "groupId", groupId, "name", name, "path", path, "expiresAt", expiresAt, "scopes", scopes, "error", err)
	}()

	if dt, _, err = gc.client.DeployTokens.CreateGroupDeployToken(
		groupId,
		&g.CreateGroupDeployTokenOptions{
			Name:      &name,
			ExpiresAt: expiresAt,
			Scopes:    &scopes,
		},
		g.WithContext(ctx),
	); err == nil {
		et = &TokenGroupDeploy{
			TokenWithScopes: TokenWithScopes{
				Token: Token{
					TokenID:   dt.ID,
					ParentID:  strconv.Itoa(groupId),
					Path:      path,
					Name:      name,
					Token:     dt.Token,
					TokenType: TokenTypeGroupDeploy,
					CreatedAt: g.Ptr(time.Now()),
				},
				Scopes: scopes,
			},
			Username: dt.Username,
		}
	}
	return et, err
}

func (gc *gitlabClient) CreateProjectDeployToken(ctx context.Context, path string, projectId int, name string, expiresAt *time.Time, scopes []string) (et *TokenProjectDeploy, err error) {
	var dt *g.DeployToken
	defer func() {
		gc.logger.Debug("Create project deploy token", "projectId", projectId, "name", name, "path", path, "expiresAt", expiresAt, "scopes", scopes, "error", err)
	}()
	if dt, _, err = gc.client.DeployTokens.CreateProjectDeployToken(
		projectId,
		&g.CreateProjectDeployTokenOptions{
			Name:      &name,
			ExpiresAt: expiresAt,
			Scopes:    &scopes,
		},
		g.WithContext(ctx),
	); err == nil {
		et = &TokenProjectDeploy{
			TokenWithScopes: TokenWithScopes{
				Token: Token{
					TokenID:   dt.ID,
					ParentID:  strconv.Itoa(projectId),
					Path:      path,
					Name:      name,
					Token:     dt.Token,
					TokenType: TokenTypeProjectDeploy,
					CreatedAt: g.Ptr(time.Now()),
				},
				Scopes: scopes,
			},
			Username: dt.Username,
		}
	}
	return et, err
}

func (gc *gitlabClient) RevokeGroupDeployToken(ctx context.Context, groupId, deployTokenId int) (err error) {
	defer func() {
		gc.logger.Debug("Revoke group deploy token", "groupId", groupId, "deployTokenId", deployTokenId, "error", err)
	}()

	_, err = gc.client.DeployTokens.DeleteGroupDeployToken(groupId, deployTokenId, g.WithContext(ctx))
	return err
}

func (gc *gitlabClient) RevokeProjectDeployToken(ctx context.Context, projectId, deployTokenId int) (err error) {
	defer func() {
		gc.logger.Debug("Revoke project deploy token", "projectId", projectId, "deployTokenId", deployTokenId, "error", err)
	}()

	_, err = gc.client.DeployTokens.DeleteProjectDeployToken(projectId, deployTokenId, g.WithContext(ctx))
	return err
}

func (gc *gitlabClient) Metadata(ctx context.Context) (metadata *g.Metadata, err error) {
	defer func() {
		gc.logger.Debug("Fetch metadata information", "metadata", metadata, "error", err)
	}()

	metadata, _, err = gc.client.Metadata.GetMetadata(g.WithContext(ctx))
	return metadata, err
}

func (gc *gitlabClient) CreatePipelineProjectTriggerAccessToken(ctx context.Context, path, name string, projectId int, description string, expiresAt *time.Time) (et *TokenPipelineProjectTrigger, err error) {
	var pt *g.PipelineTrigger
	defer func() {
		gc.logger.Debug("Create a pipeline project trigger access token", "path", path, "name", name, "projectId", description, "description", "error", err)
	}()

	if pt, _, err = gc.client.PipelineTriggers.AddPipelineTrigger(
		projectId,
		&g.AddPipelineTriggerOptions{Description: &description},
		g.WithContext(ctx),
	); err == nil {
		et = &TokenPipelineProjectTrigger{
			Token: Token{
				TokenID:   pt.ID,
				ParentID:  strconv.Itoa(projectId),
				Path:      path,
				Name:      name,
				Token:     pt.Token,
				TokenType: TokenTypePipelineProjectTrigger,
				CreatedAt: g.Ptr(time.Now()),
				ExpiresAt: expiresAt,
			},
		}
	}

	return et, err
}

func (gc *gitlabClient) RevokePipelineProjectTriggerAccessToken(ctx context.Context, projectId int, tokenId int) (err error) {
	defer func() {
		gc.logger.Debug("Revoke pipeline project trigger access token", "projectId", projectId, "tokenId", tokenId, "error", err)
	}()

	_, err = gc.client.PipelineTriggers.DeletePipelineTrigger(projectId, tokenId, g.WithContext(ctx))
	return err
}

func (gc *gitlabClient) GetGroupIdByPath(ctx context.Context, path string) (groupId int, err error) {
	defer func() {
		gc.logger.Debug("Get group id by path", "path", path, "groupId", groupId, "error", err)
	}()

	l := &g.ListGroupsOptions{
		Search: g.Ptr(path),
	}

	groups, _, err := gc.client.Groups.ListGroups(l, g.WithContext(ctx))
	if err != nil {
		return 0, fmt.Errorf("%v", err)
	}
	if len(groups) == 0 {
		return 0, fmt.Errorf("path '%s' not found: %w", path, ErrInvalidValue)
	}
	groupId = groups[0].ID
	return groupId, nil

}

func (gc *gitlabClient) GitlabClient(ctx context.Context) *g.Client {
	return gc.client
}

func (gc *gitlabClient) CreateGroupServiceAccountAccessToken(ctx context.Context, path string, groupId string, userId int, name string, expiresAt time.Time, scopes []string) (et *TokenGroupServiceAccount, err error) {
	var at *g.PersonalAccessToken
	defer func() {
		gc.logger.Debug("Create group service access token", "pat", at, "et", et, "path", path, "groupId", groupId, "userId", userId, "name", name, "expiresAt", expiresAt, "scopes", scopes, "error", err)
	}()
	at, _, err = gc.client.Groups.CreateServiceAccountPersonalAccessToken(groupId, userId, &g.CreateServiceAccountPersonalAccessTokenOptions{
		Name:      g.Ptr(name),
		ExpiresAt: (*g.ISOTime)(&expiresAt),
		Scopes:    &scopes,
	}, g.WithContext(ctx))
	if err == nil {
		et = &TokenGroupServiceAccount{
			TokenWithScopes: TokenWithScopes{
				Token: Token{
					TokenID:   at.ID,
					ParentID:  groupId,
					Path:      path,
					Name:      name,
					Token:     at.Token,
					TokenType: TokenTypeGroupServiceAccount,
					CreatedAt: at.CreatedAt,
					ExpiresAt: (*time.Time)(at.ExpiresAt),
				},
				Scopes: scopes,
			},
			UserID: userId,
		}
	}
	return et, err
}

func (gc *gitlabClient) CreateUserServiceAccountAccessToken(ctx context.Context, username string, userId int, name string, expiresAt time.Time, scopes []string) (et *TokenUserServiceAccount, err error) {
	defer func() {
		gc.logger.Debug("Create user service access token", "et", et, "username", username, "userId", userId, "name", name, "expiresAt", expiresAt, "scopes", scopes, "error", err)
	}()
	var etp *TokenPersonal
	etp, err = gc.CreatePersonalAccessToken(ctx, username, userId, name, expiresAt, scopes)
	if err == nil && etp != nil {
		et = &TokenUserServiceAccount{
			TokenWithScopes: TokenWithScopes{
				Token: Token{
					TokenID:   etp.TokenID,
					ParentID:  etp.ParentID,
					Path:      etp.Path,
					Name:      etp.Name,
					Token:     etp.Token.Token,
					TokenType: TokenTypeUserServiceAccount,
					CreatedAt: etp.CreatedAt,
					ExpiresAt: etp.ExpiresAt,
				},
				Scopes: etp.Scopes,
			},
		}
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
		_, err = c.PersonalAccessTokens.RevokePersonalAccessTokenSelf(g.WithContext(ctx))
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
		_, err = c.PersonalAccessTokens.RevokePersonalAccessTokenSelf(g.WithContext(ctx))
	}

	return err
}

func (gc *gitlabClient) CurrentTokenInfo(ctx context.Context) (et *TokenConfig, err error) {
	var pat *g.PersonalAccessToken
	defer func() { gc.logger.Debug("Current token info", "token", et, "error", err) }()
	if pat, _, err = gc.client.PersonalAccessTokens.GetSinglePersonalAccessToken(g.WithContext(ctx)); err == nil {
		et = &TokenConfig{
			TokenWithScopes: TokenWithScopes{
				Token: Token{
					TokenID:   pat.ID,
					Name:      pat.Name,
					Token:     pat.Token,
					TokenType: TokenTypePersonal,
					CreatedAt: pat.CreatedAt,
					ExpiresAt: (*time.Time)(pat.ExpiresAt),
				},
				Scopes: pat.Scopes,
			},
			UserID: pat.UserID,
		}
		// Set an artificial expiry date one year after creation if none is set.
		// This addresses issue #178 where a token could be issued without an expiry.
		// Note: As of GitLab 16.x, all tokens should have an expiry set.
		if et.ExpiresAt == nil {
			newDate := pat.CreatedAt.AddDate(1, 0, -2)
			et.ExpiresAt = &newDate
		}
	}
	return et, err
}

func (gc *gitlabClient) RotateCurrentToken(ctx context.Context) (token *TokenConfig, currentEntryToken *TokenConfig, err error) {
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

	token = &TokenConfig{
		TokenWithScopes: TokenWithScopes{
			Token: Token{
				TokenID:   pat.ID,
				ParentID:  "",
				Path:      usr.Username,
				Name:      pat.Name,
				Token:     pat.Token,
				TokenType: TokenTypePersonal,
				CreatedAt: pat.CreatedAt,
				ExpiresAt: (*time.Time)(pat.ExpiresAt),
			},
			Scopes: pat.Scopes,
		},
		UserID: pat.UserID,
	}

	gc.config.Token = token.Token.Token
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

	u, _, err := gc.client.Users.ListUsers(l, g.WithContext(ctx))
	if err != nil {
		return 0, fmt.Errorf("%v", err)
	}
	if len(u) == 0 {
		return 0, fmt.Errorf("username '%s' not found: %w", username, ErrInvalidValue)
	}
	userId = u[0].ID
	return userId, nil
}

func (gc *gitlabClient) CreatePersonalAccessToken(ctx context.Context, username string, userId int, name string, expiresAt time.Time, scopes []string) (et *TokenPersonal, err error) {
	var at *g.PersonalAccessToken
	defer func() {
		gc.logger.Debug("Create personal access token", "pat", at, "et", et, "username", username, "userId", userId, "name", name, "expiresAt", expiresAt, "scopes", scopes, "error", err)
	}()
	if at, _, err = gc.client.Users.CreatePersonalAccessToken(userId, &g.CreatePersonalAccessTokenOptions{
		Name:      g.Ptr(name),
		ExpiresAt: (*g.ISOTime)(&expiresAt),
		Scopes:    &scopes,
	}, g.WithContext(ctx)); err == nil {
		et = &TokenPersonal{
			TokenWithScopes: TokenWithScopes{
				Token: Token{
					TokenID:   at.ID,
					Path:      username,
					Name:      name,
					Token:     at.Token,
					TokenType: TokenTypePersonal,
					CreatedAt: at.CreatedAt,
					ExpiresAt: (*time.Time)(at.ExpiresAt),
				},
				Scopes: scopes,
			},
			UserID: userId,
		}
	}
	return et, err
}

func (gc *gitlabClient) CreateGroupAccessToken(ctx context.Context, groupId string, name string, expiresAt time.Time, scopes []string, accessLevel AccessLevel) (et *TokenGroup, err error) {
	var at *g.GroupAccessToken
	defer func() {
		gc.logger.Debug("Create group access token", "gat", at, "et", et, "groupId", groupId, "name", name, "expiresAt", expiresAt, "scopes", scopes, "accessLevel", accessLevel, "error", err)
	}()
	var al = new(g.AccessLevelValue)
	*al = g.AccessLevelValue(accessLevel.Value())
	if at, _, err = gc.client.GroupAccessTokens.CreateGroupAccessToken(groupId, &g.CreateGroupAccessTokenOptions{
		Name:        g.Ptr(name),
		Scopes:      &scopes,
		ExpiresAt:   (*g.ISOTime)(&expiresAt),
		AccessLevel: al,
	}, g.WithContext(ctx)); err == nil {
		et = &TokenGroup{
			TokenWithScopesAndAccessLevel: TokenWithScopesAndAccessLevel{
				Token: Token{
					TokenID:   at.ID,
					ParentID:  groupId,
					Path:      groupId,
					Name:      name,
					Token:     at.Token,
					TokenType: TokenTypeGroup,
					CreatedAt: at.CreatedAt,
					ExpiresAt: (*time.Time)(at.ExpiresAt),
				},
				Scopes:      scopes,
				AccessLevel: accessLevel,
			},
		}
	}
	return et, err
}

func (gc *gitlabClient) CreateProjectAccessToken(ctx context.Context, projectId string, name string, expiresAt time.Time, scopes []string, accessLevel AccessLevel) (et *TokenProject, err error) {
	var at *g.ProjectAccessToken
	defer func() {
		gc.logger.Debug("Create project access token", "gat", at, "et", et, "projectId", projectId, "name", name, "expiresAt", expiresAt, "scopes", scopes, "accessLevel", accessLevel, "error", err)
	}()
	var al = new(g.AccessLevelValue)
	*al = g.AccessLevelValue(accessLevel.Value())
	if at, _, err = gc.client.ProjectAccessTokens.CreateProjectAccessToken(projectId, &g.CreateProjectAccessTokenOptions{
		Name:        g.Ptr(name),
		Scopes:      &scopes,
		ExpiresAt:   (*g.ISOTime)(&expiresAt),
		AccessLevel: al,
	}, g.WithContext(ctx)); err == nil {
		et = &TokenProject{
			TokenWithScopesAndAccessLevel: TokenWithScopesAndAccessLevel{
				Token: Token{
					TokenID:   at.ID,
					ParentID:  projectId,
					Path:      projectId,
					Name:      name,
					Token:     at.Token,
					TokenType: TokenTypeProject,
					CreatedAt: at.CreatedAt,
					ExpiresAt: (*time.Time)(at.ExpiresAt),
				},
				Scopes:      scopes,
				AccessLevel: accessLevel,
			},
		}
	}
	return et, err
}

func (gc *gitlabClient) RevokePersonalAccessToken(ctx context.Context, tokenId int) (err error) {
	defer func() {
		gc.logger.Debug("Revoke personal access token", "tokenId", tokenId, "error", err)
	}()
	var resp *g.Response
	resp, err = gc.client.PersonalAccessTokens.RevokePersonalAccessToken(tokenId, g.WithContext(ctx))
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
	resp, err = gc.client.ProjectAccessTokens.RevokeProjectAccessToken(projectId, tokenId, g.WithContext(ctx))
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
	resp, err = gc.client.GroupAccessTokens.RevokeGroupAccessToken(groupId, tokenId, g.WithContext(ctx))
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
	if strings.TrimSpace(config.BaseURL) == "" {
		err = errors.Join(err, fmt.Errorf("gitlab base url: %w", ErrInvalidValue))
	}

	if strings.TrimSpace(config.Token) == "" {
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
