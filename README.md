Vault Plugin for Gitlab Access Token
------------------------------------
[![Go Report Card](https://goreportcard.com/badge/github.com/ilijamt/vault-plugin-secrets-gitlab)](https://goreportcard.com/report/github.com/ilijamt/vault-plugin-secrets-gitlab)
[![Codecov](https://img.shields.io/codecov/c/gh/ilijamt/vault-plugin-secrets-gitlab)](https://app.codecov.io/gh/ilijamt/vault-plugin-secrets-gitlab)
[![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/ilijamt/vault-plugin-secrets-gitlab)](go.mod)
[![GitHub](https://img.shields.io/github/license/ilijamt/vault-plugin-secrets-gitlab)](LICENSE)

This is a standalone backend plugin for use with Hashicorp Vault. This plugin allows for Gitlab to generate personal,
project and group access tokens. This was created so we can automate the creation/revocation of access tokens
through Vault.

## Quick Links

- Vault Website: [https://www.vaultproject.io]
- Gitlab Personal Access Tokens: [https://docs.gitlab.com/ee/api/personal_access_tokens.html]
- Gitlab Project Access Tokens: [https://docs.gitlab.com/ee/api/project_access_tokens.html]
- Gitlab Group Access Tokens: [https://docs.gitlab.com/ee/api/group_access_tokens.html]

## Getting Started

This is a [Vault plugin](https://www.vaultproject.io/docs/plugins/plugin-architecture#plugin-catalogs)
and is meant to work with Vault. This guide assumes you have already installed Vault
and have a basic understanding of how Vault works.

Otherwise, first read this guide on how to [get started with Vault](https://www.vaultproject.io/intro/getting-started/install.html).

To learn specifically about how plugins work, see documentation on [Vault plugins](https://www.vaultproject.io/docs/plugins/plugin-architecture#plugin-catalog).

### Setup

Before we can use this plugin we need to create an access token that will have rights to do what we need to.

## Security Model

The current authentication model requires providing Vault with a Gitlab Token. 

## Examples

### Setup

Before we can use the plugin we need to register and enable it in Vault.

```shell
vault plugin register \
  -sha256=$(sha256sum path/to/plugin/directory/gitlab | cut -d " " -f 1) \
  -command=vault-plugin-secrets-gitlab \
  secret gitlab

vault secrets enable gitlab
```

### Config

Due to how Gitlab manages expiration the minimum is 24h and maximum is 365 days. As per
[non-expiring-access-tokens](https://docs.gitlab.com/ee/update/deprecations.html#non-expiring-access-tokens) and
[Remove ability to create deprecated non-expiring access tokens](https://gitlab.com/gitlab-org/gitlab/-/issues/392855).
Since Gitlab 16.0 the ability to create non expiring token has been removed.

The command bellow will set up the config backend with a max TTL of 48h.

```shell
$ vault write gitlab/config max_ttl=48h base_url=https://gitlab.example.com token=gitlab-super-secret-token
```

### Roles

This will create three roles, one of each type.

```shell
$ vault write gitlab/roles/personal name=personal-token-name path=username scopes="read_api" token_type=personal token_ttl=24h
$ vault write gitlab/roles/project name=project-token-name path=group/project scopes="read_api" access_level=guest token_type=project token_ttl=24h
$ vault write gitlab/roles/group name=group-token-name path=group/subgroup scopes="read_api" access_level=developer token_type=group token_ttl=24h
```

### Get access tokens

#### Personal

```shell
$ vault read gitlab/token/personal
Key                Value
---                -----
lease_id           gitlab/token/personal/0FrzLFkRKaUNZSfa6WfFqjWK
lease_duration     20h1m37s
lease_renewable    false
access_level       n/a
created_at         2023-08-31T03:58:23.069Z
expires_at         2023-09-01T00:00:00Z
name               vault-generated-personal-access-token-227cb38b
path               username
scopes             [read_api]
token              7mbpSExz7ruyw1QgTjL-

$ vault lease revoke gitlab/token/personal/0FrzLFkRKaUNZSfa6WfFqjWK
All revocation operations queued successfully!
```

#### Group
```shell
$ vault read gitlab/token/group
Key                Value
---                -----
lease_id           gitlab/token/group/LqmL1MtuIlJ43N8q2L975jm8
lease_duration     20h14s
lease_renewable    false
access_level       developer
created_at         2023-08-31T03:59:46.043Z
expires_at         2023-09-01T00:00:00Z
name               vault-generated-group-access-token-913ab1f9
path               group/subgroup
scopes             [read_api]
token              rSYv4zwgP-2uaFEAsZyd

$ vault lease revoke gitlab/token/group/LqmL1MtuIlJ43N8q2L975jm8
All revocation operations queued successfully!
```

#### Project

```shell
$ vault read gitlab/token/project
Key                Value
---                -----
lease_id           gitlab/token/project/ZMSOrOHiP77l5kjWXq3zizPA
lease_duration     19h59m6s
lease_renewable    false
access_level       guest
created_at         2023-08-31T04:00:53.613Z
expires_at         2023-09-01T00:00:00Z
name               vault-generated-project-access-token-842113a6
path               group/project
scopes             [read_api]
token              YfRu42VaGGrxshKKwtma

$ vault lease revoke gitlab/token/project/ZMSOrOHiP77l5kjWXq3zizPA
All revocation operations queued successfully!
```

### Revoke all created tokens by this plugin
```shell
$ vault lease revoke -prefix gitlab/
All revocation operations queued successfully!
```

### Force rotation of the main token
If the original token that has been supplied to the backend is not expired. We can use the endpoint bellow
to force a rotation of the main token. This would create a new token with the same expiration as the original token.

```shell
vault put gitlab/config/rotate
```

## TODO

* [ ] Add tests against real Gitlab instance
