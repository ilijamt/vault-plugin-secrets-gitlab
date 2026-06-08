FROM scratch
LABEL org.opencontainers.image.source="https://github.com/ilijamt/vault-plugin-secrets-gitlab"
LABEL org.opencontainers.image.description="Vault/OpenBao secrets plugin for GitLab access tokens"
LABEL org.opencontainers.image.licenses="MIT"
COPY vault-plugin-secrets-gitlab /vault-plugin-secrets-gitlab
ENTRYPOINT ["/vault-plugin-secrets-gitlab"]
