package gitlab

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	secretAccessTokenType = "access_tokens"
)

var (
	fieldSchemaAccessTokens = map[string]*framework.FieldSchema{
		"name": {
			Type:         framework.TypeString,
			DisplayAttrs: &framework.DisplayAttributes{Name: "Token name"},
		},
		"token": {
			Type:         framework.TypeString,
			DisplayAttrs: &framework.DisplayAttributes{Name: "Token"},
		},
		"path": {
			Type:         framework.TypeString,
			DisplayAttrs: &framework.DisplayAttributes{Name: "Path"},
		},
		"scopes": {
			Type:         framework.TypeStringSlice,
			DisplayAttrs: &framework.DisplayAttributes{Name: "Scopes"},
		},
		"access_level": {
			Type:         framework.TypeString,
			DisplayAttrs: &framework.DisplayAttributes{Name: "Access Level"},
		},
		"expires_at": {
			Type:         framework.TypeTime,
			DisplayAttrs: &framework.DisplayAttributes{Name: "Expires At"},
		},
	}
)

func secretAccessTokens(b *Backend) *framework.Secret {
	return &framework.Secret{
		Type:   secretAccessTokenType,
		Fields: fieldSchemaAccessTokens,
		Revoke: b.secretAccessTokenRevoke,
	}
}

func (b *Backend) secretAccessTokenRevoke(ctx context.Context, req *logical.Request, _ *framework.FieldData) (*logical.Response, error) {
	var config *EntryConfig
	var err error
	config, err = getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return logical.ErrorResponse(ErrBackendNotConfigured.Error()), nil
	}

	var secret = req.Secret
	if secret == nil {
		return nil, fmt.Errorf("secret: %w", ErrNilValue)
	}

	var tokenId int
	tokenId, err = convertToInt(req.Secret.InternalData["token_id"])
	if err != nil {
		return nil, fmt.Errorf("token_id: %w", err)
	}

	var parentId = req.Secret.InternalData["parent_id"].(string)
	var tokenType TokenType
	var tokenTypeValue = req.Secret.InternalData["token_type"].(string)
	var gitlabRevokesToken, _ = strconv.ParseBool(req.Secret.InternalData["gitlab_revokes_token"].(string))
	var vaultRevokesToken = !gitlabRevokesToken
	tokenType, err = TokenTypeParse(tokenTypeValue)
	if err != nil {
		// shouldn't be possible to hit due to the guards in the creation of the roles
		return nil, fmt.Errorf("%s: %w", tokenTypeValue, ErrUnknownTokenType)
	}

	if vaultRevokesToken {
		var client Client
		client, err = b.getClient(ctx, req.Storage)
		if err != nil {
			return nil, fmt.Errorf("revoke token: %w", err)
		}

		switch tokenType {
		case TokenTypePersonal:
			err = client.RevokePersonalAccessToken(tokenId)
		case TokenTypeProject:
			err = client.RevokeProjectAccessToken(tokenId, parentId)
		case TokenTypeGroup:
			err = client.RevokeGroupAccessToken(tokenId, parentId)
		}

		if err != nil && !errors.Is(err, ErrAccessTokenNotFound) {
			return logical.ErrorResponse("failed to revoke token"), fmt.Errorf("revoke token: %w", err)
		}
	}

	event(ctx, b.Backend, "token-revoke", map[string]string{
		"lease_id":             secret.LeaseID,
		"path":                 req.Secret.InternalData["path"].(string),
		"name":                 req.Secret.InternalData["name"].(string),
		"token_id":             strconv.Itoa(tokenId),
		"token_type":           tokenTypeValue,
		"gitlab_revokes_token": strconv.FormatBool(gitlabRevokesToken),
	})

	return nil, nil
}
