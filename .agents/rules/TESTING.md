# Testing And Validation Rules

## Scope

Apply these rules for all local validation, test additions, diagnostics changes, and completion reporting.

## Rules

1. Run the smallest relevant test set first, then broaden to package-level or full-suite tests for final validation.
2. For provider logic changes, run tests in `internal/provider/`; for cross-cutting changes, run `go test ./...`.
3. When changing acceptance behavior, run acceptance tests with the standard repository workflow rather than ad-hoc commands.
4. Add or update tests alongside behavior changes. Include success paths and at least one failure/diagnostic path when applicable.
5. Do not claim tests passed unless they were executed in the current workspace.
6. If tests are skipped or blocked, state exactly why and what remains unverified.
7. Keep test fixtures deterministic and avoid coupling tests to external state outside the repository’s integration harness.

## Integration Harness Workflow

Use the repository Make targets and scripts as the only supported harness workflow.

1. Setup: run `make integration-test-setup` to start the local gateway, supporting MCP test servers, and seeded test entities.
2. Test: run `make integration-test` to execute acceptance tests with `TF_ACC=1` and harness-provided environment variables.
3. Full lifecycle: use `make integration-test-all` for setup + test + guaranteed teardown via shell trap.
4. Teardown: run `make integration-test-teardown` to stop harness processes and remove `tmp/` artifacts.
5. Debugging rule: if setup fails, inspect `tmp/*.log` first, then fix harness assumptions before changing provider logic.

## Integration Execution Rules

1. For acceptance changes, prefer `make integration-test-all` over manually exported env vars.
2. For targeted debugging, run `make integration-test-setup`, then execute focused `go test -v -run ... ./internal/provider`, then run `make integration-test-teardown`.
3. Never leave harness processes running after a test session.
4. Treat files in `tmp/` as ephemeral runtime state; do not commit them.

## Updating Harness And Acceptance Suite

Apply these steps when upgrading `go-contextforge` or changing provider surface area (add/remove/update data sources/resources).

1. Upgrade path:
   - Update dependencies (`go.mod`, `go.sum`) and resolve compile-time API breakages in provider code first.
   - Run `go test ./...` before acceptance testing to catch fast feedback failures.
2. Harness contract checks:
   - Verify setup script API calls and payload shapes still match upstream service behavior.
   - Verify the pinned gateway package/version used by the harness (`uvx --from ...`) matches the service version targeted by the provider and SDK.
   - Verify seeded entities still create successfully and ID files in `tmp/` are written for tests that require them.
   - If test prerequisites change, update `scripts/integration-test-setup.sh`, `scripts/integration-test-teardown.sh`, and Make targets together.
3. Data source/resource change checks:
   - Register new/removed implementations in `internal/provider/provider.go`.
   - Add/update/remove acceptance tests under `internal/provider/*_test.go`.
   - If a new test needs seeded objects, add creation logic and ID file output in setup script.
   - If a test consumes values via environment variables, update the `make integration-test` exports accordingly.
4. Completion gate:
   - Run `make integration-test-all` after harness/test updates.
   - Report what passed, what skipped, and any known upstream blockers.

## Reporting Rule

When finishing a task, report exactly which test commands were run and whether they passed, failed, or were not run.
