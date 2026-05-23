# Phase 9 Implementation Plan

## Summary

This file is the execution roadmap, progress marker, and handoff aid for Cartulary Phase 9: keyboard contract, clipboard and bulk-edit behaviors, Notes, Indicators, Parties, Assessments, Task Requests, Decisions, coordination surfaces, and remaining analyst-work workbook surfaces.

`docs/guides/cartulary_implementation_testing_guide.md` §11, Phase 9, is the controlling implementation-scope reference for this plan. Normative implementation-conformance behavior remains owned by Core 00 through Core 04, especially Core 01 §7.4 and §19, Core 02 §10 and §19, Core 03 §2, §11, §13, and §16-§20, and Core 04 §2.

This planning artifact does not implement Phase 9 behavior. It is intentionally root-level so coding agents can find the current roadmap during handoff or interrupted implementation sessions.

Authority model:
- Core 00 through Core 04 own implementation-conformance behavior.
- Core 05 owns claim-bearing timed or fixture-sensitive publication only.
- `PHASE9_IMPLEMENTATION_PLAN.md` is a planning artifact only.
- Generated files, generated ledgers, generated schedules, visual goldens, support-only tests, retained run artifacts, and previous phase plans are not behavior authorities.
- `docs/domain.md` is a vocabulary and concept reference for terminology-sensitive work. It does not replace the owner specs.

Current repo status after Sprint 8 audit:
- `tools/phase_registry.json` includes Phase 9 as `active`.
- `tools/phase9_test_map.json` exists and represents every authoritative Phase 9 row exactly once.
- `docs/testing/phase9_coverage_ledger.md` exists and is generated from the Phase 9 manifest.
- Direct Sprint 1 through Sprint 8 row evidence is present for keyboard/grid anchors, clipboard paste, bulk edit, stable paste-anchor translation, Timeline paste, Host/Identity entity-origin paste, Notes, Indicators, Parties, party text-plus-link flows, append-only Compromise Assessments, timestamp scalar validation, Task Requests, Decisions, required coordination surfaces, optional standardized surfaces, and registry closure.
- Sprint 9 final public-wrapper validation, retained-run maintenance, `make check`, and handoff passed on 2026-05-21. Final retained roots and decisions are recorded in the Sprint 9 section.
- No Phase 9 blocker sentinel or out-of-Phase-9 final-gate blocker remains recorded after the Sprint 9 final gate.
- The broad `browser-e2e-webserver-backed` aggregate passed in the Sprint 8 audit.

## Phase Objective

By Phase 9 exit, users must be able to stay in the workbook while using the full base keyboard contract, pasting or bulk-editing tabular data, creating and linking Notes, working with canonical Indicators, creating and linking Parties without losing raw text, recording append-only Compromise Assessments, managing Task Requests and Decisions as bounded coordination records, and using `comm_log`, `handoff`, `status_review`, and `lesson` as workbook-native coordination surfaces.

Phase 9 must also close the remaining required surface registry behavior for Notes and coordination artifacts. If the implementation exposes standardized Findings, Investigative Queries, or Forensic Keywords surfaces, those exposed surfaces must behave as workbook-native standardized optional surfaces and must not replace any required base surface.

User-observable exit state:
- Keyboard navigation and shortcuts operate on workbook cells and return focus predictably to Cartulary anchors.
- Clipboard paste and bulk edit produce visible workbook mutations without hidden macro semantics.
- Notes, Indicators, Parties, Assessments, Task Requests, Decisions, and coordination surfaces can be discovered, opened, created where allowed, queried, edited where writable, filtered or grouped where declared, and used without leaving the workbook interaction model.
- Text-plus-link party flows preserve raw source text independently from `party_id` links.
- Timestamp and direct-reference fields fail closed on invalid authoritative values while preserving invalid drafts as client-local state where Core 03 requires it.
- Optional standardized surfaces are either directly covered when exposed or explicitly marked `N/A` with owner-section justification.

## Implementation Scope

In scope:
- Phase 9 ownership manifest, coverage ledger, generated schedules, selectable row owners, and placeholder or blocker sentinel rows.
- Full keyboard contract for Arrow keys, Enter, Shift+Enter, Tab, Ctrl+V, Ctrl+K, Space, Alt+H, and Esc on workbook surfaces.
- Grid-adapter focus-anchor behavior for keyboard navigation, paste, save-state, presence, and conflict marker continuity.
- Clipboard paste into Timeline, including tabular ingest, known-column mapping, unknown-column preservation, source-origin distinctions, ordered mutations, conflict batching boundaries, sorted or filtered row anchoring, fill-down, and multi-row tag assignment.
- Notes as artifact-backed `artifact_type='note'` rows exposed through `cartulary.view.notes.v1`.
- Indicators as canonical `indicator` rows with distinct source-bound observations and lifecycle history.
- Parties system view and text-plus-link flows, including explicit create-from-text, link-existing, clear-text, clear-link, clear-both, exact-match reuse, and same-incident direct-reference validation.
- Compromise Assessments as append-only assessment rows with closed assessment-state vocabulary, confidence-band behavior, and timestamp validation.
- Task Requests and Decisions as first-class workbook-native surfaces with bounded lifecycle rules, owner semantics, queue fields, direct-reference decision links, support refs, and fail-closed contradiction handling.
- Coordination surfaces `comm_log`, `handoff`, `status_review`, and `lesson` as workbook-native artifact-backed surfaces with minimum create signals, required defaults, collection contracts, and projection-backed query fields.
- Manual relationship collection fields on assessments, tasks, decisions, and coordination surfaces preserving `confidence=null` and rejecting client-supplied `confidence`.
- Registry closure for Notes and required coordination surfaces under the artifact-backed variant registry and base view-schema registry.
- Browser-functional evidence for required Phase 9 workbook workflows.

Out of scope:
- Treating this plan, prior phase plans, generated ledgers, generated schedules, visual goldens, support-only tests, or retained run artifacts as behavior authority.
- File-based CSV or XLSX import beyond clipboard paste. That belongs to the Import Extension Profile.
- Claim-bearing timing, benchmark, or fixture-sensitive publication. That belongs to Core 05.
- Generalized workflow engines, generalized approval gates, mandatory Timeline approval fields, required per-edit coordination rituals, record-specific ACLs, or hidden sub-workspaces.
- A separate Notes storage silo or Notes-only persistence model.
- A deployment-global contact directory, CRM, party merge, or phone-number normalization model.
- Snapshot/reporting routes, immutable snapshot query semantics, release publication, or export redaction behavior.
- Pack-dependent workbook surfaces such as ATT&CK, D3FEND, or VERIS overlays unless a later owner spec defines them.

Optional-if-exposed:
- `cartulary.view.findings.v1`
- `cartulary.view.investigative_queries.v1`
- `cartulary.view.forensic_keywords.v1`

If any optional standardized surface is exposed in the base build, Phase 9 must cover the relevant `U-9-09` and `E-9-07` behavior and include `AC-285..AC-287` in the claimed surface set. If a surface is omitted, the implementation-owned coverage record must mark the omitted ACs `N/A`, record the omission, and cite the owner sections supporting that interpretation.

Carried-forward compatibility boundaries:
- Phase 8 handoff contracts are usable substrate only: stable view-query validation, canonical `meta.query`, stable view-schema discovery keys, startup and saved-view addressing, presentation-only group headers, full-row and sparse-patch wire contracts, exact-token search, and strict-prefix search.
- Phase 8 did not implement Phase 9 keyboard or clipboard workflows.
- Phase 8 did not complete Notes, Indicators, Parties, Assessments, Task Requests, Decisions, or coordination-surface workflow obligations except where those surfaces consume shared Phase 8 query, view-schema, row-patch, or saved-view behavior.
- Existing Phase 4 smoke evidence for Parties, coordination routes, Assessments, Indicators, and generic workbook surfaces is support substrate only. Phase 9 exit requires direct Phase 9 authoritative rows.

## Sprint Checklist

| Done | Sprint | Primary validation | Blockers | Follow-up notes |
| --- | --- | --- | --- | --- |
| [x] | 0. Ownership manifest and harness setup | `make phase-map-check`, `make phase-ledger-drift`, `make phase-schedule-drift`, `make phase-test-name-check` | Harness setup is complete; current blocker sentinels remain only for rows not yet replaced by later sprint evidence. | Manifest, registry entry, generated ledger, generated schedule updates, support-only carryover guards, and Sprint 0 blocker boundaries are in place. |
| [x] | 1. Keyboard and grid anchors | `make frontend-unit`, `make browser-e2e-webserver-backed`, `make phase-slice PHASE=phase9`, `git diff --check` | No Sprint 1-owned blocker remains; aggregate targets still fail on later Phase 9 blocker sentinels and one outside-Phase-9 browser regression. | Direct Sprint 1 evidence now covers keyboard command mapping, Cartulary grid anchors, live workbook shortcuts, and shared grid keyboard anchor semantics. |
| [x] | 2. Clipboard and bulk paste | `make backend-unit`, `make backend-integration`, `make frontend-unit`, `make phase-slice PHASE=phase9`, `make service-backed-slice PHASE=phase9`, `make phase-ledger-drift`, `git diff --check` | No Sprint 2-owned blocker remains; full Phase 9 remains incomplete while later blocker rows remain, and broad browser aggregate status is still affected by outside-Phase-9 `E-4-04`. | Covers shared tabular-ingest planning, Timeline paste, Host/Identity entity-origin paste, fill-down, tags, conflict grouping, sorted/filtered anchoring, file-import separation, and generated-ledger traceability. |
| [x] | 3. Notes and Indicators | `make backend-store`, `make backend-integration`, `make browser-e2e-webserver-backed`, `make phase-slice PHASE=phase9`, `make service-backed-slice PHASE=phase9`, `git diff --check` | No Sprint 3-owned blocker remains. This does not claim Phase 9 completion; later Phase 9 sprint rows remain separately owned. | Direct Sprint 3 evidence covers Notes as artifact-backed `note` rows, canonical Indicators with distinct observations and lifecycle history, workbook projection query, Notes linked creation, registry opening by canonical `view_schema_id`, and shared grid anchors for Notes/Indicators. |
| [x] | 4. Parties and text-plus-link flows | `make backend-unit`, `make backend-store`, `make backend-integration`, `make browser-e2e-webserver-backed`, `make phase-slice PHASE=phase9`, `make service-backed-slice PHASE=phase9`, `git diff --check` | No Sprint 4-owned blocker remains. Caveat: `U-9-11` is requested by older planning text as backend-unit evidence, but current direct evidence is mapped and executed as backend-store because same-incident active-target validation depends on store state. | Direct Sprint 4 evidence covers incident-scoped Parties, exact reuse by normalized `primary_email` or `external_ref`, stable-ID-only direct refs, independent text/ref clears, same-surface party actions, and row context preservation. |
| [x] | 5. Assessments and timestamp contract | `make backend-unit`, `make backend-store`, `make backend-integration`, `make browser-e2e-webserver-backed`, `make phase-slice PHASE=phase9`, `make service-backed-slice PHASE=phase9`, `git diff --check` | No Sprint 5-owned blocker remains. This does not claim Phase 9 completion; Task Request, Decision, coordination, optional standardized surface, and remaining non-assessment relationship rows remain later Phase 9 work. | Direct Sprint 5 evidence covers append-only Compromise Assessments, closed `assessment_state`, deterministic confidence bands, timestamp scalar validation and clearability, assessment support-ref confidence rejection, assessment projection reads, and browser-visible invalid timestamp draft behavior. |
| [x] | 6. Task Requests and Decisions | `make backend-unit`, `make backend-store`, `make backend-integration`, `make browser-e2e-webserver-backed`, focused `go test ./internal/modules/workbook -run 'Phase9Sprint6|Direct.*Decoder|Owner.*Null|Supersede.*Inconsistent'`, `make agent-finalize`, `git diff --check` | No Sprint 6-owned blocker remains. This does not claim Phase 9 completion; coordination, optional standardized surface, registry closure, and final gate rows remain later Phase 9 work. | Direct Sprint 6 evidence covers Task Request lifecycle and queue fields, Decision lifecycle and supersession, direct decision-reference set/clear, task/decision relationship confidence rejection, projection-backed workbook reads, browser-visible Task/Decision flows, explicit supersession fail-closed behavior for inconsistent source or target state, and wire-level owner null-clear rejection. |
| [x] | 7. Coordination surfaces | `make backend-unit`, `make backend-store`, `make backend-integration`, `make browser-e2e-webserver-backed`, `make phase-slice PHASE=phase9`, `make service-backed-slice PHASE=phase9`, `git diff --check` | No Sprint 7-owned blocker remains. This does not claim Phase 9 completion; optional standardized surfaces, registry closure, and the final phase gate remain separately owned. | Direct Sprint 7 evidence covers `comm_log`, `handoff`, `status_review`, and `lesson` as workbook-native artifact-backed coordination surfaces with minimum create signals, projection-backed reads, collection/reference semantics, timestamp behavior, and browser-visible workbook workflows. |
| [x] | 8. Optional surfaces and registry closure | `make backend-unit`, `make backend-store`, `make browser-e2e-webserver-backed`, `make phase-slice PHASE=phase9`, `make service-backed-slice PHASE=phase9`, `make phase-ledger-drift`, `git diff --check` | No Sprint 8-owned blocker remains. This does not claim Phase 9 completion; Sprint 9 final gate and handoff remain separately owned. | Current build exposes all three optional standardized surfaces; direct evidence covers `U-9-09`, `U-9-13`, `E-9-07`, and `E-9-08`. |
| [x] | 9. Final phase gate and handoff | public Phase 9 wrappers, drift checks, `make agent-finalize`, `make check`, `git diff --check` | No Sprint 9 blocker remains. | Final roots, generated-artifact status, optional-surface status, and Phase 10 follow-on notes are recorded below. |

