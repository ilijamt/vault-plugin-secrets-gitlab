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
	name := data.Get("config_name").(string)
	l := p.lock(name)
	l.Lock()
	defer l.Unlock()

	config := new(modelConfig.EntryConfig)
	warnings, err := config.UpdateFromFieldData(data)
	if err != nil {
		return nil, err
	}
	config.Name = name

	if _, err = p.updateConfigClientInfo(ctx, config); err != nil {
		return nil, err
	}

	if err = p.b.SaveConfig(ctx, req.Storage, config); err != nil {
		return nil, err
	}

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

	p.b.DeleteClient(name)
	lrd := config.LogicalResponseData(p.b.Flags().ShowConfigToken)
	p.b.Logger().Debug("Wrote new config", "lrd", lrd, "warnings", warnings)
	return &logical.Response{Data: lrd, Warnings: warnings}, nil
}
