# Source Of Truth Rules

## Purpose

Use these rules to keep agent outputs accurate and resistant to staleness.

## Rules

1. Treat implementation files as the source of truth for behavior. Prefer reading code and tests over relying on memory.
2. Do not document or assert exhaustive inventories (all fields, all tests, all resources) unless the task explicitly requires it.
3. Prefer durable guidance over volatile facts. If a fact can change often, point to where to verify it.
4. Verify claims by inspection before stating them. For behavior claims, read the relevant file; for execution claims, run the command.
5. When updating docs, describe workflows and decision rules, not snapshots of current state.
6. If upstream library behavior matters, validate against the local dependency version in `go.mod` and the local clone when needed.

## Minimum Verification Checklist

1. Confirm touched behavior in code under `internal/provider/` (and `internal/tfconv/` when conversion logic is involved).
2. Confirm test expectations in the corresponding `*_test.go` files.
3. Run relevant tests before reporting completion.
