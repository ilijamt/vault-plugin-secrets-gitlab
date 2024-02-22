package gitlab

import (
	"context"
	"crypto/rand"
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
	pathTokenRolesHelpSyn  = ``
	pathTokenRolesHelpDesc = ``

	PathTokenRoleStorage = "token"
)

var (
	fieldSchemaTokenRole = map[string]*framework.FieldSchema{
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
	var role *entryRole
	var roleName string

	if roleName = data.Get("role_name").(string); roleName == "" {
		return logical.ErrorResponse("missing role name"), nil
	}

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

	buf := make([]byte, 4)
	_, _ = rand.Read(buf)
	var token *EntryToken
	var name = strings.ToLower(fmt.Sprintf("vault-generated-%s-access-token-%x", role.TokenType.String(), buf))
	var expiresAt time.Time
	var startTime = time.Now().UTC()

	var client Client
	var gitlabRevokesTokens = role.GitlabRevokesTokens
	var vaultRevokesTokens = !role.GitlabRevokesTokens

	expiresAt = startTime.Add(role.TTL)
	if gitlabRevokesTokens {
		_, expiresAt, _ = calculateGitlabTTL(role.TTL, startTime)
	}

	client, err = b.getClient(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	switch role.TokenType {
	case TokenTypeGroup:
		if token, err = client.CreateGroupAccessToken(role.Path, name, expiresAt, role.Scopes, role.AccessLevel); err != nil {
			return nil, err
		}
	case TokenTypeProject:
		if token, err = client.CreateProjectAccessToken(role.Path, name, expiresAt, role.Scopes, role.AccessLevel); err != nil {
			return nil, err
		}
	case TokenTypePersonal:
		var userId int
		userId, err = client.GetUserIdByUsername(role.Path)
		if err != nil {
			return nil, err
		}
		if token, err = client.CreatePersonalAccessToken(role.Path, userId, name, expiresAt, role.Scopes); err != nil {
			return nil, err
		}
	default:
		return logical.ErrorResponse("invalid token type"), fmt.Errorf("%s: %w", role.TokenType.String(), ErrUnknownTokenType)
	}

	token.RoleName = role.RoleName
	token.GitlabRevokesToken = role.GitlabRevokesTokens

	if vaultRevokesTokens {
		token.ExpiresAt = &expiresAt
	}

	var secretData, secretInternal = token.SecretResponse()
	resp = b.Secret(secretAccessTokenType).Response(secretData, secretInternal)

	resp.Secret.MaxTTL = role.TTL
	resp.Secret.TTL = role.TTL
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
		Fields:          fieldSchemaTokenRole,
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