## Evidence Layer Matrix

Every authoritative Phase 9 row must have exactly one authoritative row owner in `tools/phase9_test_map.json` before Phase 9 exit. Support-only evidence may be listed in manifest notes, but support-only evidence does not complete any Phase 9 row.

| Row | Evidence layer | Phase 9 claim intent |
| --- | --- | --- |
| `U-9-01` | `frontend_unit` | Keyboard command contract maps required keys without hidden macro behavior. |
| `U-9-02` | `backend_unit` | Paste and bulk ingest group one visible user action and preserve entity binding. |
| `U-9-03` | `backend_store` | Notes are artifact-backed `note` rows exposed through Notes. |
| `U-9-04` | `backend_store` | Indicators remain canonical rows with separate observations and lifecycle history. |
| `U-9-05` | `backend_store` | Party exact-match reuse and raw text preservation behave correctly. |
| `U-9-06` | `backend_store` | Assessments are append-only with closed state vocabulary and deterministic bands. |
| `U-9-07` | `backend_store` | Tasks and Decisions are bounded workbook surfaces, not a workflow engine. |
| `U-9-08` | `backend_store` | Coordination surfaces satisfy minimum create and projection-field behavior. |
| `U-9-09` | conditional `backend_store` or `N/A` | Optional standardized surfaces behave correctly only when exposed. |
| `U-9-10` | `backend_unit` | `timestamp_instant_v1` accepts, clears, and rejects exactly as owned. |
| `U-9-11` | `backend_store` | Party direct refs accept stable IDs only, clear with `value=null`, and fail closed on non-same-incident or inactive targets. |
| `U-9-12` | `backend_store` | Manual relationship collections reject client confidence and preserve authoritative `null`. |
| `U-9-13` | `backend_unit` | Artifact-backed variant registry preserves Notes and coordination identities. |
| `U-9-GRID-01` | `frontend_unit` | Keyboard navigation updates Cartulary anchors, not vendor selection alone. |
| `I-9-01` | `backend_integration` | Multi-row paste persists ordered mutations and origin distinctions. |
| `I-9-02` | `backend_integration` | Required surfaces persist and query through workbook projections. |
| `I-9-03` | `backend_integration` | Party-link helper fields update hidden links without overwriting text. |
| `I-9-GRID-01` | `frontend_unit` | Sorted or filtered paste translates visual order to stable anchors. |
| `E-9-01` | `browser_functional` | Required keyboard shortcuts work in the grid without module switching. |
| `E-9-02` | `browser_functional` | 20x5 Timeline paste creates or updates rows, preserves identity/selection, and presents grouped conflict navigation. |
| `E-9-03` | `browser_functional` | Notes tab supports in-grid creation and record linking. |
| `E-9-04` | `browser_functional` | Party create/link/clear flows keep raw text and preserve same-surface row context. |
| `E-9-05` | `browser_functional` | Assessment sequences remain distinguishable in history and filters. |
| `E-9-06` | `browser_functional` | Task, Decision, and coordination surfaces stay workbook-native. |
| `E-9-07` | conditional `browser_functional` or `N/A` | Optional standardized surfaces are covered when exposed. |
| `E-9-08` | `browser_functional` | Registry exposes required identities; optional surfaces are additive. |
| `E-9-GRID-01` | `browser_functional` | Grid semantics stay shared across required system views, including Host/Identity entity-origin clipboard paste. |

## Global References

Controlling guide and row set:
- `docs/guides/cartulary_implementation_testing_guide.md`, section 11, Phase 9.
- `U-9-01..U-9-13`, `U-9-GRID-01`.
- `I-9-01..I-9-03`, `I-9-GRID-01`.
- `E-9-01..E-9-08`, `E-9-GRID-01`.

Owner anchors:
- Core 00: `docs/spec/00_document_set_status_and_precedence.md`, especially status, precedence, conformance model, and contract-owner matrix.
- Core 01: `docs/spec/01_architecture_storage_and_view_contracts.md`, §7.4, §18A, §18B, and §19.
- Core 02: `docs/spec/02_domain_model_schema_and_history.md`, §10 and §19.
- Core 03: `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`, §2, §11, §13, and §16-§20.
- Core 04: `docs/spec/04_security_deployment_and_conformance.md`, §2.
- Core 05: `docs/spec/05_claim_publication_and_benchmark_reproducibility.md`, claim publication only.

Phase 9 AC groups:
- `AC-003`, `AC-005`, `AC-018`, `AC-068..AC-090`, `AC-112`, `AC-116..AC-122`, `AC-137..AC-145`, `AC-278..AC-280`, `AC-285..AC-287`, `AC-300..AC-304`, `AC-313..AC-319`, `AC-354`, `AC-395..AC-397`, `AC-410`.

Expected manifest and ledger paths:
- `tools/phase_registry.json`
- `tools/phase9_test_map.json`
- `docs/testing/phase9_coverage_ledger.md`
- Generated schedule artifacts under `tools/`, including `tools/scheduler_manifest.json` and `tools/execution_topology_render_index.json` when generation updates them.

Generated-boundary rules:
- Do not hand-edit `internal/gen/**`.
- Do not hand-edit `packages/protocol-ts/src/generated/**`.
- Do not hand-edit `pnpm-lock.yaml` or `go.sum`.
- Treat `contracts/*` as repo-local derived contract artifacts downstream of the normative core; edit them only as owner-driven contract updates.
- Keep codegen drift separate from migration drift.
- Regenerate ledgers and schedules only through canonical commands.

Relevant Phase 8 handoff substrate:
- Stable view-query validation and canonical `meta.query`.
- Stable view-schema discovery keys for later keyboard, clipboard, and surface controls.
- Direct addressing of required base surfaces through `sheet_ref.kind='view_schema'`.
- Saved-view addressing through `sheet_ref.kind='saved_view'`.
- Presentation-only group headers that cannot emit mutations.
- Full-row and sparse-patch contracts preserving hidden writable fields, authoritative nulls, stable `record_id`, `row_version`, `changed_field_keys[]`, and `affected_views[]`.
- Exact-token and strict-prefix search semantics.

Existing areas to inspect before implementation:
- `packages/grid-adapter/src`
- `packages/view-contracts/src`
- `contracts/view-schemas`
- `apps/web/src/WorkbookShell.tsx`
- `apps/web/src/workbookQuery.ts`
- `apps/web/e2e`
- `internal/modules/workbook`
- `internal/modules/assessments`
- `internal/modules/entities`
- `internal/modules/evidence`
- `internal/modules/links`
- `internal/modules/projections`
- `internal/modules/records`
- `internal/modules/viewschemas`
- `internal/platform/viewquery`
- `internal/platform/viewschema`
- `db/migrations`
- `db/queries`
- `contracts/openapi/cartulary.openapi.yaml`
- `docs/domain.md`

## Sprint 0. Ownership Manifest and Harness Setup

Objective: Establish Phase 9 ownership, selection, ledgers, schedules, and explicit placeholders or blocker sentinels before feature work.

Relevant IDs: all `U-9-*`, `U-9-GRID-01`, `I-9-*`, `I-9-GRID-01`, `E-9-*`, and `E-9-GRID-01`.

Files or areas to inspect:
- `tools/phase_registry.json`
- Existing `tools/phase*_test_map.json` files for manifest shape.
- Existing `docs/testing/phase*_coverage_ledger.md` generated ledgers for expected output shape.
- `scripts/check-phase-maps.sh`
- `scripts/render-phase-ledgers.mjs`
- `scripts/render-execution-topology-artifacts.mjs`
- `scripts/check-phase-test-names.mjs`
- Sprint 0 sentinel files by layer:
  - `internal/modules/workbook/phase9_sprint0_sentinel_test.go`
  - `internal/modules/entities/phase9_sprint0_sentinel_test.go`
  - `internal/modules/assessments/phase9_sprint0_sentinel_test.go`
  - `apps/web/src/WorkbookShell.phase9.sentinel.test.tsx`
  - `apps/web/e2e/phase9.sentinel.spec.ts`

