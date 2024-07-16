package gitlab

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
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
			},
		},
	}
}

func (b *Backend) checkAndRotateConfigToken(ctx context.Context, request *logical.Request, config *EntryConfig) error {
	var err error
	b.Logger().Debug("Running checkAndRotateConfigToken")

	if time.Until(config.TokenExpiresAt) > config.AutoRotateBefore {
		b.Logger().Debug("Nothing to do it's not yet time to rotate the token")
		return nil
	}

	_, err = b.pathConfigTokenRotate(ctx, request, &framework.FieldData{})
	return err
}

func (b *Backend) pathConfigTokenRotate(ctx context.Context, request *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	b.Logger().Debug("Running pathConfigTokenRotate")
	var config *EntryConfig
	var client Client
	var err error

	b.lockClientMutex.RLock()
	if config, err = getConfig(ctx, request.Storage); err != nil {
		b.lockClientMutex.RUnlock()
		b.Logger().Error("Failed to fetch configuration", "error", err.Error())
		return nil, err
	}
	b.lockClientMutex.RUnlock()

	if config == nil {
		// no configuration yet so we don't need to rotate anything
		return logical.ErrorResponse(ErrBackendNotConfigured.Error()), nil
	}

	if client, err = b.getClient(ctx, request.Storage); err != nil {
		return nil, err
	}

	var entryToken *EntryToken
	entryToken, _, err = client.RotateCurrentToken()
	if err != nil {
		b.Logger().Error("Failed to rotate main token", "err", err)
		return nil, err
	}

	config.Token = entryToken.Token
	config.TokenId = entryToken.TokenID
	config.Scopes = entryToken.Scopes
	if entryToken.ExpiresAt != nil {
		config.TokenExpiresAt = *entryToken.ExpiresAt
	}
	if entryToken.CreatedAt != nil {
		config.TokenExpiresAt = *entryToken.CreatedAt
	}
	b.lockClientMutex.Lock()
	defer b.lockClientMutex.Unlock()
	err = saveConfig(ctx, *config, request.Storage)
	if err != nil {
		b.Logger().Error("failed to store configuration for revocation", "err", err)
		return nil, err
	}

	event(ctx, b.Backend, "config-token-rotate", map[string]string{
		"path":       "config",
		"expires_at": entryToken.ExpiresAt.Format(time.RFC3339),
		"created_at": entryToken.CreatedAt.Format(time.RFC3339),
		"scopes":     strings.Join(entryToken.Scopes, ", "),
		"token_id":   strconv.Itoa(entryToken.TokenID),
		"name":       entryToken.Name,
	})

	b.SetClient(nil)
	return config.Response(), nil
}
