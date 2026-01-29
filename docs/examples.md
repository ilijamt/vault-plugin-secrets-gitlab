Examples
========

## Config

Due to how Gitlab manages expiration the minimum is 24h and maximum is 365 days. As per
[non-expiring-access-tokens](https://docs.gitlab.com/ee/update/deprecations.html#non-expiring-access-tokens) and
[Remove ability to create deprecated non-expiring access tokens](https://gitlab.com/gitlab-org/gitlab/-/issues/392855).
Since Gitlab 16.0 the ability to create non expiring token has been removed.

If you use Vault to manage the tokens the minimal TTL you can use is `1h`, by setting `gitlab_revokes_token=false`.

The command below will set up the config backend with a max TTL of 48h.

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

## Roles

This will create multiple roles

```shell
$ vault write gitlab/roles/personal name='{{ .role_name }}-{{ .token_type }}-{{ randHexString 4 }}' path=username scopes="read_api" token_type=personal ttl=48h
$ vault write gitlab/roles/project name='{{ .role_name }}-{{ .token_type }}-{{ randHexString 4 }}' path=group/project scopes="read_api" access_level=guest token_type=project ttl=48h
$ vault write gitlab/roles/group name='{{ .role_name }}-{{ .token_type }}-{{ randHexString 4 }}' path=group/subgroup scopes="read_api" access_level=developer token_type=group ttl=48h
$ vault write gitlab/roles/sa name='{{ .role_name }}-{{ .token_type }}-{{ randHexString 4 }}' path=service_account_00b069cb73a15d0a7ba8cd67a653599c scopes="read_api" token_type=user-service-account ttl=24h
$ vault write gitlab/roles/ga name='{{ .role_name }}-{{ .token_type }}-{{ randHexString 4 }}' path=345/service_account_00b069cb73a15d0a7ba8cd67a653599c scopes="read_api" token_type=group-service-account ttl=24h
$ vault write gitlab/roles/personal-dynamic-path name='{{ .role_name }}-{{ .token_type }}-{{ randHexString 4 }}' path='ilija-.*' dynamic_path=true scopes="read_api" token_type=personal ttl=48h
```

### User service accounts

The service account users from Gitlab 16.1 are for all purposes users that don't use seats. So creating a service account and setting the path to the service account user would work the same as on a real user. More information can be found on https://docs.gitlab.com/ee/api/users.html#create-service-account-user.

```shell
$ curl --request POST --header "PRIVATE-TOKEN: $GITLAB_TOKEN" "https://gitlab/api/v4/service_accounts" | jq .
{
  "id": 63,
  "username": "service_account_964b157dcff9bcd87dc7c0837f9c47e9",
  "name": "Service account user"
}
```

### Group service accounts

The service account users from Gitlab 16.1 are for all purposes users that don't use seats. More information can be found on https://docs.gitlab.com/ee/api/group_service_accounts.html#create-a-service-account-user.

```shell
$ curl --request POST --header "PRIVATE-TOKEN: $GITLAB_TOKEN" "https://gitlab/api/v4/groups/345/service_accounts" | jq .
{
  "id": 61,
  "username": "service_account_group_345_c468757e6df2fc104de54ea470539bb5",
  "name": "Service account user"
}
```

## Revoke all created tokens by this plugin

```shell
$ vault lease revoke -prefix gitlab/
All revocation operations queued successfully!
```

## Force rotation of the main token

If the original token that has been supplied to the backend is not expired. We can use the endpoint below
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

Important: Token will be shown only after rotation, and it will not be shown again. (Unless the plugin is started with the `-show-config-token` flag)
