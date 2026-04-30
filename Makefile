SHELL := /bin/bash

GO ?= go
VAULT ?= vault

COVER_PROFILE ?= $(BUILD_DIR)/coverage.out
COVER_HTML ?= $(BUILD_DIR)/coverage.html
COVER_PKG ?= github.com/ilijamt/vault-plugin-secrets-gitlab/...

TEST_PKG ?= ./...

# Default build tags for tests; can be overridden:
#   make test TAGS="unit,local"
TAGS ?= unit,local,saas,selfhosted

# Default test args; can be overridden:
#   make test TEST_ARGS="-run TestFoo -v"
TEST_ARGS ?=

BUILD_DIR ?= build

# GitLab versions to run integration tests against. Auto-discovered from
# tests/integration/testdata/local/<version>/ subdirectories. Override to test
# a single version, e.g.:
#   make test GITLAB_VERSIONS=17.11.7
GITLAB_VERSIONS ?= $(shell ls -1 tests/integration/testdata/local 2>/dev/null | sort)

# Per-version binary coverage directory root (Go covdir format). Each test run
# writes its own subdir; we merge them at the end via `go tool covdata textfmt`.
COVER_DATA_DIR ?= $(BUILD_DIR)/covdata

# URL of the local GitLab instance booted by local-env/docker-compose.yml.
# The compose file always exposes the single web service on port 8080, so this
# rarely changes; override only if you've remapped the host port.
GITLAB_URL ?= http://localhost:8080

PLUGIN_CMD ?= vault-plugin-secrets-gitlab
PLUGIN_BIN ?= gitlab

VAULT_PLUGIN_DIR ?= ./run/plugins
VAULT_ROOT_TOKEN ?= root-token
VAULT_ADDR ?= http://127.0.0.1:8200

PLUGIN_NAME ?= gitlab
PLUGIN_TYPE ?= secret

.PHONY: test coverage clean clean-coverage build vault-plugin-enable vault-dev check-go check-vault generate-mocks fetch-token-timestamps fetch-token-timestamps-saas fetch-token-timestamps-selfhosted

check-go:
	@command -v "$(GO)" >/dev/null 2>&1 || { \
		echo "ERROR: required binary '$(GO)' not found in PATH. Install Go or set GO=<path-to-go>."; \
		exit 1; \
	}

check-vault:
	@command -v "$(VAULT)" >/dev/null 2>&1 || { \
		echo "ERROR: required binary '$(VAULT)' not found in PATH. Install Vault or set VAULT=<path-to-vault>."; \
		exit 1; \
	}

clean:
	rm -rf $(BUILD_DIR) $(VAULT_PLUGIN_DIR)

test: coverage

coverage: check-go clean-coverage
	@if [ -z "$(GITLAB_VERSIONS)" ]; then \
		echo "ERROR: no GitLab versions found under tests/integration/testdata/local/. Set GITLAB_VERSIONS or populate testdata."; \
		exit 1; \
	fi
	mkdir -p $(BUILD_DIR)
	@for v in $(GITLAB_VERSIONS); do \
		echo "=== Testing against GitLab $$v ==="; \
		mkdir -p $(COVER_DATA_DIR)/$$v; \
		GITLAB_VERSION=$$v GITLAB_URL=$(GITLAB_URL) $(GO) test $(TEST_PKG) -cover -coverpkg=$(COVER_PKG) \
			-race -tags $(TAGS) -count 1 $(TEST_ARGS) \
			-args -test.gocoverdir=$(abspath $(COVER_DATA_DIR))/$$v || exit $$?; \
	done
	$(GO) tool covdata textfmt \
		-i=$$(echo $(COVER_DATA_DIR)/* | tr ' ' ',') \
		-o $(COVER_PROFILE)
	$(GO) tool cover -html=$(COVER_PROFILE) -o $(COVER_HTML)

clean-coverage:
	rm -f $(COVER_PROFILE) $(COVER_HTML)
	rm -rf $(COVER_DATA_DIR)

build: check-go
	mkdir -p $(BUILD_DIR)
	$(GO) build -trimpath -o $(BUILD_DIR)/$(PLUGIN_BIN) ./cmd/$(PLUGIN_CMD)

vault-plugin-enable: check-vault
	export VAULT_ADDR=$(VAULT_ADDR)
	export VAULT_TOKEN=$(VAULT_ROOT_TOKEN)
	$(VAULT) secrets enable -path="$(PLUGIN_NAME)" "$(PLUGIN_NAME)"

vault-dev: check-vault clean build
	mkdir -p $(VAULT_PLUGIN_DIR)
	cp -f $(BUILD_DIR)/$(PLUGIN_BIN) $(VAULT_PLUGIN_DIR)/$(PLUGIN_BIN)
	$(VAULT) server -dev -dev-root-token-id=$(VAULT_ROOT_TOKEN) -dev-plugin-dir=$(shell pwd)/$(VAULT_PLUGIN_DIR)

fetch-token-timestamps: fetch-token-timestamps-saas fetch-token-timestamps-selfhosted

fetch-token-timestamps-saas:
	@command -v jq >/dev/null 2>&1 || { echo "ERROR: 'jq' not found in PATH."; exit 1; }
	@test -n "$$GITLAB_COM_TOKEN" || { echo "ERROR: GITLAB_COM_TOKEN is not set."; exit 1; }
	curl --fail --silent --header "PRIVATE-TOKEN: $$GITLAB_COM_TOKEN" \
		"https://gitlab.com/api/v4/personal_access_tokens/self" \
		| jq -rj '.created_at' > tests/integration/testdata/gitlab-com

fetch-token-timestamps-selfhosted:
	@command -v jq >/dev/null 2>&1 || { echo "ERROR: 'jq' not found in PATH."; exit 1; }
	@test -n "$$GITLAB_SERVICE_ACCOUNT_TOKEN" || { echo "ERROR: GITLAB_SERVICE_ACCOUNT_TOKEN is not set."; exit 1; }
	@test -n "$$GITLAB_SERVICE_ACCOUNT_URL" || { echo "ERROR: GITLAB_SERVICE_ACCOUNT_URL is not set."; exit 1; }
	curl --fail --silent --header "PRIVATE-TOKEN: $$GITLAB_SERVICE_ACCOUNT_TOKEN" \
		"$$GITLAB_SERVICE_ACCOUNT_URL/api/v4/personal_access_tokens/self" \
		| jq -rj '.created_at' > tests/integration/testdata/gitlab-selfhosted
