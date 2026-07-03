# Test Harness Simplification Refactor Tracker

## 1. Status and Authority

| Field | Value |
| --- | --- |
| Target path | `tools/harness` |
| Output path | `docs/handoffs/test-harness-module-refactor-tracker.md` |
| Active iteration | Simplification/removal iteration after historical NI-00 through NI-05 and WS-5 completion. |
| Current harness files | 249 tracked files under `tools/harness`; 0 untracked files from `git ls-files --others --exclude-standard tools/harness`. |
| Current task surface | 136 targets: 94 public, 31 internal helper, 11 check-internal. |
| Current status | SI-00 rebaseline complete. SI-01 standalone output-policy test cleanup complete in this pass. SI-02 is the next implementation slice. |
| Public behavior posture | Public Make targets, command IDs, schemas, output modes, result roots, failure taxonomy, scheduler semantics, readiness/cache semantics, and retained artifacts remain frozen unless an item is explicitly spec-gated. |
| Product boundary posture | Core 00 through Core 04 own product behavior. Do not move auth, workbook behavior, projections, saved views, product routes, WebSocket semantics, record models, or generated protocol contracts into `tools/harness`. |
| Generated artifact posture | Do not hand-edit generated roots or generated harness outputs. Owner-input changes require Make-owned generation and drift validation. |

Owner documents and inputs used for this iteration:

| Document or input | Role |
| --- | --- |
| `docs/testing-harness-nlspec.md` | Owns harness mechanics and public contracts. |
| `docs/domain.md` | Confirms harness vocabulary is implementation-support unless promoted by an owner document. |
| Core 00 through Core 04 | Own product behavior boundaries. |
| Core 05 | Applies only to claim-bearing timed, benchmark, fixture-sensitive, or publication evidence. |
| `docs/handoffs/cartulary_modular_refactor_planning_framework.md` | Refactor posture and repo-truth guidance. |
| `tools/task_surface_manifest.json` | Current task-surface mirror and owner input. |
| `tools/execution_topology_manifest.json` | Current topology/scheduler owner input. |
| Relevant `tools/harness` sources | Live implementation truth for simplification candidates. |

## 2. Historical Closure

NI-00 through NI-05 plus WS-5 are historical completed work. Do not re-execute those slices.

Completed historical work includes:

- The NI-00/NI-01 tracker refresh and generic JSON helper ownership move.
- WS-1 backend/browser shard-planner row normalization.
- WS-2 `core/test-output` adapter extraction.
- WS-3 spec-gated legacy simplification, including public `db-down` removal and generated task-surface refresh.
- WS-4 validation-only harness import-boundary enforcement.
- WS-5 final validation and handoff, including broad `make check`.

Stale historical text superseded by this tracker:

- Any 224, 237, or 238 file-count baseline.
- Any active NI-00 through NI-05 or WS-1 through WS-5 remediation instruction.
- The old "optional follow-up" for `tools/harness/core/tests/test-tool-output-real-targets.sh`; SI-01 closed that item.

## 3. Rebaseline and Tracker Gate

SI-00 is complete for this pass.

Live inventory:

- `git ls-files tools/harness`: passed; 249 tracked files.
- `git ls-files --others --exclude-standard tools/harness`: passed; 0 untracked files.
- Subtree counts: backend 26, browser 31, core 44, frontend 23, generated-artifacts 19, planning 26, readiness 18, scheduler 37, static-analysis 22, test-support 3.

Task-surface validation:

- `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'`: passed.
- Current target-class counts are public 94, check-internal 11, internal helper 31.

Required tracker-creation gates:

- `make check-harness-smoke`: passed.
- `make harness-contract`: passed.
- `make lint-markdown`: passed after the tracker edit.

Entry criteria for any remaining slice:

1. Update this tracker with the selected slice state before editing implementation files.
2. Re-run the narrow baseline needed by that slice.
3. Stop if baseline failures are not classified as pre-existing, environmental, or intentionally slice-blocking.
4. Identify generated owner inputs before touching any generator-related path.

Exit criteria for every implementation slice:

