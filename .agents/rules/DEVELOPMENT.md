# Provider Development Rules

## Scope

Apply these rules when adding or changing provider configuration, data sources, resources, schema attributes, or API mappings.

## Rules

1. Follow Terraform Plugin Framework lifecycle patterns consistently: `Metadata`, `Schema`, `Configure`, and CRUD/read methods.
2. Keep configuration precedence explicit: environment defaults first, then HCL overrides, then validation.
3. Define schema intent precisely. Use `Required`, `Optional`, `Computed`, and plan modifiers to match API behavior and avoid drift.
4. Keep API-to-Terraform type conversion centralized. Reuse `internal/tfconv` helpers and add helpers there when conversion logic is repeated.
5. Prefer direct API read endpoints when available. If only list APIs exist, implement list+filter with pagination and clear not-found handling.
6. Keep diagnostics actionable: explain what failed, where, and what user input or environment condition likely caused it.
7. Preserve backward compatibility where possible. Avoid breaking schema changes unless explicitly required, and document migration impact.
8. Register every new data source/resource in `internal/provider/provider.go` and ensure import/state behavior is covered by tests.
9. Keep read-after-write behavior explicit. If an API returns partial update responses, follow updates with a read to normalize state.
