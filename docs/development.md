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
(`paths`, `e2e`, `saas`, `serviceaccount`). Each test replays an HTTP cassette
(`go-vcr`) recorded against a real GitLab instance.

Replays are self-contained: they need no setup or extra files. Each test reads
the deterministic "current time" it needs (for token TTL/expiry/rotation math)
from its own cassette: the `created_at` of the config token reported by `GET
/personal_access_tokens/self`, and falls back to the `expires_at` recorded in
the request body for the direct client tests. Run the suites with:

```shell
make test                                 # all versions, all tags
GITLAB_VERSION=18.11.2 go test -tags "paths e2e serviceaccount" ./tests/integration/...
```

Recording new cassettes talks to a real GitLab instance. The `paths`, `e2e` and
`serviceaccount` suites record against the per-version local stack provisioned by
`local-env/` (see [`local-env/README.md`](../local-env/README.md)), which also
writes the per-version token set to `tests/integration/testdata/tokens.<version>.json`.
That file is a recording-only artifact (git-ignored): replays fall back to a
placeholder token because the cassette matcher ignores authentication headers.
The `saas` suite records against `gitlab.com` and needs `GITLAB_COM_TOKEN`.