1. Update this tracker before starting the next slice.
2. Preserve public target names, command IDs, schemas, output modes, result roots, artifact contracts, scheduler semantics, failure taxonomy, readiness/cache behavior, and product boundaries.
3. Record validation commands, pass/fail status, retained roots when emitted, skipped checks with reasons, and the exact next action.

## 4. Active Work Items

### SI-01: Standalone Real-Target Output Test Cleanup

- Disposition/status: Complete in this pass.
- Belongs in: tests and tracker.
- Remediation: Removed the stale `lint-shell` `[RESULT]` expectation from `tools/harness/core/tests/test-tool-output-real-targets.sh`; kept `lint-shell` public output unchanged. While validating the adjusted script, also narrowed stale expectations for the current `lint` aggregate artifact roles, current `explain-target` invalid-usage diagnostic, and the test's own synthetic run ID length.
- Rationale: The old test encoded compatibility assumptions that no longer match the NLSpec/current target contracts. `lint-shell` summary-mode success is retained in artifacts with empty stdout/stderr; `lint` is a run-level aggregate; invalid target selection is classified as a usage error without nearest-candidate metadata.
- Expected long-term benefit: The standalone diagnostic now tests current contracts instead of preserving removed output details.
- Compatibility or migration impact: No public target behavior changed. Only private test expectations changed.
- Risks of leaving unresolved: The known follow-up would keep producing false failures, and additional stale rows would remain hidden behind the first failure.
- Validation criteria: Direct summary-mode `lint-shell` probe exits 0 with 0 stdout and 0 stderr; `bash tools/harness/core/tests/test-tool-output-real-targets.sh` passes; `make check-harness-smoke` passes; `make harness-contract` passes; `make lint-markdown` passes after tracker update.
- Rollback or containment: Restore the previous test rows only if public output behavior is intentionally restored by owner documents.

### SI-02: Delete Unused Private Re-Export Shims

- Disposition/status: Remove now; not started.
- Belongs in: implementation, tests, tracker.
- Remediation: After a fresh `rg` confirms no live references, delete `tools/harness/backend/runner/go-shards.mjs`, `tools/harness/backend/duration/baselines.mjs`, `tools/harness/backend/drift/manifests.mjs`, and `tools/harness/browser/runtime/paths.mjs`.
- Rationale: These are compatibility leftovers from earlier module moves, not durable subsystem contracts.
- Expected long-term benefit: Clearer ownership and fewer private paths future code can accidentally depend on.
- Compatibility or migration impact: Private import compatibility only. No public Make or schema changes.
- Risks of leaving unresolved: Stale paths continue to invite new imports and make ownership harder to reason about.
- Validation criteria: `rg` each path and exported symbol before deletion; `make check-harness-smoke`; `make harness-contract`; `make lint-scripts`; `make lint-markdown`.
- Rollback or containment: Restore the deleted shim file if a hidden repo-local caller is found.

### SI-03: Retire Generated-Artifacts JSON-Shape Shim

- Disposition/status: Owner-input gated; not started.
- Belongs in: implementation, generated outputs, tests, tracker.
- Remediation: Update owner inputs/backing-script references from `tools/harness/generated-artifacts/json-shape.mjs` to `tools/harness/core/json-shape.mjs`, then remove the shim and refresh generated outputs only through Make-owned targets.
- Rationale: Generic JSON shape helpers are core harness contracts, not generated-artifact behavior.
- Expected long-term benefit: Removes stale generated-artifacts ownership from task-surface/topology metadata.
- Compatibility or migration impact: Generated/topology references make this unsafe as a plain delete. Public behavior must remain unchanged.
- Risks of leaving unresolved: The task surface continues to advertise a compatibility shim as a backing script.
- Validation criteria: Make-owned generation/drift path; `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'`; `make generated-artifact-policy-check`; `make json-shape-check`; `make check-harness-smoke`; `make harness-contract`; `make lint-markdown`.
- Rollback or containment: Revert owner-input and generated-output changes together.

### SI-04: Retire Executable-Only Shard Planner Compatibility Paths