Test-first sequence:
1. Create `tools/phase9_test_map.json` with every authoritative Phase 9 row and exactly one authoritative execution dependency per row.
2. Add Phase 9 to `tools/phase_registry.json` as `planned` until selectable row symbols exist.
3. Add placeholder or blocker-sentinel row owners only where the manifest tooling requires a discoverable test before behavior exists.
4. Move Phase 9 to `active` only after public wrappers can select coherent Phase 9 rows.
5. Generate ledgers and schedules through canonical commands.
6. Add `forbidden_id_files` for known support-only carryover files so previous phase evidence cannot claim Phase 9 IDs.

Implementation tasks:
- Add Phase 9 registry entry with manifest path `tools/phase9_test_map.json` and ledger path `docs/testing/phase9_coverage_ledger.md`.
- Encode the Evidence Layer Matrix in the manifest.
- Record authority model and Phase 8 substrate boundaries in manifest notes.
- Declare optional surface rows as conditional in row notes; do not pre-claim optional coverage.
- Ensure placeholder or sentinel names and titles include exact row IDs per harness rules.
- Generate ledgers and schedules.

Validation commands:
- `make phase-map-check`
- `make explain-phase PHASE=phase9`
- `make phase-ledgers`
- `make phase-ledger-drift`
- `make phase-schedules`
- `make phase-schedule-drift`
- `make phase-test-name-check`
- `make target-plan-json`
- `git diff --check`

Deliverables:
- `tools/phase9_test_map.json`
- Updated `tools/phase_registry.json`
- Generated `docs/testing/phase9_coverage_ledger.md`
- Generated schedule updates:
  - `tools/scheduler_manifest.json`
  - `tools/execution_topology_render_index.json`
- Intentional blocker-sentinel tests:
  - This list records the Sprint 0 seed state. Later sprints may replace individual sentinel owners with direct evidence; the current Phase 9 manifest and generated ledger remain the source of truth for current row ownership.
  - `internal/modules/workbook/phase9_sprint0_sentinel_test.go` covers `U-9-02`, `U-9-03`, `U-9-07`, `U-9-08`, `U-9-09`, `U-9-10`, `U-9-11`, `U-9-12`, `U-9-13`, `I-9-01`, `I-9-02`, and `I-9-03`.
  - `internal/modules/entities/phase9_sprint0_sentinel_test.go` covers `U-9-04` and `U-9-05`.
  - `internal/modules/assessments/phase9_sprint0_sentinel_test.go` covers `U-9-06`.
  - `apps/web/src/WorkbookShell.phase9.sentinel.test.tsx` initially covered `U-9-01`, `U-9-GRID-01`, and `I-9-GRID-01`; after Sprint 1 replacements, it remains the blocker owner for `I-9-GRID-01`.
  - `apps/web/e2e/phase9.sentinel.spec.ts` initially covered `E-9-01` through `E-9-08` and `E-9-GRID-01`; later sprints replaced those browser blocker claims with direct browser evidence. Use `tools/phase9_test_map.json` and `docs/testing/phase9_coverage_ledger.md` for current browser row ownership.

Sprint 0 validation record:
- `make phase-map-check`: initially failed because the new active Phase 9 ledger was missing.
  - Artifact root: `.cartulary/test-results/20260517T024138Z-p1619183`
  - Failure class: `harness`
  - Failure reason: `unknown_failure`
  - Message: `active phase9 ledger missing: docs/testing/phase9_coverage_ledger.md`
  - Ownership: Sprint 0-owned ordering issue.
  - Corrective action: generated the ledger through `make phase-ledgers`.
- `make phase-map-check`: passed after ledger generation.
  - Artifact root: `.cartulary/test-results/20260517T024152Z-p1619835`
- `make explain-phase PHASE=phase9`: passed and discovered Phase 9 with 27 authoritative rows.
- `make phase-ledgers`: passed.
  - Artifact root: `.cartulary/test-results/20260517T024201Z-p1620338`
- `make phase-ledger-drift`: passed.
  - Artifact root: `.cartulary/test-results/20260517T024204Z-p1620547`
- `make phase-schedules`: passed.
  - Artifact root: `.cartulary/test-results/20260517T024208Z-p1620825`
- `make phase-schedule-drift`: passed.
  - Artifact root: `.cartulary/test-results/20260517T024244Z-p1621838`
- `make phase-test-name-check`: passed.
  - Artifact root: `.cartulary/test-results/20260517T024247Z-p1622080`
- `make target-plan-json`: passed and emitted coherent Phase 9 target-plan rows.
- `git diff --check`: passed.

Sprint 0 status:
- Complete for harness setup and historical seed-state ownership.
- Phase 9 remains `active` because public wrappers can select coherent Phase 9 rows.
- Some Sprint 0 blocker sentinels have been replaced by direct later-sprint evidence; remaining blocker sentinels intentionally prevent broader Phase 9 completion claims until the relevant sprint replaces them with direct behavior evidence.

Risks:
- Prematurely marking Phase 9 active without selectable rows.
- Letting Phase 4 or Phase 8 support evidence claim Phase 9 completion.
- Forgetting conditional `N/A` handling for optional standardized surfaces.
- Adding generated ledgers or schedules by hand.

Exit criteria:
- Phase 9 is discoverable through `make explain-phase PHASE=phase9`.
- Manifest and ledger drift checks pass.
- All placeholder or sentinel rows are visible, intentional, and not described as behavior completion while they remain sentinels.
- Phase 9 is active but incomplete while any blocker sentinel remains.

## Sprint 1. Keyboard Contract and Grid Anchors

Objective: Implement and prove required keyboard behavior and Cartulary focus-anchor updates across workbook grid surfaces.

Relevant IDs: `U-9-01`, `U-9-GRID-01`, `E-9-01`, `E-9-GRID-01`; support context from `AC-003`, `AC-005`, and `AC-047`.

Files or areas to inspect:
- `packages/grid-adapter/src`
- `apps/web/src/WorkbookShell.tsx`
- Existing frontend workbook tests under `apps/web/src`
- Existing browser workflow specs under `apps/web/e2e`
- Focus-anchor symbols: `GridCellAnchor`, `resolveGridCellAnchor`, and `navigateGridCellAnchor` in `packages/grid-adapter/src/core.ts`; `formatWorkbookFocusAnchor`, `useWorkbookGridFocus`, and `rowCellTestId` wiring in `apps/web/src/WorkbookShell.tsx`.
- Keyboard-command symbols: `WorkbookKeyboardCommand`, `WorkbookKeyboardAvailability`, and `mapWorkbookKeyboardCommand` in `apps/web/src/workbookKeyboard.ts`.
- Direct Sprint 1 test files: `apps/web/src/workbookKeyboard.test.ts`, `apps/web/src/GridAdapter.phase9.anchor.test.ts`, and `apps/web/e2e/phase9.keyboard.spec.ts`.

Test-first sequence:
1. Add frontend unit tests for Arrow, Enter, Shift+Enter, Tab, Ctrl+V, Ctrl+K, Space, Alt+H, and Esc command mapping.
2. Add grid-adapter tests proving navigation updates or intentionally clears Cartulary anchors rather than vendor selection state alone.
3. Add browser test coverage for keyboard shortcuts on a live workbook grid without route/module switching.
4. Keep tests independent from later paste and surface workflow implementation except where command dispatch needs a sentinel.

Implementation tasks:
- Normalize keyboard event handling around stable `record_id` and `field_key` anchors.
- Make Enter, Shift+Enter, Tab, and Arrow navigation preserve valid anchors or intentionally clear them.
- Route Ctrl+V into the paste pipeline without hidden macro behavior.
- Route Ctrl+K, Space, Alt+H, and Esc to same-surface link/preview/history/inspector-close behavior where the active surface supports it.
- Ensure disabled or unavailable actions fail closed without changing authoritative row state.
- Preserve active editor drafts across valid focus movement where existing save-state rules require it.

Validation commands:
- `make frontend-unit`
- `make browser-e2e-webserver-backed`
- `make phase-slice PHASE=phase9`
- `git diff --check`

Deliverables:
- Direct authoritative evidence for `U-9-01`: `apps/web/src/workbookKeyboard.test.ts`.
- Direct authoritative evidence for `U-9-GRID-01`: `apps/web/src/GridAdapter.phase9.anchor.test.ts`.
- Direct authoritative browser evidence for `E-9-01`: `apps/web/e2e/phase9.keyboard.spec.ts`.
- Browser shared-grid evidence for `E-9-GRID-01`: `apps/web/e2e/phase9.keyboard.spec.ts`.
- Ctrl+V evidence is paste-intent and routing/sentinel behavior only; Sprint 1 does not claim Sprint 2 clipboard paste, multi-cell paste, fill-down, sorted or filtered paste anchoring, or bulk-edit semantics.

Sprint 1 validation record:
- `make frontend-unit`: failed overall because the later Phase 9 `I-9-GRID-01` blocker sentinel remains selected.
  - Artifact root: `.cartulary/test-results/20260517T154853Z-p203958`
  - Sprint 1-owned status: `U-9-01` and `U-9-GRID-01` frontend-unit evidence passed.
  - Failure ownership: later Phase 9 / Sprint 2, not Sprint 1.
- `make browser-e2e-webserver-backed`: failed overall because later Phase 9 browser-functional sentinels and an outside-Phase-9 browser regression remain selected.
  - Artifact root: `.cartulary/test-results/20260517T154929Z-p205996`
  - Sprint 1-owned status: `E-9-01` and `E-9-GRID-01` evidence in `apps/web/e2e/phase9.keyboard.spec.ts` passed.
  - Failure ownership: `E-9-02` through `E-9-08` are later Phase 9 blocker sentinels; `E-4-04` is outside Phase 9.
- `make phase-slice PHASE=phase9`: failed overall because later Phase 9 sentinels remain selected.
  - Artifact root: `.cartulary/test-results/20260517T155126Z-p214602`
  - Sprint 1-owned status: direct Sprint 1 frontend and browser rows passed.
  - Failure ownership: later Phase 9, not Sprint 1.
- `git diff --check`: passed.

Sprint 1 status:
- Complete for row-level Sprint 1 claimability.
- Phase 9 remains active and incomplete while later blocker sentinels remain.
- Existing Sprint 1 evidence proves keyboard command dispatch, adapter-owned anchor translation by `record_id` and `field_key`, live workbook shortcut behavior without module switching, and shared grid keyboard anchors across representative workbook grid surfaces.

Risks:
- Treating RDG or vendor selection state as the authoritative focus anchor.
- Keyboard shortcuts interfering with browser or screen-reader defaults in ways the product does not own.
- Adding shortcut behavior that creates hidden macros or broad workflow automation.

Exit criteria:
- Required keyboard commands are covered by direct Phase 9 row evidence.
- Focus anchors remain stable by `record_id` and `field_key`, or are intentionally cleared in declared cases.
- No Phase 9 keyboard behavior depends on Phase 8 support-only evidence.

## Sprint 2. Clipboard Paste and Bulk Editing

Objective: Implement and prove clipboard paste, bulk edit, fill-down, multi-row tag assignment, and stable anchor translation.

