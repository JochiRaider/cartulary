# Phase 7 Implementation Plan

## Summary

This file is the execution roadmap and progress marker for Cartulary Phase 7: reviewer-facing history, delete or restore, and rollback.

`docs/guides/cartulary_implementation_testing_guide.md` section 9.7 is the controlling implementation-scope reference for this phase. Normative behavior remains owned by the core documents, especially Core 01 section 3.3.4.2, Core 01 section 3.3.5, Core 02 sections 14 and 15, Core 03 section 10, and Core 04 sections 2 and 3.

This planning artifact does not implement Phase 7 behavior. It is intentionally root-level so agents can find it quickly during handoff or interrupted implementation sessions. No README update is required for discoverability.

Current repo status after Sprint 6 visual remediation: Phase 7 is listed as `active` in `tools/phase_registry.json`; `GET /api/v1/records/{record_id}/history`, `DELETE /api/v1/records/{record_id}`, `POST /api/v1/records/{record_id}/restore`, and `POST /api/v1/records/{record_id}/rollback` are implemented and registered; every authoritative Phase 7 row in `tools/phase7_test_map.json` has direct passing evidence; Phase 7 public wrappers pass; generated ledger, generated schedules, generated drift checks, `browser-e2e-visual`, and `make check` are clean. Non-Phase 7 Phase 3, Phase 5, and Phase 6 visual blockers from `.cartulary/test-results/20260512T001400Z-p61609` were closed by Timeline action-cell layout remediation, deterministic Phase 6 visual row ordering, and canonical visual golden refresh. Owner decision: merge-specific rollback, typed tag rollback, Phase 8 saved-view behavior, and Phase 9 keyboard/clipboard behavior remain explicit non-claims until their owning scope exists and passes direct evidence.

## Phase Objective

Phase 7 completes the reviewer-visible history and destructive-operation surface over the mutation substrate introduced by earlier phases. By phase exit, a maintainer must be able to open row-centric history, inspect attributed changes, soft-delete and restore first-class records with tombstone concurrency, roll back legal history targets without rewriting prior history, and prove retained history stays stable across delete or restore cycles, rollback, incident closure, and service restart.

The phase does not introduce mutation history from scratch. Existing mutation paths already need to emit attributed change sets, mutation entries, row revisions, projections, idempotency results, optimistic concurrency outcomes, and collaboration events. Phase 7 makes that substrate reviewer-facing and fills the delete, restore, rollback, retained-history, and destructive-lock gaps.

## Implementation Scope

In scope:
- Record-scoped history read: `GET /api/v1/records/{record_id}/history`.
- First-class record soft-delete and restore: `DELETE /api/v1/records/{record_id}` and `POST /api/v1/records/{record_id}/restore`.
- Reviewer rollback: `POST /api/v1/records/{record_id}/rollback` with `target.kind` values `history_entry`, `change_set`, and `row_restore`.
- Stable opaque `history_entry_ref` values for single-entry-addressable logical history items.
- Canonical `available_rollback_actions[]` ordering: `history_entry`, `change_set`, `row_restore`.
- Append-only `change_sets`, `change_set_mutations`, and `record_revisions` consequences for delete, restore, and rollback.
- Destructive-operation lock precedence for restore, rollback, and merge through a shared protected-set helper.
- Workbook UI affordances for reviewer history, legal rollback actions, soft-delete, restore, and whole-row restore.
- Phase 7 executable manifest, ledgers, generated schedules, focused tests, public phase slices, and handoff notes.

Out of scope unless an owner decision pulls it forward:
- Typed tag creation or tag mutation semantics beyond the rollback evidence needed for existing implemented tag targets.
- New saved-view behavior, sorting, filtering, grouping, startup-surface selection, and snapshot-stable view-query pagination; those remain Phase 8.
- Remaining workbook-native surfaces and the full keyboard/clipboard contract; those remain Phase 9 except for the minimal history access needed by `E-7-01`.
- Any public history-purge route, operator history-retention horizon, client-visible lock acquisition route, manual unlock route, or destructive-operation queue.
- Incident portability selector preservation across deployments. Phase 7 owns retained history inside the current deployment only.

## Sprint Checklist

| Done | Sprint | Validation | Blockers | Follow-up Notes |
| --- | --- | --- | --- | --- |
| [x] | 0. Phase 7 ownership manifest and harness setup | [x] audit-complete | None for manifest creation; Phase 7 was activated during Sprint 1 remediation. | Sprint 6 confirmed no Phase 7 sentinel or placeholder tests remain in the authoritative row inventory. |
| [x] | 1. Record-history read contract | [x] remediated | `AC-184` and `AC-185` are Phase 8-owned support references and are not claimed as completed Phase 7 evidence. | History route is record-scoped, not view-scoped; adjunct target-family completeness is deferred to later Phase 7 work. |
| [x] | 2. Soft-delete and restore | [x] audit-complete | None for Sprint 2 product behavior. The first-class record adapter matrix is covered for current `records.record_type` values. | Ordinary delete remains on optimistic concurrency and does not take destructive-operation locks; restore now uses the shared Sprint 4 destructive-operation lock helper for its one-record protected set. |
| [x] | 3. Rollback request, single-entry reversal, and change-set reversal | [x] audit-complete | None for Sprint 3 product behavior. Owner decision: tag rollback remains Phase 8-owned and merge-specific rollback remains unclaimed until reversible merge substrate and tests exist. | Source is `rollback`; previous history rows are never mutated in place. Sprint 4 extends executable action advertisement with `row_restore` when legal. |
| [x] | 4. Whole-row restore, retained history, and locks | [x] backend evidence complete | `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260511T193520Z-sprint4-service-backed-slice` failed in duration-baseline maintenance because that slice has no retained browser timing entries for earlier browser rows; rerun without `RESULTS_DIR` passed and skipped duration refresh. | Whole-row restore is backend-only Sprint 4 scope; Sprint 5 reviewer workbook UI/browser workflow remains unclaimed. |
| [x] | 5. Reviewer workbook UI and browser evidence | [x] implemented | `make browser-e2e-webserver-backed` first failed in `E-3-03` because the added History action exceeded the fixed Timeline action-cell height; compacting Inspect and History into one row remediated the overlap and the rerun passed. | Browser evidence covers reviewer flows over server-provided selectors and actions; typed tag rollback, merge-specific rollback, Phase 8 saved views, Phase 9 keyboard/clipboard, and full Phase 7 exit remain non-claims. |
| [x] | 6. Phase gate, ledgers, schedules, baselines, and handoff cleanup | [x] evidence-backed exit and visual remediation complete | Non-Phase 7 `browser-e2e-visual` rows `V-3-GRID-03`, `V-5-GRID-02`, `V-6-GRID-02`, and `V-6-GRID-03` failed in `.cartulary/test-results/20260512T001400Z-p61609`, then passed after remediation. | Phase 7 public wrappers, `test-fast`, `agent-finalize`, generated drift, ledger drift, schedule drift, `browser-e2e-visual`, `git diff --check`, and `make check` passed after Sprint 6 cleanup. |

## Global References

- Controlling guide: `docs/guides/cartulary_implementation_testing_guide.md`, `Phase 7`, `U-7-01..U-7-07`, `I-7-01..I-7-04`, `E-7-01..E-7-04`.
- Phase-owned ACs: `AC-007`, `AC-010..AC-012`, `AC-181..AC-183`, `AC-187`, `AC-215..AC-218`, `AC-353`, `AC-383..AC-385`.
- Ambiguous support/shared ACs: `AC-184` and `AC-185` appear in the Phase 7 `U-7-01` guide row, but the guide coverage index assigns them to Phase 8. Phase 7 treats them as support references only and does not claim them as completed Phase 7 evidence.
- Primary REQs: `REQ-01-048..REQ-01-056`, `REQ-01-071..REQ-01-082`, `REQ-01-089..REQ-01-111`, `REQ-01-561..REQ-01-563`, `REQ-02-205..REQ-02-220`, `REQ-02-238..REQ-02-242`, `REQ-03-101`, `REQ-03-138..REQ-03-144`, `REQ-03-261..REQ-03-262`, `REQ-04-021..REQ-04-024`, `REQ-04-036..REQ-04-039`.
- Existing mutation substrate: `change_sets`, `change_set_mutations`, `record_revisions`, `records`, first-class source tables, route-scoped idempotency records, projection refresh paths, and `platformws.Hub.PublishRecordChange`.
- Existing route roots to extend or coordinate: `internal/app/runtime.go` registers module routes; `internal/modules/workbook` owns generic record patch and conflict resolution; `internal/modules/timeline` owns Timeline reviewer actions; `internal/modules/entities` owns merge; `internal/modules/revisions` currently owns reusable change-set insertion helpers and should become the Phase 7 route/service home unless implementation inspection proves a narrower owner.
- Existing frontend areas to extend: `apps/web/src/WorkbookShell.tsx`, workbook surface tests in `apps/web/src`, and browser specifications in `apps/web/e2e`.
- Generated boundaries: do not hand-edit `internal/gen/**` or `packages/protocol-ts/src/generated/**`; if generated code changes are required, edit source contracts or SQL and run the canonical generation target.

