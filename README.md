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

## GitLab

- GitLab CE/EE
- gitlab.com (cannot use personal access token)
- Dedicated Instance (cannot use personal access token)

### Setup

Before we can use this plugin we need to create an access token that will have rights to do what we need to.

## Security Model

The current authentication model requires providing Vault with a Gitlab Token. 

## Configuration

### Config

|         Property          | Required | Default value | Sensitive | Description                                                                                                                                   |
|:-------------------------:|:--------:|:-------------:|:---------:|:----------------------------------------------------------------------------------------------------------------------------------------------|
|           token           |   yes    |      n/a      |    yes    | The token to access Gitlab API, it will not show when you do a read, as it's a sensitive value. Instead it will display it's SHA1 hash value. |
|         base_url          |   yes    |      n/a      |    no     | The address to access Gitlab                                                                                                                  |
|     auto_rotate_token     |    no    |      no       |    no     | Should we autorotate the token when it's close to expiry? (Experimental)                                                                      |
|    auto_rotate_before     |    no    |      24h      |    no     | How much time should be remaining on the token validity before we should rotate it? Minimum can be set to 24h and maximum to 730h             |

### Role

|       Property       | Required | Default value | Sensitive | Description                                                                                                          |
|:--------------------:|:--------:|:-------------:|:---------:|:---------------------------------------------------------------------------------------------------------------------|
|         path         |   yes    |      n/a      |    no     | Project/Group path to create an access token for. If the token type is set to personal then write the username here. |
|         name         |   yes    |      n/a      |    no     | The name of the access token                                                                                         |
|         ttl          |   yes    |      n/a      |    no     | The TTL of the token                                                                                                 |
|     access_level     |  no/yes  |      n/a      |    no     | Access level of access token (only required for Group and Project access tokens)                                     |
|        scopes        |    no    |      []       |    no     | List of scopes                                                                                                       |
|      token_type      |   yes    |      n/a      |    no     | Access token type                                                                                                    |
| gitlab_revokes_token |    no    |      no       |    no     | Gitlab revokes the token when it's time. Vault will not revoke the token when the lease expires                      |

#### ttl

Depending on `gitlab_revokes_token` the TTL will change.

* `true` - 24h <= ttl <= 365 days
* `false` - 1h <= ttl <= 365 days

#### access_level 

It's not required if `token_type` is set to `personal`. 

For a list of available roles check https://docs.gitlab.com/ee/user/permissions.html

#### scopes

Depending on the type of token you have different scopes:

* `Personal` - https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html#personal-access-token-scopes
* `Project` - https://docs.gitlab.com/ee/user/project/settings/project_access_tokens.html#scopes-for-a-project-access-token
* `Group` - https://docs.gitlab.com/ee/user/group/settings/group_access_tokens.html#scopes-for-a-group-access-token

#### token_types

Can be 

* personal
* project
* group

#### gitlab_revokes_token

This is a flag that doesn't expire the token when the token used to create the credentials expire.
When the vault token used to create gitlab credentials with a TTL longer than the vault token, the new gitlab credentials will expire at the same time with the parent.
Setting this up will not call the revoke endpoint on gitlab.

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

If you use Vault to manage the tokens the minimal TTL you can use is `1h`, by setting `gitlab_revokes_token=false`.

The command bellow will set up the config backend with a max TTL of 48h.

```shell
$ vault write gitlab/config base_url=https://gitlab.example.com token=gitlab-super-secret-token auto_rotate_token=false auto_rotate_before=48h
$ vault read gitlab/config
Key                   Value
---                   -----
auto_rotate_before    48h0m0s
auto_rotate_token     false
base_url              https://gitlab.example.com
token_id              107
token_expires_at      2025-03-29T00:00:00Z
token_sha1_hash       1014647cd9bbf359d926fcacdf78e184db9dbedc
```

You may also need to configure the Max/Default TTL for a token that can be issued by setting:

Max TTL: `1 year`
Default TTL: `1 week`

```shell
$ vault secrets tune -max-lease-ttl=8784h -default-lease-ttl=168h gitlab/
```

