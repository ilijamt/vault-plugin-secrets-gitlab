package gitlab

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/mitchellh/mapstructure"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
)

const (
	PathConfigFlags = "flags"
)

var FieldSchemaFlags = map[string]*framework.FieldSchema{
	"show_config_token": {
		Type:         framework.TypeBool,
		Description:  "Should we display the token value for the roles?",
		Default:      false,
		DisplayAttrs: &framework.DisplayAttributes{Name: "Show Config Token"},
	},
}

func (b *Backend) pathFlagsRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (lResp *logical.Response, err error) {
	b.lockFlagsMutex.RLock()
	defer b.lockFlagsMutex.RUnlock()
	var flagData map[string]any
	err = mapstructure.Decode(b.flags, &flagData)
	return &logical.Response{Data: flagData}, err
}

func (b *Backend) pathFlagsUpdate(ctx context.Context, req *logical.Request, data *framework.FieldData) (lResp *logical.Response, err error) {
	b.lockFlagsMutex.Lock()
	defer b.lockFlagsMutex.Unlock()

	var eventData = make(map[string]string)

	if showConfigToken, ok := data.GetOk("show_config_token"); ok {
		b.flags.ShowConfigToken = showConfigToken.(bool)
		eventData["show_config_token"] = strconv.FormatBool(b.flags.ShowConfigToken)
	}

	event.Event(ctx, b.Backend, operationPrefixGitlabAccessTokens, "flags-write", eventData)

	var flagData map[string]any
	err = mapstructure.Decode(b.flags, &flagData)
	return &logical.Response{Data: flagData}, err
}

func pathFlags(b *Backend) *framework.Path {
	var operations = map[logical.Operation]framework.OperationHandler{
		logical.ReadOperation: &framework.PathOperation{
			Callback: b.pathFlagsRead,
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

	if b.flags.AllowRuntimeFlagsChange {
		operations[logical.UpdateOperation] = &framework.PathOperation{
			Callback: b.pathFlagsUpdate,
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
			OperationPrefix: operationPrefixGitlabAccessTokens,
		},
		Operations: operations,
	}
}

const pathFlagsHelpSynopsis = `Flags for the plugin.`

const pathFlagsHelpDescription = ``