## Dependencies and Prerequisites

- Phase 3 mutation, history substrate, Timeline lifecycle, idempotency, row-version, projection, and collaboration behavior must remain valid support evidence.
- Phase 4 mention, merge, entity, relationship, and implemented workbook-surface mutation targets must preserve enough reversible mutation-entry detail for Phase 7 history and rollback.
- Phase 5 evidence attach and evidence association history must provide target-level before/after data for single-entry rollback where the owner sections require it.
- Phase 6 collaboration replay and `record_changed` semantics must be reused; Phase 7 must emit ordinary replayable record-change events, not a special rollback event family.
- The first-class record dispatch registry must use stable `record_id` and `record_type` authority from the record envelope rather than visible sheet labels, projection table names, or React component names.
- Public contract changes must land through owner-driven contract artifacts and generated outputs, with drift resolved by the canonical command surface.

## Public Interfaces and Deliverables

Expected public route additions:
- `GET /api/v1/records/{record_id}/history`
- `DELETE /api/v1/records/{record_id}`
- `POST /api/v1/records/{record_id}/restore`
- `POST /api/v1/records/{record_id}/rollback`

Expected schema additions:
- History envelope with `incident_id`, `record_id`, current `row_version`, `deleted`, `items[]`, and `meta.paging`.
- History item with `actor_user_id`, `committed_at`, `operation`, `diff_summary`, `change_set_id`, `reversible`, `available_rollback_actions[]`, optional `history_entry_ref`, and optional `revision_no`.
- Delete and restore request bodies with `base_row_version`, `client_txn_id`, and optional normalized `reason`.
- Delete and restore summaries with `record_id`, `incident_id`, `row_version`, `deleted`, `deleted_at`, `deleted_by_user_id`, and `change_set_id`.
- Rollback request body with `base_row_version`, `client_txn_id`, optional normalized `reason`, and a closed `target` union for `history_entry`, `change_set`, and `row_restore`.
- Rollback summary with `incident_id`, `record_id`, `row_version`, echoed `target`, `target_change_set_id` when known, `rollback_change_set_id`, and canonical ascending `affected_record_ids[]`.
- Error contract coverage for `record_deleted_use_restore`, `record_already_deleted`, `record_not_deleted`, `invalid_rollback_request`, `rollback_target_not_found`, `rollback_precondition_failed`, and existing `record_locked` retryability.

Expected persistence deliverables:
- Stable retained-history materialization from immutable `change_sets`, mutation entries, and row revisions.
- Stable opaque `history_entry_ref` strategy. Default implementation choice: persist a public opaque ref or durable mapping for eligible mutation entries; do not derive refs from transient row order, database primary keys exposed in reversible form, signing-key rotation, or non-durable in-memory state.
- Append-only delete, restore, rollback, and whole-row restore history. No prior `change_sets`, `change_set_mutations`, or `record_revisions` rows are rewritten or removed.
- Shared destructive-operation locking over the `records` envelope using canonical ascending `record_id` order and fail-fast `FOR UPDATE NOWAIT` semantics or an equivalent transaction-scoped Postgres mechanism with identical observable behavior.

Expected test and planning artifacts:
- `tools/phase7_test_map.json`
- `docs/testing/phase7_coverage_ledger.md`
- Phase 7 backend unit, backend integration/store, frontend unit, and browser-functional tests.
- Generated schedule updates from `make phase-schedules`.
- This plan updated with actual artifact roots and blocker status as implementation progresses.

## Sprint 0. Phase 7 Ownership Manifest and Harness Setup

Objective: Establish Phase 7 test ownership before feature work so TDD rows can be selected by repo tooling.

Status: Audit-complete. Phase 7 was activated during Sprint 1 remediation so public phase wrappers can execute the manifest. Sprint 0 established the manifest, generated ledger, generated schedule checks, and initial selectable row symbols. Sprint 6 replaced the initial selectable placeholder coverage with real assertions before exit.

Relevant IDs: all `U-7-*`, `I-7-*`, `E-7-*`; `tools/phase7_test_map.json`; `docs/testing/phase7_coverage_ledger.md`.

Files and areas:
- `tools/phase_registry.json`
- `tools/phase7_test_map.json`
- `docs/testing/phase7_coverage_ledger.md`
- `tools/service_backed_schedule_manifest.json`
- `tools/check_schedule_manifest.json`
- Backend tests under the Phase 7 owning modules, expected to center on `internal/modules/revisions` plus integration tests that exercise `workbook`, `timeline`, `entities`, `evidence`, and `collaboration` paths as support.
- Frontend tests under `apps/web/src`.
- Browser tests under `apps/web/e2e/phase7.history.spec.ts` or the locally chosen equivalent.

Test-first sequence:
1. Done: manifest rows exist for every `U-7-01..U-7-07`, `I-7-01..I-7-04`, and `E-7-01..E-7-04` guide row before feature implementation.
2. Done for Sprint 0 and closed by Sprint 6: initial non-behavioral row-selection coverage made Phase 7 selectable; Sprint 6 confirmed those placeholders were replaced by real assertions before exit.
3. Done: support-only carryover files are declared as `forbidden_id_files` so earlier phase evidence cannot accidentally claim Phase 7 IDs.
4. Done: the Phase 7 coverage ledger and schedules were regenerated after the manifest was added.
5. Done during Sprint 1 remediation: Phase 7 is active so public phase wrappers execute Phase 7 rows. Sprint 5 implemented the UI/browser rows, and Sprint 6 confirmed they are real browser-functional assertions rather than scope sentinels.

Implementation tasks:
- Done: `tools/phase7_test_map.json` has non-empty `claim` and `out_of_scope` text for every authoritative row.
- Done: every row declares its execution dependency explicitly. Authoritative `U-7-*` rows use `backend_store`, `I-7-*` rows use `backend_integration`, and `E-7-*` rows use `browser_functional`; the frontend unit row is supplemental only.
- Done: Phase 3, Phase 4, Phase 5, and Phase 6 support boundaries are recorded in manifest notes rather than re-owning their existing IDs.
- Done: test stubs exist in the final expected file locations so `make explain-phase PHASE=phase7` can discover symbols and row titles.
- Done: generated ledgers and schedules were produced by canonical commands; do not hand-edit generated outputs.

Validation commands:
- `make phase-map-check`
- `make explain-phase PHASE=phase7`
- `make phase-ledgers`
- `make phase-ledger-drift`
- `make phase-schedules`
- `make phase-schedule-drift`
- `make target-plan-json`
- `make phase-test-name-check`
- `git diff --check`

Deliverables:
- Complete: `tools/phase7_test_map.json`.
- Complete: `docs/testing/phase7_coverage_ledger.md`.
- Complete: generated service-backed and check schedule verification.
- Complete: selectable Phase 7 backend symbols, supplemental frontend row, and browser rows.

Risks and open questions:
- Phase activation changes task-surface behavior. Sprint 1 remediation intentionally activated Phase 7; Sprint 6 evidence now treats the Sprint 5 browser rows only as browser-functional reviewer workflow evidence, not as backend delete/restore/rollback substitutes.
- Newly declared service-backed rows require explicit duration baselines after successful uncontaminated evidence; missing baselines must not be hidden by fallback weights.
- `AC-184` and `AC-185` remain recorded as ambiguous Phase 8-assigned support references, not completed Phase 7 claims.
- Typed tag rollback remains a Phase 8 owner dependency; Sprint 6 records current record-tag history entries as visible but not reversible and does not claim Phase 8 typed tag creation or mutation behavior.
- Closed by Sprint 6: no placeholder tests remain in the authoritative Phase 7 manifest inventory.

Exit criteria:
- Met: `make explain-phase PHASE=phase7` reports the manifest, ledger path, execution dependencies, service requirements, and target coverage.
- Met: `make phase-ledger-drift` passes after ledger generation.
- Met: `make phase-schedule-drift` passes after schedule generation.
- Met: Phase 7 row IDs appear only in authoritative Phase 7 files or approved manifest support references.
- Met: `make phase-map-check`, `make phase-ledgers`, `make phase-schedules`, `make target-plan-json`, `make phase-test-name-check`, and `git diff --check` passed during the Sprint 0 audit.

## Sprint 1. Record-History Read Contract

Objective: Implement `GET /api/v1/records/{record_id}/history` as a record-scoped retained-history read route with deterministic ordering, pagination, tombstone concurrency, rollback metadata, and stable single-entry selectors.

Status: Remediated for Sprint 1 audit gaps. The read route is implemented, Phase 7 is active, and Sprint 1 remains scoped to `GET /api/v1/records/{record_id}/history`.

