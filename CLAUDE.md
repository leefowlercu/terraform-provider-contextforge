# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Table of Contents

- [Project Overview](#project-overview)
- [Architecture](#architecture)
  - [Provider Configuration Flow](#provider-configuration-flow)
  - [Provider Implementation Pattern](#provider-implementation-pattern)
  - [Implemented Data Sources](#implemented-data-sources)
  - [Binary Versioning](#binary-versioning)
- [Development Commands](#development-commands)
  - [Building and Testing](#building-and-testing)
  - [Integration Testing](#integration-testing)
  - [Debugging the Provider](#debugging-the-provider)
- [CI/CD](#cicd)
- [Release Process](#release-process)
- [Key Files and Locations](#key-files-and-locations)
- [Important Implementation Details](#important-implementation-details)
  - [Environment Variable Naming](#environment-variable-naming)
  - [Provider Protocol Version](#provider-protocol-version)
  - [Adding Data Sources or Resources](#adding-data-sources-or-resources)
  - [Test Organization](#test-organization)

## Project Overview

This is a Terraform provider for IBM ContextForge MCP Gateway built using the Terraform Plugin Framework v1.16.1. The provider communicates with the ContextForge MCP Gateway API via the `go-contextforge` v0.6.0 client library.

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

### Implemented Data Sources

**contextforge_gateway** - Retrieves gateway information by ID (`internal/provider/data_source_gateway.go`)

The gateway data source provides read-only access to ContextForge MCP Gateway resources:

- **Type conversions** handled by `internal/tfconv` package for complex types:
  - `map[string]any` → `types.Dynamic` for heterogeneous maps (capabilities, oauth_config, metadata)
  - `[]map[string]string` → `types.List` of `types.Map` for auth headers
  - Timestamps → `types.String` in RFC3339 format
- **Key attributes**: id, name, url, transport, description, enabled, reachable, created_at, updated_at
- **Authentication fields**: auth_type, auth_token, auth_username, auth_password, oauth_config
- **Organizational fields**: team_id, team, owner_email, visibility, tags

**Acceptance tests** verify the data source with a real gateway created during integration test setup (see `internal/provider/data_source_gateway_test.go`).

**contextforge_server** - Retrieves server information by ID (`internal/provider/data_source_server.go`)

The server data source provides read-only access to ContextForge virtual server resources:

- **Type conversions** handled for complex types:
  - `[]int` → `[]int64` for associated_resources and associated_prompts lists
  - Nested metrics object with performance statistics
  - Timestamps → `types.String` in RFC3339 format
- **Key attributes**: id, name, description, icon, is_active, associated_tools, associated_resources, associated_prompts, associated_a2a_agents, metrics
- **Metrics fields**: total_executions, successful_executions, failed_executions, failure_rate, min_response_time, max_response_time, avg_response_time, last_execution_time
- **Organizational fields**: team_id, team, owner_email, visibility, tags

**Acceptance tests** verify the data source with a real server created during integration test setup (see `internal/provider/data_source_server_test.go`).

**contextforge_resource** - Retrieves resource information by ID (`internal/provider/data_source_resource.go`)

The resource data source provides read-only access to ContextForge resource entities:

- **Type conversions** handled for complex types:
  - `*FlexibleID` → `types.String` using `.String()` method for ID field
  - `*int` → `types.Int64` using `tfconv.Int64Ptr()` for size and version fields
  - Nested metrics object with performance statistics
  - Timestamps → `types.String` in RFC3339 format
- **Key attributes**: id, uri, name, description, mime_type, size, is_active, metrics
- **Metrics fields**: total_executions, successful_executions, failed_executions, failure_rate, min_response_time, max_response_time, avg_response_time, last_execution_time
- **Organizational fields**: team_id, team, owner_email, visibility, tags
- **Unique characteristics**: Uses List() API endpoint and filters by ID (no dedicated metadata endpoint available)

**Acceptance tests** verify the data source with a real resource created during integration test setup (see `internal/provider/data_source_resource_test.go`).

**contextforge_tool** - Retrieves tool information by ID (`internal/provider/data_source_tool.go`)

The tool data source provides read-only access to ContextForge tool resources:

- **Type conversions** handled for complex types:
  - `map[string]any` → `types.Dynamic` for input_schema (JSON Schema definition)
  - Timestamps → `types.String` in RFC3339 format
- **Key attributes**: id, name, description, input_schema, enabled
- **Organizational fields**: tags, team_id, visibility
- **Unique characteristic**: Simplest data source with no nested metrics, associations, or authentication fields

**Acceptance tests** verify the data source with a real tool created during integration test setup (see `internal/provider/data_source_tool_test.go`).

**contextforge_team** - Retrieves team information by ID (`internal/provider/data_source_team.go`)

The team data source provides read-only access to ContextForge team resources:

- **Type conversions** handled for standard types:
  - `*int` → `types.Int64` using `tfconv.Int64Ptr()` for max_members field
  - Timestamps → `types.String` in RFC3339 format using snake_case field names (created_at, updated_at)
- **Key attributes**: id, name, slug, description, is_personal, visibility, max_members, member_count, is_active, created_by
- **Timestamps**: created_at, updated_at
- **Unique characteristic**: Uses snake_case timestamps (created_at, updated_at) unlike other resources

**Acceptance tests** verify the data source with a real team created during integration test setup (see `internal/provider/data_source_team_test.go`).

**contextforge_agent** - Retrieves agent information by ID (`internal/provider/data_source_agent.go`)

The agent data source provides read-only access to ContextForge A2A agent resources:

- **Type conversions** handled for complex types:
  - `map[string]any` → `types.Dynamic` for capabilities and config fields
  - Nested metrics object with performance statistics
  - Timestamps → `types.String` in RFC3339 format
- **Key attributes**: id, name, description, endpoint_url, enabled, capabilities, config, metrics
- **Metrics fields**: total_requests, successful_requests, failed_requests, failure_rate, min_response_time, max_response_time, avg_response_time, last_request_time
- **Organizational fields**: team_id, team, owner_email, visibility, tags
- **Unique characteristic**: Has two Dynamic type fields (capabilities, config) and uses agent-specific metrics field names (total_requests vs total_executions)

**Acceptance tests** verify the data source with a real agent created during integration test setup (see `internal/provider/data_source_agent_test.go`).

**contextforge_prompt** - Retrieves prompt information by ID (`internal/provider/data_source_prompt.go`)

The prompt data source provides read-only access to ContextForge prompt resources:

- **Type conversions** handled for complex types:
  - Integer ID (types.Int64) - only data source using integer ID instead of string
  - Nested arguments list with promptArgumentModel objects (name, description, required)
  - Nested metrics object with performance statistics
  - Timestamps → `types.String` in RFC3339 format
- **Key attributes**: id, name, description, template, arguments, is_active, tags, metrics
- **Arguments structure**: List of objects with name, description, required fields
- **Metrics fields**: total_executions, successful_executions, failed_executions, failure_rate, min_response_time, max_response_time, avg_response_time, last_execution_time
- **Organizational fields**: team_id, team, owner_email, visibility
- **Unique characteristics**:
  - Only data source with integer ID (types.Int64)
  - Uses List() API endpoint and filters by ID (no dedicated Get metadata endpoint available)
  - Has nested arguments list for prompt parameters

**Acceptance tests** verify the data source with a real prompt created during integration test setup (see `internal/provider/data_source_prompt_test.go`).

**Resources**: No managed resources are currently implemented. The `Resources()` method returns `nil`.

### Binary Versioning

Version is injected at build time via goreleaser's ldflags:
- `main.go` defines `var version = "<current_version>"` (replaced during build)
- Goreleaser sets: `-X main.version={{.Version}}`
- Version is passed to `provider.New(version)` during initialization

## Development Commands

**Prerequisites**: Go 1.25.3 or later

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

**Integration test infrastructure:**

The setup script creates a complete test environment:

1. **ContextForge Gateway** (port 8000)
   - SQLite database: `tmp/contextforge-test.db`
   - Process ID: `tmp/contextforge-test.pid`
   - Logs: `tmp/contextforge-test.log`
   - JWT token: `tmp/contextforge-test-token.txt` (7-day expiration)
   - Admin: `admin@test.local` / `testpassword123`
   - JWT secret: `test-secret-key-for-integration-testing`

2. **MCP Time Server** (port 8002)
   - Provides real MCP endpoint for gateway connectivity validation
   - Started via `mcpgateway.translate` wrapper: `uvx mcp-server-time --local-timezone=UTC`
   - Process ID: `tmp/time-server.pid`
   - Logs: `tmp/time-server.log`
   - Why needed: ContextForge validates gateway connectivity during creation

3. **Test Resources**
   - **Test Gateway**: Created via ContextForge API pointing to time server
     - URL: `http://localhost:8002/sse`
     - Transport: SSE (Server-Sent Events)
     - Name: "test-time-server"
     - Description: "Test gateway for integration tests"
     - Gateway ID saved: `tmp/contextforge-test-gateway-id.txt`
   - **Test Server**: Virtual MCP server for testing
     - Name: "test-server"
     - Description: "Test server for integration tests"
     - Server ID saved: `tmp/contextforge-test-server-id.txt`
   - **Test Tool**: MCP tool for testing
     - Name: "test-tool"
     - Description: "Test tool for integration tests"
     - Tool ID saved: `tmp/contextforge-test-tool-id.txt`
   - **Test Resource**: MCP resource for testing
     - URI: "test://integration/resource"
     - Name: "test-resource"
     - Description: "Test resource for integration tests"
     - Resource ID saved: `tmp/contextforge-test-resource-id.txt`
   - **Test Team**: Team for testing
     - Name: "test-team"
     - Slug: "test-team"
     - Description: "Test team for integration tests"
     - Team ID saved: `tmp/contextforge-test-team-id.txt`
   - **Test Agent**: A2A agent for testing
     - Name: "test-agent"
     - Endpoint URL: "http://localhost:9000/agent"
     - Description: "Test agent for integration tests"
     - Agent ID saved: `tmp/contextforge-test-agent-id.txt`
   - **Test Prompt**: Prompt for testing
     - Name: "test-prompt"
     - Template: "Hello {{name}}, this is a test prompt."
     - Description: "Test prompt for integration tests"
     - Prompt ID saved: `tmp/contextforge-test-prompt-id.txt`
   - Used by acceptance tests to verify data source functionality

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

**Provider Address Configuration:**
- **Production default**: `registry.terraform.io/hashicorp/contextforge` (defined in `main.go`)
- **Development override**: Use `TF_PROVIDER_ADDRESS` environment variable to test with alternate namespaces
- The example above uses `leefowlercu` namespace for local development/testing
- When the provider is published to the Terraform Registry, it will use the `hashicorp` namespace

## CI/CD

The project uses GitHub Actions for automated testing (`.github/workflows/test.yml`):

**Triggers:**
- Push to `master` branch
- Pull requests (ignoring changes to README.md, CHANGELOG.md, and docs/)

**Workflow Jobs:**

1. **Unit Tests** (`unit-test`)
   - Runs on ubuntu-latest with 15-minute timeout
   - Executes `go test` with coverage
   - Verifies the provider builds successfully

2. **Acceptance Tests** (`acceptance-test`)
   - Runs on ubuntu-latest with 60-minute timeout
   - Installs Python 3.11 and uv for gateway management
   - Executes full integration test lifecycle:
     - `make integration-test-setup` - Starts ContextForge gateway and MCP time server
     - `make integration-test` - Runs acceptance tests with `TF_ACC=1`
     - `make integration-test-teardown` - Cleans up test infrastructure
   - Uploads gateway logs as artifacts on failure

## Release Process

Releases are automated via goreleaser (`.goreleaser.yml`) and managed through Makefile targets:

```bash
# Automated version bumping and release preparation
make release-patch  # Bump patch version (X.Y.Z -> X.Y.Z+1)
make release-minor  # Bump minor version (X.Y.Z -> X.Y+1.0)
make release-major  # Bump major version (X.Y.Z -> X+1.0.0)

# Manual release with specific version
make release-prep VERSION=vX.Y.Z

# Direct goreleaser usage (requires GPG_FINGERPRINT environment variable)
goreleaser release --clean
```

**Release workflow:**
- `scripts/bump-version.sh` - Calculates next version based on git tags
- `scripts/prepare-release.sh` - Updates CHANGELOG.md, creates git tag, runs goreleaser

**Build configuration**:
- CGO disabled for portability
- Multi-platform builds: linux, darwin, windows, freebsd
- Architectures: amd64, 386, arm, arm64 (darwin/386 excluded)
- Binary naming: `terraform-provider-contextforge_v{version}`
- Checksums and GPG signatures generated automatically

## Key Files and Locations

```
main.go                                      - Provider entrypoint, gRPC server setup
internal/provider/provider.go                - Provider implementation (all lifecycle methods)
internal/provider/doc.go                     - Comprehensive package documentation with implementation patterns
internal/provider/data_source_gateway.go     - Gateway data source implementation
internal/provider/data_source_gateway_test.go - Gateway acceptance tests
internal/provider/data_source_server.go      - Server data source implementation
internal/provider/data_source_server_test.go - Server acceptance tests
internal/provider/data_source_resource.go    - Resource data source implementation
internal/provider/data_source_resource_test.go - Resource acceptance tests
internal/provider/data_source_tool.go        - Tool data source implementation
internal/provider/data_source_tool_test.go   - Tool acceptance tests
internal/provider/provider_test.go           - Shared test utilities (testAccProtoV6ProviderFactories, testAccPreCheck)
internal/tfconv/convert.go                   - Type conversion utilities for Terraform Plugin Framework
scripts/integration-test-setup.sh            - Gateway startup automation, MCP server, test gateway, server, tool, and resource creation
scripts/integration-test-teardown.sh         - Gateway cleanup
scripts/bump-version.sh                      - Version bumping utility (used by release targets)
test/terraform/                              - Manual testing Terraform configurations
go.mod                                       - Dependencies (Plugin Framework 1.16.1, go-contextforge 0.6.0)
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

### Test Organization

**All tests are co-located in `internal/provider/`** following Terraform provider conventions:

- `provider_test.go` - Shared test utilities:
  - `testAccProtoV6ProviderFactories` - Provider factory for acceptance tests
  - `testAccPreCheck(t *testing.T)` - Environment variable validation
  - Both are unexported (lowercase) as they're package-internal

- `data_source_<name>_test.go` - Acceptance tests for data sources:
  - Tests run with `TF_ACC=1` environment variable
  - Use `testAccProtoV6ProviderFactories` for provider instantiation
  - Use `testAccPreCheck(t)` in PreCheck field
  - Example: `data_source_gateway_test.go` with 4 tests covering basic lookup, error cases, and attribute verification

**No separate test packages**: Unlike some projects, this provider does not use `internal/data` or `internal/acctest` packages. All data sources, resources, and tests reside in `internal/provider/` to avoid import cycles.
