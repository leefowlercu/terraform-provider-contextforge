# Makefile for terraform-provider-contextforge

# Variables
BINARY_NAME=terraform-provider-contextforge
VERSION?=dev
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
GOBIN?=$(shell go env GOPATH)/bin
INSTALL_PATH=$(GOBIN)

# Integration test configuration
CONTEXTFORGE_ADDR=http://localhost:8000
CONTEXTFORGE_TOKEN_FILE=./tmp/contextforge-test-token.txt

# Default target
.DEFAULT_GOAL := help

# Build the provider binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME)

# Install the provider locally for manual testing
install: build
	@echo "Installing provider to $(INSTALL_PATH)..."
	@mkdir -p $(INSTALL_PATH)
	@cp $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Provider installed successfully"

# Run unit tests
test:
	@echo "Running unit tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -rf dist/
	@echo "Clean complete"

# Integration test setup - start local ContextForge gateway
integration-test-setup:
	@echo "Setting up integration test environment..."
	@bash scripts/integration-test-setup.sh

# Integration test teardown - stop local ContextForge gateway
integration-test-teardown:
	@echo "Tearing down integration test environment..."
	@bash scripts/integration-test-teardown.sh

# Run integration tests (assumes service is running)
integration-test:
	@echo "Running integration tests..."
	@if [ ! -f $(CONTEXTFORGE_TOKEN_FILE) ]; then \
		echo "Error: Token file not found. Run 'make integration-test-setup' first."; \
		exit 1; \
	fi
	@export TF_ACC=1 && \
	export CONTEXTFORGE_ADDR=$(CONTEXTFORGE_ADDR) && \
	export CONTEXTFORGE_TOKEN=$$(cat $(CONTEXTFORGE_TOKEN_FILE)) && \
	go test -v -timeout 30m ./...

# Run full integration test lifecycle (setup -> test -> teardown)
integration-test-all:
	@echo "Running full integration test lifecycle..."
	@$(MAKE) integration-test-setup
	@trap '$(MAKE) integration-test-teardown' EXIT; \
	$(MAKE) integration-test

# Verify goreleaser is installed and config is valid
goreleaser-check:
	@echo "Checking goreleaser installation..."
	@if ! command -v goreleaser &> /dev/null; then \
		echo "Error: goreleaser not found"; \
		echo "Install with: go install github.com/goreleaser/goreleaser/v2@latest"; \
		exit 1; \
	fi
	@echo "Validating goreleaser configuration..."
	@goreleaser check

# Create a local snapshot release for testing (does not publish)
goreleaser-snapshot:
	@echo "Creating snapshot release..."
	goreleaser release --snapshot --clean

# Verify prerequisites before release
release-check: goreleaser-check
	@echo "Checking git status..."
	@if ! git diff-index --quiet HEAD --; then \
		echo "Error: Working directory has uncommitted changes"; \
		exit 1; \
	fi
	@if ! git diff-index --cached --quiet HEAD --; then \
		echo "Error: Staging area has uncommitted changes"; \
		exit 1; \
	fi
	@echo "All release checks passed"

# Bump patch version (X.Y.Z -> X.Y.Z+1) and prepare release
release-patch: release-check
	@bash scripts/bump-version.sh patch
	@$(MAKE) release-prep VERSION=$$(cat .next-version)

# Bump minor version (X.Y.Z -> X.Y+1.0) and prepare release
release-minor: release-check
	@bash scripts/bump-version.sh minor
	@$(MAKE) release-prep VERSION=$$(cat .next-version)

# Bump major version (X.Y.Z -> X+1.0.0) and prepare release
release-major: release-check
	@bash scripts/bump-version.sh major
	@$(MAKE) release-prep VERSION=$$(cat .next-version)

# Prepare release with specific version
release-prep:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION required"; \
		echo "Usage: make release-prep VERSION=vX.Y.Z"; \
		exit 1; \
	fi
	@bash scripts/prepare-release.sh $(VERSION)

# Full release workflow (runs release-check then release-prep)
release: release-check
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION required"; \
		echo "Usage: make release VERSION=vX.Y.Z"; \
		echo ""; \
		echo "Or use automated version bumping:"; \
		echo "  make release-patch  # Increment patch version"; \
		echo "  make release-minor  # Increment minor version"; \
		echo "  make release-major  # Increment major version"; \
		exit 1; \
	fi
	@$(MAKE) release-prep VERSION=$(VERSION)

# Display help information
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build & Test:"
	@echo "  build                    - Build the provider binary"
	@echo "  install                  - Install the provider locally for manual testing"
	@echo "  test                     - Run unit tests"
	@echo "  clean                    - Clean build artifacts"
	@echo ""
	@echo "Integration Testing:"
	@echo "  integration-test-setup   - Start local ContextForge gateway for integration testing"
	@echo "  integration-test-teardown - Stop local ContextForge gateway and clean up"
	@echo "  integration-test         - Run integration tests (requires running gateway)"
	@echo "  integration-test-all     - Run full integration test lifecycle (setup -> test -> teardown)"
	@echo ""
	@echo "Release Management:"
	@echo "  goreleaser-check         - Verify goreleaser installation and config"
	@echo "  goreleaser-snapshot      - Create local snapshot release for testing"
	@echo "  release-check            - Verify prerequisites before release"
	@echo "  release-patch            - Bump patch version and create release (X.Y.Z -> X.Y.Z+1)"
	@echo "  release-minor            - Bump minor version and create release (X.Y.Z -> X.Y+1.0)"
	@echo "  release-major            - Bump major version and create release (X.Y.Z -> X+1.0.0)"
	@echo "  release-prep VERSION=vX.Y.Z - Prepare release with specific version"
	@echo "  release VERSION=vX.Y.Z   - Full release workflow"
	@echo ""
	@echo "Prerequisites for releases:"
	@echo "  - goreleaser: go install github.com/goreleaser/goreleaser/v2@latest"
	@echo "  - GITHUB_TOKEN environment variable (for GitHub releases)"
	@echo "  - GPG_FINGERPRINT environment variable (for signing)"
	@echo ""
	@echo "Quick start:"
	@echo "  make release-patch       # Create patch release (bug fixes)"
	@echo "  make release-minor       # Create minor release (new features)"
	@echo "  make release-major       # Create major release (breaking changes)"
	@echo ""
	@echo "Other:"
	@echo "  help                     - Display this help message"

.PHONY: build install test clean integration-test-setup integration-test-teardown integration-test integration-test-all \
        goreleaser-check goreleaser-snapshot release-check release-patch release-minor release-major release-prep release help
