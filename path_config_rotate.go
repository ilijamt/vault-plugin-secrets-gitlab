package gitlab

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"net/http"
	"strings"
	"time"
)

func pathConfigTokenRotate(b *Backend) *framework.Path {
	return &framework.Path{
		HelpSynopsis:    strings.TrimSpace(pathConfigHelpSynopsis),
		HelpDescription: strings.TrimSpace(pathConfigHelpDescription),
		Pattern:         fmt.Sprintf("%s/rotate$", PathConfigStorage),
		Fields:          fieldSchemaConfig,
		DisplayAttrs: &framework.DisplayAttributes{
			OperationPrefix: operationPrefixGitlabAccessTokens,
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.UpdateOperation: &framework.PathOperation{
				Callback:     b.pathConfigTokenRotate,
				DisplayAttrs: &framework.DisplayAttributes{OperationVerb: "configure"},
				Summary:      "Rotate the main Gitlab Access Token.",
				Responses: map[int][]framework.Response{
					http.StatusNoContent: {{
						Description: http.StatusText(http.StatusNoContent),
					}},
				},
			},
		},
	}
}

func (b *Backend) checkAndRotateConfigToken(ctx context.Context, request *logical.Request, config *entryConfig) error {
	var client Client
	var err error

	if client, err = b.getClient(ctx, request.Storage); err != nil {
		return err
	}

	if config.TokenExpiresAt.IsZero() {
		var entryToken *EntryToken
		// we need to fetch the token expiration information
		entryToken, err = client.MainTokenInfo()
		if err != nil {
			return err
		}
		// and update the information so we can do the checks
		config.TokenExpiresAt = *entryToken.ExpiresAt
		err = func() error {
			b.lockClientMutex.Lock()
			defer b.lockClientMutex.Unlock()
			return saveConfig(ctx, *config, request.Storage)
		}()
		if err != nil {
			return err
		}
	}

	if time.Until(config.TokenExpiresAt) > config.AutoRotateBefore {
		b.Logger().Debug("Nothing to do it's not yet time to rotate the token")
		return nil
	}

	_, err = b.pathConfigTokenRotate(ctx, request, &framework.FieldData{})
	return err
}

func (b *Backend) pathConfigTokenRotate(ctx context.Context, request *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	b.lockClientMutex.Lock()
	defer b.lockClientMutex.Unlock()

	var config *entryConfig
	var client Client
	var err error

	if config, err = getConfig(ctx, request.Storage); err != nil {
		return nil, err
	}
	if config == nil {
		// no configuration yet so we don't need to rotate anything
		return logical.ErrorResponse(ErrBackendNotConfigured.Error()), nil
	}

	if client, err = b.getClient(ctx, request.Storage); err != nil {
		return nil, err
	}

	var entryToken *EntryToken
	entryToken, err = client.RotateMainToken()
	if err != nil {
		return nil, err
	}

	config.Token = entryToken.Token

	err = saveConfig(ctx, *config, request.Storage)
	if err != nil {
		return nil, err
	}

	event(ctx, b.Backend, "config-token-rotate", map[string]string{
		"path": "config",
	})

	return nil, nil
}