Relevant IDs:
- `U-7-01`, `U-7-02`, `U-7-07`, `I-7-02`, `I-7-04`
- Browser follow-up: `E-7-01`
- `REQ-01-048..REQ-01-056`, `REQ-01-561..REQ-01-563`
- `REQ-02-205..REQ-02-218`, `REQ-02-238..REQ-02-242`
- `REQ-03-138..REQ-03-140`
- `AC-007`, `AC-215`, `AC-216`, `AC-383..AC-385`
- Ambiguous guide references: `AC-184`, `AC-185`

Grep references:
- `change_sets`
- `change_set_mutations`
- `record_revisions`
- `records`
- `Pagination`
- `history_entry_ref`
- `available_rollback_actions`
- `record_revisions_record_lookup_idx`

Files and areas:
- `contracts/openapi/cartulary.openapi.yaml`
- `contracts/errors/index.json`
- `internal/modules/revisions`
- `internal/app/runtime.go`
- `internal/platform/pagination`
- Source-record adapter registry for current first-class `record_type` values.
- Focused tests under `internal/modules/revisions` and integration tests that seed existing mutation paths.

Test-first sequence:
1. `U-7-01` asserts the response envelope includes `incident_id`, `record_id`, current `row_version`, `deleted`, newest-first `items[]`, current tombstone row version for deleted rows, deterministic same-change-set ordering, and canonical `available_rollback_actions[]`.
2. `U-7-02` asserts `history_entry_ref` is present only for exactly one reversible mutation target, is opaque, survives repeated reads, and is not a direct storage primary key.
3. `U-7-07` asserts retained history remains visible and fully paginatable after delete or restore cycles, rollback, incident closure, and service restart.
4. `I-7-02` asserts pagination is bound to `record_id`, cursor replay against a different record fails closed, invalid aliases fail with `invalid_pagination_request`, and item order is preserved across pages.
5. `I-7-04` asserts previously issued `history_entry_ref` values for older single-entry-addressable items survive restart and later state changes.

Implementation tasks:
- Add the history route to the Phase 7 route owner and register it from `internal/app/runtime.go`.
- Implement authorization as record visibility: caller without visibility receives `404`; visible caller may read history according to incident membership.
- Resolve record metadata from the authoritative `records` envelope, then load source-specific current row state only as needed for `row_version` and `deleted`.
- Materialize logical row-centric history from `change_sets`, `change_set_mutations`, and `record_revisions`; do not read projection tables as authoritative history.
- Serialize `items[]` newest-first by committed change-set order, then a deterministic logical item order within the same change set.
- Use `available_rollback_actions[]` only from `history_entry`, `change_set`, and `row_restore`, serialized in that order with unavailable actions omitted.
- Expose `revision_no` only when whole-row restore is legal for that logical item.
- Implement `history_entry_ref` as a durable opaque selector. Default: persist a generated selector or selector mapping for eligible mutation entries when first exposed, then reuse it for the retained-history lifetime.
- Keep items visible when they later become non-reversible; express current legality through `reversible=false` and `available_rollback_actions[]=[]`.
- Use the existing pagination registry with a record-history scope binding that includes `record_id` and rejects query aliases such as `page`, `offset`, `page_size`, and `block_size`.
- Declare the OpenAPI path parameter `record_id` and use a history-specific response meta schema that requires `meta.paging`.
- Add direct route evidence for unauthenticated history reads and invalid `limit` values `-1`, non-integer, and greater than `500`.

Validation commands:
- `go test ./internal/modules/revisions -run 'TestPhase7_.*(U_7_01|U_7_02|U_7_07)'`
- `go test ./internal/modules/revisions ./internal/modules/workbook ./internal/modules/timeline -run 'TestPhase7_.*(I_7_02|I_7_04)'`
- `make phase-slice PHASE=phase7`
- `make service-backed-slice PHASE=phase7`
- `git diff --check`

Deliverables:
- Complete: OpenAPI history route and response schemas, including required `record_id` path parameter and required history `meta.paging`.
- Complete: history query store/service.
- Complete: durable `history_entry_ref` selector strategy.
- Complete: unit and integration coverage for `U-7-01`, `U-7-02`, `U-7-07`, `I-7-02`, and history portions of `I-7-04`.

Risks and open questions:
- `AC-184` and `AC-185` are listed on `U-7-01` but assigned to Phase 8 in the coverage index. Sprint 1 records them as support references only and does not claim them as completed Phase 7 evidence.
- Fixture-seeded delete, restore, and rollback-like state changes are accepted only for Sprint 1 read-route retained-history invariants. Sprint 2 implemented the real delete/restore route behavior, Sprint 3 implemented executable rollback for reversible non-merge substrate, and Sprint 4 implemented executable whole-row restore.
- Adjunct mutation-target-family history completeness is deferred to later Phase 7 rollback/UI work; Sprint 1 claims row-backed history read evidence only.
- A transient `history_entry_ref` derived from result order or in-memory state would fail retained-history and restart requirements.

Exit criteria:
- History reads are stable across repeated reads, pagination, restart, delete/restore, rollback, and incident closure.
- Soft-deleted records expose the tombstone `row_version` needed by restore.
- The route does not require or accept `incident_id`, `view_schema_id`, or display labels as behavior inputs.

## Sprint 2. Soft-Delete and Restore

Objective: Implement first-class record soft-delete and restore with role gates, route-scoped idempotency, row-version preconditions, append-only history, projection invalidation, and collaboration events.

Status: Audit-complete. The Sprint 2 audit rerun on 2026-05-11 found no product blockers or new evidence gaps; see `docs/testing/phase7_sprint2_soft_delete_restore_audit.md`.

Relevant IDs:
- `U-7-03`, `U-7-04`, `I-7-01`, `I-7-03`, `E-7-03`
- `REQ-01-071..REQ-01-082`
- `REQ-04-021..REQ-04-024`
- `AC-181..AC-183`, `AC-215`, `AC-216`, `AC-218`

Grep references:
- `deleted_at`
- `deleted_by_user_id`
- `row_version_conflict`
- `record_deleted_use_restore`
- `record_already_deleted`
- `record_not_deleted`
- `PublishRecordChange`
- `change_kind = remove`
- `change_kind = invalidate`

Files and areas:
- `contracts/openapi/cartulary.openapi.yaml`
- `contracts/errors/index.json`
- `internal/modules/revisions`
- `internal/modules/workbook`
- `internal/modules/timeline`
- `internal/modules/entities`
- `internal/modules/evidence`
- Projection refresh code for all currently implemented source-record adapters.
- Collaboration fixtures in `internal/modules/collaboration` or existing WebSocket test helpers.

Completed test sequence:
1. Complete: `U-7-03` asserts delete request shape, role gates, stale row-version failure, idempotent replay, divergent `client_txn_id` reuse conflict, already-deleted failure, patch-after-delete `record_deleted_use_restore`, and adapter matrix coverage.
2. Complete: `U-7-04` asserts restore request shape, reviewer/admin role gate, tombstone `row_version` requirement, not-deleted failure, idempotent replay, destructive-lock precedence, and append-only prior history preservation.
3. Complete: `I-7-01` asserts delete and restore atomically update source rows, record envelope, projections, history rows, and emitted collaboration events.
4. Complete: `I-7-03` asserts stale restore fails closed and does not mutate current row state.
5. Support fixture complete: `E-7-03` has HTTP and WebSocket support coverage for browser soft-delete/restore validation. Sprint 5 now owns the completed reviewer workbook UI evidence for this row.

Implemented behavior:
- Delete and restore route contracts are present. Both accept `base_row_version`, `client_txn_id`, and optional normalized `reason`.
- Delete is routed through ordinary optimistic concurrency. Ordinary soft-delete remains outside the destructive-operation lock family.
- Restore uses the shared Sprint 4 transaction-scoped `FOR UPDATE NOWAIT` helper for the one-record protected set required by the current profile.
- Delete roles are `editor`, `reviewer`, or `admin`. Restore roles are `reviewer` or `admin`. Visible underprivileged callers receive `403`; callers without visibility receive `404`.
- Route-scoped idempotency uses `(actor_user_id, record_id, client_txn_id)` and compares normalized reason, with omission, explicit `null`, and normalized-empty reason equivalent.
- Delete sets envelope soft-delete state and source tombstone state where applicable, increments row version, creates one attributed `change_set`, appends reversible mutation entries, appends a `record_revisions` row with `operation = soft_delete`, and removes the row from ordinary view queries.
- Restore clears only current soft-delete state, increments row version, creates one attributed `change_set`, appends reversible mutation entries, appends a `record_revisions` row with `operation = restore`, and makes the row eligible for ordinary view queries again.
- Delete and restore preserve prior history in place. They do not hard-delete revisions, change sets, blobs, source rows, or link rows.
- Delete emits ordinary replayable `record_changed` events with `change_kind='remove'`; restore emits ordinary replayable `record_changed` events with `change_kind='invalidate'`.
- Projection refresh/invalidation covers current implemented first-class record adapters.

