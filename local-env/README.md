## Local Environment

Runs a local GitLab CE instance via Docker Compose for integration testing. The compose file has a single `web` service whose image tag is taken from `GITLAB_IMAGE_TAG` (default `17.11.7-ce.0`), so only one version runs at a time. Switch versions by re-running `initial-setup.sh` with a different argument — it tears down the volumes and brings up the new tag.

**Prerequisites:** Docker, Terraform, `curl`, `jq`

### Setup

```bash
# default: 17.11.7
bash initial-setup.sh

# explicit version
bash initial-setup.sh 17.11.7
bash initial-setup.sh 18.11.2
```

This kills any existing containers, brings up GitLab on `localhost:8080` for the selected version, creates an admin access token on `root`, runs Terraform (`tf/`) to provision users/groups/projects/tokens, and writes the generated tokens to `../tests/integration/testdata/tokens.<version>.json`. A volume backup at `backup.<version>.tar` is created at the end. Initial setup can take several minutes while GitLab boots.

### Backup and Restore

```bash
bash backup-volumes.sh 17.11.7   # save current state to backup.17.11.7.tar
bash restore-volumes.sh 17.11.7  # reset to post-setup state from backup.17.11.7.tar
```

The version argument selects which backup tarball to read/write. Use `restore` to reset the GitLab instance after running tests or making changes.

### Recording cassettes

Tests record HTTP fixtures into `tests/integration/testdata/{unit,local}/<version>/`. Tokens for that version live in `tests/integration/testdata/tokens.<version>.json`. To record against a fresh setup:

```bash
# pinned (17.11.7)
bash initial-setup.sh 17.11.7
GITLAB_VERSION=17.11.7 GITLAB_URL=http://localhost:8080 \
  make test TAGS=unit,local GITLAB_VERSIONS=17.11.7

# switch to 18.11.2
bash initial-setup.sh 18.11.2
GITLAB_VERSION=18.11.2 GITLAB_URL=http://localhost:8080 \
  make test TAGS=unit,local GITLAB_VERSIONS=18.11.2
```

Plain `make test` (no extra env) replays all per-version cassettes already on disk and produces a merged coverage report under `build/coverage.{out,html}`. Per-version binary coverage is preserved under `build/covdata/<version>/` for inspection via `go tool covdata percent -i=build/covdata/<version>`.

### Terraform

The Terraform code is split into a shared module and per-version roots:

- `tf/_shared/` — Terraform module containing all the resource definitions (users, groups, projects, access tokens). It does not declare any provider configuration; `path.root` is used for outputs so each caller writes its own `tokens.json` next to itself.
- `tf/<version>/` — root module per GitLab version. Each one declares its own `required_providers` (`versions.tf`) so the `gitlab` provider can be pinned independently (e.g. `~> 17.10` for `17.11.7`, `~> 18.11` for `18.11.2`) and configures the provider (`main.tf`), then calls `module "shared" { source = "../_shared" }`. Each root keeps its own `.terraform/`, `terraform.tfstate`, and `tokens.json`.

`initial-setup.sh` `cd`s into `tf/<version>/` based on the version arg.

To re-run a single version's Terraform independently:

```bash
cd tf/17.11.7 && terraform init && terraform apply
```

To support a new GitLab version:

```bash
cp -r tf/17.11.7 tf/<new-version>
# edit tf/<new-version>/versions.tf to bump the gitlab provider constraint
```

Resources shared by all versions belong in `tf/_shared/`; resources that need to differ go in the per-version root.

### Ports

| Port | Purpose |
|------|---------|
| 8080 | HTTP    |
| 8443 | HTTPS   |
| 2224 | SSH     |

Root password: `Iem3oe_lohy1`
