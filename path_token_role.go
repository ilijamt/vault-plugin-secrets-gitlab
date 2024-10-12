package gitlab

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/locksutil"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	pathTokenRolesHelpSyn  = `Generate an access token based on the specified role`
	pathTokenRolesHelpDesc = `This path allows you to generate an access token based on a predefined role. You must create a role beforehand in /roles/ path,
whose parameters are used to generate an access token based on a predefined role.`

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
	var token *EntryToken
	var expiresAt time.Time
	var startTime = time.Now().UTC()

	name, err = TokenName(role)
	if err != nil {
		return nil, fmt.Errorf("error generating token name: %w", err)
	}

	var client Client
	var gitlabRevokesTokens = role.GitlabRevokesTokens
	var vaultRevokesTokens = !role.GitlabRevokesTokens

	_, expiresAt, _ = calculateGitlabTTL(role.TTL, startTime)

	client, err = b.getClient(ctx, req.Storage, role.ConfigName)
	if err != nil {
		return nil, err
	}

	switch role.TokenType {
	case TokenTypeGroup:
		b.Logger().Debug("Creating group access token for role", "path", role.Path, "name", name, "expiresAt", expiresAt, "scopes", role.Scopes, "accessLevel", role.AccessLevel)
		token, err = client.CreateGroupAccessToken(role.Path, name, expiresAt, role.Scopes, role.AccessLevel)
	case TokenTypeProject:
		b.Logger().Debug("Creating project access token for role", "path", role.Path, "name", name, "expiresAt", expiresAt, "scopes", role.Scopes, "accessLevel", role.AccessLevel)
		token, err = client.CreateProjectAccessToken(role.Path, name, expiresAt, role.Scopes, role.AccessLevel)
	case TokenTypePersonal:
		var userId int
		userId, err = client.GetUserIdByUsername(role.Path)
		if err == nil {
			b.Logger().Debug("Creating personal access token for role", "path", role.Path, "userId", userId, "name", name, "expiresAt", expiresAt, "scopes", role.Scopes)
			token, err = client.CreatePersonalAccessToken(role.Path, userId, name, expiresAt, role.Scopes)
		}
	case TokenTypeUserServiceAccount:
		var userId int
		if userId, err = client.GetUserIdByUsername(role.Path); err == nil {
			b.Logger().Debug("Creating user service account access token for role", "path", role.Path, "userId", userId, "name", name, "expiresAt", expiresAt, "scopes", role.Scopes)
			token, err = client.CreateUserServiceAccountAccessToken(role.Path, userId, name, expiresAt, role.Scopes)
		}
	case TokenTypeGroupServiceAccount:
		var serviceAccount, groupId string
		{
			parts := strings.Split(role.Path, "/")
			groupId, serviceAccount = parts[0], parts[1]
		}

		var userId int
		if userId, err = client.GetUserIdByUsername(serviceAccount); err == nil {
			b.Logger().Debug("Creating group service account access token for role", "path", role.Path, "groupId", groupId, "userId", userId, "name", name, "expiresAt", expiresAt, "scopes", role.Scopes)
			token, err = client.CreateGroupServiceAccountAccessToken(role.Path, groupId, userId, name, expiresAt, role.Scopes)
		}
	default:
		return logical.ErrorResponse("invalid token type"), fmt.Errorf("%s: %w", role.TokenType.String(), ErrUnknownTokenType)
	}

	if err != nil || token == nil {
		return nil, cmp.Or(err, fmt.Errorf("%w: token is nil", ErrNilValue))
	}

	token.ConfigName = cmp.Or(role.ConfigName, DefaultConfigName)
	token.RoleName = role.RoleName
	token.GitlabRevokesToken = role.GitlabRevokesTokens

	if vaultRevokesTokens {
		// since vault is controlling the expiry we need to override here
		// and make the expiry time accurate
		expiresAt = startTime.Add(role.TTL)
		token.ExpiresAt = &expiresAt
	}

	var secretData, secretInternal = token.SecretResponse()
	resp = b.Secret(SecretAccessTokenType).Response(secretData, secretInternal)

	resp.Secret.MaxTTL = role.TTL
	resp.Secret.TTL = role.TTL
	resp.Secret.IssueTime = startTime
	if gitlabRevokesTokens {
		resp.Secret.TTL = token.ExpiresAt.Sub(*token.CreatedAt)
	}

	event(ctx, b.Backend, "token-write", map[string]string{
		"path":         fmt.Sprintf("%s/%s", PathRoleStorage, roleName),
		"name":         name,
		"parent_id":    role.Path,
		"ttl":          resp.Secret.TTL.String(),
		"role_name":    roleName,
		"token_id":     strconv.Itoa(token.TokenID),
		"token_type":   role.TokenType.String(),
		"scopes":       strings.Join(role.Scopes, ","),
		"access_level": role.AccessLevel.String(),
	})
	return resp, nil
}

func pathTokenRoles(b *Backend) *framework.Path {
	return &framework.Path{
		HelpSynopsis:    strings.TrimSpace(pathTokenRolesHelpSyn),
		HelpDescription: strings.TrimSpace(pathTokenRolesHelpDesc),
		Pattern:         fmt.Sprintf("%s/%s", PathTokenRoleStorage, framework.GenericNameWithAtRegex("role_name")),
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
