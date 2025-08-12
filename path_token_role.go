package gitlab

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/locksutil"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
	token2 "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

const (
	pathTokenRolesHelpSyn  = `Generate an access token based on the specified role`
	pathTokenRolesHelpDesc = `
This path allows you to generate an access token based on a predefined role. The role must be created beforehand in 
the ^roles/(?P<role_name>\w(([\w-.@]+)?\w)?)$ path, where its parameters, such as token permissions, scopes, and 
expiration, are defined. When you request an access token through this path, Vault will use the predefined 
role's parameters to create a new access token.`

	PathTokenRoleStorage = "token"
)

var (
	FieldSchemaTokenRole = map[string]*framework.FieldSchema{
		"role_name": {
			Type:        framework.TypeString,
			Description: "Role name",
			Required:    true,
		},
	}
)

func (b *Backend) pathTokenRoleCreate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	var resp *logical.Response
	var err error
	var role *EntryRole
	var roleName = data.Get("role_name").(string)

	lock := locksutil.LockForKey(b.roleLocks, roleName)
	lock.RLock()
	defer lock.RUnlock()

	role, err = getRole(ctx, roleName, req.Storage)
	if err != nil {
		return nil, fmt.Errorf("error getting role: %w", err)
	}
	if role == nil {
		return nil, fmt.Errorf("%s: %w", roleName, ErrRoleNotFound)
	}

	b.Logger().Debug("Creating token for role", "role_name", roleName, "token_type", role.TokenType.String())
	defer b.Logger().Debug("Created token for role", "role_name", roleName, "token_type", role.TokenType.String())

	var name string
	var token token2.Token
	var expiresAt time.Time
	var startTime = utils.TimeFromContext(ctx).UTC()

	name, err = utils.TokenName(role)
	if err != nil {
		return nil, fmt.Errorf("error generating token name: %w", err)
	}

	var client Client
	var gitlabRevokesTokens = role.GitlabRevokesTokens
	var vaultRevokesTokens = !role.GitlabRevokesTokens

	_, expiresAt, _ = utils.CalculateGitlabTTL(role.TTL, startTime)

	client, err = b.getClient(ctx, req.Storage, role.ConfigName)
	if err != nil {
		return nil, err
	}

	switch role.TokenType {
	case token2.TypeGroup:
		b.Logger().Debug("Creating group access token for role", "path", role.Path, "name", name, "expiresAt", expiresAt, "scopes", role.Scopes, "accessLevel", role.AccessLevel)
		token, err = client.CreateGroupAccessToken(ctx, role.Path, name, expiresAt, role.Scopes, role.AccessLevel)
	case token2.TypeProject:
		b.Logger().Debug("Creating project access token for role", "path", role.Path, "name", name, "expiresAt", expiresAt, "scopes", role.Scopes, "accessLevel", role.AccessLevel)
		token, err = client.CreateProjectAccessToken(ctx, role.Path, name, expiresAt, role.Scopes, role.AccessLevel)
	case token2.TypePersonal:
		var userId int
		userId, err = client.GetUserIdByUsername(ctx, role.Path)
		if err == nil {
			b.Logger().Debug("Creating personal access token for role", "path", role.Path, "userId", userId, "name", name, "expiresAt", expiresAt, "scopes", role.Scopes)
			token, err = client.CreatePersonalAccessToken(ctx, role.Path, userId, name, expiresAt, role.Scopes)
		}
	case token2.TypeUserServiceAccount:
		var userId int
		if userId, err = client.GetUserIdByUsername(ctx, role.Path); err == nil {
			b.Logger().Debug("Creating user service account access token for role", "path", role.Path, "userId", userId, "name", name, "expiresAt", expiresAt, "scopes", role.Scopes)
			token, err = client.CreateUserServiceAccountAccessToken(ctx, role.Path, userId, name, expiresAt, role.Scopes)
		}
	case token2.TypeGroupServiceAccount:
		var serviceAccount, groupId string
		{
			parts := strings.Split(role.Path, "/")
			groupId, serviceAccount = parts[0], parts[1]
		}

		var userId int
		if userId, err = client.GetUserIdByUsername(ctx, serviceAccount); err == nil {
			b.Logger().Debug("Creating group service account access token for role", "path", role.Path, "groupId", groupId, "userId", userId, "name", name, "expiresAt", expiresAt, "scopes", role.Scopes)
			token, err = client.CreateGroupServiceAccountAccessToken(ctx, role.Path, groupId, userId, name, expiresAt, role.Scopes)
		}
	case token2.TypeProjectDeploy:
		var projectId int
		if projectId, err = client.GetProjectIdByPath(ctx, role.Path); err == nil {
			token, err = client.CreateProjectDeployToken(ctx, role.Path, projectId, name, &expiresAt, role.Scopes)
		}
	case token2.TypeGroupDeploy:
		var groupId int
		if groupId, err = client.GetGroupIdByPath(ctx, role.Path); err == nil {
			token, err = client.CreateGroupDeployToken(ctx, role.Path, groupId, name, &expiresAt, role.Scopes)
		}
	case token2.TypePipelineProjectTrigger:
		var projectId int
		if projectId, err = client.GetProjectIdByPath(ctx, role.Path); err == nil {
			token, err = client.CreatePipelineProjectTriggerAccessToken(ctx, role.Path, name, projectId, name, &expiresAt)
		}
	default:
		return logical.ErrorResponse("invalid token type"), fmt.Errorf("%s: %w", role.TokenType.String(), errs.ErrUnknownTokenType)
	}

	if err != nil || token == nil {
		return nil, cmp.Or(err, fmt.Errorf("%w: token is nil", errs.ErrNilValue))
	}

	token.SetConfigName(cmp.Or(role.ConfigName, DefaultConfigName))
	token.SetRoleName(role.RoleName)
	token.SetGitlabRevokesToken(role.GitlabRevokesTokens)

	if vaultRevokesTokens {
		// since vault is controlling the expiry we need to override here
		// and make the expiry time accurate
		expiresAt = startTime.Add(role.TTL)
		token.SetExpiresAt(&expiresAt)
	}

	resp = b.Secret(SecretAccessTokenType).Response(token.Data(), token.Internal())

	resp.Secret.MaxTTL = role.TTL
	resp.Secret.TTL = role.TTL
	resp.Secret.IssueTime = startTime
	if gitlabRevokesTokens {
		resp.Secret.TTL = token.TTL()
	}

	_ = event.Event(
		ctx, b.Backend, operationPrefixGitlabAccessTokens, "token-write",
		token.Event(map[string]string{"path": fmt.Sprintf("%s/%s", PathRoleStorage, roleName)}),
	)
	return resp, nil
}

