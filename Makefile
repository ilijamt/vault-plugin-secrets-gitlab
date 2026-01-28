SHELL := /bin/bash

GO ?= go

COVER_PROFILE ?= coverage.out
COVER_HTML ?= coverage.html

# Default build tags for tests; can be overridden:
#   make test TAGS="unit,local"
TAGS ?= unit,local,saas,selfhosted

# Default test args; can be overridden:
#   make test TEST_ARGS="-run TestFoo -v"
TEST_ARGS ?=

.PHONY: test coverage clean-coverage

test: coverage

coverage:
	$(GO) test ./... -coverprofile=$(COVER_PROFILE) -race -tags $(TAGS) -count 1 $(TEST_ARGS)
	$(GO) tool cover -html=$(COVER_PROFILE) -o $(COVER_HTML)

clean-coverage:
	rm -f $(COVER_PROFILE) $(COVER_HTML)