Validation commands:
- Passed: `make task-guide ROLE=feature-dev PHASE=phase7`.
- Passed: `make explain-phase PHASE=phase7`.
- Passed: `make explain-target TARGET=backend-store DETAIL=rows`.
- Passed: `make explain-target TARGET=backend-integration DETAIL=rows`.
- Passed: `go test ./internal/modules/revisions -run 'TestPhase7_.*(U_7_03|U_7_04)'`.
- Passed: `go test ./internal/modules/revisions ./internal/modules/workbook ./internal/modules/collaboration -run 'TestPhase7_.*(I_7_01|I_7_03)'`.
- Passed: `make phase-slice PHASE=phase7`; run root `.cartulary/test-results/20260511T125928Z-p88290`.
- Passed: `make service-backed-slice PHASE=phase7`; run root `.cartulary/test-results/20260511T125946Z-p90925`.
- Passed: `make test-fast`; run root `.cartulary/test-results/20260511T125958Z-p93527`.
- Passed: `make lint`; run root `.cartulary/test-results/20260511T130054Z-p2340`.
- Passed: `make generate-drift`; run root `.cartulary/test-results/20260511T130140Z-p4916`.
- Passed: `make phase-ledger-drift`; run root `.cartulary/test-results/20260511T130144Z-p5530`.
- Passed: `make phase-schedule-drift`; run root `.cartulary/test-results/20260511T130147Z-p5774`.
- Passed: `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260511T124218Z-p46750`; run root `.cartulary/test-results/20260511T130202Z-p6101`.
- Passed: `git diff --check`.

Deliverables:
- Complete: delete and restore OpenAPI contracts.
- Complete: route-level decoder and idempotency hashing.
- Complete: source-record dispatch registry for all current first-class record types.
- Complete: projection and collaboration invalidation for delete/restore.
- Complete: unit and integration coverage for `U-7-03`, `U-7-04`, delete/restore parts of `I-7-01`, and restore part of `I-7-03`.
- Complete: Sprint 2 audit rerun recorded in `docs/testing/phase7_sprint2_soft_delete_restore_audit.md`.
- Complete: `agent-finalize` refreshed duration baselines and generated scheduler render artifacts from retained full `check` run `.cartulary/test-results/20260511T124218Z-p46750`.

Risks and open questions:
- Closed for Sprint 2: first-class delete/restore dispatch is explicit for current `records.record_type` values. Rollback adapter completeness remains a later Phase 7 concern.
- Closed for Sprint 2: delete/restore route behavior is limited to first-class record envelopes and is not reused for individual links, tags, mentions, observations, or evidence associations.
- Follow-up resolved in Sprint 4: restore, rollback, and merge use the shared destructive-operation helper over the `records` envelope.

Exit criteria:
- Met: delete and restore preserve append-only history and source attribution.
- Met: ordinary queries hide soft-deleted rows and include restored rows only through current view contracts.
- Met: collaboration subscribers observe ordinary `record_changed` messages, not a Phase 7-specific event family.
- Met: Sprint 2 validation and drift checks passed, with generated duration/scheduler maintenance artifacts updated by `agent-finalize`.

## Sprint 3. Rollback Request, Single-Entry Reversal, and Change-Set Reversal

Objective: Implement rollback route admission, single-entry rollback for history items that map to exactly one reversible mutation target, and whole-`change_set` rollback that reverses every currently reversible mutation entry in reverse deterministic entry order.

Status: Audit-complete for Sprint 3 backend scope. This claim covers rollback route admission, single-entry rollback for currently reversible `history_entry` targets, and whole-`change_set` rollback for currently reversible mutation entries. It explicitly does not claim `row_restore`, tag rollback, merge-specific rollback, retained-history restart or closure invariants, merge-lock integration, reviewer workbook UI, or browser reviewer workflows.

Relevant IDs:
- `U-7-05`, `U-7-06`, `I-7-01`, `I-7-03`, `E-7-02`
- `REQ-01-089..REQ-01-111`
- `REQ-02-212..REQ-02-218`
- `REQ-03-141..REQ-03-144`
- `REQ-04-036..REQ-04-039`
- `AC-010`, `AC-216`, `AC-217` for the `change_set` rollback portion only, `AC-218`, `AC-353`

Grep references:
- `invalid_rollback_request`
- `rollback_target_not_found`
- `rollback_precondition_failed`
- `target_not_reversible`
- `entry_requires_change_set`
- `dependent_later_changes`
- `stale_target`
- `source = 'rollback'`
- `history_entry_ref`
- `target.kind = 'change_set'`

Files and areas:
- `contracts/openapi/cartulary.openapi.yaml`
- `contracts/errors/index.json`
- `internal/modules/revisions`
- `internal/modules/entities`
- `internal/modules/evidence`
- `internal/modules/timeline`
- Existing mutation-entry helpers and target-specific inverse operations.

Test-first sequence:
1. `U-7-05` asserts rollback request shape accepts only `history_entry`, `change_set`, and `row_restore`; rejects unknown top-level members, unknown target members, missing selectors, and wrong selector JSON types with `invalid_rollback_request`.
2. `U-7-05` asserts successful rollback creates a new `change_set` with `source='rollback'` and does not mutate prior history.
3. `U-7-05` asserts `change_set` rollback reverses every currently reversible mutation entry in reverse deterministic entry order and commits one new rollback change set.
4. `U-7-06` asserts lock failure wins before stale-precondition evaluation for rollback when a protected-set lock is held.
5. `I-7-01` asserts single-entry and whole-change-set rollback atomically update source rows, projections, history rows, and collaboration events.
6. `I-7-03` asserts stale single-entry and whole-change-set rollback fail closed and leave current row state unchanged.
7. `E-7-02` later proves a reviewer can roll back one mistaken link, mention resolution, or evidence association without reverting later unrelated edits on the same row.

Implementation tasks:
- Add rollback route contract and decoder with a closed `target` union.
- Require current incident role `reviewer` or `admin`; caller without visibility receives `404`, visible underprivileged caller receives `403`.
- Reject rollback against currently soft-deleted records with `record_deleted_use_restore`.
- Resolve `history_entry_ref` only through the history selector mapping; do not accept storage primary keys or client-derived mutation-entry IDs.
- Compute the protected record set for the selected rollback target before evaluating stale row-version or target preconditions, then acquire locks through the shared destructive-operation helper.
- Use idempotency key `(actor_user_id, record_id, client_txn_id)` and normalized request comparison, including target and normalized reason.
- Implement single-entry inverse operations for mutation targets already implemented and carrying reversible before/after data. At minimum, cover row-field edits, current record-link mutations needed by existing supersede/evidence flows, existing mention resolution/dismiss/restore mutations, and existing evidence association mutations when the mutation substrate can prove a safe inverse.
- Implement `target.kind='change_set'` rollback by loading the target change set through visible record history and reversing every currently reversible mutation entry in reverse deterministic entry order.
- Treat `target.kind='row_restore'` as Sprint 3 request-shape admission only; whole-row restore execution remains Sprint 4.
- Advertise `available_rollback_actions[]` with `change_set` only when the server can execute whole-change-set rollback under current preconditions.
- Return `rollback_target_not_found` when the selector is not visible in the current record history.
- Return `rollback_precondition_failed` with the owner-defined reason code when the target exists but cannot be safely reversed against current state.
- Commit rollback by creating a new attributed `change_set` with `source='rollback'`, appending inverse mutation entries, incrementing affected first-class record row versions, appending `record_revisions`, updating projections, and publishing ordinary `record_changed` events.

Validation commands:
- `go test ./internal/modules/revisions -run 'TestPhase7_.*(U_7_05|U_7_06)'`
- `go test ./internal/modules/revisions ./internal/modules/workbook ./internal/modules/timeline ./internal/modules/evidence -run 'TestPhase7_.*(I_7_01|I_7_03)'`
- `make phase-slice PHASE=phase7`
- `make service-backed-slice PHASE=phase7`
- `git diff --check`

Deliverables:
- Complete: rollback OpenAPI request and response schemas.
- Complete: rollback error-family contract entries.
- Complete: durable single-entry rollback implementation for implemented reversible target families.
- Complete: whole-change-set rollback for currently reversible mutation entries.
- Complete: backend coverage for `U-7-05`, `U-7-06`, `I-7-01`, and `I-7-03`, including selector visibility negatives, hidden existing-record authorization, mention resolution/dismiss/restore rollback, supersede-link rollback, evidence association rollback, stale rollback, and lock precedence.

Risks and open questions:
- Resolved owner decision: tag rollback appears in `E-7-02` and `AC-010`, but typed tags are Phase 8 scope. Sprint 3 records tag rollback as a non-claim and returns `target_not_reversible` for current tag mutation entries.
- If earlier mutations did not persist reversible before/after values for a target family, do not infer rollback from projections or visible labels. Add the substrate or mark the target non-reversible with explicit current legality.
- Resolved owner decision: merge rollback may require additional mutation-entry detail beyond the current Phase 4 merge substrate. Sprint 3 does not claim merge-specific `AC-217`; add merge-specific reversible substrate and tests before claiming it.

