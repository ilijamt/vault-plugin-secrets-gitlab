package secret

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	g "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/logging"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

const (
	SecretAccessTokenType = "access_tokens"
)

type secretBackend interface {
	backend.Logging
	backend.ClientReader
	backend.EventSender
}

var (
	// FieldSchemaAccessTokens defines the field schema for access token secrets.
	FieldSchemaAccessTokens = map[string]*framework.FieldSchema{
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

// NewSecret creates a framework.Secret for access tokens with the revoke handler
// wired through the provided secretBackend interface.
func NewSecret(b secretBackend) *framework.Secret {
	return &framework.Secret{
		Type:   SecretAccessTokenType,
		Fields: FieldSchemaAccessTokens,
		Revoke: revokeAccessToken(b),
	}
}

func leaseValue[T any](secret *logical.Secret, key string) (value T, err error) {
	value, ok := secret.InternalData[key].(T)
	if !ok {
		return value, fmt.Errorf("lease %s: %w", key, errs.ErrInvalidValue)
	}
	return value, nil
}

func revokeAccessToken(b secretBackend) framework.OperationFunc {
	return func(ctx context.Context, req *logical.Request, _ *framework.FieldData) (*logical.Response, error) {
		var err error

		if req.Storage == nil {
			return nil, fmt.Errorf("storage: %w", errs.ErrNilValue)
		}

		var secret = req.Secret
		if secret == nil {
			return nil, fmt.Errorf("secret: %w", errs.ErrNilValue)
		}

		var tokenId int64
		tokenId, err = utils.ConvertToInt64(req.Secret.InternalData["token_id"])
		if err != nil {
			return nil, fmt.Errorf("token_id: %w", err)
		}

		var configName string
		if configName, err = leaseValue[string](secret, "config_name"); err != nil {
			return nil, err
		}

		var gitlabRevokesToken bool
		if gitlabRevokesToken, err = leaseValue[bool](secret, "gitlab_revokes_token"); err != nil {
			return nil, err
		}
		var parentId string
		if parentId, err = leaseValue[string](secret, "parent_id"); err != nil {
			return nil, err
		}
		var tokenTypeValue string
		if tokenTypeValue, err = leaseValue[string](secret, "token_type"); err != nil {
			return nil, err
		}
		var vaultRevokesToken = !gitlabRevokesToken

		if vaultRevokesToken {
			var tokenType token.Type
			if tokenType, err = token.ParseType(tokenTypeValue); err != nil {
				return nil, fmt.Errorf("lease token_type: %w", err)
			}

			var userId int64
			if tokenType == token.TypeGroupServiceAccount || tokenType == token.TypeProjectServiceAccount {
				if userId, err = utils.ConvertToInt64(req.Secret.InternalData["user_id"]); err != nil {
					return nil, fmt.Errorf("lease user_id: %w", err)
				}
			}

			logging.New(b.Logger()).Debug("Revoking token", logging.Fields{TokenID: tokenId, TokenType: tokenTypeValue, ConfigName: configName})
			var client g.Client
			client, err = b.GetClientByName(ctx, req.Storage, configName)
			if err != nil {
				return nil, fmt.Errorf("revoke token cannot get client got %s config: %w", configName, err)
			}

			switch tokenType {
			case token.TypePersonal:
				err = client.RevokePersonalAccessToken(ctx, tokenId)
			case token.TypeProject:
				err = client.RevokeProjectAccessToken(ctx, tokenId, parentId)
			case token.TypeGroup:
				err = client.RevokeGroupAccessToken(ctx, tokenId, parentId)
			case token.TypeUserServiceAccount:
				err = client.RevokeUserServiceAccountAccessToken(ctx, tokenId)
			case token.TypeGroupServiceAccount:
				err = client.RevokeGroupServiceAccountAccessToken(ctx, tokenId, parentId, userId)
			case token.TypeProjectServiceAccount:
				err = client.RevokeProjectServiceAccountAccessToken(ctx, tokenId, parentId, userId)
			case token.TypePipelineProjectTrigger:
				err = client.RevokePipelineProjectTriggerAccessToken(ctx, tokenId, parentId)
			case token.TypeGroupDeploy:
				err = client.RevokeGroupDeployToken(ctx, tokenId, parentId)
			case token.TypeProjectDeploy:
				err = client.RevokeProjectDeployToken(ctx, tokenId, parentId)
			}

			if err != nil && !errors.Is(err, errs.ErrAccessTokenNotFound) {
				return logical.ErrorResponse("failed to revoke token"), fmt.Errorf("revoke token: %w", err)
			}
		}

		var path, _ = req.Secret.InternalData["path"].(string)
		var name, _ = req.Secret.InternalData["name"].(string)
		_ = b.SendEvent(ctx, eventRevoke, map[string]string{
			"lease_id":             secret.LeaseID,
			"path":                 path,
			"name":                 name,
			"token_id":             strconv.FormatInt(tokenId, 10),
			"token_type":           tokenTypeValue,
			"config_name":          configName,
			"gitlab_revokes_token": strconv.FormatBool(gitlabRevokesToken),
		})

		return nil, nil
	}
}
