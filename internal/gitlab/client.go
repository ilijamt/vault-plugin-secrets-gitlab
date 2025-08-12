package gitlab

import (
	"context"
	"time"

	g "gitlab.com/gitlab-org/api/client-go"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/models"
	t "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

type Client interface {
	GitlabClient(ctx context.Context) *g.Client
	Valid(ctx context.Context) bool
	Metadata(ctx context.Context) (*g.Metadata, error)
	CurrentTokenInfo(ctx context.Context) (*models.TokenConfig, error)
	RotateCurrentToken(ctx context.Context) (newToken *models.TokenConfig, oldToken *models.TokenConfig, err error)
	CreatePersonalAccessToken(ctx context.Context, username string, userId int, name string, expiresAt time.Time, scopes []string) (*models.TokenPersonal, error)
	CreateGroupAccessToken(ctx context.Context, groupId string, name string, expiresAt time.Time, scopes []string, accessLevel t.AccessLevel) (*models.TokenGroup, error)
	CreateProjectAccessToken(ctx context.Context, projectId string, name string, expiresAt time.Time, scopes []string, accessLevel t.AccessLevel) (*models.TokenProject, error)
	RevokePersonalAccessToken(ctx context.Context, tokenId int) error
	RevokeProjectAccessToken(ctx context.Context, tokenId int, projectId string) error
	RevokeGroupAccessToken(ctx context.Context, tokenId int, groupId string) error
	GetUserIdByUsername(ctx context.Context, username string) (int, error)
	GetGroupIdByPath(ctx context.Context, path string) (int, error)
	GetProjectIdByPath(ctx context.Context, path string) (int, error)
	CreateGroupServiceAccountAccessToken(ctx context.Context, group string, groupId string, userId int, name string, expiresAt time.Time, scopes []string) (*models.TokenGroupServiceAccount, error)
	CreateUserServiceAccountAccessToken(ctx context.Context, username string, userId int, name string, expiresAt time.Time, scopes []string) (*models.TokenUserServiceAccount, error)
	RevokeUserServiceAccountAccessToken(ctx context.Context, token string) error
	RevokeGroupServiceAccountAccessToken(ctx context.Context, token string) error
	CreatePipelineProjectTriggerAccessToken(ctx context.Context, path, name string, projectId int, description string, expiresAt *time.Time) (*models.TokenPipelineProjectTrigger, error)
	RevokePipelineProjectTriggerAccessToken(ctx context.Context, projectId int, tokenId int) error
	CreateProjectDeployToken(ctx context.Context, path string, projectId int, name string, expiresAt *time.Time, scopes []string) (et *models.TokenProjectDeploy, err error)
	RevokeProjectDeployToken(ctx context.Context, projectId, deployTokenId int) (err error)
	CreateGroupDeployToken(ctx context.Context, path string, groupId int, name string, expiresAt *time.Time, scopes []string) (et *models.TokenGroupDeploy, err error)
	RevokeGroupDeployToken(ctx context.Context, groupId, deployTokenId int) (err error)
}
