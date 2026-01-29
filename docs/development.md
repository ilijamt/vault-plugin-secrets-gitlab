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