Relevant IDs: `U-9-02`, `I-9-01`, `I-9-GRID-01`, `E-9-02`, `E-9-GRID-01`; support context from `AC-003`, `AC-040`, `AC-125`, `AC-126`, and `AC-201`.

Files or areas to inspect:
- `internal/modules/workbook`
- `internal/modules/imports/tabularingest`
- `internal/modules/entities`
- `internal/modules/timeline`
- `internal/platform/viewquery`
- `internal/modules/links`
- `apps/web/src/WorkbookShell.tsx`
- `apps/web/src/workbookQuery.ts`
- `packages/grid-adapter/src`
- `apps/web/e2e/phase9.sentinel.spec.ts`
- `apps/web/e2e/phase9.keyboard.spec.ts`
- Raw-capture provenance is persisted as structured clipboard/import source-column metadata owned by the shared ingest and Timeline persistence path.

Test-first sequence:
1. Backend unit tests cover the shared tabular-ingest contract, quoted/newline CSV, explicit single-row CSV parsing, change-set grouping, fill-down request shaping, and multi-row tag assignment request shaping.
2. Backend integration tests cover multi-row Timeline paste, ordered mutations, unknown-column preservation, `mention_origin` versus `entity_origin`, Host/Identity exact-match reuse or stub creation, and bulk mutation batches.
3. Frontend unit tests cover sorted or filtered paste translating visual ranges to stable anchors, group-row rejection, vendor-coordinate rejection, rendered paste dispatch, and comma-only scalar Ctrl+V behavior.
4. Browser tests cover representative 20x5 Timeline paste, grouped conflict navigation, and Host/Identity entity-origin paste through the shared grid path.
5. Manifest and ledger checks cover Playwright `titles[]` traceability and row-level `claim_status`.

Implementation tasks:
- Parse TSV/CSV clipboard payloads without treating file import as implemented.
- Map known columns to stable `field_key` values.
- Preserve unknown columns in the owner-defined raw-capture structure.
- Apply entity binding according to active field `entity_binding_mode`.
- Treat single-line comma-only default Ctrl+V as scalar text; explicit API CSV ingest remains supported.
- Commit successful non-conflicting paste writes as one visible change set with ordered mutation entries.
- Keep same-field conflicts outside the committed paste batch until explicit resolution.
- Translate visible paste order through sorted or filtered result sets to stable `record_id` or create-row anchors.
- Ensure grouped or presentation-only rows cannot become mutation targets.
- Use the same frontend grid dispatcher for Timeline, Hosts, and Identities. Timeline sends record or create targets with row versions; Hosts and Identities send create-only targets and rely on backend `entity_origin` exact-match reuse.

Validation commands:
- `make backend-unit`
- `make backend-integration`
- `make frontend-unit`
- `make browser-e2e-webserver-backed`
- `make phase-ledgers`
- `make phase-ledger-drift`
- `make phase-slice PHASE=phase9`
- `make service-backed-slice PHASE=phase9`
- `make agent-finalize`
- `git diff --check`

Deliverables:
- Direct authoritative evidence for `U-9-02`, `I-9-01`, `I-9-GRID-01`, `E-9-02`, and `E-9-GRID-01`.
- Updated manifest notes distinguishing clipboard paste from Import Extension Profile import.
- Manifest rows with `claim_status`: Sprint 2 direct rows are `implemented`; remaining later rows stay `blocked`.
- Playwright `titles[]` for multi-scenario browser rows, including Timeline conflict navigation and Host/Identity entity-origin paste.

Sprint 2 validation record:
- `make backend-unit`: passed.
  - Artifact root: `.cartulary/test-results/20260518T002640Z-p673399`
  - Sprint 2-owned status: `U-9-02` backend-unit evidence passed.
- `make backend-integration`: passed.
  - Artifact root: `.cartulary/test-results/20260518T002657Z-p675854`
  - Sprint 2-owned status: `I-9-01` backend-integration evidence passed.
- `make frontend-unit`: passed.
  - Artifact root: `.cartulary/test-results/20260518T002640Z-p673441`
  - Sprint 2-owned status: `I-9-GRID-01` frontend-unit evidence passed.
- `make browser-e2e-webserver-backed`: failed overall on outside-Phase-9 `E-4-04`.
  - Artifact root: `.cartulary/test-results/20260518T003227Z-p686779`
  - Sprint 2-owned status: Phase 9 authoritative browser-functional evidence passed, including `E-9-02` and `E-9-GRID-01`.
  - Failure ownership: outside Phase 9 / Phase 4 auto-resolve behavior, not Sprint 2.
- `make phase-slice PHASE=phase9`: passed.
  - Artifact root: `.cartulary/test-results/20260518T003349Z-p694667`
  - Sprint 2-owned status: 5/5 selected work units passed, 43 tests, no failures.
- `make service-backed-slice PHASE=phase9`: passed.
  - Artifact root: `.cartulary/test-results/20260518T003918Z-p706604`
  - Sprint 2-owned status: 3/3 selected work units passed, 25 tests, no failures.
- `make phase-ledger-drift`: passed.
  - Artifact root: `.cartulary/test-results/20260518T004647Z-p724344`
  - Sprint 2-owned status: generated Phase 9 ledger reflects `tools/phase9_test_map.json`.
- `git diff --check`: passed.
- `git diff --cached --check`: passed as an extra hygiene check because the tree already had staged changes.
- `make phase-ledgers` and `make agent-finalize`: not run during the audit because they can refresh generated maintenance artifacts, and the audit scope prohibited regeneration.

Sprint 2 status:
- Complete for row-level Sprint 2 claimability.
- `U-9-02`, `I-9-01`, `I-9-GRID-01`, `E-9-02`, and relevant `E-9-GRID-01` evidence is direct, Phase 9-owned, and passing through the row-selected Phase 9 targets.
- File-based CSV/XLSX import remains unclaimed and separate from clipboard paste.
- Support-only tabular-ingest and Timeline helper tests remain support-only; prior Phase 4, Phase 6, and Phase 8 evidence remains support-only and cannot complete Sprint 2.
- Traceability follow-up: direct Sprint 2 tests in files named `*.sentinel*` are valid through manifest titles, but the names are weaker than the current behavior they now contain.
- Coverage follow-up: `U-9-02` unit evidence directly exercises `mention_origin`; `entity_origin` behavior is directly proven at integration and browser layers.

Risks:
- Collapsing paste into per-cell visible actions rather than one visible user action.
- Retargeting paste writes by visible row index after sort/filter changes.
- Losing unknown-column remnants.
- Treating bulk edit as hidden macro execution.
- Treating ordinary comma text as tabular paste and splitting analyst-entered scalar values.
- Letting a green Phase 9 wrapper imply full Phase 9 completion while blocker rows remain.

Exit criteria:
- Clipboard and bulk edit evidence passes in direct Sprint 2 Phase 9 rows.
- Same-field conflict batching behavior is tested.
- Host/Identity paste has browser-functional evidence under `E-9-GRID-01`.
- Generated ledgers and phase-slice summaries expose incomplete Phase 9 claim status while blocked rows remain.
- File-based import remains unclaimed.

## Sprint 3. Notes and Indicators

Objective: Complete Notes as an artifact-backed built-in sheet and Indicators as canonical rows with observations and lifecycle history.

Relevant IDs: `U-9-03`, `U-9-04`, `I-9-02`, `E-9-03`, `E-9-08`, `E-9-GRID-01`; support context from `AC-068..AC-070`, `AC-112`, `AC-116..AC-122`.

Completion status: complete for Sprint 3 as of the read-only audit retained on 2026-05-18. The audit found direct Phase 9 evidence for each Sprint 3 row and did not count Phase 4 smoke coverage, support-only tests, generated ledgers, generated schedules, retained run artifacts, or broad aggregate green status as behavior authority. This status does not claim Phase 9 completion.

Files or areas to inspect:
- `contracts/view-schemas/cartulary.view.notes.v1.json`
- `contracts/view-schemas/cartulary.view.indicators.v1.json`
- `internal/modules/workbook`
- `internal/modules/entities`
- `internal/modules/projections`
- `internal/platform/viewschema`
- `apps/web/src/WorkbookShell.tsx`
- `apps/web/e2e/phase4Helpers.ts`
- Notes persistence is not separated into a Notes-only module; Notes create, linked-note create, tags, history, and projection query run through shared workbook artifact paths in `internal/modules/workbook/mutation_store.go`, `internal/modules/workbook/store.go`, and `db/migrations/00016_phase9_linked_notes_projection.sql`.
- Indicator canonical storage, observation storage, lifecycle interval storage, and projection rollups run through `internal/modules/entities/indicator_store.go`; canonical Indicator create is routed through `internal/modules/entities/routes.go` for `cartulary.view.indicators.v1`.

Completed evidence:
1. `internal/modules/workbook/phase9_notes_indicators_test.go::TestPhase9_NotesAreArtifactBackedRows_U_9_03` proves Notes create/query use artifact-backed `artifact_type='note'` rows, shared tags and revision history, no Notes-specific table, linked-note `record_links`, and `cartulary.view.notes.v1` query exposure.
2. `internal/modules/entities/phase9_indicators_test.go::TestPhase9_IndicatorsCanonicalObservationLifecycle_U_9_04` proves canonical indicator identity, duplicate canonical create reuse, source-bound observations as distinct occurrences, lifecycle intervals as distinct history, and projection rollups.
3. `internal/modules/workbook/phase9_notes_indicators_test.go::TestPhase9_NotesAndIndicatorsQueryThroughWorkbookProjections_I_9_02` proves Notes and Indicators persist and query through workbook projections.
4. `apps/web/e2e/phase9.sentinel.spec.ts::Phase 9 E-9-03 Notes tab creates artifact-backed linked notes` proves Notes built-in tab creation and record linking.
5. `apps/web/e2e/phase9.sentinel.spec.ts::Phase 9 E-9-08 Notes and Indicators open by canonical view schema IDs` proves Sprint 3 registry opening through canonical identities.
6. `apps/web/e2e/phase9.keyboard.spec.ts::Phase 9 E-9-GRID-01 shared grid keyboard anchors stay stable across workbook cells` proves the Notes/Indicators portions of shared grid anchor semantics used by Sprint 3.

Implementation tasks:
- Ensure Notes create and linked-note actions use the shared artifact model.
- Preserve Note title/body/tag semantics through artifact rows, record history, and projections.
- Ensure Indicators query canonical indicator rows and expose observation/lifecycle pivots without collapsing source occurrences into canonical identity.
- Keep indicator create-only identity fields immutable where owner contracts require it.
- Keep Notes and Indicators visible/openable through canonical `view_schema_id` identities.

Validation commands:
- `make backend-store`
- `make backend-integration`
- `make browser-e2e-webserver-backed`
- `make phase-slice PHASE=phase9`
- `make service-backed-slice PHASE=phase9`
- `git diff --check`

