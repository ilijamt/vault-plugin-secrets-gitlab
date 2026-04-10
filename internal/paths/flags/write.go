package flags

import (
	"context"
	"strconv"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/mitchellh/mapstructure"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
)

func (p *Provider) pathFlagsUpdate(ctx context.Context, req *logical.Request, data *framework.FieldData) (lResp *logical.Response, err error) {
	var eventData = make(map[string]string)

	p.b.UpdateFlags(func(f *flags.Flags) {
		if showConfigToken, ok := data.GetOk("show_config_token"); ok {
			f.ShowConfigToken = showConfigToken.(bool)
			eventData["show_config_token"] = strconv.FormatBool(f.ShowConfigToken)
		}
	})

	_ = p.b.SendEvent(ctx, eventWrite, eventData)

	var flagData map[string]any
	err = mapstructure.Decode(p.b.Flags(), &flagData)
	return &logical.Response{Data: flagData}, err
}