Exit criteria:
- Rollback route admits only documented selector shapes.
- Single-entry rollback uses server-provided selectors and never client-inferred labels or storage identifiers.
- Whole-change-set rollback uses server-visible change-set selectors and reverses legal entries in reverse deterministic order.
- Successful rollback is visible as a new attributed history item and through ordinary collaboration events.

## Sprint 4. Whole-Row Restore, Retained History, and Locks

Objective: Complete row-backed snapshot restore, retained-history invariants, no-purge guarantees, and shared destructive-operation lock precedence.

Status: Implemented for backend Sprint 4 scope on 2026-05-11. Whole-row restore, shared restore/rollback/merge lock precedence, retained-history backend evidence, generated ledgers, and schedules have passed the validation listed below. Sprint 5 reviewer workbook UI and browser workflow evidence remains unclaimed.

Relevant IDs:
- `U-7-05`, `U-7-06`, `U-7-07`, `I-7-01`, `I-7-04`, `E-7-04`
- `REQ-01-089..REQ-01-111`
- `REQ-01-561..REQ-01-563`
- `REQ-02-205..REQ-02-220`
- `REQ-02-238..REQ-02-242`
- `REQ-03-101`, `REQ-03-141..REQ-03-144`
- `AC-011`, `AC-012`, `AC-187`, row-restore portions of `AC-217`, `AC-218`, `AC-353`, `AC-383..AC-385`

Grep references:
- `target.kind = 'row_restore'`
- `restore_to_revision_no`
- `record_locked`
- `FOR UPDATE NOWAIT`
- `records`
- `record_revisions`
- `MergeEntity`

Files and areas:
- `internal/modules/revisions`
- `internal/modules/entities/merge_store.go`
- `internal/modules/entities/merge_api.go`
- `internal/modules/entities/routes.go`
- `contracts/openapi/cartulary.openapi.yaml`
- `contracts/errors/index.json`
- Any migration needed for durable selector mapping or lock support.

Test-first sequence:
1. `U-7-05` asserts `row_restore` restores only row-backed fields to the selected `record_revisions` snapshot and does not recreate or delete links, tags, mentions, observations, or evidence associations.
2. `U-7-06` asserts restore, rollback, and merge share identical destructive-lock precedence: active lock returns `record_locked` with `retryable=true` before stale row-version or route-specific preconditions; after lock release, stale inputs fall through to ordinary downstream errors.
3. `U-7-07` and `I-7-04` assert retained-history visibility and selector stability across incident closure, delete or restore cycles, rollback, restart, and ordinary background work.
4. `E-7-04` later proves whole-row restore creates a new attributed revision and moves the visible row back to the selected historical snapshot without erasing prior history.

Implementation tasks:
- Implement `target.kind='row_restore'` by restoring only authoritative row-backed fields for the addressed `record_id` from the selected revision snapshot.
- For row restore, do not implicitly recreate/delete non-row-backed `record_links`, `record_tags`, `entity_mentions`, `indicator_observations`, or evidence associations.
- Implement a shared destructive lock helper over the `records` envelope. Acquire locks in canonical ascending `record_id` order and fail fast with `record_locked` if any protected record cannot be locked.
- Refactor merge to use the same protected-set lock helper or an equivalent shared path so `AC-187` and `AC-353` are not route-specific accidents.
- For rollback, compute the protected set from the selected target's current inverse effect, acquire locks, re-read current state, then evaluate `base_row_version`, target reversibility, and target-specific preconditions.
- Preserve all pre-existing history items and `history_entry_ref` values. Later target ineligibility must change only `reversible`, `available_rollback_actions[]`, and rollback failure behavior.
- Verify the current-profile route inventory and configuration expose no history-purge route and no history-retention-horizon setting for extant incidents.

Validation commands:
- `go test ./internal/modules/revisions ./internal/modules/entities -run 'TestPhase7_.*(U_7_05|U_7_06|U_7_07)'`
- `go test ./internal/modules/revisions ./internal/modules/entities ./internal/modules/collaboration -run 'TestPhase7_.*(I_7_01|I_7_04)'`
- `make backend-integration`
- `make phase-slice PHASE=phase7`
- `make service-backed-slice PHASE=phase7`
- `git diff --check`

Deliverables:
- Complete: whole-row restore from `record_revisions.after_json` row-backed snapshots.
- Complete: shared destructive-operation lock helper over `records` and merge integration.
- Complete: retained-history and no-purge evidence in backend Phase 7 rows.
- Complete: service-backed coverage for lock precedence, stale fail-closed behavior, and restart stability.
- Evidence roots:
  - `make generate`: `.cartulary/test-results/20260511T193209Z-p16774`
  - `make phase-ledgers`: `.cartulary/test-results/20260511T193215Z-p17367`
  - `make phase-schedules`: `.cartulary/test-results/20260511T193215Z-p17366`
  - `make phase-ledger-drift`: `.cartulary/test-results/20260511T193816Z-p32732`
  - `make phase-schedule-drift`: `.cartulary/test-results/20260511T193816Z-p32955`
  - `make json-shape-check`: `.cartulary/test-results/20260511T193816Z-p32954`
  - `CARTULARY_SUPPRESS_CHILD_SUCCESS=1 CARTULARY_CHECK_SCHEDULER_SKIP_PREREQUISITES=1 CARTULARY_TEST_RUN_ID=20260511T194500Z-sprint4-backend-integration make backend-integration`: `.cartulary/test-results/20260511T194500Z-sprint4-backend-integration`
  - `CARTULARY_SUPPRESS_CHILD_SUCCESS=1 CARTULARY_CHECK_SCHEDULER_SKIP_PREREQUISITES=1 CARTULARY_TEST_RUN_ID=20260511T194530Z-sprint4-phase-slice make phase-slice PHASE=phase7`: `.cartulary/test-results/20260511T194530Z-sprint4-phase-slice`
  - `CARTULARY_SUPPRESS_CHILD_SUCCESS=1 CARTULARY_CHECK_SCHEDULER_SKIP_PREREQUISITES=1 CARTULARY_TEST_RUN_ID=20260511T194600Z-sprint4-service-backed-slice make service-backed-slice PHASE=phase7`: `.cartulary/test-results/20260511T194600Z-sprint4-service-backed-slice`
  - `make go-test-duration-baseline-drift RESULTS_DIR=.cartulary/test-results/20260511T194600Z-sprint4-service-backed-slice`: `.cartulary/test-results/20260511T194424Z-p53617`
  - `make agent-finalize`: `.cartulary/test-results/20260511T194329Z-p52292`
- Observed harness blockers:
  - `make backend-integration` failed with `HarnessConfigError: CARTULARY_TEST_RUN_ID "20260511T193236Z-p18597" resolves to a non-empty run root`; artifact root `.cartulary/test-results/20260511T193236Z-p18597`; rerun with explicit collision-safe harness environment passed.
  - `CARTULARY_TEST_RUN_ID=20260511T193300Z-sprint4-backend-integration make backend-integration` failed with the same non-empty run-root error; artifact root `.cartulary/test-results/20260511T193300Z-sprint4-backend-integration`; rerun with `CARTULARY_SUPPRESS_CHILD_SUCCESS=1 CARTULARY_CHECK_SCHEDULER_SKIP_PREREQUISITES=1` passed.
  - `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260511T193520Z-sprint4-service-backed-slice` failed; artifact root `.cartulary/test-results/20260511T193439Z-p29042`; observed failure was `missing observed browser entry timings` for earlier browser rows because the supplied successful run was service-backed only. `make agent-finalize` without `RESULTS_DIR` passed and skipped duration-baseline refresh.

Risks and open questions:
- Transaction lock implementation must not introduce a client-visible lock surface, lock-holder identity surface, manual unlock route, or queued destructive-operation behavior.

Exit criteria:
- Restore, rollback, and merge observe the same destructive-lock precedence.
- Retained history remains fully paginatable and stable through restart and incident closure.
- No public or operator-visible current-profile setting can narrow retained history for an extant incident.

## Sprint 5. Reviewer Workbook UI and Browser Evidence

Objective: Expose Phase 7 reviewer workflows from the workbook surface using server-provided history metadata and selectors.

Status: Audit-complete for Sprint 5 scope on 2026-05-11 after remediation. The Timeline workbook surface now exposes row history from saved rows, renders row-centric server metadata and server-advertised actions, submits rollback targets only from server-supplied selectors plus the selected stable `record_id` and current row version, and proves delete, restore, and rollback outcomes through ordinary `record_changed` collaboration/view refresh paths. The Sprint 5 audit initially found generated schedule drift after `tools/phase7_test_map.json` changed; `make phase-schedules` regenerated the schedule artifacts, `make phase-schedule-drift` passed, and `make agent-finalize` passed with duration-baseline refresh skipped because `RESULTS_DIR` was unset.

