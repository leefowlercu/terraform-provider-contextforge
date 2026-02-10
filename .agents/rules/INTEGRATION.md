# Integration Environment Rules

## Scope

Apply these rules when running or editing acceptance/integration test setup and teardown flows.

## Rules

1. Use repository automation targets/scripts for setup, test execution, and teardown to keep runs reproducible.
2. Keep integration setup idempotent: repeated runs should not require manual cleanup first.
3. Ensure teardown always runs after integration testing, including failure paths.
4. Treat generated runtime files (PIDs, logs, tokens, temp databases) as ephemeral and keep them out of commits.
5. When integration setup fails, inspect harness logs first and fix script/environment assumptions before changing provider code.
6. Use dedicated test entities created by the harness; do not rely on personal or shared external environments.
7. If modifying integration scripts, document new prerequisites and failure modes in script comments or adjacent docs.
