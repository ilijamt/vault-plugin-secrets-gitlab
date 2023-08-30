Vault Plugin for Gitlab Access Token
------------------------------------
[![Go Report Card](https://goreportcard.com/badge/github.com/ilijamt/vault-plugin-secrets-gitlab)](https://goreportcard.com/report/github.com/ilijamt/vault-plugin-secrets-gitlab)
![Codecov](https://img.shields.io/codecov/c/gh/ilijamt/vault-plugin-secrets-gitlab)
![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/ilijamt/vault-plugin-secrets-gitlab)
![GitHub](https://img.shields.io/github/license/ilijamt/vault-plugin-secrets-gitlab)

This is a standalone backend plugin for use with Hashicorp Vault. This plugin allows for Gitlab to generate personal, project and group access tokens.

## Quick Links

- Vault Website: [https://www.vaultproject.io]
- Gitlab Private Access Tokens: [https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html]
- Gitlab Group Access Tokens: [https://docs.gitlab.com/ee/api/group_access_tokens.html]

## Getting Started

This is a [Vault plugin](https://www.vaultproject.io/docs/plugins/plugin-architecture#plugin-catalogs)
and is meant to work with Vault. This guide assumes you have already installed Vault
and have a basic understanding of how Vault works.

Otherwise, first read this guide on how to [get started with Vault](https://www.vaultproject.io/intro/getting-started/install.html).

To learn specifically about how plugins work, see documentation on [Vault plugins](https://www.vaultproject.io/docs/plugins/plugin-architecture#plugin-catalog).

### Setup

Before we can use this plugin we need to create an access token that will have rights to do what we need to.

## Security Model

The current authentication model requires providing Vault with a Gitlab Token. 

## TODO

* [ ] Implement autorotation of the main token
* [ ] Add tests against real Gitlab instance