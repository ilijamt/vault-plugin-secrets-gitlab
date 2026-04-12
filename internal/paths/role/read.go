package role

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func (p *Provider) pathRolesRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	var roleName = data.Get("role_name").(string)

	lock := p.b.LockForKey("role", roleName)
	lock.RLock()
	defer lock.RUnlock()

	role, err := p.b.GetRole(ctx, req.Storage, roleName)
	if err != nil {
		return logical.ErrorResponse("error reading role"), err
	}

	if role == nil {
		return nil, nil
	}

	p.b.Logger().Debug("Role read", "role", roleName)

	return &logical.Response{
		Data: role.LogicalResponseData(),
	}, nil
}