Check https://developer.hashicorp.com/vault/docs/commands/secrets/tune for more information.

There is a periodic func that runs that is responsible for autorotation and main token expiry time. 
So in the beginning you may see  `token_expires_at n/a`. But when the function runs it will update itself 
with the correct expiry date and the corresponding `token_id`.

### Roles

This will create three roles, one of each type.

```shell
# personal access tokens can only be created by Gitlab Administrators (see https://docs.gitlab.com/ee/api/users.html#create-a-personal-access-token)
$ vault write gitlab/roles/personal name=personal-token-name path=username scopes="read_api" token_type=personal ttl=48h

$ vault write gitlab/roles/project name=project-token-name path=group/project scopes="read_api" access_level=guest token_type=project ttl=48h

$ vault write gitlab/roles/group name=group-token-name path=group/subgroup scopes="read_api" access_level=developer token_type=group ttl=48h
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
##### Service accounts
The service account users from Gitlab 16.1 are for all purposes users that don't use seats. So creating a service account and setting the path to the service account user would work the same as on a real user.
```shell
$ curl --request POST --header "PRIVATE-TOKEN: $GITLAB_TOKEN" "https://gitlab/api/v4/service_accounts" | jq .
{
  "id": 2017,
  "username": "service_account_00b069cb73a15d0a7ba8cd67a653599c",
  "name": "Service account user",
  "state": "active",
  "avatar_url": "https://secure.gravatar.com/avatar/6faa2758127182d391be18b4c1e36630?s=80&d=identicon",
  "web_url": "https://gitlab/service_account_00b069cb73a15d0a7ba8cd67a653599c"
}
```

In this case you would create a role like
```shell
$ vault write gitlab/roles/sa name=sa-name path=service_account_00b069cb73a15d0a7ba8cd67a653599c scopes="read_api" token_type=personal token_ttl=24h
$ vault read gitlab/token/sa
vault read gitlab/token/sa

Key                Value
---                -----
lease_id           gitlab/token/sa/oFI2vpUdvykvMgNum6pZReYZ
lease_duration     20h1m37s
lease_renewable    false
access_level       n/a
created_at         2023-08-31T03:58:23.069Z
expires_at         2023-09-01T00:00:00Z
name               vault-generated-personal-access-token-f6417198
path               service_account_00b069cb73a15d0a7ba8cd67a653599c
scopes             [api read_api read_repository read_registry]
token              -senkScjDo-SoGwST9PP
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
$ vault write -f gitlab/config/rotate
Key                   Value
---                   -----
auto_rotate_before    48h0m0s
auto_rotate_token     false
base_url              https://gitlab.example.com
token_expires_at      2025-03-29T00:00:00Z
token_id              110
token_sha1_hash       b8ff3f9e560f29d15f756fc92a3b1d6602aaae55
```
## Upgrading

```shell
$ vault plugin register \
  -sha256=b5fd0a3481930211a09bb944aa96a18a9eab8e594b6773b25209330d752e5f83 \
  -command=gitlab\
  -version=v0.2.4 \
  secret \
  gitlab
$ vault secrets tune -plugin-version=v0.2.4 gitlab
$ vault plugin reload -plugin gitlab
$ vault secrets list -detailed -format=json | jq '."gitlab/"'
{
   "uuid":"759239c4-5fe1-4eb0-6105-480d1d67de5e",
   "type":"gitlab",
   "description":"",
   "accessor":"gitlab_294d3aea",
   "config":{
      "default_lease_ttl":2678400,
      "max_lease_ttl":31622400,
      "force_no_cache":false
   },
   "options":null,
   "local":false,
   "seal_wrap":false,
   "external_entropy_access":false,
   "plugin_version":"v0.2.4",
   "running_plugin_version":"v0.2.4",
   "running_sha256":"b5fd0a3481930211a09bb944aa96a18a9eab8e594b6773b25209330d752e5f83",
   "deprecation_status":""
}
```
## Info

Running the logging with `debug` level will shows sensitive information in the logs.
