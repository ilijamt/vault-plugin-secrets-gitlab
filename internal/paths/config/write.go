package config

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	modelConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
)

func (p *Provider) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	var name = data.Get("config_name").(string)
	var config = new(modelConfig.EntryConfig)
	var warnings, err = config.UpdateFromFieldData(data)
	if err != nil {
		return nil, err
	}
	config.Name = name

	if _, err = p.updateConfigClientInfo(ctx, config); err != nil {
		return nil, err
	}

	p.b.ClientLock()
	defer p.b.ClientUnlock()
	var lResp *logical.Response

	if err = p.b.SaveConfig(ctx, config, req.Storage); err == nil {
		_ = p.b.SendEvent(ctx, eventWrite, map[string]string{
			"path":               fmt.Sprintf("%s/%s", backend.PathConfigStorage, name),
			"auto_rotate_token":  strconv.FormatBool(config.AutoRotateToken),
			"auto_rotate_before": config.AutoRotateBefore.String(),
			"base_url":           config.BaseURL,
			"token_id":           strconv.FormatInt(config.TokenId, 10),
			"created_at":         config.TokenCreatedAt.Format(time.RFC3339),
			"expires_at":         config.TokenExpiresAt.Format(time.RFC3339),
			"scopes":             strings.Join(config.Scopes, ", "),
			"type":               config.Type.String(),
			"config_name":        config.Name,
		})

		p.b.SetClient(nil, name)
		lrd := config.LogicalResponseData(p.b.Flags().ShowConfigToken)
		p.b.Logger().Debug("Wrote new config", "lrd", lrd, "warnings", warnings)
		lResp = &logical.Response{Data: lrd, Warnings: warnings}
	}

	return lResp, err
}
