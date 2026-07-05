# test-harness-phase-accounting Next Refactor Tracker

## 1. Status and Authority

- Target path: `tools/harness/phase-accounting`
- Target label: `test-harness-phase-accounting`
- Tracker path: `docs/handoffs/test-harness-phase-accounting-module-refactor-tracker.md`
- Status: next-iteration tracker written after the 2026-07-05 remediation completed.
- Current repository posture at inspection: clean worktree, no live unsupported legacy helper files from the current harness import-boundary registry, and no open blocker carried forward from the prior tracker.
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
- Relevant schema evidence: `tools/schemas/cartulary.frontend_phase_registry.v2.schema.json`, `tools/schemas/cartulary.frontend_visual_fixture_registry.v2.schema.json`, `tools/schemas/cartulary.phase_slice_plan.v1.schema.json`

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
- The next meaningful debt is structural growth and readability debt, not incomplete prior remediation.
- `frontend-phase-manifest.mjs` still combines registry loading, phase-map shape and semantic validation, owner-ref validation, guide restatement checks, freshness digesting, visual fixture validation, ledger rendering, and scenario grep helpers in one large module.
- `phase-slice-plan.mjs` still assembles base phase rows, backend shard adaptation, browser stage adaptation, diagnostics guidance, scheduler resource limits, and work-unit serialization in one broad facade.
- `phase-manifest.mjs` remains a broad compatibility facade and CLI dispatcher for selection, verification, fixture policy, policy exceptions, and phase listing.
- Frontend target refs validate command-ID shape, but the next iteration should add early parity against task-surface metadata so target and command drift fail during phase-map validation rather than retained-root audit.
- Growth is currently spec-owned and capped: `cartulary.frontend_phase_registry.v2` requires exactly 12 frontend phases, and `cartulary.frontend_visual_fixture_registry.v2` requires exactly 21 visual fixtures with fixture IDs through `FE-VFIX-21`. Accepting `FE-P12+` or `FE-VFIX-22+` requires a spec/schema-first change.

## 5. Compatibility Freeze

Future implementation sessions must preserve these unless a workstream explicitly performs a spec-first change:

- Make target names: especially `phase-slice`, `service-backed-slice`, `frontend-evidence-audit`, `phase-ledgers`, `phase-ledger-drift`, `phase-schedules`, `phase-schedule-drift`, `phase-map-check`, `phase-test-name-check`, and `harness-contract`
- Stable `command_id` values, including `cartulary.harness.command.phase_slice.v1`, `cartulary.harness.command.service_backed_slice.v1`, and `cartulary.harness.command.frontend_evidence_audit.v1`
- Schema IDs: `cartulary.phase_slice_plan.v1`, `cartulary.frontend_phase_registry.v2`, `cartulary.frontend_phase_test_map.v3`, `cartulary.frontend_row_accounting.v3`, `cartulary.frontend_visual_fixture_registry.v2`, and `cartulary.frontend_evidence_audit_summary.v1`
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

- Remediation: perform a spec-first workstream to decide the next frontend registry and visual fixture growth contract before accepting `FE-P12+` or `FE-VFIX-22+`. After the spec/schema change, derive phase and fixture ranges from owner data or schema helpers rather than hard-coded counts and error strings.
- Areas: specification, schemas, implementation, tests, generated metadata.
- Rationale: current v2 schemas and implementation intentionally cap the frontend registry at 12 phases and the visual fixture registry at 21 fixtures.
- Expected long-term benefit: adding future frontend phases or fixtures becomes owner-data work instead of scattered validator edits.
- Compatibility or migration impact: current v2 behavior remains frozen until the owner spec adopts a replacement contract. Any schema-ID/version change must update schema attachments, freshness digests, ledgers, and generated metadata together.
- Risk of leaving unresolved: future phase growth will be brittle, likely fail late, and encourage tactical bypasses of current validators.
- Validation criteria: `make json-shape-check`, `make phase-map-check`, `make phase-ledger-drift`, `make phase-schedule-drift`, `make generate-drift`, and `make generated-artifact-policy-check`.

### G-02 Frontend Phase-Manifest Module Split

- Remediation: keep `tools/harness/phase-accounting/frontend-phase-manifest.mjs` as the compatibility facade and CLI, but split internals into owner-local modules for registry loading, phase-map validation, owner-ref/root-context validation, freshness digesting, visual fixture validation, ledger rendering, and scenario grep helpers.
- Areas: implementation, tests, generated metadata only if referenced backing paths move.
- Rationale: one large file currently owns several distinct concerns with different validation and caller patterns.
- Expected long-term benefit: smaller modules make frontend phase growth, visual fixture changes, and ledger/freshness work easier to reason about and test.
- Compatibility or migration impact: exported facade functions and CLI commands must remain stable. Public schema IDs and retained artifacts must not change.
- Risk of leaving unresolved: validation, rendering, and selection rules will continue to accrete in a single high-blast-radius module.
- Validation criteria: `make phase-map-check`, `make json-shape-check`, `make harness-contract`, `make lint-scripts`, plus targeted `node --check` for changed modules.

