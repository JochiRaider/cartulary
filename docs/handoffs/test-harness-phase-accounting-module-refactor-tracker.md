# test-harness-phase-accounting Next Refactor Tracker

## 1. Status and Authority

- Target path: `tools/harness/phase-accounting`
- Target label: `test-harness-phase-accounting`
- Tracker path: `docs/handoffs/test-harness-phase-accounting-module-refactor-tracker.md`
- Status: 2026-07-05 remediation implementation completed through WS-08 final validation and handoff.
- Current repository posture at final handoff: harness remediation changes are implemented and validated; `docs/reporting-subsystem-nlspec.md` contains an unrelated pre-existing/concurrent worktree modification that this effort did not edit or revert.
- Owner posture: this tracker identifies future harness structural refactors. It does not authorize product behavior changes or public harness contract changes except where a workstream is explicitly marked spec-first.

Authority and source hierarchy:

1. `docs/testing-harness-nlspec.md` owns harness mechanics: command invocation, target selection, scheduling, fixture lifecycle, service ownership, artifact emission, cleanup, summary emission, and harness verification gates.
2. `docs/domain.md` owns vocabulary and boundary interpretation only. Harness terms remain implementation-support terms unless an owner spec promotes them.
3. Core 00 through Core 04 own product behavior. Core 05 applies only to claim-bearing publication or benchmark boundaries.
4. Generated task/schedule artifacts and generated Make includes are downstream of owner inputs and must not become behavior owners.
5. The prior tracker is the controlling historical artifact for completed remediation, not a reason to preserve unsupported private behavior.

## 2. Inspected Sources

Current tracker inspection covered:

- Prior tracker: `docs/handoffs/test-harness-phase-accounting-module-refactor-tracker.md`
- Harness and vocabulary owners: `docs/testing-harness-nlspec.md`, `docs/domain.md`
- Generated and command metadata: `tools/generated_artifact_policy.json`, `tools/task_surface_manifest.json`, `tools/task_surface.generated.mk`, `tools/execution_topology_manifest.json`, `tools/execution_topology_render_index.json`, `tools/scheduler_manifest.json`, `tools/browser_e2e_batch_manifest.json`
- Phase-accounting implementation and tests: all current files under `tools/harness/phase-accounting`
- Relevant adjacent harness areas: `tools/harness/scheduler`, `tools/harness/backend`, `tools/harness/diagnostics`, `tools/harness/output/test-output`, and `tools/harness/tests`
- Relevant schema evidence: `tools/schemas/cartulary.frontend_phase_registry.v3.schema.json`, `tools/schemas/cartulary.frontend_visual_fixture_registry.v3.schema.json`, retained historical v2 schema attachments, and `tools/schemas/cartulary.phase_slice_plan.v1.schema.json`

## 3. Historical Baseline

Completed remediation from the prior tracker remains accepted unless validation proves regression:

- Runtime execution for `phase-slice` and `service-backed-slice` was moved out of the phase-accounting CLI and into scheduler-owned `tools/harness/scheduler/phase-slice-execution.mjs`.
- Frontend row-accounting now consumes normalized test-output observations from `tools/harness/output/test-output/frontend-row-evidence.mjs`; the frontend phase-accounting facade no longer re-exports test-output indexes.
- Target-plan smoke coverage was split into owner-aligned diagnostics, backend, and phase-accounting wrappers behind the unchanged `harness-smoke-print-target-plan` aggregate.
- Retained-root `frontend-evidence-audit` validation was completed against broad/support/visual/a11y retained roots in the prior remediation.
- Generated metadata changes from that remediation were refreshed through Make-owned generation and verified by drift checks.

Do not re-open those items in a future session unless a current validation target fails or a current owner spec changes.

## 4. Current Findings

