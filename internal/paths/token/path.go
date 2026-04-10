package token

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/secret"
)

const (
	PathTokenRoleStorage = "token"

	pathTokenRolesHelpSyn  = `Generate an access token based on the specified role`
	pathTokenRolesHelpDesc = `
This path allows you to generate an access token based on a predefined role. The role must be created beforehand in 
the ^roles/(?P<role_name>\w(([\w-.@]+)?\w)?)$ path, where its parameters, such as token permissions, scopes, and 
expiration, are defined. When you request an access token through this path, Vault will use the predefined 
role's parameters to create a new access token.`
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

// tokenBackend defines the narrow interface this provider needs.
type tokenBackend interface {
	backend.Logging
	backend.RoleLocker
	backend.RoleStore
	backend.ClientProvider
	backend.EventSender
}

// Provider implements backend.PathProvider for the token role path.
type Provider struct {
	b      tokenBackend
	secret *framework.Secret
}

func (p *Provider) Name() string { return "token" }

// New creates a new token path provider.
// The secret parameter is the framework.Secret for access tokens (injected, not from the interface).
func New(b tokenBackend, s *framework.Secret) *Provider {
	return &Provider{b: b, secret: s}
}

// Paths returns the framework paths for token generation.
func (p *Provider) Paths() []*framework.Path {
	return []*framework.Path{p.pathTokenRoles()}
}

func (p *Provider) pathTokenRoles() *framework.Path {
	return &framework.Path{
		HelpSynopsis:    strings.TrimSpace(pathTokenRolesHelpSyn),
		HelpDescription: strings.TrimSpace(pathTokenRolesHelpDesc),
		Pattern:         fmt.Sprintf("%s/%s%s", PathTokenRoleStorage, framework.GenericNameRegex("role_name"), framework.OptionalParamRegex("path")),
		Fields:          FieldSchemaTokenRole,
		DisplayAttrs: &framework.DisplayAttributes{
			OperationPrefix: paths.OperationPrefixGitlabAccessTokens,
			OperationSuffix: "generate",
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: p.pathTokenRoleCreate,
				Summary:  "Create an access token based on a predefined role",
				DisplayAttrs: &framework.DisplayAttributes{
					OperationVerb:   "generate",
					OperationSuffix: "credentials",
				},
				Responses: map[int][]framework.Response{
					http.StatusOK: {{
						Description: http.StatusText(http.StatusOK),
						Fields:      secret.FieldSchemaAccessTokens,
					}},
				},
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: p.pathTokenRoleCreate,
				Summary:  "Create an access token based on a predefined role",
				DisplayAttrs: &framework.DisplayAttributes{
					OperationSuffix: "credentials-with-parameters",
					OperationVerb:   "generate-with-parameters",
				},
				Responses: map[int][]framework.Response{
					http.StatusOK: {{
						Description: http.StatusText(http.StatusOK),
						Fields:      secret.FieldSchemaAccessTokens,
					}},
				},
			},
		},
	}
}