Relevant IDs:
- `E-7-01`, `E-7-02`, `E-7-03`, `E-7-04`
- Frontend support for `U-7-*` rows where useful
- `REQ-01-048..REQ-01-056`, `REQ-01-071..REQ-01-082`, `REQ-01-089..REQ-01-111`, `REQ-01-250..REQ-01-277`
- `REQ-03-138..REQ-03-144`, `REQ-03-261..REQ-03-262`
- `AC-007`, `AC-010..AC-012`, `AC-215..AC-218`

Grep references:
- `WorkbookShell`
- `record_changed`
- `affected_views`
- `change_kind`
- `Conflict`
- `row history`
- `rollback`

Files and areas:
- `apps/web/src/WorkbookShell.tsx`
- `apps/web/src/WorkbookShell.phase7.test.tsx`
- `apps/web/e2e/phase7.history.spec.ts`
- Existing browser test harness and web-server-backed orchestration.

Completed test-first sequence:
1. Complete: prerequisite reruns passed for whole-row restore, retained-history stability, and shared destructive-operation lock behavior before frontend edits.
2. Complete: `apps/web/src/WorkbookShell.phase7.test.tsx` covers history opening from a selected saved row, server-advertised action rendering, selector-only rollback request construction, delete or restore row-version source selection, and ordinary socket remove/invalidate continuity.
3. Complete: `E-7-01` opens history from a selected Timeline workbook row and asserts actor, timestamp, operation, diff summary, and legal rollback actions render from server metadata.
4. Complete: `E-7-02` rolls back an attached-evidence record-link mutation through a server-supplied `history_entry_ref`, proves another client observes an ordinary `record_changed` `invalidate` event for the affected Timeline view, and proves a later unrelated summary edit remains intact.
5. Complete: `E-7-03` soft-deletes and restores a row using current and tombstone row versions and verifies another client observes ordinary `record_changed` `remove` on delete and `invalidate` on restore.
6. Complete: `E-7-04` performs whole-row restore from a server-advertised `row_restore` action, verifies the request uses only `restore_to_revision_no`, proves the visible Timeline summary returns to the selected historical snapshot, and verifies prior history remains visible after the new attributed revision with actor attribution and timestamp metadata.

Implemented behavior:
- Added shared frontend Phase 7 history types and the selector helper `buildRecordRollbackTargetFromHistoryAction`.
- Added a one-click saved-row History action in the Timeline actions column. The action opens an inline inspector history section, not a generalized approval workflow or modal-only flow.
- History rendering is row-centric and uses `/api/v1/records/{record_id}/history`; it displays actor, timestamp, operation, diff summary, change-set id, current row version, deleted state, and only server-advertised legal action controls.
- Rollback request targets are built only from `available_rollback_actions[]`, `history_entry_ref`, `change_set_id`, and `revision_no`; the client does not infer legality from labels, diff text, item order, SQL names, projection storage names, or visible workbook structure.
- Soft-delete and restore controls use the selected row's current row version or the tombstone row version returned by server history metadata.
- Delete, restore, rollback, and row-restore requests use ordinary route calls and then ordinary view/history refresh behavior. Other clients observe delete, restore, and rollback through the existing `record_changed` collaboration stream.
- The Timeline action layout was compacted so the additional History affordance does not break the pre-existing Phase 3 fixed-row action-cell interactions.
- `apps/web/playwright.shared.config.ts` now binds the default owned web stack to the same frontend and backend ports used by direct Playwright defaults, so direct `pnpm playwright test ...` commands wait on the stack they actually start.

Validation commands and outcomes:
- Passed: `go test ./internal/modules/revisions -run 'TestPhase7_RollbackSelectorUnion_U_7_05|TestPhase7_RetainedHistoryInvariants_U_7_07|TestPhase7_DestructiveOperationLocks_U_7_06'`.
- Passed: `go test ./internal/modules/revisions -run 'TestPhase7_DeleteRestoreRollbackAtomicConsequences_I_7_01|TestPhase7_RetainedHistoryAcrossRestartAndClosure_I_7_04'`.
- Passed: `tmp/node-runtime/bin/pnpm --dir apps/web exec vitest run src/WorkbookShell.phase7.test.tsx`.
- Audit rerun passed: `make frontend-unit`; run root `.cartulary/test-results/20260511T233055Z-p49413`.
- Audit rerun passed: `make frontend-typecheck`; run root `.cartulary/test-results/20260511T233105Z-p50338`.
- Audit rerun passed: `make lint-biome`; run root `.cartulary/test-results/20260511T233109Z-p50626`.
- Audit rerun passed: `FORCE_COLOR=0 NO_COLOR=1 tmp/node-runtime/bin/pnpm --dir apps/web exec playwright test e2e/phase7.history.spec.ts`; 4 tests passed.
- Passed: `FORCE_COLOR=0 NO_COLOR=1 tmp/node-runtime/bin/pnpm --dir apps/web exec playwright test e2e/phase3.spec.ts -g "E-3-03 drives review, demotion, and supersede through the visible workbook surface"` after compacting Timeline actions.
- Audit rerun passed: `make browser-e2e-webserver-backed`; run root `.cartulary/test-results/20260511T233132Z-p52358`; 37 tests passed.
- Passed: `make browser-e2e-duration-baselines RESULTS_DIR=.cartulary/test-results/20260511T225805Z-p78524`; run root `.cartulary/test-results/20260511T225921Z-p84996`.
- Audit remediation passed: `make agent-finalize`; run root `.cartulary/test-results/20260511T233530Z-p64256`, with duration-baseline refresh explicitly skipped because `RESULTS_DIR` was unset.
- Passed: `make phase-ledgers`; run root `.cartulary/test-results/20260511T231250Z-p13622`.
- Audit rerun passed: `make phase-ledger-drift`; run root `.cartulary/test-results/20260511T233540Z-p65285`.
- Audit remediation passed: `make phase-schedules`; run root `.cartulary/test-results/20260511T233512Z-p63713`.
- Audit remediation passed: `make phase-schedule-drift`; run root `.cartulary/test-results/20260511T233527Z-p64017`, and rerun after `agent-finalize` passed at `.cartulary/test-results/20260511T233551Z-p65605`.
- Audit rerun passed: `make phase-slice PHASE=phase7`; run root `.cartulary/test-results/20260511T233211Z-p56764`; 44 tests passed.
- Audit rerun passed: `make service-backed-slice PHASE=phase7`; run root `.cartulary/test-results/20260511T233230Z-p59589`; 44 tests passed.
- Audit rerun passed: `git diff --check`.

Observed blockers and remediations:
- Direct Phase 7 Playwright initially timed out waiting for the web server because the default config waited on `127.0.0.1:4173` while the owned stack chose a dynamic frontend port. Remediation: bind the default owned-stack frontend port to the configured public origin port.
- Direct Phase 7 Playwright then timed out waiting for API readiness at `127.0.0.1:8080` because the owned stack chose a dynamic backend port while helpers defaulted to `8080`. Remediation: bind the default owned-stack backend port to the configured API origin port.
- Early browser spec attempts failed because the mention-resolution target selected by the test did not advertise a current `history_entry` rollback action and because assertions used the Timeline field key instead of the scalar workbook key for summary cell test ids. Remediation: use the already-implemented attached-evidence record-link rollback family and assert visible Timeline summary cells with the stable `summary` workbook key.
- `make lint-biome` failed during development for formatting, an array-index React key, and an unused import. Remediation: format the edited files, key history list items by server metadata, and remove the unused import.
- `make phase-schedules` failed after the `E-7-02` browser title changed because `tools/browser_e2e_duration_baselines.json` still referenced the old title. Remediation: align the title, regenerate schedules, then refresh browser duration baselines from the successful retained browser run.
- `make browser-e2e-webserver-backed` first failed in pre-existing `E-3-03` because adding History as a fifth vertical action overflowed the fixed Timeline action-cell height and row controls intercepted clicks. Remediation: place Inspect and History in a compact top row; the focused Phase 3 test and the full browser target then passed.
- `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260511T225805Z-p78524` failed at `.cartulary/test-results/20260511T225853Z-p83257` because the browser-only run root did not contain harness-smoke timing summaries. Remediation: rerun `make agent-finalize` without `RESULTS_DIR`, recording the intended duration-baseline skip, and refresh browser duration baselines separately from the successful retained browser run.
- Sprint 5 audit initially failed `make phase-schedule-drift` at `.cartulary/test-results/20260511T233256Z-p62610` because `tools/phase7_test_map.json` changed after the generated schedule artifacts. Remediation: `make phase-schedules` passed at `.cartulary/test-results/20260511T233512Z-p63713`, `make phase-schedule-drift` passed at `.cartulary/test-results/20260511T233527Z-p64017`, and a post-`agent-finalize` drift rerun passed at `.cartulary/test-results/20260511T233551Z-p65605`.

