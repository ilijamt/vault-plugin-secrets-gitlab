package role

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/logging"
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

	logging.New(p.b.Logger()).Debug("Role read", logging.Fields{RoleName: roleName})

	return &logical.Response{
		Data: role.LogicalResponseData(),
	}, nil
}
