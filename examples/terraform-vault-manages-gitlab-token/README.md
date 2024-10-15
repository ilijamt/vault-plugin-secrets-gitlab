Terraform with Patch Values
---------------------------

```shell
export TF_VAR_gitlab_base_url="http://localhost:8080"
export TF_VAR_gitlab_token="glpat-secret-random-token"
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=root
```

```shell
❯ terraform plan -out plan

Terraform used the selected providers to generate the following execution plan. Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # vault_generic_endpoint.mount_default_config will be created
  + resource "vault_generic_endpoint" "mount_default_config" {
      + data_json            = (sensitive value)
      + disable_delete       = true
      + disable_read         = false
      + id                   = (known after apply)
      + ignore_absent_fields = true
      + path                 = "gitlab/config/default"
      + write_data           = (known after apply)
      + write_data_json      = (known after apply)
      + write_fields         = [
          + "base_url",
          + "auto_rotate_token",
          + "auto_rotate_before",
          + "type",
          + "scopes",
        ]
    }

Plan: 1 to add, 0 to change, 0 to destroy.

──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────

Saved the plan to: plan

To perform exactly these actions, run the following command to apply:
    terraform apply "plan"
❯ terraform apply plan
vault_generic_endpoint.mount_default_config: Creating...
vault_generic_endpoint.mount_default_config: Creation complete after 0s [id=gitlab/config/default]

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.
```

After that we have a configuration endpoint in Vault 

```shell
❯ vault list gitlab/config
Keys
----
default

❯ vault read gitlab/config/default
Key                   Value
---                   -----
auto_rotate_before    48h0m0s
auto_rotate_token     true
base_url              http://localhost:8080
name                  default
scopes                api, read_api, read_user, sudo, admin_mode, create_runner, k8s_proxy, read_repository, write_repository, ai_features, read_service_ping
token_created_at      2024-07-11T18:53:26Z
token_expires_at      2025-07-11T00:00:00Z
token_id              1
token_sha1_hash       9441e6e07d77a2d5601ab5d7cac5868d358d885c
type                  self-managed
```