- Disposition/status: Remove now with generated-owner updates; not started.
- Belongs in: implementation, tests, generated outputs, tracker.
- Remediation: Replace executable-only compatibility behavior in `backend/go-shard-plan.mjs` and `browser/browser-shard-plan.mjs` with explicit CLI entrypoints. Move browser discovery/duration-maintenance CLI behavior to a planning-owned CLI, update Make node-tool references and tests, then remove matching import-boundary allow-list entries.
- Rationale: Backend/browser planner libraries should be pure over explicit inputs; discovery belongs at CLI/planning boundaries.
- Expected long-term benefit: Reduces cross-subsystem coupling and makes future phase growth easier.
- Compatibility or migration impact: Public target names, output classes, schemas, and artifacts must remain unchanged. Generated/task-surface metadata updates must be Make-owned.
- Risks of leaving unresolved: Executable compatibility paths keep private dynamic discovery alive and weaken import-boundary enforcement.
- Validation criteria: Targeted backend and browser planner tests; `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'`; required generation/drift checks; `make check-harness-smoke`; `make harness-contract`; `make lint-scripts`; broaden to browser E2E only if planner behavior changes.
- Rollback or containment: Restore old entrypoints and allow-list entries as one slice.

### SI-05: Remove Private Retired-Diagnostic Branches and Tests

- Disposition/status: Remove now after fresh spec search; not started.
- Belongs in: implementation, tests, tracker.
- Remediation: Remove private tests and bespoke branches that only preserve retired diagnostics, including `CARTULARY_ALLOW_EMPTY_MANIFEST_SELECTION` rejection and direct `dev-services.sh db-down` compatibility checks, unless a fresh owner-doc search finds a continuing requirement.
- Rationale: Removed private commands/env vars should not remain maintained as if they were contracts.
- Expected long-term benefit: Less compatibility burden and fewer misleading affordances.
- Compatibility or migration impact: No public Make target compatibility impact expected.
- Risks of leaving unresolved: Maintainers keep preserving obsolete private behavior.
- Validation criteria: Affected shell tests; `make check-harness-smoke`; `make harness-contract`; `make lint-shell`; `make lint-markdown`.
- Rollback or containment: Restore the diagnostic branch/test only if an owner document requires it.

### SI-06: Audit `core/test-output` Legacy Readers and Fallbacks

- Disposition/status: Needs more evidence; audit-only first slice.
- Belongs in: implementation, tests, tracker, and possibly specification.
- Remediation: Audit legacy readers and fallback paths in `tools/harness/core/test-output/cli.mjs`. Remove only readers that are not NLSpec diagnostic support and not needed for retained-run inspection.
- Rationale: `core/test-output/cli.mjs` remains a high-risk mixed-responsibility module after adapter extraction.
- Expected long-term benefit: Lower output/accounting complexity without weakening retained evidence.
- Compatibility or migration impact: High. Output modes, retained artifacts, and historical diagnostic readers are public/spec-sensitive.
- Risks of leaving unresolved: Core output code remains difficult to reason about and future accounting changes stay brittle.
- Validation criteria: First slice is audit-only plus tracker update. Any later edit requires focused output-mode tests, `make harness-contract`, `make check-harness-smoke`, and retained-run fixture checks.
- Rollback or containment: Keep changes small and revert reader removal independently of adapter cleanup.

## 5. Explicit Keeps and Spec-Gated Removals

Keep with rationale:

- Scheduler common v9 schemas that are still referenced by current v10 schemas.
- Closed-shape negative tests for current schemas and owner inputs.
- Retired public env passthrough deny-list entries in `make-node-tools.mjs`, because they prevent accidental child-process inheritance.
- Playwright apt fallback, because it still supports Linux/WSL readiness setup.
- `LINT_SHELL_STRICT`, because it remains part of the current public shell-lint behavior.
- Retained-run `ALLOW_OLDER_RESULTS_DIR`, because it is NLSpec-owned finalizer behavior.

Spec-gated removal only:

- Historical diagnostic schemas/readers such as frontend row accounting v1/v2, visual fixture registry v1, accessibility summary v1, and `web_e2e_stack.v2`.
- Any public target, command ID, output class, schema ID, artifact role, result-root behavior, scheduler event/resource shape, failure taxonomy, readiness/cache behavior, or retained-run finalizer contract.

## 6. Sequencing and Risks

Required order:

