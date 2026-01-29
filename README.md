Vault Plugin for Gitlab Access Token
=====================================
[![Go Report Card](https://goreportcard.com/badge/github.com/ilijamt/vault-plugin-secrets-gitlab)](https://goreportcard.com/report/github.com/ilijamt/vault-plugin-secrets-gitlab)
[![Codecov](https://img.shields.io/codecov/c/gh/ilijamt/vault-plugin-secrets-gitlab)](https://app.codecov.io/gh/ilijamt/vault-plugin-secrets-gitlab)
[![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/ilijamt/vault-plugin-secrets-gitlab)](go.mod)
[![GitHub](https://img.shields.io/github/license/ilijamt/vault-plugin-secrets-gitlab)](LICENSE)
[![Release](https://img.shields.io/github/release/ilijamt/vault-plugin-secrets-gitlab.svg)](https://github.com/ilijamt/vault-plugin-secrets-gitlab/releases/latest)

This is a standalone backend plugin for use with Hashicorp Vault. This plugin allows for Gitlab to generate personal,
project and group access tokens. This was created so we can automate the creation/revocation of access tokens
through Vault.

**IMPORTANT**: Upgrading to >= 0.7.x will require you to revoke, remove all the paths, and remove the mount path. This is required because the paths internally have changed to accomodate config per role.

## Security model

The current authentication model requires providing Vault with a Gitlab Token.

## GitLab support

- GitLab CE/EE - Self Managed (tested against 17.10.3)
- gitlab.com (cannot use personal access token, and user service account)
- Dedicated Instance (cannot use personal access token, and user service account)

## Quick links

- Vault Website - https://www.vaultproject.io
- Gitlab Personal Access Tokens - https://docs.gitlab.com/ee/api/personal_access_tokens.html
- Gitlab Project Access Tokens - https://docs.gitlab.com/ee/api/project_access_tokens.html
- Gitlab Group Access Tokens - https://docs.gitlab.com/ee/api/group_access_tokens.html
- Gitlab User Service Account Tokens - https://docs.gitlab.com/api/user_service_accounts/
- Gitlab Group Service Account Tokens - https://docs.gitlab.com/ee/api/group_service_accounts.html
- Gitlab Pipeline Project Trigger Tokens - https://docs.gitlab.com/ee/api/pipeline_triggers.html
- Gitlab Group/Project Deploy Tokens - https://docs.gitlab.com/ee/user/project/deploy_tokens

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
- [Upgrade guidance](./docs/upgrading.md)
- [Local development](./docs/development.md)

## Info

Running the logging with `debug` level will show sensitive information in the logs.
