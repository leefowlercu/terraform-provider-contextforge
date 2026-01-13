# Terraform Provider ContextForge

[![Build Status](https://github.com/leefowlercu/terraform-provider-contextforge/actions/workflows/test.yml/badge.svg)](https://github.com/leefowlercu/terraform-provider-contextforge/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/leefowlercu/terraform-provider-contextforge)](https://goreportcard.com/report/github.com/leefowlercu/terraform-provider-contextforge)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

Terraform provider for IBM ContextForge MCP Gateway management. Manages virtual servers, gateways, tools, resources, agents, and prompts for the ContextForge MCP Gateway service.

**Current Version**: v0.3.0 ([CHANGELOG.md](CHANGELOG.md))

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Requirements](#requirements)
- [Quick Start](#quick-start)
- [Provider Configuration](#provider-configuration)
  - [Authentication](#authentication)
  - [Configuration Example](#configuration-example)
- [Data Sources](#data-sources)
  - [contextforge_agent](#contextforge_agent)
  - [contextforge_gateway](#contextforge_gateway)
  - [contextforge_prompt](#contextforge_prompt)
  - [contextforge_resource](#contextforge_resource)
  - [contextforge_server](#contextforge_server)
  - [contextforge_team](#contextforge_team)
  - [contextforge_tool](#contextforge_tool)
- [Resources](#resources)
  - [contextforge_agent](#contextforge_agent-resource)
  - [contextforge_gateway](#contextforge_gateway-resource)
  - [contextforge_resource](#contextforge_resource-resource)
  - [contextforge_server](#contextforge_server-resource)
  - [contextforge_tool](#contextforge_tool-resource)
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

The provider is built using the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) v1.16.1 and communicates with the ContextForge MCP Gateway API via the [go-contextforge](https://github.com/leefowlercu/go-contextforge) v0.8.1 client library. The `internal/tfconv` package handles type conversions between the client library's Go types and Terraform Plugin Framework types, including conversions for heterogeneous maps, authentication headers, and RFC3339 timestamps.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.25 (for development)
- Access to a ContextForge MCP Gateway instance

## Quick Start

1. Configure the provider in your Terraform configuration:

```hcl
terraform {
  required_providers {
    contextforge = {
      source  = "registry.terraform.io/hashicorp/contextforge"
      version = "~> 0.2"
    }
  }
}

provider "contextforge" {
  address = "https://contextforge.example.com"
  token   = var.contextforge_token
}
```

2. Set your environment variables:

```bash
export CONTEXTFORGE_ADDR="https://contextforge.example.com"
export CONTEXTFORGE_TOKEN="your-jwt-token"
```

3. Create a resource:

```hcl
resource "contextforge_server" "example" {
  name        = "my-mcp-server"
  description = "My MCP virtual server"
}
```

4. Apply the configuration:

```bash
terraform init
terraform plan
terraform apply
```

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
      version = "~> 0.2"
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

### contextforge_agent

Retrieves information about an existing ContextForge A2A agent by ID.

**Example Usage:**

```hcl
data "contextforge_agent" "example" {
  id = "agent-id-12345"
}

output "agent_info" {
  value = {
    name         = data.contextforge_agent.example.name
    endpoint_url = data.contextforge_agent.example.endpoint_url
    enabled      = data.contextforge_agent.example.enabled
  }
}

output "agent_metrics" {
  value = data.contextforge_agent.example.metrics
}
```

**Key Attributes:**

- `id` - (Required) The unique identifier of the agent to retrieve
- `name` - Agent name
- `description` - Agent description
- `endpoint_url` - Agent endpoint URL
- `enabled` - Whether the agent is enabled
- `capabilities` - Agent capabilities (dynamic object)
- `config` - Agent configuration (dynamic object)
- `metrics` - Nested object with performance metrics (total_requests, successful_requests, failed_requests, failure_rate, response times)
- `tags` - Agent tags
- `team_id` - Team ID
- `visibility` - Visibility setting (public, private, etc.)
- `created_at` - Agent creation timestamp
- `updated_at` - Agent last update timestamp

See the Terraform Registry documentation for the complete attribute reference.

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

### contextforge_prompt

Retrieves information about an existing ContextForge prompt by ID.

**Example Usage:**

```hcl
data "contextforge_prompt" "example" {
  id = 123
}

output "prompt_info" {
  value = {
    name      = data.contextforge_prompt.example.name
    template  = data.contextforge_prompt.example.template
    is_active = data.contextforge_prompt.example.is_active
  }
}

output "prompt_arguments" {
  value = data.contextforge_prompt.example.arguments
}

output "prompt_metrics" {
  value = data.contextforge_prompt.example.metrics
}
```

**Key Attributes:**

- `id` - (Required) The unique identifier of the prompt to retrieve (integer)
- `name` - Prompt name
- `description` - Prompt description
- `template` - Prompt template with parameter placeholders
- `arguments` - List of prompt arguments/parameters (name, description, required)
- `is_active` - Whether the prompt is active
- `tags` - Prompt tags
- `metrics` - Nested object with performance metrics (total_executions, successful_executions, failed_executions, failure_rate, response times)
- `team_id` - Team ID
- `team` - Team name
- `owner_email` - Owner email address
- `visibility` - Visibility setting (public, private, etc.)
- `created_at` - Prompt creation timestamp
- `updated_at` - Prompt last update timestamp

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

### contextforge_server

Retrieves information about an existing ContextForge server by ID.

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

### contextforge_team

Retrieves information about an existing ContextForge team by ID.

**Example Usage:**

```hcl
data "contextforge_team" "example" {
  id = "team-id-12345"
}

output "team_info" {
  value = {
    name         = data.contextforge_team.example.name
    slug         = data.contextforge_team.example.slug
    is_personal  = data.contextforge_team.example.is_personal
    is_active    = data.contextforge_team.example.is_active
    member_count = data.contextforge_team.example.member_count
  }
}
```

**Key Attributes:**

- `id` - (Required) The unique identifier of the team to retrieve
- `name` - Team name
- `slug` - Team slug (URL-friendly identifier)
- `description` - Team description
- `is_personal` - Whether this is a personal team
- `visibility` - Visibility setting (public, private, etc.)
- `max_members` - Maximum number of members allowed in the team
- `member_count` - Current number of members in the team
- `is_active` - Whether the team is active
- `created_by` - Email address of the user who created the team
- `created_at` - Team creation timestamp (RFC3339 format)
- `updated_at` - Team last update timestamp (RFC3339 format)

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

The provider supports full CRUD operations for the following managed resources.

### contextforge_agent (Resource)

Manages a ContextForge A2A (Agent-to-Agent) agent resource.

**Example Usage:**

```hcl
resource "contextforge_agent" "example" {
  name         = "my-agent"
  endpoint_url = "https://agent.example.com/api"
  description  = "My A2A agent"
  enabled      = true
  tags         = ["production", "a2a"]
}
```

**Required Attributes:**

- `name` - Agent name
- `endpoint_url` - Agent endpoint URL

**Optional Attributes:**

- `description` - Agent description
- `agent_type` - Agent type
- `protocol_version` - Protocol version
- `config` - Agent configuration (dynamic object)
- `auth_type` - Authentication type
- `enabled` - Whether the agent is enabled
- `tags` - List of tags
- `team_id` - Team ID
- `visibility` - Visibility setting

**Read-Only Attributes:**

- `id` - Agent unique identifier
- `slug` - URL-friendly identifier
- `capabilities` - Agent capabilities (dynamic object)
- `reachable` - Whether the agent is reachable
- `metrics` - Performance metrics object
- `created_at`, `updated_at` - Timestamps

### contextforge_gateway (Resource)

Manages a ContextForge MCP Gateway resource.

**Example Usage:**

```hcl
resource "contextforge_gateway" "example" {
  name        = "my-gateway"
  url         = "http://localhost:8080/sse"
  transport   = "SSE"
  description = "My MCP gateway"
  enabled     = true
  tags        = ["production"]
}
```

**Required Attributes:**

- `name` - Gateway name
- `url` - Gateway endpoint URL (must be reachable)
- `transport` - Transport protocol (SSE, HTTP, STDIO, STREAMABLEHTTP)

**Optional Attributes:**

- `description` - Gateway description
- `enabled` - Whether the gateway is enabled
- `auth_type` - Authentication type
- `auth_username`, `auth_password` - Basic authentication credentials
- `auth_token` - Bearer token authentication
- `auth_header_key`, `auth_header_value` - Custom header authentication
- `oauth_config` - OAuth configuration (dynamic object)
- `tags` - List of tags
- `team_id` - Team ID
- `visibility` - Visibility setting

**Read-Only Attributes:**

- `id` - Gateway unique identifier
- `slug` - URL-friendly identifier
- `reachable` - Whether the gateway is reachable
- `capabilities` - Gateway capabilities (dynamic object)
- `created_at`, `updated_at`, `last_seen` - Timestamps

### contextforge_resource (Resource)

Manages a ContextForge resource entity.

**Example Usage:**

```hcl
resource "contextforge_resource" "example" {
  uri         = "config://app/settings"
  name        = "app-settings"
  content     = jsonencode({ theme = "dark", language = "en" })
  description = "Application settings"
  mime_type   = "application/json"
  tags        = ["config"]
}
```

**Required Attributes:**

- `uri` - Resource URI
- `name` - Resource name
- `content` - Resource content (write-only, not returned by API)

**Optional Attributes:**

- `description` - Resource description
- `mime_type` - MIME type of the resource
- `tags` - List of tags
- `team_id` - Team ID (can only be set at creation)
- `visibility` - Visibility setting (can only be set at creation)

**Read-Only Attributes:**

- `id` - Resource unique identifier
- `size` - Resource size in bytes
- `is_active` - Whether the resource is active
- `metrics` - Performance metrics object
- `created_at`, `updated_at` - Timestamps

### contextforge_server (Resource)

Manages a ContextForge virtual server resource.

**Example Usage:**

```hcl
resource "contextforge_server" "example" {
  name        = "my-server"
  description = "My MCP virtual server"
  icon        = "https://example.com/icon.png"
  tags        = ["production"]

  associated_tools     = ["tool-id-1", "tool-id-2"]
  associated_resources = ["resource-id-1"]
  associated_prompts   = ["prompt-id-1"]
}
```

**Required Attributes:**

- `name` - Server name

**Optional Attributes:**

- `description` - Server description
- `icon` - Server icon URL
- `associated_tools` - List of associated tool IDs
- `associated_resources` - List of associated resource IDs
- `associated_prompts` - List of associated prompt IDs
- `associated_a2a_agents` - List of associated A2A agent IDs
- `tags` - List of tags
- `team_id` - Team ID
- `visibility` - Visibility setting

**Read-Only Attributes:**

- `id` - Server unique identifier
- `is_active` - Whether the server is active
- `metrics` - Performance metrics object (total_executions, successful_executions, failed_executions, failure_rate, response times)
- `created_at`, `updated_at` - Timestamps

### contextforge_tool (Resource)

Manages a ContextForge tool resource.

**Example Usage:**

```hcl
resource "contextforge_tool" "example" {
  name        = "my-tool"
  description = "My MCP tool"
  enabled     = true
  tags        = ["utility"]

  input_schema = jsonencode({
    type = "object"
    properties = {
      message = {
        type        = "string"
        description = "The message to process"
      }
    }
    required = ["message"]
  })
}
```

**Required Attributes:**

- `name` - Tool name

**Optional Attributes:**

- `description` - Tool description
- `input_schema` - JSON Schema defining tool input parameters (dynamic)
- `enabled` - Whether the tool is enabled
- `tags` - List of tags
- `team_id` - Team ID
- `visibility` - Visibility setting

**Read-Only Attributes:**

- `id` - Tool unique identifier
- `created_at`, `updated_at` - Timestamps

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
3. Creates test resources (gateway, server, tool, resource, team, agent, prompt)
4. Runs integration tests with `TF_ACC=1`
5. Tears down the gateway after tests complete

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
   - Test Gateway, Server, Tool, Resource, Team, Agent, and Prompt
   - IDs saved to `tmp/contextforge-test-*-id.txt` files

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
| `integration-test-all` | Run full integration test lifecycle (setup -> test -> teardown) |
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
- [go-contextforge](https://github.com/leefowlercu/go-contextforge) v0.8.1 - Go client library for ContextForge MCP Gateway
