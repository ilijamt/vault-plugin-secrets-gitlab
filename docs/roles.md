Roles
=====

## Role

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

### path

#### token_type is personal

Format of the path is `{username}` example `admin`.

#### token_type is project

Format of the path is the full path of the project for example `group/project` or `group/subgroup/project`

#### token_type is group

Format of the path is the full path of the project for example `group` or `group/subgroup`

#### token_type is user-service-account

Format of the path is `{username}` example `service_account_65c74d39b4f71fc3fdc72330fce28c28`.

#### token_type is group-service-account

Format of the path is `{groupId}/{serviceAccountName}` example `265/service_account_65c74d39b4f71fc3fdc72330fce28c28`.

#### token_type is project-deploy

Format of the path is the full path of the project for example `group/project` or `group/subgroup/project`

#### token_type is group-deploy

Format of the path is the full path of the project for example `group` or `group/subgroup`

#### token_type is pipeline-project-trigger

Format of the path is the full path of the project for example `group/project` or `group/subgroup/project`

### name

When generating a token, you have control over the token's name by using templating. The name is constructed using Go's [text/template](https://pkg.go.dev/text/template), which allows for dynamic generation of names based on available data. You can refer to Go's [text/template](https://pkg.go.dev/text/template#hdr-Examples) documentation for examples and guidance on how to use it effectively.

Important: GitLab does not permit duplicate token names. If your template doesn't ensure unique names, token generation will fail.

Here are some examples of effective token name templates:

* `vault-generated-{{ .token_type }}-access-token-{{ randHexString 4 }}`
* `{{ .role_name }}-{{ .token_type }}-{{ randHexString 4 }}`

#### Data

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

#### Functions

You can also use the following functions within your template:

* `randHexString(bytes int) string` - Generates a random hexadecimal string with the specified number of bytes.
* `stringsJoin(elems []string, sep string) string` - joins a list of `elems` strings with a `sep`
* `yesNoBool(in bool) string` - just return `yes` if `in` is true otherwise it returns `no`
* `timeNowFormat(layout string) string` - layout is a go time format string layout
* `stringsSplit(elems string, sep string) string` - splits a string `elems` with a `sep`
* `trimSpace(s string) string` - trims the space from a string
* `stringsReplace(s, old, new string, n int) string` - runs replace on the string

### ttl

Depending on `gitlab_revokes_token` the TTL will change.

* `true` - 24h <= ttl <= 365 days
* `false` - 1h <= ttl <= 365 days

### access_level

It's not required if `token_type` is set to `personal`, `pipeline-project-trigger`, `project-deploy`, `group-deploy`.

For a list of available roles check https://docs.gitlab.com/ee/user/permissions.html

### scopes

It's not required if `token_type` is set to `pipeline-project-trigger`.

Depending on the type of token you have different scopes:

* Personal - https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html#personal-access-token-scopes
* Project - https://docs.gitlab.com/ee/user/project/settings/project_access_tokens.html#scopes-for-a-project-access-token
* Group - https://docs.gitlab.com/ee/user/group/settings/group_access_tokens.html#scopes-for-a-group-access-token
* Deploy - https://docs.gitlab.com/ee/user/project/deploy_tokens/#scope

### token_types

Can be

* personal
* project
* group
* user-service-account
* group-service-account
* pipeline-project-trigger
* project-deploy
* group-deploy

### gitlab_revokes_token

This is a flag that doesn't expire the token when the token used to create the credentials expire.
When the vault token used to create gitlab credentials with a TTL longer than the vault token, the new gitlab credentials will expire at the same time with the parent.
Setting this up will not call the revoke endpoint on gitlab.
