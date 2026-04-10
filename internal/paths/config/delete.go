package config

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
)

func (p *Provider) pathConfigDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	p.b.ClientLock()
	defer p.b.ClientUnlock()
	var err error
	var name = data.Get("config_name").(string)

	if config, err := p.b.GetConfig(ctx, req.Storage, name); err == nil {
		if config == nil {
			return logical.ErrorResponse(errs.ErrBackendNotConfigured.Error()), nil
		}

		if err = req.Storage.Delete(ctx, fmt.Sprintf("%s/%s", backend.PathConfigStorage, name)); err == nil {
			_ = p.b.SendEvent(ctx, eventDelete, map[string]string{
				"path": fmt.Sprintf("%s/%s", backend.PathConfigStorage, name),
			})
			p.b.SetClient(nil, name)
		}
	}

	return nil, err
}
