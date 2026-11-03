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
	sdklogging "github.com/hashicorp/vault/sdk/helper/logging"
	g "gitlab.com/gitlab-org/api/client-go/v3"
	"golang.org/x/time/rate"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/logging"
	modelConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
	modelToken "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/token"
	t "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

type gitlabClient struct {
	client     *g.Client
	httpClient *http.Client
	config     *modelConfig.EntryConfig
	log        logging.Logger
}

func revokeResult(kind string, resp *g.Response, err error) error {
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("%s: %w", kind, errs.ErrAccessTokenNotFound)
	}
	return err
}

// Deferred log calls also run on failure, when et is nil.
func tokenLogFields[T any, P interface {
	*T
	LogFields() logging.Fields
}](et P) logging.Fields {
	if et == nil {
		return logging.Fields{}
	}
	return et.LogFields()
}

func (gc *gitlabClient) GetProjectIdByPath(ctx context.Context, path string) (projectId int64, err error) {
	defer func() {
		gc.log.Debug("Get project id by path", logging.Fields{Path: path, ProjectID: fmt.Sprint(projectId), Err: err})
	}()

	projectId = -1
	var project *g.Project
	if project, _, err = gc.client.Projects.GetProject(path, &g.GetProjectOptions{}, g.WithContext(ctx)); err == nil {
		projectId = project.ID
	}

	return projectId, err
}

func (gc *gitlabClient) CreateGroupDeployToken(ctx context.Context, path string, groupId int64, name string, expiresAt *time.Time, scopes []string) (et *modelToken.TokenGroupDeploy, err error) {
	var dt *g.DeployToken
	defer func() {
		gc.log.Debug("Create group deploy token", logging.Fields{GroupID: fmt.Sprint(groupId), Name: name, Path: path, ExpiresAt: expiresAt, Scopes: scopes, Err: err})
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
		et = &modelToken.TokenGroupDeploy{
			TokenWithScopes: modelToken.TokenWithScopes{
				Token: modelToken.Token{
					TokenID:   dt.ID,
					ParentID:  strconv.FormatInt(groupId, 10),
					Path:      path,
					Name:      name,
					Token:     dt.Token,
					TokenType: t.TypeGroupDeploy,
					CreatedAt: new(time.Now()),
				},
				Scopes: scopes,
			},
			Username: dt.Username,
		}
	}
	return et, err
}

func (gc *gitlabClient) CreateProjectDeployToken(ctx context.Context, path string, projectId int64, name string, expiresAt *time.Time, scopes []string) (et *modelToken.TokenProjectDeploy, err error) {
	var dt *g.DeployToken
	defer func() {
		gc.log.Debug("Create project deploy token", logging.Fields{ProjectID: fmt.Sprint(projectId), Name: name, Path: path, ExpiresAt: expiresAt, Scopes: scopes, Err: err})
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
		et = &modelToken.TokenProjectDeploy{
			TokenWithScopes: modelToken.TokenWithScopes{
				Token: modelToken.Token{
					TokenID:   dt.ID,
					ParentID:  strconv.FormatInt(projectId, 10),
					Path:      path,
					Name:      name,
					Token:     dt.Token,
					TokenType: t.TypeProjectDeploy,
					CreatedAt: new(time.Now()),
				},
				Scopes: scopes,
			},
			Username: dt.Username,
		}
	}
	return et, err
}

func (gc *gitlabClient) RevokeGroupDeployToken(ctx context.Context, tokenId int64, groupId string) (err error) {
	defer func() {
		gc.log.Debug("Revoke group deploy token", logging.Fields{TokenID: tokenId, GroupID: groupId, Err: err})
	}()
	resp, err := gc.client.DeployTokens.DeleteGroupDeployToken(groupId, tokenId, g.WithContext(ctx))
	return revokeResult("group deploy", resp, err)
}

func (gc *gitlabClient) RevokeProjectDeployToken(ctx context.Context, tokenId int64, projectId string) (err error) {
	defer func() {
		gc.log.Debug("Revoke project deploy token", logging.Fields{TokenID: tokenId, ProjectID: projectId, Err: err})
	}()
	resp, err := gc.client.DeployTokens.DeleteProjectDeployToken(projectId, tokenId, g.WithContext(ctx))
	return revokeResult("project deploy", resp, err)
}

func (gc *gitlabClient) Metadata(ctx context.Context) (metadata *g.Metadata, err error) {
	defer func() {
		gc.log.Debug("Fetch metadata information", logging.Fields{Err: err, Extra: map[string]any{"metadata": metadata}})
	}()

	metadata, _, err = gc.client.Metadata.GetMetadata(g.WithContext(ctx))
	return metadata, err
}

