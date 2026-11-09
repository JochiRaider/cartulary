# Tooling Instructions

Supplement the [root procedure](../AGENTS.md) for repository tooling and harness or verification-routing development, including work outside `tools`. Paths in code spans are repository-relative.

- The adopted [Testing Harness NLSpec](../docs/testing-harness-nlspec.md) owns harness mechanics. Read relevant clauses when changing those mechanics; schemas, policies, and other machine inputs project that owner.
- Verification routing belongs to `contracts/verification/**`, `tools/test_catalog_owner.json`, and `tools/test_families/**`.
- `tools/task_surface_owner.json` is the authored task-surface owner; `tools/task_surface_manifest.json` and generated Make includes are projections. `tools/execution_topology_manifest.json` owns authored execution topology and declares generated outputs.
- When changed owner inputs affect generated outputs, update those inputs first and regenerate through `make generate`. Verify affected projections with `make source-check CHECK=generated`; select `CHECK=schema` only when changed schemas or structured inputs require it. Reuse equivalent validation already completed on compatible source, including validation performed by a necessary finalizer. Generated outputs never substitute for authored owners.
- Tooling changes follow the root targeted-check stopping rule. Editing harness code alone does not require a whole owner suite, all source checks, or `agent-finalize`. Select broader validation or finalization only for a concrete affected contract or artifact that narrower checks do not cover.
- Tool versions come from `tools/toolchain_pins.json`; `make source-check CHECK=toolchain` checks production/bootstrap inputs, not prose mirrors. Do not duplicate pin or public-target inventories.
- For ordinary test execution or diagnosis, use the [test-operations skill](../.agents/skills/cartulary-test-ops/SKILL.md); it supplies the workflow independently of this development guidance.
