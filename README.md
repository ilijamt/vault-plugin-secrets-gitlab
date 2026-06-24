Vault Plugin for Gitlab Access Token
=====================================
[![Go Report Card](https://goreportcard.com/badge/github.com/ilijamt/vault-plugin-secrets-gitlab)](https://goreportcard.com/report/github.com/ilijamt/vault-plugin-secrets-gitlab)
[![Codecov](https://img.shields.io/codecov/c/gh/ilijamt/vault-plugin-secrets-gitlab)](https://app.codecov.io/gh/ilijamt/vault-plugin-secrets-gitlab)
[![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/ilijamt/vault-plugin-secrets-gitlab)](go.mod)
[![GitHub](https://img.shields.io/github/license/ilijamt/vault-plugin-secrets-gitlab)](LICENSE)
[![Release](https://img.shields.io/github/release/ilijamt/vault-plugin-secrets-gitlab.svg)](https://github.com/ilijamt/vault-plugin-secrets-gitlab/releases/latest)

This is a standalone backend plugin for use with Hashicorp Vault/OpenBao. It lets you automate the
creation and revocation of GitLab personal, project, and group access tokens through Vault.

**IMPORTANT**: Upgrading to >= 0.7.x will require you to revoke, remove all the paths, and remove the mount path. This is required because the paths internally have changed to accommodate config per role.

## Security model

The current authentication model requires providing Vault with a Gitlab Token.

## GitLab Support

- **GitLab CE/EE (Self-Managed)**
  - 17.11.7 (tested)
  - 18.11.5 (tested)
  - 19.0.2 (tested)
  - 19.1.1 (tested)
- **GitLab.com**
  - Personal access tokens and user service accounts are not supported
- **GitLab Dedicated**
  - Personal access tokens and user service accounts are not supported

## Quick links

- Vault Website – https://www.vaultproject.io
- OpenBao Website - https://openbao.org/

## Token types

> **All tiers** = Free + Premium + Ultimate · **All offerings** = GitLab.com + Self-Managed + Dedicated

| Token type | Tier | Offering | Status |
| --- | --- | --- | --- |
| [Personal Access Tokens](https://docs.gitlab.com/api/personal_access_tokens/) | All tiers | All offerings | GA |
| [Project Access Tokens](https://docs.gitlab.com/api/project_access_tokens/) | All tiers | All offerings | GA |
| [Group Access Tokens](https://docs.gitlab.com/api/group_access_tokens/) | All tiers | All offerings | GA |
| [User/Group/Project Service Account Tokens](https://docs.gitlab.com/api/service_accounts/)¹ | All tiers | All offerings | GA |
| [Pipeline Project Trigger Tokens](https://docs.gitlab.com/api/pipeline_triggers/) | All tiers | All offerings | GA |
| [Group/Project Deploy Tokens](https://docs.gitlab.com/user/project/deploy_tokens/) | All tiers | All offerings | GA |

¹ Service accounts on GitLab Free are capped: up to 100 per top-level group on GitLab.com, or 100 per instance on Self-Managed. Premium and Ultimate are unlimited.

### What each `token_type` does

Set `token_type` on the role. The `path` format and which fields apply depend on the type.

| `token_type` | Issues | `path` format | `scopes` | `access_level` | Min GitLab² |
| --- | --- | --- | :---: | :---: | :---: |
| `personal` | A personal access token for an existing user | `{username}` | yes | n/a | all |
| `project` | A project access token (project bot user) | `group/project` (or nested) | yes | yes | 13.10 |
| `group` | A group access token (group bot user) | `group` (or `group/subgroup`) | yes | yes | 14.7 |
| `user-service-account` | A PAT for an existing instance-level service account | `{username}` | yes | n/a | 16.1 |
| `group-service-account` | A PAT for an existing group/subgroup service account | `{groupId}/{serviceAccountName}` | yes | n/a | 16.1 |
| `project-service-account` | A PAT for an existing project service account | `{projectId}/{serviceAccountName}` | yes | n/a | 18.11 |
| `pipeline-project-trigger` | A pipeline trigger token | `group/project` (or nested) | n/a | n/a | all |
| `project-deploy` | A project deploy token | `group/project` (or nested) | yes | n/a | 12.9 |
| `group-deploy` | A group deploy token | `group` (or `group/subgroup`) | yes | n/a | 12.9 |

² Minimum GitLab version where the feature exists. `all` means it predates every supported version. The plugin itself is tested on 17.11.7, 18.11.5, 19.0.2 and 19.1.1.

### What this plugin does and does not do

| Capability | Supported | Notes |
| --- | :---: | --- |
| Create a token on demand for a role and return it via Vault | yes | |
| Revoke the token when the Vault lease expires | yes | |
| Let GitLab expire the token by TTL | yes | set `gitlab_revokes_token=true` |
| Auto-rotate the config token used to talk to GitLab | yes | set `auto_rotate_token` |
| Create the service account, user, project, or group | no | must already exist in GitLab |
| Rotate an already-issued token in place | no | request a new token by reading the role again (you or your automation drive this; Vault does not do it on its own) |
| Create `user-service-account` on GitLab.com (SaaS) or Dedicated | no | use `group-service-account` or `project-service-account` |

## Getting started

This is a [Vault plugin](https://www.vaultproject.io/docs/plugins/plugin-architecture#plugin-catalogs)
and is meant to work with Vault. This guide assumes you have already installed Vault
and have a basic understanding of how Vault works.

Otherwise, first read this guide on how to [get started with Vault](https://www.vaultproject.io/intro/getting-started/install.html).

To learn specifically about how plugins work, see documentation on [Vault plugins](https://www.vaultproject.io/docs/plugins/plugin-architecture#plugin-catalog).

### Quickstart

Register the plugin binary and enable it:

```shell
vault plugin register \
  -sha256=$(sha256sum path/to/plugin/directory/gitlab | cut -d " " -f 1) \
  -command=vault-plugin-secrets-gitlab \
  secret gitlab

vault secrets enable gitlab
```

Configure the backend and verify the config:

```shell
vault write gitlab/config/default base_url=https://gitlab.example.com token=gitlab-super-secret-token auto_rotate_token=false auto_rotate_before=48h type=self-managed
vault read gitlab/config/default
```

Create a role and request a token:

```shell
vault write gitlab/roles/personal name='{{ .role_name }}-{{ .token_type }}-{{ randHexString 4 }}' path=username scopes="read_api" token_type=personal ttl=48h
vault read gitlab/token/personal/username
```

## [Documentation](./docs/index.md)

- [Path overview and endpoint patterns](./docs/paths.md)
- [Runtime flags](./docs/flags.md)
- [Backend configuration](./docs/configuration.md)
- [Role configuration and templating](./docs/roles.md)
- [End-to-end examples](./docs/examples.md)
- [Install as an OpenBao OCI plugin](./docs/openbao-oci.md)
- [Upgrade guidance](./docs/upgrading.md)
- [Local development](./docs/development.md)

## Info

Running the logging with `debug` level will show sensitive information in the logs.
