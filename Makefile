.DEFAULT_TARGET=help
VERSION:=$(shell cat VERSION)

# Image URL to use all building/pushing image targets
IMG ?= ghcr.io/busser/murmur:$(VERSION)

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Format source code.
	go fmt ./...

.PHONY: vet
vet: ## Vet source code.
	go vet ./...

.PHONY: test
test: ## Run unit tests.
	go test ./...

.PHONY: test-e2e
test-e2e: ## Run all tests, including end-to-end tests.
	go test -tags=e2e ./...

##@ Build

.PHONY: build
build: fmt vet ## Build murmur binary.
	go build -o bin/murmur

##@ Release

.PHONY: release
release: test ## Release a new version.
	git tag -a "$(VERSION)" -m "$(VERSION)"
	git push origin "$(VERSION)"
	GITHUB_TOKEN=$$(gh auth token) goreleaser release --clean --release-notes=docs/release-notes/$(VERSION).md

.PHONY: release-dry-run
release-dry-run: ## Test the release process without publishing.
	goreleaser release --snapshot --clean --skip=publish --release-notes=docs/release-notes/$(VERSION).md
