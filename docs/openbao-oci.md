Install as an OpenBao OCI plugin
================================

The plugin is published as a multi-arch OCI image so OpenBao can pull and
register it from a
[declarative `plugin` block](https://openbao.org/docs/configuration/plugins/),
instead of you placing a binary in the plugin directory. Images are available
from release `v0.x.y` onward.

The image lives at `ghcr.io/ilijamt/vault-plugin-secrets-gitlab`. Each release
is tagged `vX.Y.Z`, and `latest` points at the newest release. Builds cover
`linux/amd64` and `linux/arm64`.

The image is `FROM scratch` and holds only the static plugin binary at
`/vault-plugin-secrets-gitlab`. That binary is the same one attached to the
GitHub release, repackaged rather than recompiled, so its SHA256 matches the
value in the release's `*_BINARY_SHA256SUMS` asset.

## OpenBao server configuration

```hcl
# Allow OpenBao to download and register declaratively configured plugins.
plugin_auto_download = true
plugin_auto_register = true

plugin "secret" "gitlab" {
  image       = "ghcr.io/ilijamt/vault-plugin-secrets-gitlab"
  version     = "v0.x.y"
  binary_name = "vault-plugin-secrets-gitlab"
  sha256sum   = "<sha256 from the release's BINARY_SHA256SUMS>"
}
```

`binary_name` must match the file inside the image (`vault-plugin-secrets-gitlab`).
`sha256sum` is checked against the extracted binary, not the image digest, so
copy the value for your platform from the release's
`vault-plugin-secrets-gitlab_<version>_BINARY_SHA256SUMS` asset.

Running plugins from OCI images needs a container runtime on the OpenBao host.
See the
[OpenBao plugin documentation](https://openbao.org/docs/configuration/plugins/)
for runtime setup and the rest of the `plugin_*` directives.