Deliverables:
- Direct authoritative evidence for `U-9-03`, `U-9-04`, `I-9-02`, `E-9-03`, and the Notes/Indicators portions of `E-9-08` and `E-9-GRID-01`.
- Retained validation roots from the Sprint 3 audit:
  - `make backend-store`: `.cartulary/test-results/20260518T022943Z-p936004`
  - `make backend-integration`: `.cartulary/test-results/20260518T023509Z-p946945`
  - `make browser-e2e-webserver-backed`: `.cartulary/test-results/20260518T024044Z-p958235`
  - `make phase-slice PHASE=phase9`: `.cartulary/test-results/20260518T024202Z-p965722`
  - `make service-backed-slice PHASE=phase9`: `.cartulary/test-results/20260518T024749Z-p979405`
  - `git diff --check`: passed without retained harness root.
- Targeted retained evidence roots inside the Phase 9 wrapper run:
  - Backend store Notes/Indicators: `.cartulary/test-results/20260518T024202Z-p965722/backend-store/backend-store-phase9-notes-and-indicators-evidence`
  - Backend integration Notes/Indicators: `.cartulary/test-results/20260518T024202Z-p965722/backend-integration/backend-integration-phase9-notes-and-indicators-evidence`
  - Browser Phase 9 authoritative selection: `.cartulary/test-results/20260518T024044Z-p958235/browser-e2e-webserver-backed/browser-e2e-functional-phase9-authoritative`

Risks:
- Treating `artifact` as a synonym for Notes.
- Reusing indicator observation rows as canonical indicators.
- Letting Phase 4 route smoke tests stand in for Phase 9 surface workflows.

Exit criteria:
- Notes and Indicators direct Phase 9 rows pass.
- Notes remains artifact-backed and required through `cartulary.view.notes.v1`.
- Indicators retain canonical row versus observation/lifecycle separation.

## Sprint 4. Parties and Text-Plus-Link Flows

Objective: Complete incident-scoped Parties and source-preserving party text plus hidden direct-reference link workflows.

Relevant IDs: `U-9-05`, `U-9-11`, `I-9-03`, `E-9-04`; support context from `AC-277..AC-280` and `AC-315..AC-319`.

Files or areas to inspect:
- `contracts/view-schemas/cartulary.view.parties.v1.json`
- `internal/modules/entities`
- `internal/modules/evidence`
- `internal/modules/workbook`
- `internal/modules/revisions`
- `internal/platform/viewschema`
- `apps/web/src/WorkbookShell.tsx`
- `apps/web/e2e`
- `internal/modules/workbook/mutation_api.go` for direct-reference decode and `value=null` clear validation.
- `internal/modules/workbook/mutation_store.go` for exact Party reuse and same-incident active-target validation.
- `internal/modules/revisions/delete_restore_store.go` for active incoming Party-reference delete guards.

Evidence sequence:
1. Backend store `U-9-05` evidence proves Party create, incident scoping, raw text preservation, and exact-match reuse by normalized `primary_email` or `external_ref` only.
2. Direct-reference `U-9-11` evidence proves exact stable Party IDs, rejects non-ID inputs and non-direct clear shapes, accepts direct-write `value=null`, and rejects foreign, deleted, wrong-type, and deployment-user targets. This evidence currently runs in the backend-store layer rather than backend-unit because active target validation depends on persisted incident state.
3. Backend integration `I-9-03` evidence proves hidden `*_party_id` fields update without overwriting requester, collector, source, audience, or attendee text.
4. Browser-functional `E-9-04` evidence proves create-from-text, link-existing, clear-link, clear-text, and clear-both flows remain same-surface and preserve visible row context where exposed.

Completed behavior:
- Keep `party` as an incident-scoped first-class record, not a deployment user or global contact.
- Reuse exact active same-incident Parties only by normalized `primary_email`, or failing that, `external_ref`.
- Preserve requester, collector, source, audience, and attendee text independently from party refs.
- Use ordinary record patch direct writes with `value=null` for direct-reference clears.
- Reject non-direct-write clear shapes, fuzzy IDs, labels, emails as refs, whitespace-padded IDs, arrays, objects, booleans, numbers, foreign incident parties, deleted party targets, wrong-type records, and deployment-user IDs.
- Keep same-surface focus, scroll, and originating row context after party actions where the browser flow is implemented.

Validation commands:
- `make backend-unit`: passed; retained root `.cartulary/test-results/20260518T165456Z-p2237222`.
- `make backend-store`: passed; retained root `.cartulary/test-results/20260518T165511Z-p2238234`.
- `make backend-integration`: passed; retained root `.cartulary/test-results/20260518T170032Z-p2248234`.
- `make browser-e2e-webserver-backed`: passed; retained root `.cartulary/test-results/20260518T170607Z-p2259417`.
- `make phase-slice PHASE=phase9`: passed; retained root `.cartulary/test-results/20260518T170722Z-p2266675`.
- `make service-backed-slice PHASE=phase9`: passed; retained root `.cartulary/test-results/20260518T171258Z-p2279272`.
- `git diff --check`: passed without retained harness root.

Deliverables:
- Direct authoritative evidence for `U-9-05`, `U-9-11`, `I-9-03`, and `E-9-04`.
- Targeted Sprint 4 evidence inside the Phase 9 wrapper run:
  - Backend store Parties evidence: `.cartulary/test-results/20260518T170722Z-p2266675/backend-store/backend-store-phase9-parties-evidence`
  - Backend integration Parties evidence: `.cartulary/test-results/20260518T170722Z-p2266675/backend-integration/backend-integration-phase9-parties-evidence`
  - Browser Phase 9 authoritative selection: `.cartulary/test-results/20260518T170722Z-p2266675/browser-e2e-webserver-backed/browser-e2e-functional-phase9-authoritative`

Risks:
- Auto-creating or auto-linking parties from ordinary text entry.
- Clearing source text when clearing `party_id`, or clearing `party_id` when clearing text.
- Treating email or display name as a submitted direct-reference scalar.
- Introducing party merge or global contact behavior.
- Treating the `U-9-11` backend-store evidence layer as a backend-unit row without either splitting decoder-only unit coverage or updating row-layer expectations.

Exit criteria:
- Party text-plus-link semantics are directly evidenced.
- Raw source text and direct party refs remain independently controlled.
- Party references inherit incident authorization without record-specific ACLs.
- Sprint 4 is complete for direct behavior evidence. This does not claim full Phase 9 completion.

## Sprint 5. Assessments and Timestamp Contract

Objective: Complete Compromise Assessments as append-only rows and close direct timestamp validation behavior across Phase 9 surfaces.

Relevant IDs: `U-9-06`, `U-9-10`, `U-9-12`, `I-9-02`, `E-9-05`; support context from `AC-018`, `AC-080..AC-084`, `AC-300..AC-304`, `AC-354`, and `AC-395..AC-397`.

Files or areas to inspect:
- `contracts/view-schemas/cartulary.view.assessments.v1.json`
- `internal/modules/assessments`
- `internal/modules/workbook`
- `internal/platform/viewschema`
- `internal/platform/fieldnorm`
- `apps/web/src/WorkbookShell.assessments.test.tsx`
- `apps/web/e2e/phase4.workbook.assessments.spec.ts`
- `internal/platform/fieldnorm/timestamp.go` for `timestamp_instant_v1` parsing and UTC normalization.
- `apps/web/src/WorkbookShell.tsx` for assessment create drafts, band-first confidence mapping, and invalid timestamp draft preservation.
- `apps/web/e2e/phase9.sentinel.spec.ts` for the direct `E-9-05` browser workflow.

Evidence sequence:
1. Backend store `U-9-06` evidence proves append-only assessment rows, closed `assessment_state`, distinct `disproven` and `cleared`, rejection of operational-response state values, rejection of in-place semantic patches, minimum create-set enforcement, create defaults, deterministic `confidence_band` derivation, separate `assessment_state` and `confidence_band` filters, and canonical confidence-band sort order.
2. Backend unit `U-9-10` evidence proves `timestamp_instant_v1` accepts only RFC 3339 JSON strings with explicit `Z` or numeric offsets, normalizes offset-equivalent instants to UTC for equality and idempotency hashing, rejects invalid shapes including empty strings and timezone-less strings, and applies explicit JSON `null` only to clearable timestamp fields.
3. Backend store `U-9-12` evidence proves the assessment `support_refs` collection action rejects client-supplied `confidence` and persists manual `supported_by` links with `provenance='manual'` and `confidence=NULL`.
4. Backend integration `I-9-02` evidence proves assessment rows query through `assessment_grid_projection` with projected confidence band and support-link count where Sprint 5 depends on projection-backed reads.
5. Browser-functional `E-9-05` evidence proves visible assessment history across `unknown`, `suspected`, `confirmed`, `disproven`, and `cleared`, band-first confidence behavior, separate assessment filters, support-link count display, and invalid `assessment.assessed_at` draft preservation after authoritative save rejection.

Completed behavior:
- Assessment create commits only when `assessment.subject_ref`, `assessment.subject_type`, `assessment.assessment_state`, and non-empty `assessment.rationale` are present after create-time normalization.
- Context-preseeded subject values and `assessment.support_refs` may seed the create surface, but do not satisfy the minimum semantic create set by themselves.
- Existing assessment rows reject in-place semantic edits to append-only assessment fields and supporting-link semantics.
- Operational response terms such as `contained`, `isolated`, `disabled`, `reset`, and `monitored` remain outside `assessment_state`.
- Omitted `assessment.assessed_at`, `assessment.assessor`, and `assessment.confidence_score` default to commit timestamp, authenticated actor, and `NULL`.
- `confidence_score=NULL` derives `confidence_band='unset'`; exact score bands derive `low`, `medium`, and `high` at `0..39`, `40..69`, and `70..100`; canonical ascending order is `unset`, `low`, `medium`, `high`.
- Band-first editing persists `unset`, `low`, `medium`, and `high` as `NULL`, `25`, `55`, and `85`.
- `assessment.assessed_at` binds to `timestamp_instant_v1`, is `clearable=false`, and rejects explicit JSON `null`.
- Empty strings, timezone-less strings, date-only strings, and non-string JSON values fail closed for `timestamp_instant_v1`; empty string is never a timestamp clear token.
- Invalid timestamp drafts remain client-local unsaved editor state without creating a durable row or rendering the invalid draft as an authoritative persisted value.
- Assessment manual relationship collection actions preserve server-owned link confidence: client-supplied `confidence` is rejected and authoritative manual links keep `confidence=NULL`.

Validation commands:
- `make backend-unit`: passed; retained root `.cartulary/test-results/20260518T215646Z-p2813166`.
- `make backend-store`: passed; retained root `.cartulary/test-results/20260518T215700Z-p2814677`.
- `make backend-integration`: passed; retained root `.cartulary/test-results/20260518T220224Z-p2824898`.
- `make browser-e2e-webserver-backed`: passed; retained root `.cartulary/test-results/20260518T220755Z-p2836242`.
- `make phase-slice PHASE=phase9`: passed; retained root `.cartulary/test-results/20260518T220907Z-p2843650`.
- `make service-backed-slice PHASE=phase9`: passed; retained root `.cartulary/test-results/20260518T221452Z-p2856624`.
- `git diff --check`: passed without retained harness root.

