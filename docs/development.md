Local development
=================

Start vault with, this should create a dev server on port 8200

```shell
make vault-dev
```

And then enable the plugin by running:

```shell
make vault-plugin-enable
```

To configure the plugin run:

```shell
vault write gitlab/config/default base_url=http://localhost:8080/ token=glpat-wU8yWBGat-nypZcyf1LL auto_rotate_token=false auto_rotate_before=48h type=self-managed
vault write gitlab/roles/pdp name='{{ .role_name }}-{{ .token_type }}-{{ randHexString 4 }}' path='.*' dynamic_path=true scopes="read_api" token_type=personal ttl=48h
```

Then you can request the token for the role you created:

```shell
vault read gitlab/token/pdp/root
vault read gitlab/token/pdp/admin-user
vault read gitlab/token/pdp/normal-user
```

Integration tests
=================

Integration tests live in `tests/integration/` and are gated by build tags
(`unit`, `local`, `saas`, `selfhosted`). The `saas` and `selfhosted` suites
talk to real GitLab instances and need a few environment variables plus
prepared testdata files that pin the `created_at` timestamp of the token
used to authenticate.

Prepare the testdata before running the suites:

For SaaS (`gitlab.com`):

```shell
curl --silent --header "PRIVATE-TOKEN: $GITLAB_COM_TOKEN" \
  "https://gitlab.com/api/v4/personal_access_tokens/self" \
  | jq -rj '.created_at' > tests/integration/testdata/gitlab-com
```

For self-hosted:

```shell
curl --silent --header "PRIVATE-TOKEN: $GITLAB_SERVICE_ACCOUNT_TOKEN" \
  "$GITLAB_SERVICE_ACCOUNT_URL/api/v4/personal_access_tokens/self" \
  | jq -rj '.created_at' > tests/integration/testdata/gitlab-selfhosted
```

The required environment variables are:

- `GITLAB_COM_TOKEN` — personal access token for `gitlab.com` (SaaS suite).
- `GITLAB_SERVICE_ACCOUNT_URL` — host of the self-hosted GitLab instance
  (without scheme), e.g. `gitlab.example.com`.
- `GITLAB_SERVICE_ACCOUNT_TOKEN` — token for the self-hosted instance.

Or, with the env vars exported, run the Makefile targets:

```shell
make fetch-token-timestamps              # both
make fetch-token-timestamps-saas         # SaaS only
make fetch-token-timestamps-selfhosted   # self-hosted only
```