Deliverables:
- Complete: reviewer history inspector section in the Timeline workbook surface.
- Complete: explicit rollback, soft-delete, restore, and whole-row restore controls driven by server history metadata.
- Complete: frontend unit coverage for client-side history rendering, selector use, destructive row-version selection, and collaboration continuity.
- Complete: browser tests for `E-7-01..E-7-04`.
- Complete: Phase 7 manifest, generated coverage ledger, generated schedules, and browser duration baselines updated through canonical commands.

Risks and non-claims:
- Typed tag rollback remains unclaimed because typed tags are Phase 8.
- Merge-specific rollback and browser whole-change-set rollback remain unclaimed.
- Phase 8 saved-view behavior and Phase 9 workbook-native keyboard/clipboard behavior remain unclaimed.
- Sprint 5 does not claim full Phase 7 exit.

Exit criteria:
- Met: browser evidence proves the reviewer can execute Phase 7 workflows from the workbook surface.
- Met: client rollback requests use only server-supplied selectors and stable `record_id` values.
- Met: another client observes delete, restore, and rollback through ordinary `record_changed` collaboration streams without a special client-side event family.

## Sprint 6. Phase Gate, Ledgers, Schedules, Baselines, and Handoff Cleanup

Objective: Prove Phase 7 is complete, resolve generated drift, update test ledgers and schedules, refresh baselines when needed, and hand off stable prerequisites to Phase 8.

Status: Evidence-backed exit complete on 2026-05-12, with non-Phase 7 `make check` blockers recorded below.

Relevant IDs: all `U-7-*`, `I-7-*`, `E-7-*`; all Phase 7-owned ACs and unresolved support AC notes.

Files and areas:
- `PHASE7_IMPLEMENTATION_PLAN.md`
- `tools/phase7_test_map.json`
- `docs/testing/phase7_coverage_ledger.md`
- `tools/go_test_duration_baselines.json` if service-backed Go rows are added or promoted.
- Generated schedule manifests.
- Any generated OpenAPI or protocol outputs produced by `make generate`.

Test-first sequence:
1. Complete: Sprint 6 inspection found no skipped Phase 7 tests, no duplicate actual row IDs, no forbidden Phase 7 IDs in prior-phase files, and no remaining authoritative placeholder tests.
2. Complete: manifest, ledger, schedule, and baseline coverage validation ran before broader tests.
3. Complete: focused backend, frontend, browser, and public Phase 7 wrappers passed.
4. Complete: `make phase-slice PHASE=phase4` passed after Sprint 6 removed unused merge-adjacent dead code.
5. Complete: `make agent-finalize` ran without `RESULTS_DIR` because Sprint 6 did not add or promote duration inputs and no full qualifying duration-maintenance run root was supplied.

Implementation tasks:
- Complete: `tools/phase7_test_map.json` has no stale placeholder claims, no duplicate actual row IDs, and retained support/forbidden file boundaries.
- Complete: `make phase-ledgers`, `make phase-ledger-drift`, `make phase-schedules`, and `make phase-schedule-drift` passed after manifest cleanup.
- Complete: no contract, SQL, or codegen source change required `make generate`; `make generate-drift` passed.
- Complete: no duration baseline refresh was required; `make go-test-duration-baseline-coverage` passed.
- Complete: Sprint 6 removed unused dead helpers reported by Staticcheck in `internal/modules/entities/merge_store.go`, `internal/modules/revisions/rollback_store.go`, and `internal/modules/revisions/store.go`; targeted Phase 7 regressions and `make lint-go VERBOSE=1` passed afterward.
- Complete: artifact roots, blockers, non-claims, and Phase 8 handoff notes are recorded in this plan.

Validation commands:
- Passed: `make explain-phase PHASE=phase7`.
- Passed: `make phase-map-check`; run root `.cartulary/test-results/20260512T000006Z-p72458`.
- Passed: `make phase-test-name-check`; run root `.cartulary/test-results/20260512T000039Z-p73206`.
- Passed: `make phase-ledgers`; initial run root `.cartulary/test-results/20260512T000352Z-p75460`; final post-manifest-cleanup run root `.cartulary/test-results/20260512T001900Z-sprint6-phase-ledgers-final`.
- Passed: `make phase-ledger-drift`; initial run root `.cartulary/test-results/20260512T000354Z-p76037`; final post-documentation run root `.cartulary/test-results/20260512T002022Z-p96732`.
- Passed: `make phase-schedules`; initial run root `.cartulary/test-results/20260512T000352Z-p75462`; final post-manifest-cleanup run root `.cartulary/test-results/20260512T001930Z-sprint6-phase-schedules-final`.
- Passed: `make phase-schedule-drift`; initial run root `.cartulary/test-results/20260512T000354Z-p76035`; final post-documentation run root `.cartulary/test-results/20260512T002022Z-p96905`.
- Passed: `make go-test-duration-baseline-coverage`; run root `.cartulary/test-results/20260512T000352Z-p75532`.
- Passed: `go test ./internal/modules/revisions -run 'TestPhase7_.*(U_7_01|U_7_02|U_7_07)'`.
- Passed: `go test ./internal/modules/revisions ./internal/modules/workbook ./internal/modules/timeline -run 'TestPhase7_.*(I_7_02|I_7_04)'`.
- Passed: `go test ./internal/modules/revisions -run 'TestPhase7_.*(U_7_03|U_7_04|U_7_05|U_7_06)'`.
- Passed: `go test ./internal/modules/revisions ./internal/modules/workbook ./internal/modules/timeline ./internal/modules/evidence ./internal/modules/entities ./internal/modules/collaboration -run 'TestPhase7_.*(I_7_01|I_7_03|I_7_04)'`.
- Passed after dead-code cleanup: `go test ./internal/modules/entities ./internal/modules/revisions -run 'TestPhase7_.*(U_7_05|U_7_06|I_7_01|I_7_03)'`.
- Passed after dead-code cleanup: `go test ./internal/modules/revisions -run 'TestPhase7_.*(U_7_01|U_7_02|U_7_03|U_7_04|U_7_05|U_7_06|U_7_07|I_7_01|I_7_03|I_7_04)'`.
- Passed: `make frontend-unit`; run root `.cartulary/test-results/20260512T000446Z-p77905`; direct rerun after transient aggregate failure passed at `.cartulary/test-results/20260512T001249Z-p51541`.
- Passed: `make frontend-typecheck`; run root `.cartulary/test-results/20260512T000446Z-p77819`.
- Passed: `make lint-biome`; run root `.cartulary/test-results/20260512T000446Z-p77862`.
- Passed: `make browser-e2e-webserver-backed`; run root `.cartulary/test-results/20260512T000458Z-p79239`; 37 tests passed.
- Passed: `make phase-slice PHASE=phase7`; initial run root `.cartulary/test-results/20260512T000542Z-p83733`; post-cleanup run root `.cartulary/test-results/20260512T001054Z-p35075`.
- Passed: `make service-backed-slice PHASE=phase7`; initial run root `.cartulary/test-results/20260512T000601Z-p86580`; post-cleanup run root `.cartulary/test-results/20260512T001115Z-p38330`.
- Passed earlier-phase regression: `make phase-slice PHASE=phase4`; run root `.cartulary/test-results/20260512T001137Z-p41143`.
- Passed: `make agent-finalize`; initial run root `.cartulary/test-results/20260512T000628Z-p89425`; post-cleanup run root `.cartulary/test-results/20260512T001204Z-p48043`; duration-baseline refresh intentionally skipped because `RESULTS_DIR` was unset.
- Passed: `make lint-go VERBOSE=1`; run root `.cartulary/test-results/20260512T000949Z-p31375` failed before dead-code cleanup, then the rerun passed without findings.
- Passed: `make test-fast`; initial run root `.cartulary/test-results/20260512T000635Z-p90423`; post-cleanup retry passed at `.cartulary/test-results/20260512T001302Z-p52476`. The intermediate run `.cartulary/test-results/20260512T001215Z-p49047` failed in Phase 6 `frontend-unit`, and direct `make frontend-unit` passed immediately afterward at `.cartulary/test-results/20260512T001249Z-p51541`.
- Passed: `make generate-drift`; initial run root `.cartulary/test-results/20260512T000731Z-p99140`; post-cleanup run root `.cartulary/test-results/20260512T001354Z-p60969`; final post-documentation run root `.cartulary/test-results/20260512T002022Z-p96859`.
- Closed outside Phase 7: `make check` initially failed at `.cartulary/test-results/20260512T001400Z-p61609` in non-Phase 7 `browser-e2e-visual` rows `V-3-GRID-03`, `V-5-GRID-02`, `V-6-GRID-02`, and `V-6-GRID-03`.
- Passed after visual remediation: focused visual rerun for `V-3-GRID-03|V-5-GRID-02|V-6-GRID-02|V-6-GRID-03` passed after canonical golden refresh.
- Passed after visual remediation: `make browser-e2e-visual`; run root `.cartulary/test-results/20260512T010453Z-p22559`; 11 tests passed.
- Failed as expected for a visual-only run root: `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260512T010453Z-p22559`; run root `.cartulary/test-results/20260512T010535Z-p26209`; duration maintenance reported missing functional `E-*` browser timing entries because the supplied run root contained only visual rows.
- Passed after visual remediation: `make agent-finalize`; run root `.cartulary/test-results/20260512T010552Z-p26711`; duration-baseline refresh intentionally skipped because `RESULTS_DIR` was unset.
- Passed after visual remediation: `make generate-drift`; run root `.cartulary/test-results/20260512T010615Z-p27786`.
- Passed after visual remediation: `make phase-ledger-drift`; run root `.cartulary/test-results/20260512T010615Z-p27879`.
- Passed after visual remediation: `make phase-schedule-drift`; run root `.cartulary/test-results/20260512T010615Z-p27921`.
- Passed after visual remediation: `make check`; run root `.cartulary/test-results/20260512T010623Z-p28902`; 505 tests passed.
- Passed: `git diff --check`.