Deliverables:
- Direct authoritative evidence for `U-9-06`, `U-9-10`, the assessment portion of `U-9-12`, applicable `I-9-02`, and `E-9-05`.
- Targeted Sprint 5 evidence inside the Phase 9 wrapper run:
  - Backend unit timestamp evidence: `.cartulary/test-results/20260518T220907Z-p2843650/backend-unit/backend-unit-phase9-timestamp-evidence`
  - Backend store Assessments evidence: `.cartulary/test-results/20260518T220907Z-p2843650/backend-store/backend-store-phase9-assessments-evidence`
  - Backend store assessment support-refs evidence: `.cartulary/test-results/20260518T220907Z-p2843650/backend-store/backend-store-phase9-assessment-support-refs-evidence`
  - Backend integration required-surface projection evidence: `.cartulary/test-results/20260518T220907Z-p2843650/backend-integration/backend-integration-phase9-required-surface-projection-evidence`
  - Browser Phase 9 authoritative selection: `.cartulary/test-results/20260518T220907Z-p2843650/browser-e2e-webserver-backed/browser-e2e-functional-phase9-authoritative`

Risks:
- Overwriting prior assessment history.
- Using operational response vocabulary as assessment state.
- Treating empty string as timestamp clear.
- Silently coercing timezone-less timestamps.

Exit criteria:
- Assessment and timestamp direct evidence passes.
- Invalid timestamp behavior is visibly client-local until corrected, discarded, or explicitly cleared where clearable.
- Manual relationship confidence remains authoritative `null`.
- Sprint 5 is complete for direct behavior evidence. This does not claim full Phase 9 completion.

## Sprint 6. Task Requests and Decisions

Objective: Complete Task Requests and Decisions as workbook-native bounded lifecycle surfaces with owner, queue, support, and direct-reference semantics.

Relevant IDs: `U-9-07`, `U-9-11`, `U-9-12`, `I-9-02`, `E-9-06`, `E-9-GRID-01`; support context from `AC-085`, `AC-086`, `AC-137..AC-145`, `AC-313`, `AC-314`, and `AC-315..AC-319`.

Files or areas to inspect:
- `contracts/view-schemas/cartulary.view.task_requests.v1.json`
- `contracts/view-schemas/cartulary.view.decisions.v1.json`
- `internal/modules/workbook`
- `internal/modules/projections`
- `internal/modules/records`
- `internal/platform/viewschema`
- `apps/web/src/WorkbookShell.tsx`
- `apps/web/e2e`
- Task Request storage, lifecycle, relationship, and projection behavior is realized through the shared workbook store and projection paths, including `internal/modules/workbook/mutation_store.go`, `internal/modules/workbook/store.go`, `internal/modules/projections/tasks_decisions.go`, and the Task Requests view schema contract.
- Decision storage, lifecycle, relationship, explicit supersession, and projection behavior is realized through the shared workbook store and projection paths, including `internal/modules/workbook/mutation_store.go`, `internal/modules/workbook/routes.go`, `internal/modules/projections/tasks_decisions.go`, and the Decisions view schema contract.

Completion status: complete for Sprint 6 as of the residual remediation update on 2026-05-20. The implementation already matched the owner specs; the remaining work was test-only evidence hardening for explicit supersession fail-closed behavior against inconsistent source or target state and for wire-level rejection of `task.owner_user_id` null clear. This status does not claim Phase 9 completion.

Test-first sequence:
1. Add backend store tests for task lifecycle guards, owner semantics, queue fields, and minimum create signal.
2. Add backend store tests for decision lifecycle transitions, terminal states, explicit supersession, and fail-closed inconsistent machine state.
3. Add backend unit tests for `same_incident_decision_ref_v1` direct-reference semantics.
4. Add backend store tests for task and decision collection fields rejecting client `confidence`.
5. Add browser tests proving Task Requests and Decisions are workbook surfaces with queue, due-date, blocked-work, owner, support, and decision-link flows.

Completed evidence:
1. `internal/modules/workbook/phase9_task_decisions_store_test.go::TestPhase9Sprint6_TaskRequestLifecycleDecisionLinksAndProjection` proves Task Request minimum create, lifecycle, owner/default semantics, queue fields, due-date clear, blocked/done lifecycle side effects, direct decision-link set/clear, invalid decision targets, manual `record_links`, and projection-backed queue reads.
2. `internal/modules/workbook/phase9_task_decisions_store_test.go::TestPhase9Sprint6_TaskLifecycleGuardFailures_U_9_07` proves invalid Task Request lifecycle states fail closed without partial records, links, or row-version changes.
3. `internal/modules/workbook/phase9_task_decisions_store_test.go::TestPhase9Sprint6_DecisionLifecycleSupersessionAndConsistency` proves Decision minimum create, legal lifecycle transitions, terminal-state guards, support and affected refs, direct `superseded` rejection, explicit supersession, idempotent replay, executed-target supersession projection, and fail-closed inconsistent-machine behavior for ordinary grid mutations.
4. `internal/modules/workbook/phase9_task_decisions_store_test.go::TestPhase9Sprint6_SupersedeDecisionRejectsInconsistentSourceOrTarget` proves the explicit supersession action itself fails closed when either the source or target decision machine state is inconsistent, with no new supersession links, relationship links, or row-version changes.
5. `internal/modules/workbook/phase9_task_decisions_store_test.go::TestPhase9Sprint6_DecisionTerminalTransitionMatrix_U_9_07` proves terminal Decision transition behavior, including rejection of direct `superseded`.
6. `internal/modules/workbook/mutation_api_test.go::TestSupportPhase9_DirectDecisionReferenceDecoderAcceptsOnlyExactStableIDs` proves `same_incident_decision_ref_v1` accepts exact stable IDs, accepts direct `value=null` clear for `task.decision_record_id`, rejects non-ID inputs, and rejects collection-action clear shapes.
7. `internal/modules/workbook/mutation_api_test.go::TestSupportPhase9_TaskOwnerNullClearRejectedAtDecoder` proves `task.owner_user_id` rejects wire-level `value=null` with `400 invalid_mutation_payload` and `reason_code='field_not_nullable'` before store execution.
8. `internal/modules/workbook/mutation_api_test.go::TestPhase9TaskDecisionRelationshipConfidenceRejected` proves Task/Decision relationship collection payloads reject client-supplied `confidence`.
9. Browser evidence in `apps/web/e2e/phase9.sentinel.spec.ts` covers native workbook flows for Decisions and Task Requests, including support refs, affected records, supersession, queue/priority, due date, blocked reason, owner filtering, and direct decision-link clear.

Implementation tasks:
- Implement or harden Task Request create, patch, lifecycle guard, owner, due-date, blocked-reason, completion, linked-record, and decision-link behavior.
- Implement or harden Decision create, patch, lifecycle guard, support refs, affected/supersession projections, and explicit supersession behavior.
- Reject generalized workflow engine behavior and mandatory Timeline approval fields.
- Keep direct decision-link clear on the ordinary record patch route with `value=null`.
- Keep Task Requests and Decisions discoverable/openable by canonical `view_schema_id`.
- Ensure queue and owner filters use projection-backed stable fields.

Validation commands:
- `make backend-unit`
- `make backend-store`
- `make backend-integration`
- `make browser-e2e-webserver-backed`
- `make phase-slice PHASE=phase9`
- `make service-backed-slice PHASE=phase9`
- `git diff --check`

Deliverables:
- Direct authoritative evidence for `U-9-07`, decision-reference portions of `U-9-11`, task/decision portions of `U-9-12`, relevant `I-9-02` coverage, and Task/Decision portions of `E-9-06`.
- Residual remediation evidence for explicit supersession fail-closed behavior and wire-level owner null-clear rejection.
- Focused residual validation:
  - `go test ./internal/modules/workbook -run 'Phase9Sprint6|Direct.*Decoder|Owner.*Null|Supersede.*Inconsistent'`: passed.
  - `make agent-finalize`: passed with `generated=unchanged`; retained-run maintenance was skipped because `RESULTS_DIR` was unset.
  - Artifact root: `.cartulary/test-results/20260520T021445Z-p565332`.

Risks:
- Treating `approved` as a generalized row-edit approval gate.
- Allowing `done -> canceled`, `canceled -> done`, direct `superseded`, or other illegal transitions.
- Losing authoritative `record_links` representation behind convenience fields.
- Collapsing Task Requests or Decisions into separate modules.
- Allowing explicit supersession to bypass fail-closed inconsistent-machine validation.
- Letting wire-level null clears reach store-level lifecycle guards for non-clearable owner fields.

Exit criteria:
- Task Request and Decision lifecycle evidence passes directly under Phase 9.
- Queue, owner, blocked-work, due-date, support, and direct-reference semantics are covered.
- No generalized workflow engine or Timeline approval field is introduced.
- Residual Sprint 6 coverage gaps are closed by direct tests, and no specification or implementation behavior change is required.

## Sprint 7. Coordination Surfaces

Objective: Complete workbook-native coordination surfaces for `comm_log`, `handoff`, `status_review`, and `lesson`.

Relevant IDs: `U-9-08`, `U-9-10`, `U-9-12`, `I-9-02`, `E-9-06`, `E-9-GRID-01`; support context from `AC-087..AC-090`, `AC-281..AC-284`, `AC-300..AC-304`, and `AC-395..AC-397`.

Files or areas to inspect:
- `contracts/view-schemas/cartulary.view.comm_log.v1.json`
- `contracts/view-schemas/cartulary.view.handoff.v1.json`
- `contracts/view-schemas/cartulary.view.status_review.v1.json`
- `contracts/view-schemas/cartulary.view.lesson.v1.json`
- `internal/modules/workbook`
- `internal/modules/projections`
- `internal/platform/viewschema`
- `apps/web/src/WorkbookShell.tsx`
- `apps/web/e2e`
- Artifact-backed coordination storage code: `db/migrations/00007_phase4_workbook_surfaces.sql`, `db/migrations/00016_phase9_linked_notes_projection.sql`, `internal/modules/workbook/mutation_store.go`, and `internal/modules/workbook/store.go`.
- Handoff risk-ref child-row storage code: `db/migrations/00007_phase4_workbook_surfaces.sql` table `handoff_risk_refs`, plus `internal/modules/workbook/mutation_store.go` risk-ref upsert/tombstone paths and `internal/modules/workbook/store.go` risk-ref projection expression.

Test-first sequence:
1. Backend store tests cover each coordination surface minimum create signal, omitted defaults, rejected-create no-state behavior, saved-view additivity, and immutable subtype IDs.
2. Backend store tests cover projection-backed sort, semantic filter, grouping, declared query-field mapping, and canonical `handoff.ack_state` token ordering.
3. Backend store tests cover coordination collection fields, `party_ref`, `record_ref`, `risk_ref` validation, item wire shapes, duplicate coalescing, removal by `item_ref`, and source-preserving `comm_log.audience`.
4. Backend integration tests prove all four surfaces persist structured state and query through workbook projections.
5. Browser tests cover Task/Decision and coordination workbook-native flows plus queue, handoff, status review, lesson, blocked-work, next-check, follow-up, and shared-grid behavior where exposed.

