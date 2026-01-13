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

This is a Terraform provider for IBM ContextForge MCP Gateway built using the Terraform Plugin Framework v1.16.1. The provider communicates with the ContextForge MCP Gateway API via the `go-contextforge` v0.8.1 client library.

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
  - `[]string` → `types.List` (with `types.StringType`) for associated_resources and associated_prompts lists
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
  - String ID (types.String) - changed from int to string in go-contextforge v0.8.1
  - Nested arguments list with promptArgumentModel objects (name, description, required)
  - Nested metrics object with performance statistics
  - Timestamps → `types.String` in RFC3339 format
  - Tags use `contextforge.TagNames()` to convert `[]Tag` to `[]string`
- **Key attributes**: id, name, description, template, arguments, is_active, tags, metrics
- **Arguments structure**: List of objects with name, description, required fields
- **Metrics fields**: total_executions, successful_executions, failed_executions, failure_rate, min_response_time, max_response_time, avg_response_time, last_execution_time
- **Organizational fields**: team_id, team, owner_email, visibility
- **Unique characteristics**:
  - Uses List() API endpoint and filters by ID (no dedicated Get metadata endpoint available)
  - Has nested arguments list for prompt parameters

**Acceptance tests** verify the data source with a real prompt created during integration test setup (see `internal/provider/data_source_prompt_test.go`).

### Implemented Resources

**contextforge_gateway** - Manages gateway resources (`internal/provider/resource_gateway.go`)

The gateway resource provides full CRUD operations for ContextForge MCP Gateway resources:

- **Type conversions** handled by `internal/tfconv` package for complex types:
  - `map[string]any` → `types.Dynamic` for capabilities, oauth_config
  - `[]map[string]string` → `types.List` of `types.Map` for auth_headers
  - Pointer strings for most fields
- **Key attributes**: name (required), url (required), transport (required), description, enabled, auth fields (9 types), tags, team_id, visibility
- **Read-only fields**: id, reachable, capabilities, timestamps (3), metadata (16 fields), slug, version
- **Special considerations**: Update followed by GET workaround for API bug

**Acceptance tests** include basic CRUD, import, and validation tests (see `internal/provider/resource_gateway_test.go`). Note: Basic and import tests skipped due to upstream requirement that gateway URLs must be reachable.

**contextforge_tool** - Manages tool resources (`internal/provider/resource_tool.go`)

The tool resource provides full CRUD operations for ContextForge tool resources:

- **Type conversions** handled for complex types:
  - `map[string]any` → `types.Dynamic` for input_schema (JSON Schema)
  - Empty schema filtering with `isEmptyInputSchema()` helper
- **Key attributes**: name (required), description, input_schema (Dynamic), enabled, tags, team_id, visibility
- **Read-only fields**: id, timestamps (2)
- **Special considerations**: Update followed by GET workaround for API bug

**Acceptance tests** verify resource functionality (see `internal/provider/resource_tool_test.go`). Note: 4 tests skipped due to upstream bugs in tool update API.

**contextforge_server** - Manages virtual server resources (`internal/provider/resource_server.go`)

The server resource provides full CRUD operations for ContextForge virtual server resources:

- **Type conversions** handled for association fields:
  - associated_tools: `[]string` → `types.List` (with `types.StringType`)
  - associated_resources: `[]string` → `types.List` (with `types.StringType`)
  - associated_prompts: `[]string` → `types.List` (with `types.StringType`)
  - associated_a2a_agents: `[]string` → `types.List` (with `types.StringType`)
- **Key attributes**: name (required), description, icon, tags, associations (4 types), team_id, visibility
- **Read-only fields**: id, is_active, metrics (nested object with 8 fields), team, owner_email, timestamps (2), metadata (10 fields), version
- **Metrics nested object**: total_executions, successful_executions, failed_executions, failure_rate, response times (min/max/avg), last_execution_time

**Acceptance tests** verify resource functionality with 6 tests (see `internal/provider/resource_server_test.go`). Tests cover basic CRUD, complete configuration, updates, associations, import, and validation.

