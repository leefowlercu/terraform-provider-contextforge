# Agent Documentation Maintenance Rules

## Scope

Apply these rules when editing `AGENTS.md` or any file in `.agents/rules/`.

## Rules

1. Keep the root `AGENTS.md` concise and router-focused. It should direct agents to detailed rule docs, not duplicate them.
2. Use progressive disclosure: high-level routing in `AGENTS.md`, implementation-specific guidance in `.agents/rules/*.md`.
3. Write rules as durable instructions about process and decision-making, not snapshots of current repository contents.
4. Avoid long lists of current resources, fields, versions, or test names in agent guidance docs.
5. When a new workflow area appears, add a new `.agents/rules/<TOPIC>.md` file and link it from `AGENTS.md`.
6. Keep cross-links working. After edits, verify linked rule files exist and headings are sensible.
7. Prefer imperative language (`Do`, `Use`, `Avoid`) and explicit trigger conditions (`Read this when ...`).
