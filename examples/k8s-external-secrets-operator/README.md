# External Secrets Operator with Vault GitLab Plugin

This example shows how to use the [External Secrets Operator](https://external-secrets.io/) with HashiCorp Vault and `vault-plugin-secrets-gitlab` to automatically provision GitLab deploy tokens as Kubernetes Docker registry credentials.

## Prerequisites

- HashiCorp Vault with `vault-plugin-secrets-gitlab` configured
- External Secrets Operator installed in your Kubernetes cluster
- Vault Kubernetes auth method configured

## Steps

### 1. Create a role in the GitLab secrets plugin

```shell
vault write <path_to_vault_gitlab_plugin>/roles/<gitlab_token_role> \
  name="vault-{{ randHexString 4 }}" \
  path=<path_to_project> \
  scopes=read_registry \
  token_type=project-deploy \
  ttl=168h
```

### 2. Apply the Kubernetes manifest

Update the placeholders in [manifest.yaml](manifest.yaml) with your values:

| Placeholder                    | Description                                      |
|--------------------------------|--------------------------------------------------|
| `<path_to_vault_gitlab_plugin>`| Mount path of the GitLab secrets plugin in Vault |
| `<gitlab_token_role>`          | Name of the role created in step 1               |
| `<vault_url>`                  | URL of your Vault server                         |
| `<kubernetes_mount_path>`      | Mount path of the Kubernetes auth method         |
| `<kubernetes_mount_role>`      | Vault role for Kubernetes auth                   |
| `<url_to_container_registry>`  | GitLab container registry URL                    |

```shell
kubectl apply -f manifest.yaml
```

This creates:

- A `VaultDynamicSecret` generator that requests a deploy token from Vault
- An `ExternalSecret` that uses the generator to create a `kubernetes.io/dockerconfigjson` secret, refreshed every 168 hours
