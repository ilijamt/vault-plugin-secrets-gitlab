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

  # null_resource.mount_default_config_patch["auto_rotate_before"] will be created
  + resource "null_resource" "mount_default_config_patch" {
      + id       = (known after apply)
      + triggers = {
          + "auto_rotate_before" = "48h"
        }
    }

  # null_resource.mount_default_config_patch["auto_rotate_token"] will be created
  + resource "null_resource" "mount_default_config_patch" {
      + id       = (known after apply)
      + triggers = {
          + "auto_rotate_token" = "true"
        }
    }

  # null_resource.mount_default_config_patch["base_url"] will be created
  + resource "null_resource" "mount_default_config_patch" {
      + id       = (known after apply)
      + triggers = {
          + "base_url" = "http://localhost:8080"
        }
    }

  # null_resource.mount_default_config_patch["type"] will be created
  + resource "null_resource" "mount_default_config_patch" {
      + id       = (known after apply)
      + triggers = {
          + "type" = "self-managed"
        }
    }

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

Plan: 5 to add, 0 to change, 0 to destroy.

❯ terraform apply plan
vault_generic_endpoint.mount_default_config: Creating...
vault_generic_endpoint.mount_default_config: Creation complete after 0s [id=gitlab/config/default]
null_resource.mount_default_config_patch["base_url"]: Creating...
null_resource.mount_default_config_patch["auto_rotate_token"]: Creating...
null_resource.mount_default_config_patch["auto_rotate_before"]: Creating...
null_resource.mount_default_config_patch["type"]: Creating...
null_resource.mount_default_config_patch["base_url"]: Provisioning with 'local-exec'...
null_resource.mount_default_config_patch["base_url"] (local-exec): Executing: ["bash" "-c" "      vault patch gitlab/config/default base_url=http://localhost:8080 >/dev/null\n"]
null_resource.mount_default_config_patch["auto_rotate_before"]: Provisioning with 'local-exec'...
null_resource.mount_default_config_patch["type"]: Provisioning with 'local-exec'...
null_resource.mount_default_config_patch["auto_rotate_token"]: Provisioning with 'local-exec'...
null_resource.mount_default_config_patch["auto_rotate_before"] (local-exec): Executing: ["bash" "-c" "      vault patch gitlab/config/default auto_rotate_before=48h >/dev/null\n"]
null_resource.mount_default_config_patch["type"] (local-exec): Executing: ["bash" "-c" "      vault patch gitlab/config/default type=self-managed >/dev/null\n"]
null_resource.mount_default_config_patch["auto_rotate_token"] (local-exec): Executing: ["bash" "-c" "      vault patch gitlab/config/default auto_rotate_token=true >/dev/null\n"]
null_resource.mount_default_config_patch["type"]: Creation complete after 0s [id=8417009586748670144]
null_resource.mount_default_config_patch["base_url"]: Creation complete after 0s [id=3051316335689864969]
null_resource.mount_default_config_patch["auto_rotate_before"]: Creation complete after 0s [id=3174774997957690363]
null_resource.mount_default_config_patch["auto_rotate_token"]: Creation complete after 0s [id=2586021087863779131]
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

Now if we change the value 

```shell
❯ terraform plan -out plan -var "gitlab_type=saas"
vault_generic_endpoint.mount_default_config: Refreshing state... [id=gitlab/config/default]
null_resource.mount_default_config_patch["auto_rotate_before"]: Refreshing state... [id=3174774997957690363]
null_resource.mount_default_config_patch["base_url"]: Refreshing state... [id=3051316335689864969]
null_resource.mount_default_config_patch["auto_rotate_token"]: Refreshing state... [id=2586021087863779131]
null_resource.mount_default_config_patch["type"]: Refreshing state... [id=8417009586748670144]

Terraform used the selected providers to generate the following execution plan. Resource actions are indicated with the following symbols:
-/+ destroy and then create replacement

Terraform will perform the following actions:

  # null_resource.mount_default_config_patch["type"] must be replaced
-/+ resource "null_resource" "mount_default_config_patch" {
      ~ id       = "8417009586748670144" -> (known after apply)
      ~ triggers = { # forces replacement
          ~ "type" = "self-managed" -> "saas"
        }
    }

Plan: 1 to add, 0 to change, 1 to destroy.

──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────

Saved the plan to: plan

To perform exactly these actions, run the following command to apply:
    terraform apply "plan"
    
❯ terraform apply plan
null_resource.mount_default_config_patch["type"]: Destroying... [id=8417009586748670144]
null_resource.mount_default_config_patch["type"]: Destruction complete after 0s
null_resource.mount_default_config_patch["type"]: Creating...
null_resource.mount_default_config_patch["type"]: Provisioning with 'local-exec'...
null_resource.mount_default_config_patch["type"] (local-exec): Executing: ["bash" "-c" "      vault patch gitlab/config/default type=saas >/dev/null\n"]
null_resource.mount_default_config_patch["type"]: Creation complete after 0s [id=7287861734270135244]

Apply complete! Resources: 1 added, 0 changed, 1 destroyed.

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
type                  saas
```