- Public Make targets, `command_id` values, schema IDs, retained artifact paths, output shapes, failure mapping, cleanup behavior, and public input contracts are stable compatibility surfaces.
- Private helper paths may move only after callers move to declared owner facades and any task-surface/topology metadata is refreshed through Make-owned generation.
- No live unsupported legacy backend/frontend/scheduler helper files from the current import-boundary registry were found in the repository.
- The implemented remediation reduced private structural coupling while preserving public Make targets, schema IDs, retained artifact paths, output shapes, and command behavior.
- `frontend-phase-manifest.mjs` is now a compatibility facade and CLI path; owner-local frontend modules hold registry, validation, visual fixture, ledger, scenario-grep, audit-routing, and core compatibility responsibilities.
- `phase-slice-plan.mjs` now delegates base row selection/accounting and resource-limit resolution to owner-local planner helpers while preserving `cartulary.phase_slice_plan.v1`.
- `phase-manifest.mjs` is now a compatibility facade with stable ESM re-exports and direct-execution delegation to `phase-manifest-cli.mjs`; CLI command dispatch no longer lives in the facade.
- Frontend target refs now validate target existence and command-ID parity against task-surface metadata, and implemented required closure targets must have a declared audit retained-root route.
- Growth is now spec-owned and append-only: `cartulary.frontend_phase_registry.v3` accepts a contiguous declared phase range beginning at `FE-P0`, and `cartulary.frontend_visual_fixture_registry.v3` accepts a contiguous fixture range beginning at `FE-VFIX-01`. The current owner data still declares `FE-P0` through `FE-P11` and `FE-VFIX-01` through `FE-VFIX-21`; accepting `FE-P12+` or `FE-VFIX-22+` requires adding owner data, maps/ledgers, digests, and validation evidence, not a schema-count change.

## 5. Compatibility Freeze

Future implementation sessions must preserve these unless a workstream explicitly performs a spec-first change:

- Make target names: especially `phase-slice`, `service-backed-slice`, `frontend-evidence-audit`, `phase-ledgers`, `phase-ledger-drift`, `phase-schedules`, `phase-schedule-drift`, `phase-map-check`, `phase-test-name-check`, and `harness-contract`
- Stable `command_id` values, including `cartulary.harness.command.phase_slice.v1`, `cartulary.harness.command.service_backed_slice.v1`, and `cartulary.harness.command.frontend_evidence_audit.v1`
- Schema IDs: `cartulary.phase_slice_plan.v1`, `cartulary.frontend_phase_registry.v3`, `cartulary.frontend_phase_test_map.v3`, `cartulary.frontend_row_accounting.v3`, `cartulary.frontend_visual_fixture_registry.v3`, and `cartulary.frontend_evidence_audit_summary.v1`
- Retained artifact paths, including `frontend-row-accounting.json`, `frontend-evidence-audit-summary.json`, scheduler events/summaries, tool-run summaries, target summaries, generated ledgers, and phase-slice plan output
- Public input contracts for `PHASE`, `PHASE_NAMESPACE`, `ROWS`, `JSON`, `CHECK_RESULTS_DIR`, `BROWSER_SUPPORT_RESULTS_DIR`, `BROWSER_VISUAL_RESULTS_DIR`, and `BROWSER_A11Y_RESULTS_DIR`
- Failure class/reason mapping, cleanup behavior, service ownership, runtime reset behavior, and scheduler resource semantics

## 6. Out of Scope

- Product HTTP routes, WebSocket behavior, workbook shell behavior, workbook projection behavior, saved-view behavior, storage behavior, revision/history semantics, authorization, authentication, release publication behavior, and domain model changes
- SQL migrations, DB queries, generated product contracts, dependency locks, `go.sum`, `pnpm-lock.yaml`, and tool-managed installs
- Hand-editing generated roots or generated outputs, including `internal/gen/**`, `packages/protocol-ts/src/generated/**`, `packages/ui-contracts/src/generated/**`, `tools/task_surface.generated.mk`, `tools/scheduler_manifest.json`, `tools/browser_e2e_batch_manifest.json`, and `tools/execution_topology_render_index.json`
- Preserving private helper paths merely because they existed historically
- Treating visual, accessibility, design-direction, or implementation-support evidence as product conformance without a Core 05 or product-owner boundary

