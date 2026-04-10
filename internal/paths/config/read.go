package config

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	modelConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
)

func (p *Provider) pathConfigRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (lResp *logical.Response, err error) {
	p.b.ClientRLock()
	defer p.b.ClientRUnlock()

	var name = data.Get("config_name").(string)
	var config *modelConfig.EntryConfig
	if config, err = p.b.GetConfig(ctx, req.Storage, name); err == nil {
		if config == nil {
			return logical.ErrorResponse(errs.ErrBackendNotConfigured.Error()), nil
		}
		lrd := config.LogicalResponseData(p.b.Flags().ShowConfigToken)
		p.b.Logger().Debug("Reading configuration info", "info", lrd)
		lResp = &logical.Response{Data: lrd}
	}
	return lResp, err
}
