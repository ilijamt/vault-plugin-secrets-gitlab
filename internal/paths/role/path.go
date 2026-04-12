package role

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

const (
	pathRolesHelpSyn  = `Create a role with parameters that are used to generate a various access tokens.`
	pathRolesHelpDesc = `
This path allows you to create a role with predefined parameters that will be used to generate tokens for different 
access types in GitLab. The role defines the configuration for generating project, group, personal access tokens,
user service accounts, or group service accounts.`
)

var (
	FieldSchemaRoles = map[string]*framework.FieldSchema{
		"role_name": {
			Type:        framework.TypeString,
			Description: "Role name",
			Required:    true,
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "Role Name",
			},
		},
		"path": {
			Type:        framework.TypeString,
			Description: "Project/Group path to create an access token for. If the token type is set to personal then write the username here. If dynamic_path is true set then this is regex.",
			Required:    true,
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "path",
			},
		},
		"name": {
			Type:        framework.TypeString,
			Description: "The name of the access token",
			Required:    true,
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "Name",
			},
		},
		"scopes": {
			Type:        framework.TypeCommaStringSlice,
			Description: "List of scopes",
			Required:    false,
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "Scopes",
			},
			AllowedValues: utils.ToAny(token.ValidPersonalTokenScopes...),
		},
		"ttl": {
			Type:        framework.TypeDurationSecond,
			Description: "The TTL of the token",
			Required:    false,
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "Token TTL",
			},
		},
		"access_level": {
			Type:        framework.TypeString,
			Description: "access level of access token (only required for Group and Project access tokens)",
			Required:    false,
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "Access Level",
			},
			AllowedValues: utils.ToAny(token.ValidAccessLevels...),
		},
		"token_type": {
			Type:          framework.TypeString,
			Description:   "access token type",
			Required:      true,
			AllowedValues: utils.ToAny(token.ValidTokenTypes...),
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "Token Type",
			},
		},
		"gitlab_revokes_token": {
			Type:        framework.TypeBool,
			Default:     false,
			Required:    false,
			Description: `Gitlab revokes the token when it's time. Vault will not revoke the token when the lease expires.`,
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "Gitlab revokes token.",
			},
		},
		"config_name": {
			Type:        framework.TypeString,
			Default:     backend.DefaultConfigName,
			Required:    false,
			Description: "The config we use when interacting with the role, this can be specified if you want to use a specific config for the role, otherwise it uses the default one.",
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "Configuration.",
			},
		},
		"dynamic_path": {
			Type:        framework.TypeBool,
			Default:     false,
			Required:    false,
			Description: "Should path be changeable dynamically for this role?",
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "Dynamic Path.",
			},
		},
	}
)

// roleBackend defines the narrow interface this provider needs.
type roleBackend interface {
	backend.Logging
	backend.RoleLocker
	backend.RoleStore
	backend.ConfigStore
	backend.EventSender
}

// Provider implements backend.PathProvider for role paths.
type Provider struct {
	b roleBackend
}

func (p *Provider) Name() string { return "role" }

// New creates a new role path provider.
func New(b roleBackend) *Provider {
	return &Provider{b: b}
}

// Paths returns all role-related framework paths.
func (p *Provider) Paths() []*framework.Path {
	return []*framework.Path{
		p.pathListRoles(),
		p.pathRoles(),
	}
}

func (p *Provider) pathRoles() *framework.Path {
	return &framework.Path{
		HelpSynopsis:    strings.TrimSpace(pathRolesHelpSyn),
		HelpDescription: strings.TrimSpace(pathRolesHelpDesc),
		Pattern:         fmt.Sprintf("%s/%s", backend.PathRoleStorage, framework.GenericNameRegex("role_name")),
		Fields:          FieldSchemaRoles,
		DisplayAttrs: &framework.DisplayAttributes{
			OperationPrefix: paths.OperationPrefixGitlabAccessTokens,
			OperationSuffix: "role",
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.DeleteOperation: &framework.PathOperation{
				Callback: p.pathRolesDelete,
				Summary:  "Deletes a role",
				Responses: map[int][]framework.Response{
					http.StatusNoContent: {{
						Description: http.StatusText(http.StatusNoContent),
					}},
				},
			},
			logical.CreateOperation: &framework.PathOperation{
				Callback: p.pathRolesWrite,
				Summary:  "Creates a new role",
				Responses: map[int][]framework.Response{
					http.StatusNoContent: {{
						Description: http.StatusText(http.StatusNoContent),
					}},
				},
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: p.pathRolesWrite,
				Summary:  "Updates an existing role",
				Responses: map[int][]framework.Response{
					http.StatusNoContent: {{
						Description: http.StatusText(http.StatusNoContent),
					}},
				},
			},
			logical.ReadOperation: &framework.PathOperation{
				Callback: p.pathRolesRead,
				Summary:  "Reads an existing role",
				Responses: map[int][]framework.Response{
					http.StatusNotFound: {{
						Description: http.StatusText(http.StatusNotFound),
					}},
					http.StatusOK: {{
						Fields: FieldSchemaRoles,
					}},
				},
			},
		},
		ExistenceCheck: p.pathRoleExistenceCheck,
	}
}

func (p *Provider) pathRoleExistenceCheck(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
	name := data.Get("role_name").(string)
	role, err := p.b.GetRole(ctx, req.Storage, name)
	if err != nil {
		if strings.Contains(err.Error(), logical.ErrReadOnly.Error()) {
			return false, nil
		}

		return false, fmt.Errorf("error reading role: %w", err)
	}

	return role != nil, nil
}