## 7. Gap Records

### G-01 Frontend Phase and Fixture Growth Readiness

- Status: completed in WS-02 on 2026-07-05.
- Remediation: adopted a spec-first v3 frontend registry and visual fixture growth contract before accepting `FE-P12+` or `FE-VFIX-22+`. The implementation now derives phase and fixture ranges from owner data using contiguous append-only ID validation rather than hard-coded counts and error strings.
- Areas: specification, schemas, implementation, tests, generated metadata.
- Rationale: prior v2 schemas and implementation intentionally capped the frontend registry at 12 phases and the visual fixture registry at 21 fixtures.
- Expected long-term benefit: adding future frontend phases or fixtures becomes owner-data work instead of scattered validator edits.
- Compatibility or migration impact: current owner data migrated to `cartulary.frontend_phase_registry.v3` and `cartulary.frontend_visual_fixture_registry.v3`; v2 schema attachments remain retained historical diagnostics. Guide, map, registry, visual fixture, and freshness digests were refreshed.
- Risk of leaving unresolved: resolved for schema-count growth; future phase/fixture additions still require owner data, maps/ledgers, and validation evidence.
- Validation criteria: passed `make json-shape-check`, `make phase-map-check`, `make phase-ledger-drift`, `make phase-schedule-drift`, `make generate-drift`, and `make generated-artifact-policy-check`.

### G-02 Frontend Phase-Manifest Module Split

- Status: completed in WS-03 on 2026-07-05.
- Remediation: keep `tools/harness/phase-accounting/frontend-phase-manifest.mjs` as the compatibility facade and CLI, but split internals into owner-local modules for registry loading, phase-map validation, owner-ref/root-context validation, freshness digesting, visual fixture validation, ledger rendering, and scenario grep helpers.
- Areas: implementation, tests, generated metadata only if referenced backing paths move.
- Rationale: one large file currently owns several distinct concerns with different validation and caller patterns.
- Expected long-term benefit: smaller modules make frontend phase growth, visual fixture changes, and ledger/freshness work easier to reason about and test.
- Compatibility or migration impact: exported facade functions and CLI commands must remain stable. Public schema IDs and retained artifacts must not change.
- Risk of leaving unresolved: resolved for the public facade split; future hardening can move more private helper bodies out of `frontend/phase-manifest-core.mjs` without changing callers.
- Validation criteria: passed `node --check` for changed modules, `make phase-map-check`, `make json-shape-check`, `make harness-contract`, `make lint-scripts`, `make phase-schedules`, `make phase-schedule-drift`, `make generate-drift`, and `make generated-artifact-policy-check`.

### G-03 Phase-Slice Planner Decomposition

- Status: completed in WS-04 on 2026-07-05.
- Remediation: keep `phase-slice-plan.mjs`, `frontend-phase-slice-plan.mjs`, and `phase-slice-cli.mjs` public/helper behavior stable, but split row selection, backend shard adaptation, browser stage adaptation, resource-limit resolution, and work-unit serialization behind small owner-local helpers.
- Areas: implementation, tests, generated metadata if backing script paths change.
- Rationale: the base planner still imports and coordinates phase rows, backend target/shard planning, browser stage metadata, diagnostics guidance, scheduler resources, and plan contract serialization.
- Expected long-term benefit: adding future phase namespaces, runners, or resource policies does not require editing one cross-owner planner.
- Compatibility or migration impact: preserve `cartulary.phase_slice_plan.v1`, JSON plan output, scheduler summary behavior, target summaries, and public target behavior for `phase-slice` and `service-backed-slice`.
- Risk of leaving unresolved: resolved for row-selection/accounting and resource-limit helper extraction; browser/backend unit helpers remain behind the stable planner facade and can be narrowed further without public contract changes.
- Validation criteria: passed `node --check` for changed planner modules, `make phase-slice PHASE=phase4 JSON=1`, `make service-backed-slice PHASE=phase4 JSON=1`, frontend namespace slice JSON checks, `make run-harness-smoke-extended`, `make harness-contract`, `make lint-scripts`, `make json-shape-check`, and `make phase-schedule-drift`.

