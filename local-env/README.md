## Local Environment

To run tests against a real GitLab instance, follow the steps below.

### Initial Setup

1. **Run the setup script:**

   This command will set up a GitLab instance that is fully configured for testing locally.

```bash
bash initial-setup.sh
```

   **Note:** Setting up the GitLab instance might take some time. After the setup, a complete backup of the PostgreSQL database will be created to facilitate quick restoration if needed.

### Restoring the Environment

If you need to restore the GitLab instance back to its original configuration, use the following command:

```bash
bash restore-volumes.sh
```