1. SI-00 rebaseline and tracker gate: complete.
2. SI-01 standalone real-target output test cleanup: complete.
3. SI-02 unused private shim deletion.
4. SI-03 generated-artifacts JSON-shape shim retirement.
5. SI-04 executable shard-planner compatibility retirement.
6. SI-05 retired private diagnostics cleanup.
7. SI-06 `core/test-output` legacy-reader audit.

Dependencies and sequencing notes:

- SI-02 is the narrowest next slice and should run before larger planner or generated-output work.
- SI-03 must not hand-edit generated outputs; use owner inputs and Make generation.
- SI-04 should follow SI-02 so import-boundary and direct-reference scans are cleaner.
- SI-05 can move later if SI-04 exposes a dependency, but it must still start with a fresh owner-doc search.
- SI-06 is audit-first because retained artifact/output compatibility risk is high.

Primary risks:

- Accidental public contract drift from simplifying private code.
- Hidden generated/topology references to compatibility shims.
- Removing diagnostic historical readers that the NLSpec still allows.
- Expanding `tools/harness` into product behavior instead of harness mechanics.

## 7. Validation Plan

Tracker creation and SI-01 validation for this pass:

- `git ls-files tools/harness`: passed.
- `git ls-files --others --exclude-standard tools/harness`: passed.
- `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'`: passed.
- `CARTULARY_OUTPUT_MODE=summary make lint-shell`: passed with empty stdout and stderr.
- `bash tools/harness/core/tests/test-tool-output-real-targets.sh`: passed after SI-01 cleanup.
- `make check-harness-smoke`: passed.
- `make harness-contract`: passed.
- `make lint-markdown`: passed after the tracker edit.

Future slice validation:

- SI-02: reference scans, `make check-harness-smoke`, `make harness-contract`, `make lint-scripts`, `make lint-markdown`.
- SI-03: Make-owned generation/drift path, task-surface report, generated-artifact policy, JSON shape check, smoke, contract, Markdown.
- SI-04: targeted backend/browser planner tests, task-surface report, generation/drift if owner inputs change, smoke, contract, scripts lint, browser E2E only if behavior changes.
- SI-05: affected shell tests, smoke, contract, shell lint, Markdown.
- SI-06: audit-only first; later edits require output-mode, retained-run, smoke, and contract checks.

Final handoff criteria after the chosen implementation sequence:

- Run `make agent-finalize`.
- Run `make check`.
- If no successful full warm-check `RESULTS_DIR` is supplied, report retained-run maintenance as skipped because `RESULTS_DIR` was unset.

## 8. Handoff Log

| Date | Slice | Work | Commands | Result | Next action |
| --- | --- | --- | --- | --- | --- |
| 2026-07-03 | SI-00 | Rebaselined tracker after historical NI/WS completion. | `git ls-files tools/harness`; `git ls-files --others --exclude-standard tools/harness`; `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'`; `make check-harness-smoke`; `make harness-contract`. | Passed. Live baseline is 249 tracked harness files, 0 untracked harness files, and task surface 94 public / 31 internal helper / 11 check-internal. | SI-01. |
| 2026-07-03 | SI-01 | Removed stale standalone output-policy expectations for `lint-shell`, `lint`, invalid `explain-target`, and synthetic run ID length. | `CARTULARY_OUTPUT_MODE=summary make lint-shell`; `bash -n tools/harness/core/tests/test-tool-output-real-targets.sh`; `bash tools/harness/core/tests/test-tool-output-real-targets.sh`; `make check-harness-smoke`; `make harness-contract`; `make lint-markdown`. | Passed. The known `lint-shell` follow-up is closed without changing public `lint-shell` output. | SI-02. |

Future handoff rows must record command exit status, retained run root when emitted, relevant failure artifact when failed, skipped checks with reason, and exact next action.

## 9. Exact Next Action

Start SI-02. Before deleting anything, run `rg` for each candidate shim path and exported symbol:

- `tools/harness/backend/runner/go-shards.mjs`
- `tools/harness/backend/duration/baselines.mjs`
- `tools/harness/backend/drift/manifests.mjs`
- `tools/harness/browser/runtime/paths.mjs`

If the fresh scan confirms no live references, delete only those unused private shims, then run SI-02 validation and update this tracker before starting SI-03.
