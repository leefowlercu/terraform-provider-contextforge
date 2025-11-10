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

# Display help information
help:
	@echo "Available targets:"
	@echo "  build                    - Build the provider binary"
	@echo "  install                  - Install the provider locally for manual testing"
	@echo "  test                     - Run unit tests"
	@echo "  clean                    - Clean build artifacts"
	@echo "  integration-test-setup   - Start local ContextForge gateway for integration testing"
	@echo "  integration-test-teardown - Stop local ContextForge gateway and clean up"
	@echo "  integration-test         - Run integration tests (requires running gateway)"
	@echo "  integration-test-all     - Run full integration test lifecycle (setup -> test -> teardown)"
	@echo "  help                     - Display this help message"

.PHONY: build install test clean integration-test-setup integration-test-teardown integration-test integration-test-all help
