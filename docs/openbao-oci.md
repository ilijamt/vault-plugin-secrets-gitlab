Install as an OpenBao OCI plugin
================================

The plugin is published as a multi-arch OCI image (`linux/amd64`, `linux/arm64`)
at `ghcr.io/ilijamt/vault-plugin-secrets-gitlab`. OpenBao can pull and register
it from a
[declarative `plugin` block](https://openbao.org/docs/configuration/plugins/),
so you don't have to drop a binary into the plugin directory yourself. Each
release is tagged `vX.Y.Z`, and `latest` points to the newest.

## OpenBao server configuration

A complete working config:

```hcl
ui           = true
cluster_addr = "http://127.0.0.1:8201"
api_addr     = "http://127.0.0.1:8200"

listener "tcp" {
  address     = "127.0.0.1:8200"
  tls_disable = true
}

storage "inmem" {}

plugin "secret" "gitlab" {
  image       = "ghcr.io/ilijamt/vault-plugin-secrets-gitlab"
  version     = "v0.12.1"
  binary_name = "vault-plugin-secrets-gitlab"
  sha256sum   = "4fb9d72f5d176201cd6c61a8f6c422e4732ebfd5377ae0df120ee2c20cc758cd"
}

plugin_directory     = "/tmp/openbao/plugins"
plugin_auto_download = true
plugin_auto_register = true
```

Set `version` to the release you want, and `sha256sum` to the matching value
from that release's `vault-plugin-secrets-gitlab_<version>_BINARY_SHA256SUMS`
asset. OpenBao checks `sha256sum` against the extracted binary, not the image
digest, and `binary_name` must match the binary inside the image.

Running plugins from OCI images needs a container runtime on the OpenBao host.
See the
[OpenBao plugin documentation](https://openbao.org/docs/configuration/plugins/)
for runtime setup and the rest of the `plugin_*` directives.
