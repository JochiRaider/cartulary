# Cartulary Repository Procedure

## Authority and editing safeguards

- Adopted subsystem NLSpecs and normative Core sections govern behavior under [Core 00 precedence](docs/spec/00_document_set_status_and_precedence.md). Typed contracts are downstream projections; tests and routing do not define requirements or prove specification completeness. Repair conflicting projections or implementation within the authorized scope; humans review projection fidelity.
- Tests, generators, runtime metadata, conformance, and release evidence must not read, stat, hash, or otherwise depend on `docs/`, README, or other Markdown.
- Do not hand-edit paths covered by `generated_roots` or `generated_files` in the [generated-artifact policy](tools/generated_artifact_policy.json), outputs declared by the [execution topology](tools/execution_topology_manifest.json), or tool-managed dependencies such as `go.sum` and `pnpm-lock.yaml`. Update authored inputs and regenerate through Make.

## Load guidance when relevant

Before planning or editing a subtree, read its applicable nested `AGENTS.md` files. Load additional instruction chains when crossing directories; avoid unrelated guidance.

- Backend work in `cmd`, `db`, or `internal`: [backend instructions](internal/AGENTS.md).
- Harness or verification-routing development anywhere, including `internal/testutil` and `contracts/verification`: [tooling instructions](tools/AGENTS.md).
- Test selection, execution, observation, or diagnosis: [test-operations skill](.agents/skills/cartulary-test-ops/SKILL.md). Ordinary test execution does not require tooling-development guidance.
- Domain changes: relevant vocabulary and owner sections in [domain](docs/domain.md). Presentation changes: relevant [design](docs/design.md) sections. Rendered UI inspection: [UI-review skill](.agents/skills/cartulary-ui-review/SKILL.md).
- Browser visual golden changes: [maintenance guide](docs/guides/cartulary_visual_golden_maintenance.md).

## Workflow by task

- Review-only work does not authorize mutations.
- Documentation-only changes, wherever located, use `make lint-markdown`; skip `format`, generation, `agent-finalize`, and product verification. Do not change product checks, generated artifacts, conformance, or release evidence for documentation maintenance.
- Implementation work, including mixed code/documentation tasks, uses `make task-guide OWNER=<owner-id>` to select narrow verification. Investigate failures before rerunning or broadening; broaden only when ownership or risk requires it. Run `make agent-finalize` before broader end-of-run verification; it can rewrite generated artifacts. Follow the test-operations skill for retained-run maintenance.

Use public Make targets from the repository root for managed development, build, dependency, generation, validation, and maintenance workflows. File inspection and ordinary Git operations are outside this requirement. Direct tool/script invocations are not canonical verification. Discover commands with `make help` / `make help ALL=1`; read versions from the [pin registry](tools/toolchain_pins.json).

## Completion reporting

Report substantive changes, verification commands/results, and skipped required checks with reasons. For failures, include the target, summary artifact or run root, and relationship to the change.