Implementation tasks:
- Keep coordination surfaces artifact-backed with distinct `artifact_type` values.
- Enforce minimum create signals and no partial state on rejection.
- Implement collection-review actions and same-incident target validation.
- Preserve `comm_log.audience` text when supplemental party refs change.
- Derive `handoff.ack_state` from `acknowledged_at`.
- Keep `handoff.open_risk_refs` as child `risk_ref` rows scoped to one handoff, not first-class risk records.
- Keep coordination surfaces workbook-native and directly addressable by canonical `view_schema_id`.

Validation commands:
- `make backend-unit`
- `make backend-store`
- `make backend-integration`
- `make browser-e2e-webserver-backed`
- `make phase-slice PHASE=phase9`
- `make service-backed-slice PHASE=phase9`
- `git diff --check`

Deliverables:
- Direct authoritative evidence for `U-9-08`, coordination portions of `U-9-10` and `U-9-12`, relevant `I-9-02` coverage, and coordination portions of `E-9-06` and `E-9-GRID-01`.
- Direct backend store evidence: `internal/modules/workbook/phase9_coordination_surfaces_test.go`.
- Direct timestamp evidence: `internal/modules/workbook/phase9_timestamp_contract_test.go::TestPhase9_TimestampInstantPatchDecoder_U_9_10`.
- Direct relationship-confidence evidence: `internal/modules/assessments/phase9_assessment_contract_test.go::TestPhase9_U_9_12_RelationshipConfidenceRejectedAndManualLinksRemainNull`.
- Direct integration evidence: `internal/modules/workbook/phase9_notes_indicators_test.go::TestPhase9_CoordinationSurfacesQueryThroughWorkbookProjections_I_9_02`.
- Direct browser evidence: `apps/web/e2e/phase9.sentinel.spec.ts::Phase 9 E-9-06 coordination workbook workflows stay native` and `apps/web/e2e/phase9.keyboard.spec.ts::Phase 9 E-9-GRID-01 shared grid keyboard anchors stay stable across workbook cells`.
- Latest audit validation artifact roots:
  - `make backend-unit`: `.cartulary/test-results/20260521T012303Z-p627398`
  - `make backend-store`: `.cartulary/test-results/20260521T012315Z-p628416`
  - `make backend-integration`: `.cartulary/test-results/20260521T012838Z-p638605`
  - `make browser-e2e-webserver-backed`: `.cartulary/test-results/20260521T013406Z-p649711`
  - `make phase-slice PHASE=phase9`: `.cartulary/test-results/20260521T013525Z-p657289`
  - `make service-backed-slice PHASE=phase9`: `.cartulary/test-results/20260521T014114Z-p671006`
  - `git diff --check`: exit `0`

Risks:
- Promoting coordination artifacts to first-class record types contrary to owner text.
- Creating separate modules instead of workbook-native surfaces.
- Letting defaults satisfy minimum create signals.
- Treating risk refs as generic `record_id` values.

Exit criteria:
- Satisfied for Sprint 7 in the current checked-out state: all required coordination surfaces have direct Phase 9 evidence.
- Satisfied for Sprint 7 in the current checked-out state: coordination create/query/edit behavior remains workbook-native and projection-backed.
- Satisfied for Sprint 7 in the current checked-out state: collections, references, risk refs, and timestamp fields obey owner contracts.
- This is not a full Phase 9 completion claim; Sprint 8 optional-surface and registry-closure rows are now directly evidenced, and Sprint 9 final gate remains separately owned.

## Sprint 8. Optional Surfaces and Registry Closure

Objective: Close the artifact-backed variant registry and optional standardized surface handling without inventing omitted behavior or preserving stale 14-only discovery expectations.

Relevant IDs: `U-9-09`, `U-9-13`, `E-9-07`, `E-9-08`; support context from `AC-285..AC-287` and `AC-410`.

Status: audited and direct-evidence satisfied for Sprint 8 in the current checked-out state. The current build intentionally exposes seventeen standardized current-profile workbook surfaces: the fourteen required pack-independent surfaces plus the implemented standardized optional surfaces `cartulary.view.findings.v1`, `cartulary.view.investigative_queries.v1`, and `cartulary.view.forensic_keywords.v1`. `cartulary.view.hypotheses.v1`, `artifact_type='hypothesis'`, and first-class hypothesis records remain excluded; hypotheses remain represented through Findings with `finding.kind='hypothesis'`.

Owner interpretation:
- Core 01 Table 7.4-A and AC-411 define the current-profile workbook-surface identity map as seventeen rows.
- Core 02 §10.4.4A closes the artifact-backed tagged-variant registry to `note`, `comm_log`, `handoff`, `status_review`, `lesson`, and `finding`.
- Core 02 §10.4.6 states the current implementation exposes Findings, Investigative Queries, and Forensic Keywords as standardized optional workbook surfaces with no reference-pack activation requirement.
- Core 03 preserves Notes, coordination surfaces, and optional standardized surfaces as workbook-native `view_schema` surfaces rather than separate modules.

