package gitlab

import (
	"context"
	"time"

	g "gitlab.com/gitlab-org/api/client-go"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/token"
	t "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

type Client interface {
	GitlabClient(ctx context.Context) *g.Client
	Valid(ctx context.Context) bool
	Metadata(ctx context.Context) (*g.Metadata, error)
	CurrentTokenInfo(ctx context.Context) (*token.TokenConfig, error)
	RotateCurrentToken(ctx context.Context) (newToken *token.TokenConfig, oldToken *token.TokenConfig, err error)
	CreatePersonalAccessToken(ctx context.Context, username string, userId int64, name string, expiresAt time.Time, scopes []string) (*token.TokenPersonal, error)
	CreateGroupAccessToken(ctx context.Context, groupId string, name string, expiresAt time.Time, scopes []string, accessLevel t.AccessLevel) (*token.TokenGroup, error)
	CreateProjectAccessToken(ctx context.Context, projectId string, name string, expiresAt time.Time, scopes []string, accessLevel t.AccessLevel) (*token.TokenProject, error)
	RevokePersonalAccessToken(ctx context.Context, tokenId int64) error
	RevokeProjectAccessToken(ctx context.Context, tokenId int64, projectId string) error
	RevokeGroupAccessToken(ctx context.Context, tokenId int64, groupId string) error
	GetUserIdByUsername(ctx context.Context, username string) (int64, error)
	GetGroupIdByPath(ctx context.Context, path string) (int64, error)
	GetProjectIdByPath(ctx context.Context, path string) (int64, error)
	CreateGroupServiceAccountAccessToken(ctx context.Context, group string, groupId string, userId int64, name string, expiresAt time.Time, scopes []string) (*token.TokenGroupServiceAccount, error)
	CreateUserServiceAccountAccessToken(ctx context.Context, username string, userId int64, name string, expiresAt time.Time, scopes []string) (*token.TokenUserServiceAccount, error)
	RevokeUserServiceAccountAccessToken(ctx context.Context, token string) error
	RevokeGroupServiceAccountAccessToken(ctx context.Context, token string) error
	CreatePipelineProjectTriggerAccessToken(ctx context.Context, path, name string, projectId int64, description string, expiresAt *time.Time) (*token.TokenPipelineProjectTrigger, error)
	RevokePipelineProjectTriggerAccessToken(ctx context.Context, projectId int64, tokenId int64) error
	CreateProjectDeployToken(ctx context.Context, path string, projectId int64, name string, expiresAt *time.Time, scopes []string) (et *token.TokenProjectDeploy, err error)
	RevokeProjectDeployToken(ctx context.Context, projectId, deployTokenId int64) (err error)
	CreateGroupDeployToken(ctx context.Context, path string, groupId int64, name string, expiresAt *time.Time, scopes []string) (et *token.TokenGroupDeploy, err error)
	RevokeGroupDeployToken(ctx context.Context, groupId, deployTokenId int64) (err error)
}