### G-04 Frontend Target-Ref Parity and Audit Routing

- Status: completed in WS-05 on 2026-07-05.
- Remediation: added early validation that frontend phase-map `targets[].target_name` exists in task-surface metadata and `targets[].command_id` equals that target's actual `command_id`. Centralized retained-root routing used by `frontend-evidence-audit` and fail phase-map validation when an implemented required target cannot be audited by the current public input contract. Adopted optional public audit inputs for `browser-e2e-a11y-preflight` and `browser-e2e-measurement` so future required closure rows for those targets have an owner-declared route.
- Areas: implementation, tests, documentation, task-surface/topology owner metadata, generated metadata.
- Rationale: current target refs validate command-ID pattern but should catch target/command drift before retained evidence is generated. Audit root routing is currently hard-coded in the audit CLI.
- Expected long-term benefit: target drift fails during cheap metadata validation instead of late retained-root audit.
- Compatibility or migration impact: existing required audit inputs remain compatible. Two new optional retained-root inputs are documented and declared in task-surface metadata; they are required only when a selected frontend phase has implemented required closure rows for the matching targets.
- Risk of leaving unresolved: resolved for task-surface target/command parity and audit-route coverage; fresh retained-root audit remains a standing validation requirement when claiming real browser-run closure.
- Validation criteria: passed `make phase-map-check`, `make json-shape-check`, `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all`, `make phase-schedules`, `make phase-schedule-drift`, `make harness-contract`, and `make run-harness-smoke-extended`.

### G-05 Phase-Manifest CLI and Facade Narrowing

- Status: completed in WS-06 on 2026-07-05.
- Remediation: kept `tools/harness/phase-accounting/phase-manifest.mjs` as the compatibility path, moved CLI command dispatch into `tools/harness/phase-accounting/phase-manifest-cli.mjs`, and narrowed the facade to stable ESM re-exports plus a direct-execution shim.
- Areas: implementation, tests, generated metadata.
- Rationale: the facade is intentionally much smaller than historical versions, but it still mixes compatibility exports with many CLI command branches.
- Expected long-term benefit: less compatibility burden and clearer ownership for future selection or verification commands.
- Compatibility or migration impact: no public target, command output, schema, retained artifact, or caller-path change unless metadata is updated through owner inputs and Make generation.
- Risk of leaving unresolved: resolved for CLI dispatch; future command additions should land in the CLI module instead of growing the facade.
- Validation criteria: passed `node --check` for changed modules, direct `node tools/harness/phase-accounting/phase-manifest.mjs list-phases`, `make phase-test-name-check`, `make phase-map-check`, `make harness-contract`, `make lint-scripts`, and `make run-harness-smoke-extended`.

## 8. Workstream Sequencing

