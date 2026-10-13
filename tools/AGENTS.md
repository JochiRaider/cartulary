# Tooling Instructions

Supplement the [root procedure](../AGENTS.md) for repository tooling and harness or verification-routing development, including work outside `tools`. Paths in code spans are repository-relative.

- The adopted [Testing Harness NLSpec](../docs/testing-harness-nlspec.md) owns harness mechanics. Read relevant clauses when changing those mechanics; schemas, policies, and other machine inputs project that owner.
- Verification routing belongs to `contracts/verification/**`, `tools/test_catalog_owner.json`, and `tools/test_families/**`.
- `tools/task_surface_owner.json` is the authored task-surface owner; `tools/task_surface_manifest.json` and generated Make includes are projections. `tools/execution_topology_manifest.json` owns authored execution topology and declares generated outputs.
- Update owner inputs first. Refresh downstream task-surface, topology, and schedule artifacts through `make generate`; verify with `make generate-drift`, `make generated-artifact-policy-check`, and `make json-shape-check` as applicable. Generated outputs never substitute for authored owners.
- Tool versions come from `tools/toolchain_pins.json`; `make toolchain-drift` checks production/bootstrap inputs, not prose mirrors. Do not duplicate pin or public-target inventories.
- For ordinary test execution or diagnosis, use the [test-operations skill](../.agents/skills/cartulary-test-ops/SKILL.md); it supplies the workflow independently of this development guidance.
