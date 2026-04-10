package flags

import (
	"net/http"
	"strings"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths"
)

const (
	PathConfigFlags = "flags"

	pathFlagsHelpSynopsis    = `Flags for the plugin.`
	pathFlagsHelpDescription = ``
)

var FieldSchemaFlags = map[string]*framework.FieldSchema{
	"show_config_token": {
		Type:         framework.TypeBool,
		Description:  "Should we display the token value for the roles?",
		Default:      false,
		DisplayAttrs: &framework.DisplayAttributes{Name: "Show Config Token"},
	},
}

// flagsBackend defines the narrow interface this provider needs.
type flagsBackend interface {
	backend.FlagsProvider
	backend.EventSender
}

// Provider implements backend.PathProvider for the flags path.
type Provider struct {
	b flagsBackend
}

func (p *Provider) Name() string { return "flags" }

// New creates a new flags path provider.
func New(b flagsBackend) *Provider {
	return &Provider{b: b}
}

// Paths returns the framework paths for the flags endpoint.
func (p *Provider) Paths() []*framework.Path {
	return []*framework.Path{p.pathFlags()}
}

func (p *Provider) pathFlags() *framework.Path {
	var operations = map[logical.Operation]framework.OperationHandler{
		logical.ReadOperation: &framework.PathOperation{
			Callback: p.pathFlagsRead,
			DisplayAttrs: &framework.DisplayAttributes{
				OperationVerb:   "read",
				OperationSuffix: "flags",
			},
			Summary: "Read the flags for the plugin.",
			Responses: map[int][]framework.Response{
				http.StatusOK: {{
					Description: http.StatusText(http.StatusOK),
					Fields:      FieldSchemaFlags,
				}},
			},
		},
	}

	if p.b.Flags().AllowRuntimeFlagsChange {
		operations[logical.UpdateOperation] = &framework.PathOperation{
			Callback: p.pathFlagsUpdate,
			DisplayAttrs: &framework.DisplayAttributes{
				OperationVerb:   "update",
				OperationSuffix: "flags",
			},
			Summary: "Update the flags for the plugin.",
			Responses: map[int][]framework.Response{
				http.StatusOK: {{
					Description: http.StatusText(http.StatusOK),
					Fields:      FieldSchemaFlags,
				}},
				http.StatusBadRequest: {{
					Description: http.StatusText(http.StatusBadRequest),
					Fields:      FieldSchemaFlags,
				}},
			},
		}
	}

	return &framework.Path{
		HelpSynopsis:    strings.TrimSpace(pathFlagsHelpSynopsis),
		HelpDescription: strings.TrimSpace(pathFlagsHelpDescription),
		Pattern:         PathConfigFlags,
		Fields:          FieldSchemaFlags,
		DisplayAttrs: &framework.DisplayAttributes{
			OperationPrefix: paths.OperationPrefixGitlabAccessTokens,
		},
		Operations: operations,
	}
}
