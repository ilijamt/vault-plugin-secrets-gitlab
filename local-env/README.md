## Local Environment

Runs a local GitLab CE 17.10.3 instance via Docker Compose for integration testing.

**Prerequisites:** Docker, Terraform, `curl`, `jq`

### Setup

```bash
bash initial-setup.sh
```

This starts GitLab on `localhost:8080`, creates an admin access token on the `root` user, provisions users/groups/projects/tokens via Terraform (`tf/`), and copies the generated tokens to `../testdata/tokens.json`. A volume backup is created automatically at the end. The initial setup can take several minutes while GitLab boots.

### Backup and Restore

```bash
bash backup-volumes.sh   # save current state to backup.tar
bash restore-volumes.sh  # reset to post-setup state from backup.tar
```

Use restore to reset the GitLab instance after running tests or making changes.

### Terraform

The `tf/` directory provisions users, groups, projects, and access tokens on the GitLab instance. To re-run independently:

```bash
cd tf && terraform init && terraform apply
```

### Ports

| Port | Purpose |
|------|---------|
| 8080 | HTTP    |
| 8443 | HTTPS   |
| 2224 | SSH     |

Root password: `Iem3oe_lohy1`