### G-03 Phase-Slice Planner Decomposition

- Remediation: keep `phase-slice-plan.mjs`, `frontend-phase-slice-plan.mjs`, and `phase-slice-cli.mjs` public/helper behavior stable, but split row selection, backend shard adaptation, browser stage adaptation, resource-limit resolution, and work-unit serialization behind small owner-local helpers.
- Areas: implementation, tests, generated metadata if backing script paths change.
- Rationale: the base planner still imports and coordinates phase rows, backend target/shard planning, browser stage metadata, diagnostics guidance, scheduler resources, and plan contract serialization.
- Expected long-term benefit: adding future phase namespaces, runners, or resource policies does not require editing one cross-owner planner.
- Compatibility or migration impact: preserve `cartulary.phase_slice_plan.v1`, JSON plan output, scheduler summary behavior, target summaries, and public target behavior for `phase-slice` and `service-backed-slice`.
- Risk of leaving unresolved: future runner or namespace additions will increase accidental scheduler/backend/browser coupling.
- Validation criteria: `make phase-slice PHASE=phase4 JSON=1`, `make service-backed-slice PHASE=phase4 JSON=1`, frontend namespace slice JSON checks, `make run-harness-smoke-extended`, `make harness-contract`, and `make lint-scripts`.

### G-04 Frontend Target-Ref Parity and Audit Routing

- Remediation: add early validation that frontend phase-map `targets[].target_name` exists in task-surface metadata and `targets[].command_id` equals that target's actual `command_id`. Centralize the retained-root routing used by `frontend-evidence-audit` and fail phase-map validation when an implemented required target cannot be audited by the current public input contract.
- Areas: implementation, tests, documentation if public input behavior changes.
- Rationale: current target refs validate command-ID pattern but should catch target/command drift before retained evidence is generated. Audit root routing is currently hard-coded in the audit CLI.
- Expected long-term benefit: target drift fails during cheap metadata validation instead of late retained-root audit.
- Compatibility or migration impact: current audit inputs remain unchanged unless an owner spec first adopts a public input change.
- Risk of leaving unresolved: new or renamed required closure targets can produce expensive retained evidence before audit reports an avoidable routing or command-ID mismatch.
- Validation criteria: `make phase-map-check`, `make frontend-evidence-audit PHASE_NAMESPACE=frontend PHASE=<active FE phase> CHECK_RESULTS_DIR=<root> BROWSER_SUPPORT_RESULTS_DIR=<root> BROWSER_VISUAL_RESULTS_DIR=<root> BROWSER_A11Y_RESULTS_DIR=<root>`, `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all`, and `make json-shape-check`.

### G-05 Phase-Manifest CLI and Facade Narrowing

- Remediation: keep `tools/harness/phase-accounting/phase-manifest.mjs` as the compatibility path, but move CLI command dispatch into owner-local command modules and keep ESM exports focused on manifest loading, selection, verification, and fixture-policy facades.
- Areas: implementation, tests, generated metadata only if backing scripts change.
- Rationale: the facade is intentionally much smaller than historical versions, but it still mixes compatibility exports with many CLI command branches.
- Expected long-term benefit: less compatibility burden and clearer ownership for future selection or verification commands.
- Compatibility or migration impact: no public target, command output, schema, retained artifact, or caller-path change unless metadata is updated through owner inputs and Make generation.
- Risk of leaving unresolved: every new selection or verification command grows the broad facade and makes import-boundary reasoning harder.
- Validation criteria: `make phase-test-name-check`, `make phase-map-check`, browser and Vitest manifest smoke through `make run-harness-smoke-extended`, and `make harness-contract`.

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
| OR-001 | Frontend phase/fixture growth is currently capped by v2 schemas and validators. | Future `FE-P12+` or `FE-VFIX-22+` cannot be accepted safely by implementation-only edits. | Spec/schema-first workstream G-01. | open |
| OR-002 | Frontend target refs are not yet fully checked against task-surface command IDs at phase-map validation time. | Target drift may fail only during retained-root audit. | Implement G-04 before adding new closure targets. | open |
| OR-003 | Broad facades remain compatibility paths with many callers. | Path movement can break backend, browser, generated-artifact, diagnostics, and test-output callers. | Characterize with WS-01, move internals behind stable facades, and refresh metadata only through WS-07. | open |
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
