# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.2.0] - 2025-11-13

### Added

- Feat: add contextforge_resource data source
- Feat: add contextforge_server data source
- Feat: add contextforge_tool data source
- Feat: add team, agent, and prompt data sources

### Fixed

- Fix: add .next-version to gitignore for release workflow

### Documentation

- Docs: update documentation for new data sources
- Docs: update documentation for resource data source
- Docs: various updates to project documentation and minor fixes

### Build

- Build: update go-contextforge to v0.6.0
- Build: update prepare-release script
- Build: update prepare-release script to handle the "unreleased" section of the existing root changelog

### Tests

- Test: add resource creation to integration test setup
- Test: add team, agent, and prompt to integration test setup

## [v0.1.0] - 2025-11-13

### Added

- Add gateway data source with attributes and type conversion utilities
- Initial commit, add provider, providerserver, test terraform module, project documentation, initial release toolchain, make targets, integration test setup & teardown scripts

### Documentation

- Document gateway data source, ci/cd, and package structure

### Build

- Update build/release toolchain
- Update build/release toolchain and initial version of provider binary
- Update goreleaser changelog commit message filters
- Update prepare release script to correctly prompt for gpg key passphrase during release prep
- Update prepare release script to handle initial commit scenarios
- Add terraform registry manifest file

### Tests

- Add mcp time server and test gateway creation to integration setup

[Unreleased]: https://github.com/leefowlercu/terraform-provider-contextforge/compare/v0.1.0...HEAD
[v0.1.0]: https://github.com/leefowlercu/terraform-provider-contextforge/releases/tag/v0.1.0

[v0.2.0]: https://github.com/leefowlercu/terraform-provider-contextforge/compare/v0.1.0...v0.2.0
