Upgrading
=========

```shell
$ vault plugin register \
  -sha256=<sha256> \
  -command=gitlab\
  -version=v0.12.1 \
  secret \
  gitlab
$ vault secrets tune -plugin-version=v0.12.1 gitlab
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
   "plugin_version":"v0.12.1",
   "running_plugin_version":"v0.12.1",
   "running_sha256":"<sha256>",
   "deprecation_status":""
}
```