Deliverables:
- Complete: all authoritative Phase 7 rows pass as real assertions.
- Complete: generated drift, phase-ledger drift, and phase-schedule drift are clean.
- Complete: no baseline update was required.
- Complete: final artifact-root notes and Phase 8 handoff notes are recorded.

Risks and open questions:
- Closed non-Phase 7 check blocker: `browser-e2e-visual` rows `V-3-GRID-03`, `V-5-GRID-02`, `V-6-GRID-02`, and `V-6-GRID-03` were remediated without broadening Phase 7 scope. `V-5-GRID-02` required Timeline action-cell layout changes; `V-6-GRID-03` required deterministic visual fixture ordering; the four accepted visual changes were refreshed through Playwright snapshot update and verified by `make browser-e2e-visual`.
- Resolved owner decisions for Phase 7 exit: tag rollback remains Phase 8-owned; `AC-184` and `AC-185` remain Phase 8-owned support references; merge-specific rollback remains unclaimed until reversible merge substrate and tests exist; Phase 8 saved-view behavior and Phase 9 keyboard/clipboard behavior remain non-claims.

Exit criteria:
- Met: all authoritative `U-7-*`, `I-7-*`, and `E-7-*` rows pass in their intended layers.
- Met: no placeholder tests remain.
- Met: public `phase-slice` and `service-backed-slice` wrappers pass for Phase 7.
- Met: touched earlier-phase regression `make phase-slice PHASE=phase4` passes.
- Met: generated drift is resolved through canonical commands.
- Met: `make agent-finalize` ran and passed after focused evidence.
- Met: `make check` passes after closing the non-Phase 7 visual blockers.

## Risks and Open Questions

1. Tag rollback is explicitly named in Phase 7 UI/AC language, but typed tags are Phase 8 scope. Sprint 6 records tag rollback as a Phase 8-owned non-claim; current record-tag history entries remain visible but intentionally non-reversible with `target_not_reversible`.
2. `AC-184` and `AC-185` appear in `U-7-01`, but the coverage index assigns them to Phase 8. Sprint 6 treats them as Phase 8-owned support references, not completed Phase 7 evidence.
3. First-class record dispatch must stay explicit. Sprint 6 confirmed the current delete/restore adapter matrix for first-class `records.record_type` values: `timeline_event`, `host`, `identity`, `party`, `indicator`, `artifact`, `task_request`, `decision`, `evidence`, and `assessment`.
4. `history_entry_ref` must be stable and opaque. Default decision: persist a durable opaque selector or selector mapping. Do not use transient order, exposed database primary keys, or a key-rotation-sensitive derived token.
5. Merge rollback may need more reversible mutation-entry detail than the current Phase 4 merge path records. Sprint 6 keeps merge-specific rollback unclaimed until reversible merge substrate and direct tests exist.
6. `record_locked` exists and is covered for restore. Rollback and merge lock work must continue to use the public error contract and must not rely on internal errors without contract coverage.
7. Retained-history invariants prohibit purge behavior for extant current-profile incidents. Do not add configuration or operator tooling that narrows visible retained history during this phase.

## Phase Validation Criteria

Phase 7 is valid only when:
- Every authoritative Phase 7 manifest row runs and passes in the intended test layer.
- Shared harnesses affected by history, mutation, authorization, collaboration, and audit changes still pass.
- All earlier phases still pass for touched behavior.
- New history/delete/restore/rollback routes have contract, authorization, idempotency, concurrency, audit, projection, and collaboration coverage.
- Mutation-bearing paths have service-backed integration coverage against real Postgres and, where applicable, object/evidence boundaries.
- Browser evidence proves the reviewer workflows through the workbook surface without relying on client-side inference of rollback legality.
- No phase claim depends on behavior explicitly deferred to Phase 8 or Phase 9 unless an owner decision pulled that behavior forward.

## Phase Exit Criteria

Phase 7 exits only when:
- `tools/phase7_test_map.json` is active, complete, and drift-free.
- `docs/testing/phase7_coverage_ledger.md` is generated and drift-free.
- Every `U-7-*`, `I-7-*`, and `E-7-*` row is backed by a real assertion.
- History pagination, tombstone row-version restore, stable `history_entry_ref`, delete, restore, single-entry rollback, change-set rollback, row restore, retained-history stability, and lock precedence all have passing evidence.
- OpenAPI, error index, generated Go, generated TypeScript, and SQL outputs are current after `make generate` when source inputs changed.
- `make phase-slice PHASE=phase7`, `make service-backed-slice PHASE=phase7`, `make test-fast`, `make agent-finalize`, and `make check` have passed or any non-Phase 7 blocker is recorded with exact target and artifact root.
- The Sprint Checklist and handoff notes in this file are updated with actual completion status and artifact roots.

## Handoff Requirements for Phase 8

Phase 8 may rely on these completed Phase 7 contracts after exit:
- Row-centric history reads are stable, paginated, and bound to `record_id`.
- First-class record delete and restore preserve append-only history and emit ordinary collaboration events.
- Single-entry and whole-change-set rollback create new attributed `change_sets` with `source='rollback'` and never rewrite prior history.
- `history_entry_ref` values are stable for the retained-history lifetime of the record in the current deployment.
- Whole-row restore restores only row-backed fields and does not implicitly recreate or delete relationship-like mutation targets.
- Destructive-operation locks are shared by restore, rollback, and merge and fail fast before stale-precondition evaluation.
- Retained history for extant records is not narrowed by incident closure, delete or restore cycles, rollback, restart, or current-profile purge settings.

Phase 8 handoff must also state:
- Tag rollback remains Phase 8-owned. Phase 7 shows current `record_tag` mutation entries in history when present but does not advertise executable rollback actions for them and returns `target_not_reversible` if addressed.
- `AC-184` and `AC-185` are treated as Phase 8-owned support references for Phase 7, not shared completed Phase 7 evidence.
- Final first-class record adapter matrix:
  - History reads are record-envelope based for first-class records with retained change-set or revision evidence.
  - Delete/restore adapter coverage is explicit for `timeline_event`, `host`, `identity`, `party`, `indicator`, `artifact`, `task_request`, `decision`, `evidence`, and `assessment`.
  - Row-level rollback target kinds are executable for `timeline_record`, `host`, `identity`, `indicator`, `assessment`, `evidence`, and generic `record` where reversible before/after mutation data exists; `row_restore` uses `record_revisions` snapshots and does not recreate or delete relationship-like adjuncts.
- Mutation target families visible in history but intentionally non-reversible: `record_tag` remains Phase 8-owned; merge-specific history remains non-reversible until merge writes reversible mutation-entry detail; unsupported/default mutation families remain non-reversible with `target_not_reversible`.
- Final Phase 7 public wrapper roots: `make phase-slice PHASE=phase7` at `.cartulary/test-results/20260512T001054Z-p35075`; `make service-backed-slice PHASE=phase7` at `.cartulary/test-results/20260512T001115Z-p38330`.
- Broader verification roots: `make test-fast` passed at `.cartulary/test-results/20260512T001302Z-p52476`; `make browser-e2e-visual` passed at `.cartulary/test-results/20260512T010453Z-p22559`; `make generate-drift` passed at `.cartulary/test-results/20260512T010615Z-p27786`; `make phase-ledger-drift` passed at `.cartulary/test-results/20260512T010615Z-p27879`; `make phase-schedule-drift` passed at `.cartulary/test-results/20260512T010615Z-p27921`; `make agent-finalize` passed at `.cartulary/test-results/20260512T010552Z-p26711`; `make check` passed at `.cartulary/test-results/20260512T010623Z-p28902`.