| Workstream | Depends on | Class | Main changes | Risks | Exit criteria | Required Make validation |
| --- | --- | --- | --- | --- | --- | --- |
| WS-00 Tracker refresh | none | docs | Replace prior tracker with this next-iteration handoff and retain completed remediation as history. | Documentation drift only. | This file states current findings, gap records, validation, generated-artifact handling, and binary criteria. | `make lint-markdown` |
| WS-01 Characterization baseline | WS-00 | validation | Run narrow current-state checks before implementation slices. | Existing unrelated harness failure may obscure refactor risk. | Baseline commands and artifacts recorded before code movement. | `make phase-map-check`, `make phase-test-name-check`, `make harness-contract`, `make run-harness-smoke-extended` |
| WS-02 Growth contract decision | WS-01 | spec-first | Decide frontend phase/fixture growth contract and schema/version posture before `FE-P12+` or `FE-VFIX-22+`. | Public schema compatibility and freshness digests. | Owner spec and schema plan accepted, or current v2 cap explicitly retained. | `make json-shape-check`, `make phase-map-check`, drift checks when inputs change |
| WS-03 Frontend manifest split | WS-01 | implementation | Split frontend registry/map/freshness/visual/ledger/grep helpers behind stable facade. | Accidental CLI or export drift. | `frontend-phase-manifest.mjs` remains the facade and callers continue to pass. | `make phase-map-check`, `make json-shape-check`, `make harness-contract`, `make lint-scripts` |
| WS-04 Phase-slice planner split | WS-01 | implementation | Extract base/frontend planner sub-helpers while preserving `cartulary.phase_slice_plan.v1`. | Plan ordering, resource claims, service-wrapper behavior, and scheduler summaries. | JSON plan, smoke, and targeted phase-slice behavior are unchanged. | `make phase-slice PHASE=phase4 JSON=1`, `make service-backed-slice PHASE=phase4 JSON=1`, `make run-harness-smoke-extended`, `make harness-contract` |
| WS-05 Target-ref and audit routing parity | WS-01 | implementation | Validate frontend target refs against task-surface command IDs and centralize audit root routing. | Public input contract accidentally changes without spec. | Drift is caught during phase-map validation; current audit inputs still work. | `make phase-map-check`, `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all`, `make json-shape-check` |
| WS-06 Phase-manifest CLI narrowing | WS-01 | implementation | Move CLI dispatch behind owner-local modules while retaining compatibility path. | Many adjacent browser/backend/generated callers use the facade. | Existing CLI commands and imports behave identically. | `make phase-test-name-check`, `make phase-map-check`, `make run-harness-smoke-extended`, `make harness-contract` |
| WS-07 Generated metadata refresh | any path or metadata move | generated | Update owner inputs first, then refresh generated outputs through Make. | Hand-editing generated artifacts or stale topology. | Drift checks pass and old paths do not remain in generated metadata. | `make phase-schedules`, `make phase-schedule-drift`, `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check` |
| WS-08 Final handoff | all implemented slices | handoff | Record files changed, verification, skipped checks, and retained artifacts. | Missing failure context for next agent. | Handoff log is current and binary criteria are checked. | `make agent-finalize`; broader `make check` only when implementation breadth warrants it |

## 9. Generated-Artifact Handling

- Do not hand-edit generated files or generated roots.
- Update owner inputs before generated outputs. Path or backing-script changes usually start in `tools/task_surface_manifest.json` and/or `tools/execution_topology_manifest.json`.
- Run `make phase-schedules` after task-surface, topology, or schedule owner inputs change.
- After generation, verify with `make phase-schedule-drift`, `make generate-drift`, `make generated-artifact-policy-check`, and `make json-shape-check`.
- Treat `tools/task_surface.generated.mk`, `tools/scheduler_manifest.json`, `tools/browser_e2e_batch_manifest.json`, and `tools/execution_topology_render_index.json` as downstream outputs.
- If generated metadata still references an old helper path after a move, the implementation slice is not complete.

## 10. Validation Matrix

| Change type | Required validation |
| --- | --- |
| Tracker-only docs update | `make lint-markdown` |
| Phase-map validation or frontend metadata validation | `make phase-map-check`, `make json-shape-check`, `make harness-contract` |
| Phase-slice planner behavior | `make phase-slice PHASE=phase4 JSON=1`, `make service-backed-slice PHASE=phase4 JSON=1`, frontend namespace slice JSON check, `make run-harness-smoke-extended`, `make harness-contract` |
| Frontend retained-evidence audit behavior | Direct `make frontend-evidence-audit ...` with broad/support/visual/a11y retained roots, plus `make run-harness-smoke-extended` |
| CLI/facade movement | `node --check` for changed modules, `make phase-test-name-check`, `make phase-map-check`, `make harness-contract`, `make lint-scripts` |
| Owner input or backing-script path movement | `make phase-schedules`, `make phase-schedule-drift`, `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check`, `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all` |
| Broad implementation pass | `make agent-finalize`; run `make check` only when breadth or risk warrants it |

