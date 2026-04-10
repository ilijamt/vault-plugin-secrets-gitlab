package flags

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/mitchellh/mapstructure"
)

func (p *Provider) pathFlagsRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (lResp *logical.Response, err error) {
	var flagData map[string]any
	err = mapstructure.Decode(p.b.Flags(), &flagData)
	return &logical.Response{Data: flagData}, err
}