func pathTokenRoles(b *Backend) *framework.Path {
	return &framework.Path{
		HelpSynopsis:    strings.TrimSpace(pathTokenRolesHelpSyn),
		HelpDescription: strings.TrimSpace(pathTokenRolesHelpDesc),
		Pattern:         fmt.Sprintf("%s/%s", PathTokenRoleStorage, framework.GenericNameRegex("role_name")),
		Fields:          FieldSchemaTokenRole,
		DisplayAttrs: &framework.DisplayAttributes{
			OperationPrefix: operationPrefixGitlabAccessTokens,
			OperationSuffix: "generate",
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathTokenRoleCreate,
				Summary:  "Create an access token based on a predefined role",
				DisplayAttrs: &framework.DisplayAttributes{
					OperationVerb:   "generate",
					OperationSuffix: "credentials",
				},
				Responses: map[int][]framework.Response{
					http.StatusOK: {{
						Description: http.StatusText(http.StatusOK),
						Fields:      fieldSchemaAccessTokens,
					}},
				},
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathTokenRoleCreate,
				Summary:  "Create an access token based on a predefined role",
				DisplayAttrs: &framework.DisplayAttributes{
					OperationSuffix: "credentials-with-parameters",
					OperationVerb:   "generate-with-parameters",
				},
				Responses: map[int][]framework.Response{
					http.StatusOK: {{
						Description: http.StatusText(http.StatusOK),
						Fields:      fieldSchemaAccessTokens,
					}},
				},
			},
		},
	}
}
