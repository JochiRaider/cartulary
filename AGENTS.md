# Cartulary Repository Procedure

## Authority and boundaries

- Adopted subsystem NLSpecs and normative Core sections govern behavior. Typed contracts are downstream projections; tests and routing do not define requirements or prove specification completeness. Repair conflicting projections or implementation within the authorized scope; humans review projection fidelity.
- Tests, generators, runtime metadata, conformance, and release evidence must not read, stat, hash, or otherwise depend on `docs/`, README, or other Markdown.
- Keep `cmd/*` composition-only, `internal/app/*` responsible for application assembly, `internal/modules/*` for domain/application behavior, and `internal/platform/*` for transport and runtime plumbing.

## Load guidance when relevant

Before planning or editing a subtree, read its applicable nested `AGENTS.md` files. Load additional instruction chains when crossing directories; avoid unrelated guidance.

- Backend work in `cmd`, `db`, or `internal`: [backend instructions](internal/AGENTS.md).
- Harness or verification-routing development anywhere, including `internal/testutil` and `contracts/verification`: [tooling instructions](tools/AGENTS.md).
- Test selection, execution, observation, or diagnosis: [test-operations skill](.agents/skills/cartulary-test-ops/SKILL.md). Ordinary test execution does not require tooling-development guidance.
- Domain changes: relevant vocabulary and owner sections in [domain](docs/domain.md). Presentation changes: relevant [design](docs/design.md) sections. Rendered UI inspection: [UI-review skill](.agents/skills/cartulary-ui-review/SKILL.md).
- Golden changes: [maintenance guide](docs/guides/cartulary_visual_golden_maintenance.md). Visual/accessibility evidence does not satisfy Core 05 publication requirements by itself.

## Editing and verification

- Do not hand-edit paths covered by `generated_roots` or `generated_files` in the [generated-artifact policy](tools/generated_artifact_policy.json), outputs declared by the [execution topology](tools/execution_topology_manifest.json), or tool-managed dependencies such as `go.sum` and `pnpm-lock.yaml`. Update authored inputs and regenerate through Make.
- Run repository operations through public Make targets from the repository root. Direct tool/script invocations are not canonical verification. Discover commands with `make help` / `make help-all`; read versions from the [pin registry](tools/toolchain_pins.json).
- Use `make task-guide ROLE=module-author OWNER=<owner-id>` to select narrow verification. Investigate failures before rerunning or broadening; broaden only when ownership or risk requires it.
- For implementation tasks, run `make agent-finalize` before broader end-of-run verification; it can rewrite generated artifacts. Supply `RESULTS_DIR=<successful full warm check run root>` for retained-run maintenance; otherwise report that maintenance was skipped because `RESULTS_DIR` was unset.
- Review-only work does not authorize mutations. Documentation-only changes use `make lint-markdown`; skip `format`, generation, `agent-finalize`, and product verification. Do not change product checks, generated artifacts, conformance, or release evidence for documentation maintenance.
- `make clean` deletes reproducible build/report outputs; `make distclean` also deletes repo-local caches and dependency installs.
- Report substantive changes, verification commands/results, and skipped required checks with reasons. For failures, include the target, summary artifact or run root, and relationship to the change.
