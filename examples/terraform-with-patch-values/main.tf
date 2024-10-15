variable "gitlab_base_url" {
  description = "GitLab base URL, eg. https://gitlab.com"
  type        = string
}

variable "gitlab_token" {
  description = "GitLab Token"
  type        = string
  sensitive   = true
}

variable "gitlab_type" {
  description = "GitLab Type can be saas, self-managed or dedicated"
  type        = string
  default     = "self-managed"
}

variable "gitlab_auto_rotate_token" {
  type    = bool
  default = true
}

variable "gitlab_auto_rotate_before" {
  type    = string
  default = "48h"
}

locals {
  vault_config_default_data = {
    token              = var.gitlab_token
    base_url           = var.gitlab_base_url
    auto_rotate_token  = var.gitlab_auto_rotate_token
    auto_rotate_before = var.gitlab_auto_rotate_before
    type               = var.gitlab_type
  }

  vault_config_default_patch_data = {
    for k, v in local.vault_config_default_data : k => v if k != "token"
  }
}

resource "vault_generic_endpoint" "mount_default_config" {
  path                 = "gitlab/config/default"
  disable_delete       = true
  ignore_absent_fields = true

  write_fields = [
    "base_url",
    "auto_rotate_token",
    "auto_rotate_before",
    "type",
    "scopes",
  ]

  data_json = jsonencode(local.vault_config_default_data)

  lifecycle {
    ignore_changes = [
      data_json
    ]
  }
}

resource "null_resource" "mount_default_config_patch" {
  for_each = local.vault_config_default_patch_data
  triggers = { (each.key) = each.value }

  provisioner "local-exec" {
    command     = <<EOT
      vault patch gitlab/config/default ${each.key}=${each.value} >/dev/null
    EOT
    interpreter = ["bash", "-c"]
  }

  depends_on = [
    vault_generic_endpoint.mount_default_config,
  ]
}