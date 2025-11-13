# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.0] - 2025-11-13

### Added

- Ufeat: add gateway data source with 77 attributes and type conversion utilities
- Ufeat: initial commit, add provider, providerserver, test terraform module, project documentation, initial release toolchain, make targets, integration test setup & teardown scripts


### Documentation

- Udocs: document gateway data source, ci/cd, and package structure


### Build

- Ubuild: update build/release toolchain
- Ubuild: update build/release toolchain
- Ubuild: update build/release toolchain and initial version of provider binary
- Ubuild: update goreleaser changelog commit message filters
- Ubuild: update prepare release script to correctly prompt for gpg key passphrase during release prep
- Ubuild: update prepare release script to handle initial commit scenarios
- Uchore: add terraform registry manifest file


### Tests

- Utest: add mcp time server and test gateway creation to integration setup
## [Unreleased]

### Added

- Initial provider implementation with IBM ContextForge MCP Gateway support
- Gateway data source with 77 attributes covering core fields, authentication, organizational metadata, timestamps, and custom metadata
- Type conversion utilities for complex Terraform types (map[string]any, timestamps, dynamic types)
- Integration test infrastructure with automated gateway lifecycle management
- CI/CD workflow for unit and acceptance tests via GitHub Actions
- Release toolchain with GoReleaser configuration for multi-platform builds
- MCP time server integration for testing gateway connectivity

### Documentation

- Provider architecture documentation in CLAUDE.md
- Gateway data source implementation details
- Integration testing setup and teardown procedures
- Package structure and development commands

[Unreleased]: https://github.com/leefowlercu/terraform-provider-contextforge/commits/dev

[v0.1.0]: https://github.com/leefowlercu/terraform-provider-contextforge/releases/tag/v0.1.0