func (gc *gitlabClient) CreatePipelineProjectTriggerAccessToken(ctx context.Context, path, name string, projectId int64, description string, expiresAt *time.Time) (et *modelToken.TokenPipelineProjectTrigger, err error) {
	var pt *g.PipelineTrigger
	defer func() {
		gc.log.Debug("Create a pipeline project trigger access token", logging.Fields{Path: path, Name: name, ProjectID: fmt.Sprint(projectId), Err: err, Extra: map[string]any{"description": description}})
	}()

	if pt, _, err = gc.client.PipelineTriggers.AddPipelineTrigger(
		projectId,
		&g.AddPipelineTriggerOptions{Description: &description},
		g.WithContext(ctx),
	); err == nil {
		et = &modelToken.TokenPipelineProjectTrigger{
			Token: modelToken.Token{
				TokenID:   pt.ID,
				ParentID:  strconv.FormatInt(projectId, 10),
				Path:      path,
				Name:      name,
				Token:     pt.Token,
				TokenType: t.TypePipelineProjectTrigger,
				CreatedAt: new(time.Now()),
				ExpiresAt: expiresAt,
			},
		}
	}

	return et, err
}

func (gc *gitlabClient) RevokePipelineProjectTriggerAccessToken(ctx context.Context, tokenId int64, projectId string) (err error) {
	defer func() {
		gc.log.Debug("Revoke pipeline project trigger access token", logging.Fields{TokenID: tokenId, ProjectID: projectId, Err: err})
	}()
	resp, err := gc.client.PipelineTriggers.DeletePipelineTrigger(projectId, tokenId, g.WithContext(ctx))
	return revokeResult("pipeline project trigger", resp, err)
}

func (gc *gitlabClient) GetGroupIdByPath(ctx context.Context, path string) (groupId int64, err error) {
	defer func() {
		gc.log.Debug("Get group id by path", logging.Fields{Path: path, GroupID: fmt.Sprint(groupId), Err: err})
	}()

	group, _, err := gc.client.Groups.GetGroup(path, nil, g.WithContext(ctx))
	if err != nil {
		return 0, fmt.Errorf("path '%s' not found: %w", path, errs.ErrInvalidValue)
	}

	groupId = group.ID
	return groupId, nil
}

func (gc *gitlabClient) GitlabClient(_ context.Context) *g.Client {
	return gc.client
}