**contextforge_agent** - Manages A2A (Agent-to-Agent) agent resources (`internal/provider/resource_agent.go`)

The agent resource provides full CRUD operations for ContextForge A2A agent resources:

- **Type conversions** handled for Dynamic fields:
  - capabilities: `map[string]any` → `types.Dynamic` (computed, read-only)
  - config: `map[string]any` ↔ `types.Dynamic` (optional, user-configurable)
  - Helper functions: `tfconv.ConvertMapToObjectValue()`, `tfconv.ConvertObjectValueToMap()`
- **Key attributes**: name (required), endpoint_url (required), description, agent_type, protocol_version, config (Dynamic), auth_type, enabled, tags, team_id, visibility
- **Read-only fields**: id, slug, capabilities (Dynamic), reachable, owner_email, metrics (nested object with 8 fields), timestamps (3), metadata (10 fields), version
- **Metrics nested object**: total_executions, successful_executions, failed_executions, failure_rate, response times (min/max/avg), last_execution_time
- **SDK types**: Uses `AgentCreate` for Create operations, `AgentUpdate` for Update operations
- **Plan modifiers**: Optional+computed fields use `UseStateForUnknown()` to prevent drift detection
- **Special considerations**:
  - No Update+GET workaround needed (API works correctly)
  - Empty config objects cause drift issues (documented limitation)

**Acceptance tests** verify resource functionality with 1 passing test (see `internal/provider/resource_agent_test.go`). Tests cover validation. Note: 4 tests skipped due to API behavior with empty config objects and computed fields showing as changed during updates/imports.

**contextforge_resource** - Manages resource resources (`internal/provider/resource_resource.go`)

The resource resource provides full CRUD operations for ContextForge resource entities:

- **Type conversions** handled for FlexibleID and integer fields:
  - ID: `*FlexibleID` → `types.String` using `.String()` method
  - Size: `*int` → `types.Int64` using `tfconv.Int64Ptr()`
  - Version: `*int` → `types.Int64` using `tfconv.Int64Ptr()`
- **Key attributes**: uri (required), name (required), content (required, write-only), description, mime_type, tags, team_id, visibility
- **Read-only fields**: id (FlexibleID), size, is_active, metrics (nested object with 8 fields), team, owner_email, timestamps (2), metadata (10 fields), version
- **Metrics nested object**: total_executions, successful_executions, failed_executions, failure_rate, response times (min/max/avg), last_execution_time
- **SDK types**: Uses `ResourceCreate` for Create operations, `ResourceUpdate` for Update operations
- **Special characteristics**:
  - Uses List() API endpoint and filters by ID (no dedicated Get() endpoint available)
  - FlexibleID can be string or integer, converted via `.String()` method
  - Content field is required for creation but not returned by API (write-only)
  - Size and IsActive are computed-only fields (backend-managed)
- **Plan modifiers**: team_id and visibility use `RequiresReplace()` (can only be set at creation)

**Acceptance tests** are implemented (see `internal/provider/resource_resource_test.go`). Tests cover basic CRUD, complete configuration, updates, import, and validation. Note: Some tests may experience transient failures during post-apply refresh due to timing/token issues.

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
   - Version: `mcp-contextforge-gateway==1.0.0b1` (pinned in setup script)
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

2a. **Additional MCP Time Servers** (ports 8003-8004)
   - Port 8003: Used by `TestAccGatewayResource_basic` for CRUD testing
   - Port 8004: Used by `TestAccGatewayResource_import` for import testing
   - Same `mcpgateway.translate` wrapper pattern as port 8002
   - Process IDs: `tmp/time-server-8003.pid`, `tmp/time-server-8004.pid`
   - Logs: `tmp/time-server-8003.log`, `tmp/time-server-8004.log`
   - Why needed: Gateway resource tests create new gateways that must be reachable

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
go.mod                                       - Dependencies (Plugin Framework 1.16.1, go-contextforge 0.8.1)
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
