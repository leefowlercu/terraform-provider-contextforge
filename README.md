# terraform-provider-contextforge

Terraform provider for IBM ContextForge MCP Gateway management. Manages virtual servers, gateways, tools, resources, and prompts for the ContextForge MCP Gateway service.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Requirements](#requirements)
- [Provider Configuration](#provider-configuration)
  - [Authentication](#authentication)
  - [Configuration Example](#configuration-example)
- [Data Sources](#data-sources)
  - [contextforge_gateway](#contextforge_gateway)
  - [contextforge_server](#contextforge_server)
  - [contextforge_resource](#contextforge_resource)
  - [contextforge_tool](#contextforge_tool)
- [Resources](#resources)
- [Development](#development)
  - [Prerequisites](#prerequisites)
  - [Building the Provider](#building-the-provider)
  - [Installing for Local Development](#installing-for-local-development)
- [Testing](#testing)
  - [Unit Tests](#unit-tests)
  - [Integration Tests](#integration-tests)
- [Makefile Targets](#makefile-targets)
- [Contributing](#contributing)

## Overview

The ContextForge Terraform Provider enables infrastructure-as-code management of IBM ContextForge MCP Gateway resources. The provider is built using the Terraform Plugin Framework and communicates with the ContextForge MCP Gateway API.

For more information about ContextForge MCP Gateway, see the [official documentation](https://github.com/IBM/mcp-context-forge).

## Architecture

The provider is built using the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) v1.16.1 and communicates with the ContextForge MCP Gateway API via the [go-contextforge](https://github.com/leefowlercu/go-contextforge) v0.6.0 client library. The `internal/tfconv` package handles type conversions between the client library's Go types and Terraform Plugin Framework types, including conversions for heterogeneous maps, authentication headers, and RFC3339 timestamps.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.25 (for development)
- Access to a ContextForge MCP Gateway instance

## Provider Configuration

### Authentication

The provider supports two configuration attributes:

- `address` - (Optional) ContextForge MCP Gateway address URL (e.g., `https://contextforge.example.com`). This is a URL with a scheme, hostname, and port but no path. Can also be set via `CONTEXTFORGE_ADDR` environment variable.
- `token` - (Optional, Sensitive) JWT token used to authenticate with the ContextForge MCP Gateway. Can also be set via `CONTEXTFORGE_TOKEN` environment variable.

Both attributes are optional in the provider configuration block but must be set via either the configuration or environment variables. Configuration values take precedence over environment variables.

### Configuration Example

```hcl
terraform {
  required_providers {
    contextforge = {
      source  = "registry.terraform.io/hashicorp/contextforge"
      version = "~> 0.1"
    }
  }
}

provider "contextforge" {
  address = "https://contextforge.example.com"
  token   = var.contextforge_token
}
```

Alternatively, use environment variables:

```bash
export CONTEXTFORGE_ADDR="https://contextforge.example.com"
export CONTEXTFORGE_TOKEN="your-jwt-token"

terraform plan
```

## Data Sources

### contextforge_gateway

Retrieves information about an existing ContextForge MCP Gateway by ID.

**Example Usage:**

```hcl
data "contextforge_gateway" "example" {
  id = "gateway-id-12345"
}

output "gateway_url" {
  value = data.contextforge_gateway.example.url
}

output "gateway_status" {
  value = {
    enabled   = data.contextforge_gateway.example.enabled
    reachable = data.contextforge_gateway.example.reachable
  }
}
```

**Key Attributes:**

- `id` - (Required) The unique identifier of the gateway to retrieve
- `name` - Gateway name
- `url` - Gateway endpoint URL
- `transport` - Transport protocol (SSE, HTTP, STDIO, STREAMABLEHTTP)
- `description` - Gateway description
- `enabled` - Whether the gateway is enabled
- `reachable` - Whether the gateway is currently reachable
- `created_at` - Gateway creation timestamp
- `updated_at` - Gateway last update timestamp

See the Terraform Registry documentation for the complete attribute reference.

### contextforge_server

Retrieves information about an existing ContextForge virtual server by ID.

**Example Usage:**

```hcl
data "contextforge_server" "example" {
  id = "server-id-12345"
}

output "server_status" {
  value = {
    name      = data.contextforge_server.example.name
    is_active = data.contextforge_server.example.is_active
  }
}

output "server_metrics" {
  value = data.contextforge_server.example.metrics
}
```

**Key Attributes:**

- `id` - (Required) The unique identifier of the server to retrieve
- `name` - Server name
- `description` - Server description
- `icon` - Server icon (URL or data URI)
- `is_active` - Whether the server is active
- `associated_tools` - Associated tool IDs
- `associated_resources` - Associated resource IDs
- `associated_prompts` - Associated prompt IDs
- `associated_a2a_agents` - Associated A2A agent IDs
- `metrics` - Nested object with performance metrics (total_executions, successful_executions, failed_executions, failure_rate, response times)
- `created_at` - Server creation timestamp
- `updated_at` - Server last update timestamp

See the Terraform Registry documentation for the complete attribute reference.

### contextforge_resource

Retrieves information about an existing ContextForge resource by ID.

**Example Usage:**

```hcl
data "contextforge_resource" "example" {
  id = "1"
}

output "resource_info" {
  value = {
    name      = data.contextforge_resource.example.name
    uri       = data.contextforge_resource.example.uri
    is_active = data.contextforge_resource.example.is_active
  }
}

output "resource_metrics" {
  value = data.contextforge_resource.example.metrics
}
```

**Key Attributes:**

- `id` - (Required) The unique identifier of the resource to retrieve
- `uri` - Resource URI
- `name` - Resource name
- `description` - Resource description
- `mime_type` - MIME type of the resource
- `size` - Resource size in bytes
- `is_active` - Whether the resource is active
- `metrics` - Nested object with performance metrics (total_executions, successful_executions, failed_executions, failure_rate, response times)
- `tags` - Resource tags
- `team_id` - Team ID
- `visibility` - Visibility setting (public, private, etc.)
- `created_at` - Resource creation timestamp
- `updated_at` - Resource last update timestamp

See the Terraform Registry documentation for the complete attribute reference.

### contextforge_tool

Retrieves information about an existing ContextForge tool by ID.

**Example Usage:**

```hcl
data "contextforge_tool" "example" {
  id = "tool-id-12345"
}

output "tool_info" {
  value = {
    name    = data.contextforge_tool.example.name
    enabled = data.contextforge_tool.example.enabled
  }
}

output "tool_input_schema" {
  value = data.contextforge_tool.example.input_schema
}
```

**Key Attributes:**

- `id` - (Required) The unique identifier of the tool to retrieve
- `name` - Tool name
- `description` - Tool description
- `input_schema` - JSON Schema defining tool input parameters
- `enabled` - Whether the tool is enabled
- `tags` - Tool tags
- `team_id` - Team ID
- `visibility` - Visibility setting (public, private, etc.)
- `created_at` - Tool creation timestamp
- `updated_at` - Tool last update timestamp

See the Terraform Registry documentation for the complete attribute reference.

## Resources

No managed resources are currently implemented. Resources for managing gateways, servers, tools, resources, and prompts will be added in future releases.

## Development

### Prerequisites

- [Go](https://golang.org/doc/install) >= 1.25
- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [uvx](https://github.com/astral-sh/uv) (for integration tests)
- [jq](https://stedolan.github.io/jq/) (for integration test scripts)

### Building the Provider

Build the provider binary:

```bash
make build
```

This creates the `terraform-provider-contextforge` binary in the project root.

### Installing for Local Development

Install the provider to your local Terraform plugins directory:

```bash
make install
```

This builds and copies the provider binary to `$GOPATH/bin`.

### Continuous Integration

The project uses GitHub Actions for automated testing. The workflow runs on:
- Push to `master` branch
- Pull requests

**Workflow Jobs:**

1. **Unit Tests** - Runs unit tests and verifies the provider builds successfully
2. **Acceptance Tests** - Runs full integration test lifecycle:
   - Starts local ContextForge gateway using `uvx`
   - Runs acceptance tests with `TF_ACC=1`
   - Cleans up gateway and test artifacts

See `.github/workflows/test.yml` for the complete workflow configuration.

## Testing

### Unit Tests

Run the unit test suite:

```bash
make test
```

### Integration Tests

Integration tests require a running ContextForge MCP Gateway instance. The test infrastructure automatically manages a local gateway using `uvx` and the `mcp-contextforge-gateway` package.

**Run the full integration test lifecycle:**

```bash
make integration-test-all
```

This target:
1. Starts a local ContextForge gateway on `http://localhost:8000`
2. Generates a JWT token for authentication
3. Runs integration tests with `TF_ACC=1`
4. Tears down the gateway after tests complete

**Manual integration test workflow:**

```bash
# Start the gateway
make integration-test-setup

# Run integration tests
make integration-test

# Stop the gateway
make integration-test-teardown
```

**Integration test infrastructure:**

The integration test setup script creates a complete test environment:

1. **ContextForge Gateway** - Launches gateway on port 8000
   - Creates admin user (`admin@test.local`)
   - Generates JWT token with 7-day expiration
   - Token stored in `tmp/contextforge-test-token.txt`
   - Gateway PID in `tmp/contextforge-test.pid`
   - Logs in `tmp/contextforge-test.log`

2. **MCP Time Server** - Starts test MCP server on port 8002
   - Provides real MCP endpoint for gateway connectivity validation
   - Uses `mcp-server-time` via `mcpgateway.translate` wrapper
   - PID stored in `tmp/time-server.pid`

3. **Test Resources** - Creates test entities for acceptance tests
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
   - Used by acceptance tests to verify data source functionality

## Makefile Targets

| Target | Description |
|--------|-------------|
| `build` | Build the provider binary |
| `install` | Install the provider locally for manual testing |
| `test` | Run unit tests |
| `clean` | Clean build artifacts and dist directory |
| `integration-test-setup` | Start local ContextForge gateway for integration testing |
| `integration-test-teardown` | Stop local ContextForge gateway and clean up |
| `integration-test` | Run integration tests (requires running gateway) |
| `integration-test-all` | Run full integration test lifecycle (setup → test → teardown) |
| `help` | Display help information |

## Contributing

Contributions are welcome. When contributing:

1. Follow the [Go Style Guide](https://google.github.io/styleguide/go/guide)
2. Ensure all tests pass (`make test`)
3. Run integration tests (`make integration-test-all`)
4. Use conventional commit format for commit messages
5. Update documentation as needed

**Documentation:**
- See [CLAUDE.md](CLAUDE.md) for detailed architecture, implementation patterns, and developer guidelines
- See [CHANGELOG.md](CHANGELOG.md) for version history and release notes

**Dependencies:**
- [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) v1.16.1
- [go-contextforge](https://github.com/leefowlercu/go-contextforge) v0.6.0 - Go client library for ContextForge MCP Gateway
