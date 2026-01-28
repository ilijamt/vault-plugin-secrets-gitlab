Vault Plugin for Gitlab Access Token
------------------------------------
[![Go Report Card](https://goreportcard.com/badge/github.com/ilijamt/vault-plugin-secrets-gitlab)](https://goreportcard.com/report/github.com/ilijamt/vault-plugin-secrets-gitlab)
[![Codecov](https://img.shields.io/codecov/c/gh/ilijamt/vault-plugin-secrets-gitlab)](https://app.codecov.io/gh/ilijamt/vault-plugin-secrets-gitlab)
[![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/ilijamt/vault-plugin-secrets-gitlab)](go.mod)
[![GitHub](https://img.shields.io/github/license/ilijamt/vault-plugin-secrets-gitlab)](LICENSE)
[![Release](https://img.shields.io/github/release/ilijamt/vault-plugin-secrets-gitlab.svg)](https://github.com/ilijamt/vault-plugin-secrets-gitlab/releases/latest)

This is a standalone backend plugin for use with Hashicorp Vault. This plugin allows for Gitlab to generate personal,
project and group access tokens. This was created so we can automate the creation/revocation of access tokens
through Vault.

**IMPORTANT**: Upgrading to >= 0.7.x will require you to revoke, remove all the paths, and remove the mount path. This is required because the paths internally have changed to accomodate config per role.

## Quick Links

- Vault Website - https://www.vaultproject.io
- Gitlab Personal Access Tokens - https://docs.gitlab.com/ee/api/personal_access_tokens.html
- Gitlab Project Access Tokens - https://docs.gitlab.com/ee/api/project_access_tokens.html
- Gitlab Group Access Tokens - https://docs.gitlab.com/ee/api/group_access_tokens.html
- Gitlab User Service Account Tokens - https://docs.gitlab.com/api/user_service_accounts/
- Gitlab Group Service Account Tokens - https://docs.gitlab.com/ee/api/group_service_accounts.html
- Gitlab Pipeline Project Trigger Tokens - https://docs.gitlab.com/ee/api/pipeline_triggers.html
- Gitlab Group/Project Deploy Tokens - https://docs.gitlab.com/ee/user/project/deploy_tokens

## Getting Started

This is a [Vault plugin](https://www.vaultproject.io/docs/plugins/plugin-architecture#plugin-catalogs)
and is meant to work with Vault. This guide assumes you have already installed Vault
and have a basic understanding of how Vault works.

Otherwise, first read this guide on how to [get started with Vault](https://www.vaultproject.io/intro/getting-started/install.html).

To learn specifically about how plugins work, see documentation on [Vault plugins](https://www.vaultproject.io/docs/plugins/plugin-architecture#plugin-catalog).

## GitLab

- GitLab CE/EE - Self Managed (tested against 17.10.3)
- gitlab.com (cannot use personal access token, and user service account)
- Dedicated Instance (cannot use personal access token, and user service account)

### Setup

Before we can use this plugin we need to create an access token that will have rights to do what we need to.

## Paths

For a list of the available endpoints you can check bellow or by running the command `vault path-help gitlab` for your version after you've mounted it.

```shell
$ vault path-help gitlab
## DESCRIPTION

The Gitlab Access token auth Backend dynamically generates private
and group tokens.

After mounting this Backend, credentials to manage Gitlab tokens must be configured
with the "^config/(?P<config_name>\w(([\w-.@]+)?\w)?)$" endpoints.

## PATHS

The following paths are supported by this backend. To view help for
any of the paths below, use the help command with any route matching
the path pattern. Note that depending on the policy of your auth token,
you may or may not be able to access certain paths.

    ^config/(?P<config_name>\w(([\w-.]+)?\w)?)$
        Configure the Gitlab Access Tokens Backend.

    ^config/(?P<config_name>\w(([\w-.]+)?\w)?)/rotate$
        Rotate the gitlab token for this configuration.

    ^config?/?$
        Lists existing configs

    ^flags$
        Flags for the plugins.

    ^roles/(?P<role_name>\w(([\w-.]+)?\w)?)$
        Create a role with parameters that are used to generate a various access tokens.

    ^roles?/?$
        Lists existing roles

    ^token/(?P<role_name>\w(([\w-.]+)?\w)?)$
        Generate an access token based on the specified role
```
## Flags

There are some flags we can specify to enable/disable some functionality in the vault plugin.


|               Flag                | Default value | Changable during runtime if `allow-runtime-flags-change` is set to `true` | Description                                                                            |
|:---------------------------------:|:-------------:|:-------------------------------------------------------------------------:|----------------------------------------------------------------------------------------|
|         show-config-token         |     false     |                                   true                                    | Display the token value when reading a config on it's endpoint like `/config/default`. |
|    allow-runtime-flags-change     |     false     |                                   false                                   | Allows you to change the flags at runtime                                              |

## Security Model

The current authentication model requires providing Vault with a Gitlab Token. 

## Configuration

### Config

|      Property      | Required | Default value | Sensitive | Description                                                                                                                                   |
|:------------------:|:--------:|:-------------:|:---------:|:----------------------------------------------------------------------------------------------------------------------------------------------|
|       token        |   yes    |      n/a      |    yes    | The token to access Gitlab API, it will not show when you do a read, as it's a sensitive value. Instead it will display it's SHA1 hash value. |
|      base_url      |   yes    |      n/a      |    no     | The address to access Gitlab                                                                                                                  |
| auto_rotate_token  |    no    |      no       |    no     | Should we autorotate the token when it's close to expiry? (Experimental)                                                                      |
| auto_rotate_before |    no    |      24h      |    no     | How much time should be remaining on the token validity before we should rotate it? Minimum can be set to 24h and maximum to 730h             |
|        type        |   yes    |      n/a      |    no     | The type of gitlab instance that we use can be one of saas, self-managed or dedicated                                                         |

### Role

|       Property       | Required | Default value | Sensitive | Description                                                                                                                                                                                                         |
|:--------------------:|:--------:|:-------------:|:---------:|:--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
|         path         |   yes    |      n/a      |    no     | Project/Group path to create an access token for. If the token type is set to personal then write the username here. If `dynamic_path` is set to true this needs to be a regex.                                     |
|         name         |   yes    |      n/a      |    no     | The name of the access token                                                                                                                                                                                        |
|         ttl          |   yes    |      n/a      |    no     | The TTL of the token                                                                                                                                                                                                |
|     access_level     |  no/yes  |      n/a      |    no     | Access level of access token (only required for Group and Project access tokens)                                                                                                                                    |
|        scopes        |    no    |      []       |    no     | List of scopes                                                                                                                                                                                                      |
|      token_type      |   yes    |      n/a      |    no     | Access token type                                                                                                                                                                                                   |
| gitlab_revokes_token |    no    |      no       |    no     | Gitlab revokes the token when it's time. Vault will not revoke the token when the lease expires                                                                                                                     |
|     config_name      |    no    |    default    |    no     | The configuration to use for the role                                                                                                                                                                               |
|     dynamic_path     |    no    |     false     |    no     | If set to true, you will be able to use the regex pattern to match the path from the role path                                                                                                                      |

#### path

##### `token_type` is `personal`

Format of the path is `{username}` example `admin`.

##### `token_type` is `project`

Format of the path is the full path of the project for example `group/project` or `group/subgroup/project`

##### `token_type` is `group`

Format of the path is the full path of the project for example `group` or `group/subgroup`

##### `token_type` is `user-service-account`

Format of the path is `{username}` example `service_account_65c74d39b4f71fc3fdc72330fce28c28`.

##### `token_type` is `group-service-account`

Format of the path is `{groupId}/{serviceAccountName}` example `265/service_account_65c74d39b4f71fc3fdc72330fce28c28`.

##### `token_type` is `project-deploy`

Format of the path is the full path of the project for example `group/project` or `group/subgroup/project`

##### `token_type` is `group-deploy`

Format of the path is the full path of the project for example `group` or `group/subgroup`

##### `token_type` is `pipeline-project-trigger`

Format of the path is the full path of the project for example `group/project` or `group/subgroup/project`

#### name

When generating a token, you have control over the token's name by using templating. The name is constructed using Go's [text/template](https://pkg.go.dev/text/template), which allows for dynamic generation of names based on available data. You can refer to Go's [text/template](https://pkg.go.dev/text/template#hdr-Examples) documentation for examples and guidance on how to use it effectively.

**Important**: GitLab does not permit duplicate token names. If your template doesn't ensure unique names, token generation will fail.

Here are some examples of effective token name templates:

* `vault-generated-{{ .token_type }}-access-token-{{ randHexString 4 }}`
* `{{ .role_name }}-{{ .token_type }}-{{ randHexString 4 }}`

##### Data

The following data points can be used within your token name template. These are derived from the role for which the token is being generated:

* path (using this with `dynamic_path` can lead to unexpected results)
* ttl
* access_level
* scopes (csv string ex: api, sudo, read_api)
* token_type
* role_name
* config_name
* gitlab_revokes_token
* unix_timestamp_utc

##### Functions

You can also use the following functions within your template:

* `randHexString(bytes int) string` - Generates a random hexadecimal string with the specified number of bytes.
* `stringsJoin(elems []string, sep string) string` - joins a list of `elems` strings with a `sep`
* `yesNoBool(in bool) string` - just return `yes` if `in` is true otherwise it returns `no`
* `timeNowFormat(layout string) string` - layout is a go time format string layout
* `stringsSplit(elems string, sep string) string` - splits a string `elems` with a `sep`
* `trimSpace(s string) string` - trims the space from a string
* `stringsReplace(s, old, new string, n int) string` - runs replace on the string

#### ttl

Depending on `gitlab_revokes_token` the TTL will change.

* `true` - 24h <= ttl <= 365 days
* `false` - 1h <= ttl <= 365 days

#### access_level 

It's not required if `token_type` is set to `personal`, `pipeline-project-trigger`, `project-deploy`, `group-deploy`.

For a list of available roles check https://docs.gitlab.com/ee/user/permissions.html

#### scopes

It's not required if `token_type` is set to `pipeline-project-trigger`.

Depending on the type of token you have different scopes:

* `Personal` - https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html#personal-access-token-scopes
* `Project` - https://docs.gitlab.com/ee/user/project/settings/project_access_tokens.html#scopes-for-a-project-access-token
* `Group` - https://docs.gitlab.com/ee/user/group/settings/group_access_tokens.html#scopes-for-a-group-access-token
* `Deploy` - https://docs.gitlab.com/ee/user/project/deploy_tokens/#scope

#### token_types

Can be 

* personal
* project
* group
* user-service-account
* group-service-account
* pipeline-project-trigger
* project-deploy
* group-deploy

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
$ vault write gitlab/config/default base_url=https://gitlab.example.com token=gitlab-super-secret-token auto_rotate_token=false auto_rotate_before=48h type=self-managed
$ vault read gitlab/config/default
Key                   Value
---                   -----
auto_rotate_before    48h0m0s
auto_rotate_token     false
base_url              http://localhost:8080
name                  default
scopes                api, read_api, read_user, sudo, admin_mode, create_runner, k8s_proxy, read_repository, write_repository, ai_features, read_service_ping
token_created_at      2024-07-11T18:53:26Z
token_expires_at      2025-07-11T00:00:00Z
token_id              1
token_sha1_hash       9441e6e07d77a2d5601ab5d7cac5868d358d885c
type                  self-managed
gitlab_version        17.5.3-ee
gitlab_revision       9d81c27eee7
gitlab_is_enterprise  true
```

After initial setup should you wish to change any value you can do so by using the patch command for example

```shell
$ vault patch gitlab/config/default type=saas auto_rotate_token=true auto_rotate_before=64h token=glpat-secret-admin-token
Key                   Value
---                   -----
auto_rotate_before    64h0m0s
auto_rotate_token     true
base_url              http://localhost:8080
name                  default
scopes                api, read_api, read_user, sudo, admin_mode, create_runner, k8s_proxy, read_repository, write_repository, ai_features, read_service_ping
token_created_at      2024-07-11T18:53:46Z
token_expires_at      2025-07-11T00:00:00Z
token_id              2
token_sha1_hash       c6e762667cadb936f0c8439b0d240661a270eba1
type                  saas
gitlab_version        17.7.0-pre
gitlab_revision       22e9474dc6b
gitlab_is_enterprise  true
```

All the config properties as defined above in the Config section can be patched.

You may also need to configure the Max/Default TTL for a token that can be issued by setting:

Max TTL: `1 year`
Default TTL: `1 week`

```shell
$ vault secrets tune -max-lease-ttl=8784h -default-lease-ttl=168h gitlab/
```

Check https://developer.hashicorp.com/vault/docs/commands/secrets/tune for more information.

There is a periodic func that runs that is responsible for autorotation and main token expiry time. 

### Roles

This will create multiple roles

```shell
$ vault write gitlab/roles/personal name='{{ .role_name }}-{{ .token_type }}-{{ randHexString 4 }}' path=username scopes="read_api" token_type=personal ttl=48h
$ vault write gitlab/roles/project name='{{ .role_name }}-{{ .token_type }}-{{ randHexString 4 }}' path=group/project scopes="read_api" access_level=guest token_type=project ttl=48h
$ vault write gitlab/roles/group name='{{ .role_name }}-{{ .token_type }}-{{ randHexString 4 }}' path=group/subgroup scopes="read_api" access_level=developer token_type=group ttl=48h
$ vault write gitlab/roles/sa name='{{ .role_name }}-{{ .token_type }}-{{ randHexString 4 }}' path=service_account_00b069cb73a15d0a7ba8cd67a653599c scopes="read_api" token_type=user-service-account ttl=24h
$ vault write gitlab/roles/ga name='{{ .role_name }}-{{ .token_type }}-{{ randHexString 4 }}' path=345/service_account_00b069cb73a15d0a7ba8cd67a653599c scopes="read_api" token_type=group-service-account ttl=24h
$ vault write gitlab/roles/personal-dynamic-path name='{{ .role_name }}-{{ .token_type }}-{{ randHexString 4 }}' path='ilija-.*' dynamic_path=true scopes="read_api" token_type=personal ttl=48h
```

#### User service accounts

The service account users from Gitlab 16.1 are for all purposes users that don't use seats. So creating a service account and setting the path to the service account user would work the same as on a real user. More information can be found on https://docs.gitlab.com/ee/api/users.html#create-service-account-user.

```shell
$ curl --request POST --header "PRIVATE-TOKEN: $GITLAB_TOKEN" "https://gitlab/api/v4/service_accounts" | jq .
{
  "id": 63,
  "username": "service_account_964b157dcff9bcd87dc7c0837f9c47e9",
  "name": "Service account user"
}
```

#### Group service accounts

The service account users from Gitlab 16.1 are for all purposes users that don't use seats. More information can be found on https://docs.gitlab.com/ee/api/group_service_accounts.html#create-a-service-account-user.

```shell
$ curl --request POST --header "PRIVATE-TOKEN: $GITLAB_TOKEN" "https://gitlab/api/v4/groups/345/service_accounts" | jq .
{
  "id": 61,
  "username": "service_account_group_345_c468757e6df2fc104de54ea470539bb5",
  "name": "Service account user"
}

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
$ vault write -f gitlab/config/default/rotate
Key                   Value
---                   -----
auto_rotate_before    48h0m0s
auto_rotate_token     false
base_url              http://localhost:8080
name                  default
scopes                api, read_api, read_user, sudo, admin_mode, create_runner, k8s_proxy, read_repository, write_repository, ai_features, read_service_ping
token                 glpat-VhoWnWJ7RYwE78dn7Nsj
token_created_at      2024-10-15T12:57:47Z
token_expires_at      2025-10-15T00:00:00Z
token_id              43
token_sha1_hash       91a91bb30f816770081c570504c5e2723bcb1f38
type                  self-managed
```

**Important**: Token will be shown only after rotation, and it will not be shown again. (Unless the plugin is started with the `-show-config-token` flag)

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

Running the logging with `debug` level will show sensitive information in the logs.
