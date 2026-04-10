package config

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	modelConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
)

func (p *Provider) pathConfigPatch(ctx context.Context, req *logical.Request, data *framework.FieldData) (lResp *logical.Response, err error) {
	var name = data.Get("config_name").(string)
	var warnings []string
	var changes map[string]string
	var config *modelConfig.EntryConfig
	config, err = p.b.GetConfig(ctx, req.Storage, name)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return logical.ErrorResponse(errs.ErrBackendNotConfigured.Error()), nil
	}

	warnings, changes, err = config.Merge(data)
	if err != nil {
		return nil, err
	}

	if _, ok := data.GetOk("token"); ok {
		if _, err = p.updateConfigClientInfo(ctx, config); err != nil {
			return nil, err
		}
	}

	p.b.ClientLock()
	defer p.b.ClientUnlock()
	if err = p.b.SaveConfig(ctx, config, req.Storage); err == nil {
		lrd := config.LogicalResponseData(p.b.Flags().ShowConfigToken)
		_ = p.b.SendEvent(ctx, eventPatch, changes)
		p.b.SetClient(nil, name)
		p.b.Logger().Debug("Patched config", "lrd", lrd, "warnings", warnings)
		lResp = &logical.Response{Data: lrd, Warnings: warnings}
	}

	return lResp, err
}
