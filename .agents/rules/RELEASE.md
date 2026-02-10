# CI And Release Rules

## Scope

Apply these rules when changing CI workflows, release automation, build metadata, or artifact generation behavior.

## Rules

1. Keep local workflow and CI workflow aligned. If one changes, update the other or explain the intentional difference.
2. Maintain a green baseline for unit and acceptance jobs before merging release-affecting changes.
3. Keep version derivation and injection automated. Do not hardcode release versions in source files.
4. When changing release scripts or goreleaser config, verify outputs still match expected provider naming and metadata conventions.
5. Prefer Makefile targets as the public interface for CI/release operations; keep script internals behind those targets.
6. Document any new release prerequisite in the script or configuration that enforces it.
7. Preserve reproducibility: avoid release steps that depend on local, undocumented machine state.
8. Do not manually author release sections in `CHANGELOG.md`; versioned changelog entries must be generated from commit history by the release workflow.
9. Do not manually add arbitrary prose under `## [Unreleased]` in `CHANGELOG.md`; keep it empty or commit-derived only.
10. If changelog updates are needed, use release automation (`make release-dry-run VERSION=vX.Y.Z` for local verification, then release targets), not ad-hoc edits.
