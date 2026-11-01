package config

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/logging"
)

func (p *Provider) pathConfigRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	name := data.Get("config_name").(string)
	l := p.b.LockForKey("config", name)
	l.RLock()
	defer l.RUnlock()
	config, err := p.b.GetConfig(ctx, req.Storage, name)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return logical.ErrorResponse(errs.ErrBackendNotConfigured.Error()), nil
	}

	lrd := config.LogicalResponseData(p.b.Flags().ShowConfigToken)
	logging.New(p.b.Logger()).Debug("Reading configuration info", logging.Fields{Extra: map[string]any{"info": config.LogicalResponseData(false)}})
	return &logical.Response{Data: lrd}, nil
}
