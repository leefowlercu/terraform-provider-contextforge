# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Table of Contents

- [Project Overview](#project-overview)
- [Architecture](#architecture)
  - [Provider Configuration Flow](#provider-configuration-flow)
  - [Provider Implementation Pattern](#provider-implementation-pattern)
  - [Binary Versioning](#binary-versioning)
- [Development Commands](#development-commands)
  - [Building and Testing](#building-and-testing)
  - [Integration Testing](#integration-testing)
  - [Debugging the Provider](#debugging-the-provider)
- [Release Process](#release-process)
- [Key Files and Locations](#key-files-and-locations)
- [Important Implementation Details](#important-implementation-details)
  - [Environment Variable Naming](#environment-variable-naming)
  - [Provider Protocol Version](#provider-protocol-version)
  - [Adding Data Sources or Resources](#adding-data-sources-or-resources)

## Project Overview

This is a Terraform provider for IBM ContextForge MCP Gateway built using the Terraform Plugin Framework v1.16.1. The provider communicates with the ContextForge MCP Gateway API via the `go-contextforge` v0.5.0 client library.

**Provider Address**: `registry.terraform.io/hashicorp/contextforge`

## Architecture

### Provider Configuration Flow

The provider implements a dual-source configuration pattern for `address` and `token`:

1. **Environment variables** are read first as defaults (`CONTEXTFORGE_ADDR`, `CONTEXTFORGE_TOKEN`)
2. **HCL configuration attributes** override environment variables if explicitly set
3. **Validation** occurs in two phases:
   - Unknown value detection (values from `terraform plan` that aren't yet known)
   - Empty value validation (both sources checked, errors raised if missing)

This pattern is implemented in `internal/provider/provider.go:Configure()`.

### Provider Implementation Pattern

The provider follows Terraform Plugin Framework conventions:

- `ContextForgeProvider` struct holds version information
- `ContextForgeProviderModel` maps HCL configuration to Go types using `tfsdk` tags
- Provider lifecycle methods: `Metadata()`, `Schema()`, `Configure()`, `DataSources()`, `Resources()`
- The `Configure()` method creates the `contextforge.Client` and stores it in both `resp.DataSourceData` and `resp.ResourceData` for downstream use

**Currently Implemented**: Provider infrastructure only. No data sources or resources are implemented yet (both return `nil`).

### Binary Versioning

Version is injected at build time via goreleaser's ldflags:
- `main.go` defines `var version = "0.1.0"`
- Goreleaser sets: `-X main.version={{.Version}}`
- Version is passed to `provider.New(version)` during initialization

## Development Commands

### Building and Testing

```bash
# Build provider binary
make build

# Install to local $GOPATH/bin for manual testing
make install

# Run unit tests
make test

# Run unit tests for a specific package
go test -v ./internal/provider

# Run a specific test
go test -v -run TestProviderSchema ./internal/provider
```

### Integration Testing

Integration tests use `TF_ACC=1` and require a running ContextForge gateway. The test infrastructure automates gateway lifecycle management using `uvx` and `mcp-contextforge-gateway`.

```bash
# Full lifecycle (setup → test → teardown)
make integration-test-all

# Manual workflow for debugging
make integration-test-setup    # Starts gateway on localhost:8000
make integration-test           # Runs tests with TF_ACC=1
make integration-test-teardown  # Stops gateway and cleans tmp/
```

**Integration test artifacts** (created in `tmp/`):
- `contextforge-test.db` - SQLite database
- `contextforge-test.pid` - Gateway process ID
- `contextforge-test.log` - Gateway logs
- `contextforge-test-token.txt` - JWT token (7-day expiration)

**Test gateway credentials**:
- Admin email: `admin@test.local`
- Admin password: `testpassword123`
- JWT secret: `test-secret-key-for-integration-testing`

### Debugging the Provider

Run provider with debugger support:

```bash
go run main.go -debug
```

Override provider address for local development:

```bash
export TF_PROVIDER_ADDRESS="registry.terraform.io/leefowlercu/contextforge"
go run main.go
```

## Release Process

Releases are automated via goreleaser (`.goreleaser.yml`):

```bash
# Requires GPG_FINGERPRINT environment variable for signing
goreleaser release --clean
```

**Build configuration**:
- CGO disabled for portability
- Multi-platform builds: linux, darwin, windows, freebsd
- Architectures: amd64, 386, arm, arm64 (darwin/386 excluded)
- Binary naming: `terraform-provider-contextforge_v{version}`
- Checksums and GPG signatures generated automatically

## Key Files and Locations

```
main.go                           - Provider entrypoint, gRPC server setup
internal/provider/provider.go     - Provider implementation (all lifecycle methods)
scripts/integration-test-setup.sh - Gateway startup automation
scripts/integration-test-teardown.sh - Gateway cleanup
test/terraform/                   - Integration test Terraform configurations
go.mod                            - Dependencies (Plugin Framework 1.16.1, go-contextforge 0.5.0)
```

## Important Implementation Details

### Environment Variable Naming

The provider uses `CONTEXTFORGE_ADDR` (not `CONTEXTFORGE_ENDPOINT`) for the gateway address. This is a URL with scheme, hostname, and port but no path (e.g., `https://contextforge.example.com`).

### Provider Protocol Version

The provider uses Terraform Plugin Protocol version 6 (`ProtocolVersion: 6` in `main.go`). This is the protocol version for Terraform Plugin Framework.

### Adding Data Sources or Resources

To add new data sources or resources:

1. Create implementation files in `internal/provider/` (e.g., `data_source_foo.go`, `resource_bar.go`)
2. Implement the `datasource.DataSource` or `resource.Resource` interface
3. Add factory function to `DataSources()` or `Resources()` method in `provider.go`
4. The `contextforge.Client` is available via type assertion from `req.ProviderData`

Example pattern:
```go
client, ok := req.ProviderData.(*contextforge.Client)
if !ok {
    resp.Diagnostics.AddError(...)
    return
}
```