## 11. Handoff Update Requirements

Every implementation session using this tracker must update this file or a successor handoff with:

- Workstream status and exact started/completed timestamp.
- Files changed, grouped by owner area.
- Public contracts intentionally preserved or spec-first changes made.
- Generated-artifact handling: owner inputs changed, Make generation run, and drift results.
- Verification commands, pass/fail status, run roots when emitted, and failure classification if any command fails.
- Skipped checks with concrete reason.
- Any user-visible blocker, retained-root requirement, or owner-spec decision still needed.

## 12. Open Risks and Blockers

| ID | Risk or blocker | Why it matters | Resolution path | Status |
| --- | --- | --- | --- | --- |
| OR-001 | Frontend phase/fixture growth was capped by v2 schemas and validators. | Future `FE-P12+` or `FE-VFIX-22+` needed a spec/schema-first path. | Completed by WS-02 with v3 append-only contiguous registry and fixture contracts. | resolved |
| OR-002 | Frontend target refs were not fully checked against task-surface command IDs at phase-map validation time. | Target drift could fail only during retained-root audit. | Completed by WS-05 with task-surface target/command parity validation and centralized audit-route checks. | resolved |
| OR-003 | Compatibility facades remain high-value public/private caller boundaries. | Path movement can still break backend, browser, generated-artifact, diagnostics, and test-output callers even after the internals were narrowed. | Completed current narrowing in WS-03 and WS-06; future movement should keep facades stable and refresh metadata only through Make. | standing watch |
| OR-004 | Direct retained-root audit requires fresh or retained broad/support/visual/a11y roots. | Audit behavior cannot be proven from phase names, generated ledgers, or historical evidence. | Produce or supply exact Make-owned roots before claiming audit behavior complete. | standing validation requirement |

## 13. Binary Completion Criteria

This tracker iteration is complete only when all of the following are true:

- The prior remediation is summarized as completed history and not repeated as open work.
- Current public compatibility surfaces are explicitly frozen.
- Out-of-scope product and generated-file boundaries are explicit.
- Each gap record includes remediation, areas, rationale, long-term benefit, compatibility or migration impact, unresolved risk, and validation criteria.
- Workstreams include dependencies, sequencing, risks, exit criteria, required Make targets, and generated-artifact handling.
- Open risks/blockers are listed with resolution paths.
- Handoff/update requirements are explicit enough for another agent to resume without rediscovering context.
- Tracker-only validation is recorded; implementation-specific validation is scoped for future slices.

## 14. Session Handoff Log

