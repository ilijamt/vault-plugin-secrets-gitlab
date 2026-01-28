package gitlab

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/locksutil"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
	role2 "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/role"
	t "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
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
		"path": {
			Type:        framework.TypeString,
			Description: "Overwrites the role path, only available if the role has dynamic-path set to true",
			Required:    false,
		},
	}
)

func (b *Backend) pathTokenRoleCreate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	var resp *logical.Response
	var err error
	var role *role2.Role
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

	// The regexp is always valid, as it is checked during role creation.
	// We only need to validate that the path is correct and matches the regexp.
	// If DynamicPath is false, the path is already validated during role creation,
	// so no additional path validation is required here.
	if role.DynamicPath {
		rx, _ := regexp.Compile(role.Path)
		rolePath := data.Get("path").(string)
		if !t.IsValidPath(rolePath, role.TokenType) {
			return logical.ErrorResponse("invalid path"), fmt.Errorf("path '%s' is not valid for token type %s: %w", rolePath, role.TokenType, errs.ErrInvalidValue)
		}
		if !rx.MatchString(rolePath) {
			return logical.ErrorResponse("path doesn't match regex"), fmt.Errorf("regexp (%s) with path '%s': %w", role.Path, rolePath, errs.ErrInvalidValue)
		}
		role.Path = rolePath
	}

	b.Logger().Debug("Creating token for role", "role_name", roleName, "token_type", role.TokenType.String())
	defer b.Logger().Debug("Created token for role", "role_name", roleName, "token_type", role.TokenType.String())

	var name string
	var token t.Token
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
	case t.TypeGroup:
		b.Logger().Debug("Creating group access token for role", "path", role.Path, "name", name, "expiresAt", expiresAt, "scopes", role.Scopes, "accessLevel", role.AccessLevel)
		token, err = client.CreateGroupAccessToken(ctx, role.Path, name, expiresAt, role.Scopes, role.AccessLevel)
	case t.TypeProject:
		b.Logger().Debug("Creating project access token for role", "path", role.Path, "name", name, "expiresAt", expiresAt, "scopes", role.Scopes, "accessLevel", role.AccessLevel)
		token, err = client.CreateProjectAccessToken(ctx, role.Path, name, expiresAt, role.Scopes, role.AccessLevel)
	case t.TypePersonal:
		var userId int64
		userId, err = client.GetUserIdByUsername(ctx, role.Path)
		if err == nil {
			b.Logger().Debug("Creating personal access token for role", "path", role.Path, "userId", userId, "name", name, "expiresAt", expiresAt, "scopes", role.Scopes)
			token, err = client.CreatePersonalAccessToken(ctx, role.Path, userId, name, expiresAt, role.Scopes)
		}
	case t.TypeUserServiceAccount:
		var userId int64
		if userId, err = client.GetUserIdByUsername(ctx, role.Path); err == nil {
			b.Logger().Debug("Creating user service account access token for role", "path", role.Path, "userId", userId, "name", name, "expiresAt", expiresAt, "scopes", role.Scopes)
			token, err = client.CreateUserServiceAccountAccessToken(ctx, role.Path, userId, name, expiresAt, role.Scopes)
		}
	case t.TypeGroupServiceAccount:
		var serviceAccount, groupId string
		{
			parts := strings.Split(role.Path, "/")
			groupId, serviceAccount = parts[0], parts[1]
		}

		var userId int64
		if userId, err = client.GetUserIdByUsername(ctx, serviceAccount); err == nil {
			b.Logger().Debug("Creating group service account access token for role", "path", role.Path, "groupId", groupId, "userId", userId, "name", name, "expiresAt", expiresAt, "scopes", role.Scopes)
			token, err = client.CreateGroupServiceAccountAccessToken(ctx, role.Path, groupId, userId, name, expiresAt, role.Scopes)
		}
	case t.TypeProjectDeploy:
		var projectId int64
		if projectId, err = client.GetProjectIdByPath(ctx, role.Path); err == nil {
			token, err = client.CreateProjectDeployToken(ctx, role.Path, projectId, name, &expiresAt, role.Scopes)
		}
	case t.TypeGroupDeploy:
		var groupId int64
		if groupId, err = client.GetGroupIdByPath(ctx, role.Path); err == nil {
			token, err = client.CreateGroupDeployToken(ctx, role.Path, groupId, name, &expiresAt, role.Scopes)
		}
	case t.TypePipelineProjectTrigger:
		var projectId int64
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
		// since vault is controlling the expiry, we need to override here
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
		ctx, b.Backend, "token-write",
		token.Event(map[string]string{"path": fmt.Sprintf("%s/%s", PathRoleStorage, roleName)}),
	)
	return resp, nil
}

func pathTokenRoles(b *Backend) *framework.Path {
	return &framework.Path{
		HelpSynopsis:    strings.TrimSpace(pathTokenRolesHelpSyn),
		HelpDescription: strings.TrimSpace(pathTokenRolesHelpDesc),
		Pattern:         fmt.Sprintf("%s/%s%s", PathTokenRoleStorage, framework.GenericNameRegex("role_name"), framework.OptionalParamRegex("path")),
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
