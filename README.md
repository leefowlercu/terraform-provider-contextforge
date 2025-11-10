# terraform-provider-contextforge

Terraform provider for IBM ContextForge MCP Gateway management. Manages virtual servers, gateways, tools, resources, and prompts for the ContextForge MCP Gateway service.

## Table of Contents

- [Overview](#overview)
- [Requirements](#requirements)
- [Provider Configuration](#provider-configuration)
  - [Authentication](#authentication)
  - [Configuration Example](#configuration-example)
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
      source  = "hashicorp/contextforge"
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

The integration test setup script:
- Launches ContextForge gateway on port 8000
- Creates an admin user (`admin@test.local`)
- Generates a JWT token with 7-day expiration
- Stores the token in `tmp/contextforge-test-token.txt`
- Stores the gateway PID in `tmp/contextforge-test.pid`
- Stores logs in `tmp/contextforge-test.log`

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

The provider uses:
- [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) v1.16.1
- [go-contextforge](https://github.com/leefowlercu/go-contextforge) v0.5.0 - Go client library for ContextForge MCP Gateway
