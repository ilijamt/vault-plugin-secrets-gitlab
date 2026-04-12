package config

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
)

func (p *Provider) pathConfigPatch(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	name := data.Get("config_name").(string)
	l := p.b.LockForKey("config", name)
	l.Lock()
	defer l.Unlock()

	config, err := p.b.GetConfig(ctx, req.Storage, name)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return logical.ErrorResponse(errs.ErrBackendNotConfigured.Error()), nil
	}

	warnings, changes, err := config.Merge(data)
	if err != nil {
		return nil, err
	}

	if _, ok := data.GetOk("token"); ok {
		if _, err = p.updateConfigClientInfo(ctx, config); err != nil {
			return nil, err
		}
	}

	if err = p.b.SaveConfig(ctx, req.Storage, config); err != nil {
		return nil, err
	}

	lrd := config.LogicalResponseData(p.b.Flags().ShowConfigToken)
	_ = p.b.SendEvent(ctx, eventPatch, changes)
	p.b.DeleteClient(name)
	p.b.Logger().Debug("Patched config", "lrd", lrd, "warnings", warnings)
	return &logical.Response{Data: lrd, Warnings: warnings}, nil
}