| Time | Session | Files changed | Commands run | Result | Next action |
| --- | --- | --- | --- | --- | --- |
| 2026-07-05 | Codex GPT-5 tracker refresh | `docs/handoffs/test-harness-phase-accounting-module-refactor-tracker.md` | Fresh inspection of prior tracker, harness/domain docs, generated policy/topology/task-surface metadata, phase-accounting implementation/tests, adjacent scheduler/backend/diagnostics/test-output files, schema caps, and `make lint-markdown`. | Next-iteration tracker written in place; markdown validation passed. | Leave implementation work for a later authorized slice. |
| 2026-07-05T00:09:10-04:00..2026-07-05T00:11:15-04:00 | WS-00/WS-01 implementation baseline | `docs/handoffs/test-harness-phase-accounting-module-refactor-tracker.md` | `git status --short`; `make phase-map-check`; `make phase-test-name-check`; `make harness-contract`; `make run-harness-smoke-extended`. | Starting worktree was clean. Baseline checks all passed before refactor edits. | Start WS-02 growth contract implementation. |
| 2026-07-05T00:11:15-04:00..2026-07-05T00:18:02-04:00 | WS-02 growth contract | `docs/testing-harness-nlspec.md`; `docs/guides/cartulary_frontend_implementation_testing_guide.md`; `tools/schemas/cartulary.frontend_phase_registry.v3.schema.json`; `tools/schemas/cartulary.frontend_visual_fixture_registry.v3.schema.json`; `tools/frontend_phase_registry.json`; `tools/frontend_phase_maps/*.json`; `tools/frontend_visual_fixture_registry.json`; `tools/harness/phase-accounting/frontend-phase-manifest.mjs`; `tools/harness/generated-artifacts/check-json-shapes.mjs`; tracker | `node --check tools/harness/phase-accounting/frontend-phase-manifest.mjs`; `node --check tools/harness/generated-artifacts/check-json-shapes.mjs`; `make json-shape-check`; `make phase-map-check`; `make phase-ledger-drift`; `make phase-schedule-drift`; `make generated-artifact-policy-check`; `make generate-drift`. | Adopted v3 append-only contiguous frontend phase and visual fixture contracts; refreshed guide/map/registry/freshness digests. All WS-02 validation passed. | Start WS-03 frontend manifest split. |
| 2026-07-05T00:18:02-04:00..2026-07-05T00:22:07-04:00 | WS-03 frontend manifest split | `tools/harness/phase-accounting/frontend-phase-manifest.mjs`; new owner-local modules under `tools/harness/phase-accounting/frontend/`; frontend registry/validation/ledger caller imports; `tools/execution_topology_render_index.json`; tracker | `node --check` for changed frontend/generated modules; `make phase-map-check`; `make json-shape-check`; `make harness-contract`; `make lint-scripts`; `make phase-schedules`; `make phase-schedule-drift`; `make generate-drift`; `make generated-artifact-policy-check`. | Compatibility path remains executable; internal callers moved to narrower registry/validation/ledger/visual/grep modules. `json-shape-check` initially reported stale schedule inputs after an import move; `make phase-schedules` refreshed downstream generated output, then drift checks passed. | Start WS-04 phase-slice planner split. |
| 2026-07-05T00:22:07-04:00..2026-07-05T00:26:54-04:00 | WS-04 phase-slice planner split | `tools/harness/phase-accounting/phase-slice-plan.mjs`; `tools/harness/phase-accounting/phase-slice-planning/row-selection.mjs`; `tools/harness/phase-accounting/phase-slice-planning/resource-limits.mjs`; tracker | `node --check` for changed planner modules; `make phase-slice PHASE=phase4 JSON=1`; `make service-backed-slice PHASE=phase4 JSON=1`; `make phase-slice PHASE_NAMESPACE=frontend PHASE=FE-P3 JSON=1`; `make service-backed-slice PHASE_NAMESPACE=frontend PHASE=FE-P3 JSON=1`; `make harness-contract`; `make lint-scripts`; `make run-harness-smoke-extended`; `make json-shape-check`; `make phase-schedule-drift`. | Extracted base row selection/accounting and phase-slice resource-limit resolution behind owner-local helpers while preserving `cartulary.phase_slice_plan.v1` behavior. All WS-04 validation passed. | Start WS-05 target-ref parity and audit routing. |
| 2026-07-05T00:26:54-04:00..2026-07-05T00:38:25-04:00 | WS-05 target-ref parity and audit routing | `docs/testing-harness-nlspec.md`; `tools/execution_topology_manifest.json`; generated task-surface/topology outputs; `tools/harness/phase-accounting/frontend/phase-manifest-core.mjs`; `tools/harness/phase-accounting/frontend/audit-routing.mjs`; `tools/harness/phase-accounting/frontend-evidence-audit-cli.mjs`; tracker | `make phase-schedules`; `make phase-map-check`; `make json-shape-check` (`.cartulary/test-results/20260705T043717Z-p54621`); `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all`; `make phase-schedule-drift` (`.cartulary/test-results/20260705T043731Z-p55508`); `make harness-contract`; `make run-harness-smoke-extended`. | Frontend phase-map validation now checks task-surface target existence and command-id parity, and implemented required closure targets must have an audit input route. Audit routing is shared between validation and `frontend-evidence-audit`; optional preflight and measurement retained-root inputs are spec/documentation/metadata declared. Synthetic retained-root audit fixture coverage passed through extended smoke. | Start WS-06 phase-manifest CLI narrowing. |
| 2026-07-05T00:38:25-04:00..2026-07-05T00:45:13-04:00 | WS-06 phase-manifest CLI narrowing | `tools/harness/phase-accounting/phase-manifest.mjs`; `tools/harness/phase-accounting/phase-manifest-cli.mjs`; `tools/execution_topology_manifest.json`; generated task-surface/topology outputs; tracker | `node --check tools/harness/phase-accounting/phase-manifest.mjs`; `node --check tools/harness/phase-accounting/phase-manifest-cli.mjs`; `node tools/harness/phase-accounting/phase-manifest.mjs list-phases`; `make phase-schedules` (`.cartulary/test-results/20260705T044259Z-p18585`); `make phase-test-name-check`; `make phase-map-check`; `make harness-contract`; `make lint-scripts`; `make run-harness-smoke-extended`. | CLI command dispatch now lives in `phase-manifest-cli.mjs`; `phase-manifest.mjs` remains the compatibility facade and direct command path. Owner metadata declares the new backing module and generated outputs were refreshed. All WS-06 validation passed. | Start WS-07 generated metadata refresh. |
| 2026-07-05T00:45:13-04:00..2026-07-05T00:45:57-04:00 | WS-07 generated metadata refresh | `tools/execution_topology_manifest.json`; generated `tools/task_surface_manifest.json`, `tools/task_surface.generated.mk`, and `tools/execution_topology_render_index.json`; tracker | `make phase-schedules` (`.cartulary/test-results/20260705T044541Z-p80473`); `make phase-ledger-drift` (`.cartulary/test-results/20260705T044541Z-p80500`); `make generated-artifact-policy-check` (`.cartulary/test-results/20260705T044541Z-p80552`); `make phase-schedule-drift` (`.cartulary/test-results/20260705T044551Z-p81177`); `make generate-drift` (`.cartulary/test-results/20260705T044551Z-p81212`); `make json-shape-check` (`.cartulary/test-results/20260705T044551Z-p81211`); `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all`. | Owner metadata changes were regenerated through Make; task-surface report emitted `check=pass`; generated artifact, phase-ledger, schedule, JSON shape, and generation drift checks passed. | Start WS-08 final validation and handoff. |
| 2026-07-05T00:45:57-04:00..2026-07-05T00:47:15-04:00 | WS-08 final validation and handoff | Tracker final status/current findings; no additional implementation files beyond previous workstreams. | `make agent-finalize` (`.cartulary/test-results/20260705T044631Z-p82961`); `make lint-markdown`; `make phase-map-check`; `make phase-test-name-check`; `make harness-contract`; `make lint-scripts`; `make json-shape-check` (`.cartulary/test-results/20260705T044705Z-p86285`); `make phase-schedule-drift` (`.cartulary/test-results/20260705T044705Z-p86318`); `make generate-drift` (`.cartulary/test-results/20260705T044705Z-p86337`); `make generated-artifact-policy-check` (`.cartulary/test-results/20260705T044705Z-p86502`); `git status --short`; `git diff --stat`. | Final validation passed. `agent-finalize` reported retained-run reuse skipped because `RESULTS_DIR` was unset. Full `make check` was not run because the implemented surface was covered by targeted harness, smoke, contract, documentation, and drift validations. Direct real retained-root `frontend-evidence-audit` with fresh browser roots was not run; synthetic retained-root audit fixture coverage passed earlier through `make run-harness-smoke-extended`. Unrelated `docs/reporting-subsystem-nlspec.md` worktree changes were observed and left untouched. | Remediation complete. |