func (gc *gitlabClient) CreateGroupServiceAccountAccessToken(ctx context.Context, path string, groupId string, userId int64, name, description string, expiresAt time.Time, scopes []string) (et *modelToken.TokenGroupServiceAccount, err error) {
	var at *g.PersonalAccessToken
	defer func() {
		gc.log.With(tokenLogFields(et)).Debug("Create group service access token", logging.Fields{Path: path, GroupID: fmt.Sprint(groupId), UserID: userId, Name: name, ExpiresAt: &expiresAt, Scopes: scopes, Err: err})
	}()
	at, _, err = gc.client.Groups.CreateServiceAccountPersonalAccessToken(groupId, userId, &g.CreateServiceAccountPersonalAccessTokenOptions{
		Name:        new(name),
		Description: new(description),
		ExpiresAt:   (*g.ISOTime)(&expiresAt),
		Scopes:      &scopes,
	}, g.WithContext(ctx))
	if err == nil {
		et = &modelToken.TokenGroupServiceAccount{
			TokenWithScopes: modelToken.TokenWithScopes{
				Token: modelToken.Token{
					TokenID:   at.ID,
					ParentID:  groupId,
					Path:      path,
					Name:      name,
					Token:     at.Token,
					TokenType: t.TypeGroupServiceAccount,
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

func (gc *gitlabClient) CreateProjectServiceAccountAccessToken(ctx context.Context, path string, projectId string, userId int64, name, description string, expiresAt time.Time, scopes []string) (et *modelToken.TokenProjectServiceAccount, err error) {
	var at *g.PersonalAccessToken
	defer func() {
		gc.log.With(tokenLogFields(et)).Debug("Create project service access token", logging.Fields{Path: path, ProjectID: fmt.Sprint(projectId), UserID: userId, Name: name, ExpiresAt: &expiresAt, Scopes: scopes, Err: err})
	}()
	at, _, err = gc.client.Projects.CreateProjectServiceAccountPersonalAccessToken(projectId, userId, &g.CreateProjectServiceAccountPersonalAccessTokenOptions{
		Name:        new(name),
		Description: new(description),
		ExpiresAt:   (*g.ISOTime)(&expiresAt),
		Scopes:      &scopes,
	}, g.WithContext(ctx))
	if err == nil {
		et = &modelToken.TokenProjectServiceAccount{
			TokenWithScopes: modelToken.TokenWithScopes{
				Token: modelToken.Token{
					TokenID:   at.ID,
					ParentID:  projectId,
					Path:      path,
					Name:      name,
					Token:     at.Token,
					TokenType: t.TypeProjectServiceAccount,
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

func (gc *gitlabClient) CreateUserServiceAccountAccessToken(ctx context.Context, username string, userId int64, name, description string, expiresAt time.Time, scopes []string) (et *modelToken.TokenUserServiceAccount, err error) {
	defer func() {
		gc.log.With(tokenLogFields(et)).Debug("Create user service access token", logging.Fields{Username: username, UserID: userId, Name: name, ExpiresAt: &expiresAt, Scopes: scopes, Err: err})
	}()
	var etp *modelToken.TokenPersonal
	etp, err = gc.CreatePersonalAccessToken(ctx, username, userId, name, description, expiresAt, scopes)
	if err == nil && etp != nil {
		et = &modelToken.TokenUserServiceAccount{
			TokenWithScopes: modelToken.TokenWithScopes{
				Token: modelToken.Token{
					TokenID:   etp.TokenID,
					ParentID:  etp.ParentID,
					Path:      etp.Path,
					Name:      etp.Name,
					Token:     etp.Token.Token,
					TokenType: t.TypeUserServiceAccount,
					CreatedAt: etp.CreatedAt,
					ExpiresAt: etp.ExpiresAt,
				},
				Scopes: etp.Scopes,
			},
		}
	}
	return et, err
}

func (gc *gitlabClient) RevokeUserServiceAccountAccessToken(ctx context.Context, tokenId int64) (err error) {
	defer func() {
		gc.log.Debug("Revoke user service account token", logging.Fields{TokenID: tokenId, Err: err})
	}()
	resp, err := gc.client.PersonalAccessTokens.RevokePersonalAccessTokenByID(tokenId, g.WithContext(ctx))
	return revokeResult("user service account", resp, err)
}

func (gc *gitlabClient) RevokeGroupServiceAccountAccessToken(ctx context.Context, tokenId int64, groupId string, userId int64) (err error) {
	defer func() {
		gc.log.Debug("Revoke group service account token", logging.Fields{TokenID: tokenId, GroupID: groupId, UserID: userId, Err: err})
	}()
	resp, err := gc.client.Groups.RevokeServiceAccountPersonalAccessToken(groupId, userId, tokenId, g.WithContext(ctx))
	return revokeResult("group service account", resp, err)
}

func (gc *gitlabClient) RevokeProjectServiceAccountAccessToken(ctx context.Context, tokenId int64, projectId string, userId int64) (err error) {
	defer func() {
		gc.log.Debug("Revoke project service account token", logging.Fields{TokenID: tokenId, ProjectID: projectId, UserID: userId, Err: err})
	}()
	resp, err := gc.client.Projects.RevokeProjectServiceAccountPersonalAccessToken(projectId, userId, tokenId, g.WithContext(ctx))
	return revokeResult("project service account", resp, err)
}

func (gc *gitlabClient) CurrentTokenInfo(ctx context.Context) (et *modelToken.TokenConfig, err error) {
	var pat *g.PersonalAccessToken
	defer func() { gc.log.With(tokenLogFields(et)).Debug("Current token info", logging.Fields{Err: err}) }()
	if pat, _, err = gc.client.PersonalAccessTokens.GetSinglePersonalAccessToken(g.WithContext(ctx)); err == nil {
		et = &modelToken.TokenConfig{
			TokenWithScopes: modelToken.TokenWithScopes{
				Token: modelToken.Token{
					TokenID:   pat.ID,
					Name:      pat.Name,
					Token:     pat.Token,
					TokenType: t.TypePersonal,
					CreatedAt: pat.CreatedAt,
					ExpiresAt: (*time.Time)(pat.ExpiresAt),
				},
				Scopes: pat.Scopes,
			},
			UserID: pat.UserID,
		}
		// GitLab allows tokens without an expiry when it does not enforce one (#178).
		if et.ExpiresAt == nil {
			newDate := pat.CreatedAt.AddDate(1, 0, -2)
			et.ExpiresAt = &newDate
		}
	}
	return et, err
}

func (gc *gitlabClient) RotateCurrentToken(ctx context.Context) (token *modelToken.TokenConfig, currentEntryToken *modelToken.TokenConfig, err error) {
	var expiresAt time.Time
	defer func() {
		gc.log.With(tokenLogFields(currentEntryToken)).Debug("Rotating current token", logging.Fields{ExpiresAt: &expiresAt})
		gc.log.With(tokenLogFields(token)).Debug("Rotated current token", logging.Fields{Err: err})
	}()

	currentEntryToken, err = gc.CurrentTokenInfo(ctx)
	if err != nil {
		return nil, nil, err
	}

	var usr *g.User
	usr, _, err = gc.client.Users.GetUser(currentEntryToken.UserID, &g.GetUserOptions{})
	if err != nil {
		return nil, nil, err
	}

	var pat *g.PersonalAccessToken
	var durationTTL = currentEntryToken.ExpiresAt.Sub(*currentEntryToken.CreatedAt)
	_, expiresAt, _ = utils.CalculateGitlabTTL(durationTTL, utils.TimeFromContext(ctx))
	pat, _, err = gc.client.PersonalAccessTokens.RotatePersonalAccessToken(
		currentEntryToken.TokenID,
		&g.RotatePersonalAccessTokenOptions{ExpiresAt: (*g.ISOTime)(&expiresAt)},
	)

	if err != nil {
		return nil, nil, err
	}

	token = &modelToken.TokenConfig{
		TokenWithScopes: modelToken.TokenWithScopes{
			Token: modelToken.Token{
				TokenID:   pat.ID,
				ParentID:  "",
				Path:      usr.Username,
				Name:      pat.Name,
				Token:     pat.Token,
				TokenType: t.TypePersonal,
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

func (gc *gitlabClient) GetUserIdByUsername(ctx context.Context, username string) (userId int64, err error) {
	defer func() {
		gc.log.Debug("Get user id by username", logging.Fields{Username: username, UserID: userId, Err: err})
	}()

	l := &g.ListUsersOptions{
		Username: new(username),
	}

	u, _, err := gc.client.Users.ListUsers(l, g.WithContext(ctx))
	if err != nil {
		return 0, fmt.Errorf("%w", err)
	}
	if len(u) == 0 {
		return 0, fmt.Errorf("username '%s' not found: %w", username, errs.ErrInvalidValue)
	}
	userId = u[0].ID
	return userId, nil
}

func (gc *gitlabClient) CreatePersonalAccessToken(ctx context.Context, username string, userId int64, name, description string, expiresAt time.Time, scopes []string) (et *modelToken.TokenPersonal, err error) {
	var at *g.PersonalAccessToken
	defer func() {
		gc.log.With(tokenLogFields(et)).Debug("Create personal access token", logging.Fields{Username: username, UserID: userId, Name: name, ExpiresAt: &expiresAt, Scopes: scopes, Err: err})
	}()
	if at, _, err = gc.client.Users.CreatePersonalAccessToken(userId, &g.CreatePersonalAccessTokenOptions{
		Name:        new(name),
		Description: new(description),
		ExpiresAt:   (*g.ISOTime)(&expiresAt),
		Scopes:      &scopes,
	}, g.WithContext(ctx)); err == nil {
		et = &modelToken.TokenPersonal{
			TokenWithScopes: modelToken.TokenWithScopes{
				Token: modelToken.Token{
					TokenID:   at.ID,
					Path:      username,
					Name:      name,
					Token:     at.Token,
					TokenType: t.TypePersonal,
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

func (gc *gitlabClient) CreateGroupAccessToken(ctx context.Context, groupId string, name, description string, expiresAt time.Time, scopes []string, accessLevel t.AccessLevel) (et *modelToken.TokenGroup, err error) {
	var at *g.GroupAccessToken
	defer func() {
		gc.log.With(tokenLogFields(et)).Debug("Create group access token", logging.Fields{GroupID: fmt.Sprint(groupId), Name: name, ExpiresAt: &expiresAt, Scopes: scopes, AccessLevel: accessLevel.String(), Err: err})
	}()
	var al = new(g.AccessLevelValue)
	*al = g.AccessLevelValue(accessLevel.Value())
	if at, _, err = gc.client.GroupAccessTokens.CreateGroupAccessToken(groupId, &g.CreateGroupAccessTokenOptions{
		Name:        new(name),
		Description: new(description),
		Scopes:      &scopes,
		ExpiresAt:   (*g.ISOTime)(&expiresAt),
		AccessLevel: al,
	}, g.WithContext(ctx)); err == nil {
		et = &modelToken.TokenGroup{
			TokenWithScopesAndAccessLevel: modelToken.TokenWithScopesAndAccessLevel{
				Token: modelToken.Token{
					TokenID:   at.ID,
					ParentID:  groupId,
					Path:      groupId,
					Name:      name,
					Token:     at.Token,
					TokenType: t.TypeGroup,
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

func (gc *gitlabClient) CreateProjectAccessToken(ctx context.Context, projectId string, name, description string, expiresAt time.Time, scopes []string, accessLevel t.AccessLevel) (et *modelToken.TokenProject, err error) {
	var at *g.ProjectAccessToken
	defer func() {
		gc.log.With(tokenLogFields(et)).Debug("Create project access token", logging.Fields{ProjectID: fmt.Sprint(projectId), Name: name, ExpiresAt: &expiresAt, Scopes: scopes, AccessLevel: accessLevel.String(), Err: err})
	}()
	var al = new(g.AccessLevelValue)
	*al = g.AccessLevelValue(accessLevel.Value())
	if at, _, err = gc.client.ProjectAccessTokens.CreateProjectAccessToken(projectId, &g.CreateProjectAccessTokenOptions{
		Name:        new(name),
		Description: new(description),
		Scopes:      &scopes,
		ExpiresAt:   (*g.ISOTime)(&expiresAt),
		AccessLevel: al,
	}, g.WithContext(ctx)); err == nil {
		et = &modelToken.TokenProject{
			TokenWithScopesAndAccessLevel: modelToken.TokenWithScopesAndAccessLevel{
				Token: modelToken.Token{
					TokenID:   at.ID,
					ParentID:  projectId,
					Path:      projectId,
					Name:      name,
					Token:     at.Token,
					TokenType: t.TypeProject,
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

func (gc *gitlabClient) RevokePersonalAccessToken(ctx context.Context, tokenId int64) (err error) {
	defer func() {
		gc.log.Debug("Revoke personal access token", logging.Fields{TokenID: tokenId, Err: err})
	}()
	resp, err := gc.client.PersonalAccessTokens.RevokePersonalAccessTokenByID(tokenId, g.WithContext(ctx))
	return revokeResult("personal", resp, err)
}

func (gc *gitlabClient) RevokeProjectAccessToken(ctx context.Context, tokenId int64, projectId string) (err error) {
	defer func() {
		gc.log.Debug("Revoke project access token", logging.Fields{TokenID: tokenId, ProjectID: projectId, Err: err})
	}()
	resp, err := gc.client.ProjectAccessTokens.RevokeProjectAccessToken(projectId, tokenId, g.WithContext(ctx))
	return revokeResult("project", resp, err)
}

func (gc *gitlabClient) RevokeGroupAccessToken(ctx context.Context, tokenId int64, groupId string) (err error) {
	defer func() {
		gc.log.Debug("Revoke group access token", logging.Fields{TokenID: tokenId, GroupID: groupId, Err: err})
	}()
	resp, err := gc.client.GroupAccessTokens.RevokeGroupAccessToken(groupId, tokenId, g.WithContext(ctx))
	return revokeResult("group", resp, err)
}

func (gc *gitlabClient) Valid(_ context.Context) bool {
	return gc.client != nil && gc.config != nil
}

var _ Client = new(gitlabClient)

func newGitlabClient(config *modelConfig.EntryConfig, httpClient *http.Client) (gc *g.Client, err error) {
	if strings.TrimSpace(config.BaseURL) == "" {
		err = errors.Join(err, fmt.Errorf("gitlab base url: %w", errs.ErrInvalidValue))
	}

	if strings.TrimSpace(config.Token) == "" {
		err = errors.Join(err, fmt.Errorf("gitlab token: %w", errs.ErrInvalidValue))
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

func NewGitlabClient(config *modelConfig.EntryConfig, httpClient *http.Client, logger hclog.Logger) (client Client, err error) {
	if config == nil {
		return nil, fmt.Errorf("configure the backend first, config: %w", errs.ErrNilValue)
	}

	if logger == nil {
		logger = sdklogging.NewVaultLoggerWithWriter(io.Discard, hclog.NoLevel)
	}

	var gc *g.Client
	if gc, err = newGitlabClient(config, httpClient); err != nil {
		return nil, err
	}

	return &gitlabClient{client: gc, config: config, log: logging.New(logger), httpClient: httpClient}, err
}
