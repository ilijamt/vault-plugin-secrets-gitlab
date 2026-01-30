SHELL := /bin/bash

GO ?= go
VAULT ?= vault

COVER_PROFILE ?= $(BUILD_DIR)/coverage.out
COVER_HTML ?= $(BUILD_DIR)/coverage.html

# Default build tags for tests; can be overridden:
#   make test TAGS="unit,local"
TAGS ?= unit,local,saas,selfhosted

# Default test args; can be overridden:
#   make test TEST_ARGS="-run TestFoo -v"
TEST_ARGS ?=

BUILD_DIR ?= build
PLUGIN_CMD ?= vault-plugin-secrets-gitlab
PLUGIN_BIN ?= gitlab

VAULT_PLUGIN_DIR ?= ./run/plugins
VAULT_ROOT_TOKEN ?= root-token
VAULT_ADDR ?= http://127.0.0.1:8200

PLUGIN_NAME ?= gitlab
PLUGIN_TYPE ?= secret

.PHONY: test coverage clean clean-coverage build vault-plugin-enable vault-dev check-go check-vault

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
	mkdir -p $(BUILD_DIR)
	$(GO) test ./... -cover -coverprofile=$(COVER_PROFILE) -race -tags $(TAGS) -count 1 $(TEST_ARGS)
	$(GO) tool cover -html=$(COVER_PROFILE) -o $(COVER_HTML)

clean-coverage:
	rm -f $(COVER_PROFILE) $(COVER_HTML)

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