Implementation and evidence owners:
- `contracts/view-schemas/index.json`
- `contracts/view-schemas`
- `internal/platform/viewschema`
- `internal/platform/viewschema/artifact_variants.go`
- `internal/modules/viewschemas/viewschemas_integration_test.go`
- `internal/modules/workbook/optional_surfaces_store_test.go`
- `apps/web/src/WorkbookShell.tsx`
- `apps/web/e2e/phase9.sentinel.spec.ts`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/domain.md`

Direct evidence:
- `U-9-13`: `internal/platform/viewschema/registry_test.go::TestPhase9_U_9_13_ArtifactVariantRegistryClosure`.
- `U-9-09`: `internal/modules/workbook/optional_surfaces_store_test.go::TestPhase9_U_9_09_OptionalStandardizedSurfacesStoreBehavior`, `TestPhase9_U_9_09_OptionalStandardizedSurfacesProjectionQueryBehavior`, and `TestPhase9_U_9_09_FindingsConfidenceBandBoundaries`.
- `E-9-07`: `apps/web/e2e/phase9.sentinel.spec.ts::Phase 9 E-9-07 optional standardized surfaces are workbook-native when exposed`.
- `E-9-08`: `apps/web/e2e/phase9.sentinel.spec.ts::Phase 9 E-9-08 required registry identities stay canonical with optional additions`.

Remediation completed for Sprint 8 handoff:
- Keep the 17-surface runtime and contract model; do not roll back optional standardized surfaces to satisfy stale 14-only expectations.
- Clarify Core 04 AC-116 so it describes the fourteen required pack-independent floor, the three implemented standardized optional artifact-backed surfaces with `required_reference_pack_keys=[]`, and the continued exclusion of pack-dependent framework overlays.
- Align `TestViewSchemasDiscoveryHTTP` with the current-profile standardized registry while preserving auth, public-resource-shape, paging, cursor-binding, singleton, and missing-schema checks.
- Keep Findings on `artifact_type='finding'` plus `finding.kind`; reject a standalone hypothesis surface or durable `artifact_type='hypothesis'`.

Behavioral constraints to preserve:
- Preserve Notes as required built-in sheet `cartulary.view.notes.v1`.
- Preserve required coordination identities as canonical workbook-native `view_schema` identities.
- Prevent optional standardized surfaces from substituting for required base surfaces.
- Findings enforce `artifact_type='finding'`, `finding.kind`, lifecycle state, confidence behavior, and support/contradiction refs.
- Investigative Queries enforce declared structured subtype, minimum create signal, immutable `query_id`, and workbook-native behavior.
- Forensic Keywords enforce declared structured subtype, minimum create signal, match-mode/case-sensitive behavior, and workbook-native behavior.

Validation commands:
- `make agent-finalize`
- `go test ./internal/modules/viewschemas -run TestViewSchemasDiscoveryHTTP -count=1`
- `make backend-integration`
- `make backend-unit`
- `make backend-store`
- `make browser-e2e-webserver-backed`
- `make phase-slice PHASE=phase9`
- `make service-backed-slice PHASE=phase9`
- `make phase-ledger-drift`
- `git diff --check`

Latest Sprint 8 audit validation artifact roots:
- `make backend-unit`: `.cartulary/test-results/20260521T170458Z-p2461060`
- `make backend-store`: `.cartulary/test-results/20260521T170736Z-p2465651`
- `make browser-e2e-webserver-backed`: `.cartulary/test-results/20260521T171304Z-p2477118`
- `make phase-slice PHASE=phase9`: `.cartulary/test-results/20260521T171424Z-p2484709`
- `make service-backed-slice PHASE=phase9`: `.cartulary/test-results/20260521T172011Z-p2498526`
- `make phase-ledger-drift`: `.cartulary/test-results/20260521T172558Z-p2511845`
- `git diff --check`: exit `0`

Prior Sprint 8 remediation validation artifact roots:
- `go test ./internal/modules/viewschemas -run TestViewSchemasDiscoveryHTTP -count=1`: pass; direct diagnostic, no retained root.
- `make agent-finalize`: pass, generated unchanged, retained-run maintenance skipped because `RESULTS_DIR` was unset; retained root `.cartulary/test-results/20260521T161250Z-p2318798`.
- `make backend-integration`: pass; retained root `.cartulary/test-results/20260521T161805Z-p2327271`.
- `make backend-unit`: pass; retained root `.cartulary/test-results/20260521T162336Z-p2338370`.
- `make backend-store`: pass; retained root `.cartulary/test-results/20260521T162344Z-p2339299`.
- `make browser-e2e-webserver-backed`: pass; retained root `.cartulary/test-results/20260521T162908Z-p2350579`.
- `make phase-slice PHASE=phase9`: pass; retained root `.cartulary/test-results/20260521T163035Z-p2358577`.
- `make service-backed-slice PHASE=phase9`: pass; retained root `.cartulary/test-results/20260521T163625Z-p2372379`.
- `make phase-ledger-drift`: pass; retained root `.cartulary/test-results/20260521T164215Z-p2385879`.
- `make check`: pass; retained root `.cartulary/test-results/20260521T164230Z-p2386596`.
- `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260521T164230Z-p2386596`: fail at `browser-e2e-duration-baselines` because the retained `check` run had no observed browser entry timing for `E-9-GRID-01`; retained root `.cartulary/test-results/20260521T164932Z-p2432429`.
- `make phase-schedule-drift`: pass after discarding the failed finalizer's partial unstaged duration/schedule delta; retained root `.cartulary/test-results/20260521T165151Z-p2438318`.
- `git diff --check`: exit `0`.

Risks:
- Reintroducing 14-only assertions in HTTP discovery or conformance text.
- Coupling standardized optional workbook surfaces to reference-pack activation.
- Marking exposed optional surfaces `N/A`.
- Adding `cartulary.view.hypotheses.v1` or `artifact_type='hypothesis'` contrary to current-profile owner text.
- Letting optional surfaces replace Notes or required coordination surfaces.

Exit criteria:
- Satisfied after remediation validation: artifact-backed registry closure is directly evidenced.
- Satisfied after remediation validation: optional standardized surfaces are directly covered as exposed workbook-native surfaces.
- Satisfied after remediation validation: required base registry remains complete and canonical, with optional standardized surfaces additive only.
- Remaining Phase 9 work is Sprint 9 final wrappers, final gate, retained-run maintenance, and handoff.

## Sprint 9. Browser Workflows, Public Wrappers, and Final Gate

Objective: Prove Phase 9 through public wrappers, service-backed slices, browser workflows, drift gates, finalizer, and handoff notes.

Relevant IDs: all Phase 9 rows.

Files or areas to inspect:
- `tools/phase9_test_map.json`
- `docs/testing/phase9_coverage_ledger.md`
- `tools/scheduler_manifest.json`
- `tools/execution_topology_render_index.json`
- `apps/web/e2e`
- Final successful Phase 9 wrapper roots are recorded under "Final Sprint 9 validation evidence" below.
- Final `make check` root is recorded under "Final Sprint 9 validation evidence" below.

Test-first sequence:
1. Remove or convert every placeholder row to direct evidence, or keep only explicit blocker sentinels that prevent false completion claims.
2. Run focused direct commands for changed backend/frontend/browser areas.
3. Run public Phase 9 wrappers.
4. Run generated drift, ledger drift, and schedule drift checks.
5. Run `make agent-finalize` before broader final verification.
6. Run `make check` or record exact non-Phase-9 blockers.

Implementation tasks:
- Audit manifest rows against actual test names and titles.
- Audit support-only tests and `forbidden_id_files`.
- Refresh generated ledgers and schedules through canonical commands.
- Record optional surface decisions.
- Record Phase 10 handoff notes without broadening Phase 9 scope.
- Record final artifact roots and any exact blockers.

Validation commands:
- `make phase-map-check`
- `make phase-test-name-check`
- `make phase-ledger-drift`
- `make phase-schedule-drift`
- `make generate-drift`
- `make migration-drift` if schema, migration, DB query, or migration-owned behavior changes.
- `make phase-slice PHASE=phase9`
- `make service-backed-slice PHASE=phase9`
- `make browser-e2e-webserver-backed`
- `make browser-e2e-visual` when Phase 9 changes visual golden-backed states.
- `make test-fast`
- `make agent-finalize`
- `make check`
- `git diff --check`

Deliverables:
- Final `tools/phase9_test_map.json`.
- Final generated `docs/testing/phase9_coverage_ledger.md`.
- Final generated schedules.
- Final retained run roots for public Phase 9 wrappers and `make check`.
- Handoff notes for Phase 10 and later extension-profile work.
- Exact non-Phase-9 blocker records if any final wrapper or gate fails.

Final Sprint 9 validation evidence, retained on 2026-05-21:
- Phase row audit: `tools/phase9_test_map.json` has 27 expected IDs, 27 authoritative rows, 0 duplicate IDs, 0 non-implemented rows, and 2 support-only Go targets. Support-only targets remain non-owning through `support_go_targets` and `forbidden_id_files`.
- `make phase-map-check`: passed; direct command emitted no retained harness root.
- `make phase-test-name-check`: passed; direct command emitted no retained harness root.
- `make phase-ledgers`: passed; retained root `.cartulary/test-results/20260521T201500Z-p2765649`; generated files changed: none.
- `make phase-schedules`: passed; retained root `.cartulary/test-results/20260521T201503Z-p2765889`; generated files changed: none.
- `make phase-ledger-drift`: passed; retained root `.cartulary/test-results/20260521T201509Z-p2766191`.
- `make phase-schedule-drift`: passed; retained root `.cartulary/test-results/20260521T201512Z-p2766404`.
- `make generate-drift`: passed; retained root `.cartulary/test-results/20260521T201516Z-p2766652`; codegen drift clean.
- `make migration-drift`: not required in final Sprint 9 execution because Sprint 9 made no schema, migration, DB-query, or migration-owned behavior changes. Migration drift remains separate from codegen drift.
- `make phase-slice PHASE=phase9`: passed; retained root `.cartulary/test-results/20260521T201527Z-p2767476`; 5/5 work units, 65 tests, 0 failed.
- `make service-backed-slice PHASE=phase9`: passed; retained root `.cartulary/test-results/20260521T202116Z-p2781359`; 3/3 work units, 48 tests, 0 failed.
- `make browser-e2e-webserver-backed`: passed; retained root `.cartulary/test-results/20260521T202704Z-p2794534`; 55 tests, 0 failed.
- `make browser-e2e-visual`: not required as a standalone Sprint 9 validation because Sprint 9 changed only handoff documentation and did not change visual-golden-backed UI states. The final `make check` gate still ran its visual work unit as part of the aggregate.
- `make test-fast`: passed; retained root `.cartulary/test-results/20260521T202829Z-p2802280`; 2/2 work units, 569 tests, 0 failed.
- `make agent-finalize`: passed; retained root `.cartulary/test-results/20260521T203501Z-p2823548`; outcome `generated=unchanged`, `files=0`, retained-run maintenance skipped because `RESULTS_DIR` was unset.
- `make check`: passed; retained root `.cartulary/test-results/20260521T203513Z-p2824720`; 114/114 work units, 638 tests, 0 failed.
- `git diff --check`: passed; no retained harness root.

Final Sprint 9 decisions:
- Optional standardized surfaces are exposed and directly evidenced, not `N/A`: Findings, Investigative Queries, and Forensic Keywords are present as additive workbook-native system views; `cartulary.view.hypotheses.v1` remains absent. Owner basis: Core 01 sections 7.4 and 19, Core 02 section 10.4.6, and Core 03 sections 16.4 and 19.
- All authoritative Phase 9 rows have direct passing evidence through the final public wrappers; no placeholder row or blocker sentinel remains required to prevent a false completion claim.
- No generated ledger, generated schedule, codegen, lockfile, or generated source changes were produced by the canonical Sprint 9 commands.
- No out-of-Phase-9 blocker was encountered in the final gate.

Phase 10 and later handoff notes:
- Claim-bearing timed or fixture-sensitive publication remains Core 05 work and is not claimed by Phase 9.
- File-based CSV/XLSX import, snapshot/reporting, release publication, export redaction, and reference-pack overlays remain outside Phase 9 unless a later owner spec or phase plan activates them.
- Future retained-run duration maintenance should be done with an explicit successful `RESULTS_DIR=<run root>` and recorded in that later phase or maintenance handoff; the Sprint 9 finalizer intentionally ran without a retained check root per the Sprint 9 sequence.

Risks:
- Treating a passing support-only test as authoritative completion.
- Leaving placeholders in a way that reads as implemented behavior.
- Updating visual goldens without owner-driven UI changes and review.
- Reporting `make check` failure without exact target, artifact root, failing row or test, and out-of-scope rationale.

Exit criteria:
- Every authoritative Phase 9 row has direct passing evidence or an explicit blocker sentinel that prevents completion claims.
- Public Phase 9 wrappers pass.
- Generated, ledger, and schedule drift are clean.
- `make agent-finalize` is recorded.
- `make check` passes or exact out-of-Phase-9 blockers are recorded.
- `git diff --check` passes.

Sprint 9 exit status: satisfied in the live repository on 2026-05-21. Phase 9 authoritative row evidence is direct and passing through the final public wrappers, support-only evidence was not counted as completion, generated and drift checks are clean, `make agent-finalize` is recorded, `make check` passed, and `git diff --check` passed.

## Phase Validation Criteria

Binary completion criteria:
- Phase 9 is present in `tools/phase_registry.json` and is selectable only after row owners exist.
- `tools/phase9_test_map.json` covers every authoritative Phase 9 guide row exactly once.
- `docs/testing/phase9_coverage_ledger.md` is generated from the manifest and is not hand-edited.
- Generated schedule artifacts are refreshed and drift-checked through canonical commands.
- All placeholders are replaced with direct row evidence, or any remaining sentinel explicitly blocks Phase 9 completion.
- `make phase-map-check` passes.
- `make phase-test-name-check` passes.
- `make phase-ledger-drift` passes.
- `make phase-schedule-drift` passes.
- `make generate-drift` passes.
- `make migration-drift` passes when Phase 9 changes schema or migration-owned behavior.
- `make phase-slice PHASE=phase9` passes.
- `make service-backed-slice PHASE=phase9` passes, unless the manifest intentionally declares no service-backed work and the public wrapper reports an explicit no-op.
- Browser-functional Phase 9 rows pass through the public browser workflow target selected by the Phase 9 manifest.
- Optional standardized surfaces are directly evidenced when exposed or marked `N/A` with owner-section justification when omitted.
- Support-only evidence is never counted as authoritative Phase 9 completion.
- `make agent-finalize` runs before final broader verification and its outcome is recorded.
- `make check` passes, or any failure outside Phase 9 is recorded with exact target, artifact root, failing row or test, and why it is outside Phase 9.
- `git diff --check` passes.

## Phase Exit Criteria

Required final commands:
- `make phase-map-check`
- `make phase-test-name-check`
- `make phase-ledger-drift`
- `make phase-schedule-drift`
- `make generate-drift`
- `make migration-drift` if schema or migration-owned behavior changed.
- `make phase-slice PHASE=phase9`
- `make service-backed-slice PHASE=phase9`
- `make browser-e2e-webserver-backed`
- `make test-fast`
- `make agent-finalize`
- `make check`
- `git diff --check`

Handoff requirements:
- Record the final `make phase-slice PHASE=phase9` artifact root.
- Record the final `make service-backed-slice PHASE=phase9` artifact root or explicit no-op summary.
- Record the final browser workflow artifact root for Phase 9 browser rows.
- Record the final `make check` artifact root.
- Record whether `make agent-finalize` ran unchanged, updated generated artifacts, skipped retained-run maintenance because `RESULTS_DIR` was unset, failed at a subtarget, or was explicitly skipped.
- Record optional standardized surface status as exposed and evidenced, or omitted and marked `N/A`.
- Record all generated files changed by canonical generation.
- Record any migration drift resolution separately from codegen drift resolution.
- Record Phase 10 handoff notes without claiming Phase 10 behavior.

Exact blocker recording:
- If `make check` fails, record the exact failing target, artifact root, failing row or test, failure class if available, and why it is or is not Phase 9-owned.
- If `make phase-slice PHASE=phase9` fails, record the exact row, command, artifact root, and whether the failure is product behavior, harness setup, service setup, or manifest selection.
- If `make service-backed-slice PHASE=phase9` fails, record the exact scheduler work unit, service suite root, row or test, and service state.
- If browser workflows fail, record the exact spec/test title, Playwright artifact root, browser stack root, and whether the failure is a Phase 9 behavior gap or an operational setup failure.
- If optional standardized surfaces are omitted, record the exact owner-section basis for `N/A`.

Phase 9 must not be described as implemented until the live repository has direct passing evidence for the relevant authoritative Phase 9 rows.
