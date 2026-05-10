# Phase 7 Implementation Plan

## Summary

This file is the execution roadmap and progress marker for Cartulary Phase 7: reviewer-facing history, delete or restore, and rollback.

`docs/guides/cartulary_implementation_testing_guide.md` section 9.7 is the controlling implementation-scope reference for this phase. Normative behavior remains owned by the core documents, especially Core 01 section 3.3.4.2, Core 01 section 3.3.5, Core 02 sections 14 and 15, Core 03 section 10, and Core 04 sections 2 and 3.

This planning artifact does not implement Phase 7 behavior. It is intentionally root-level so agents can find it quickly during handoff or interrupted implementation sessions. No README update is required for discoverability.

Current repo status at plan authoring: Phase 7 is listed as `planned` in `tools/phase_registry.json`; `tools/phase7_test_map.json` and `docs/testing/phase7_coverage_ledger.md` do not exist yet; the OpenAPI contract does not expose the Phase 7 history, delete, restore, or rollback route family yet; the `record_locked` error code exists, but the remaining Phase 7 delete, restore, rollback, history, and rollback-precondition public error families still need contract coverage.

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
| [ ] | 0. Phase 7 ownership manifest and harness setup | [ ] planned | None for manifest creation, but Phase 7 remains planned until the implementation branch intentionally activates it. | Add selectable rows first; later sprints must replace every placeholder before phase exit. |
| [ ] | 1. Record-history read contract | [ ] planned | Resolve whether `AC-184` and `AC-185` in `U-7-01` are support references or guide drift, because the coverage index assigns those ACs to Phase 8. | History route must be record-scoped, not view-scoped. |
| [ ] | 2. Soft-delete and restore | [ ] planned | First-class record dispatch must be made explicit before implementation: support every current `records` envelope type with a source adapter, or document the owner decision for any excluded current type. | Ordinary delete does not take destructive-operation locks; restore does. |
| [ ] | 3. Rollback request and single-entry reversal | [ ] planned | Tag rollback language appears in Phase 7 evidence while typed tags are Phase 8; decide whether to pull minimal tag rollback forward or mark tag-specific coverage blocked by owner decision. | Source must be `rollback`; previous history rows are never mutated in place. |
| [ ] | 4. Whole-change-set rollback, whole-row restore, retained history, and locks | [ ] planned | Confirm whether the existing Phase 4 merge substrate contains enough mutation-entry detail to satisfy merge rollback before claiming `AC-217`. | Shared destructive locks must cover restore, rollback, and merge with identical precedence. |
| [ ] | 5. Reviewer workbook UI and browser evidence | [ ] planned | UI must not infer legal rollback actions from visible labels, diff text, or storage identifiers. | Browser evidence covers reviewer flows over server-provided selectors and actions. |
| [ ] | 6. Phase gate, ledgers, schedules, baselines, and handoff cleanup | [ ] planned | Blocked until all Sprint 0-5 placeholders are replaced with real assertions and unresolved owner decisions are closed or recorded as explicit non-claims. | Run `make agent-finalize` first at end of run, then broader verification. |

## Global References

- Controlling guide: `docs/guides/cartulary_implementation_testing_guide.md`, `Phase 7`, `U-7-01..U-7-07`, `I-7-01..I-7-04`, `E-7-01..E-7-04`.
- Phase-owned ACs: `AC-007`, `AC-010..AC-012`, `AC-181..AC-183`, `AC-187`, `AC-215..AC-218`, `AC-353`, `AC-383..AC-385`.
- Ambiguous support/shared ACs: `AC-184` and `AC-185` appear in the Phase 7 `U-7-01` guide row, but the guide coverage index assigns them to Phase 8. Treat them as an open guide/owner decision before claiming them in Phase 7.
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

Status: Planned. Phase 7 is currently registered as `planned`; no Phase 7 manifest or generated ledger exists at plan authoring time.

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
1. Add manifest rows for every `U-7-01..U-7-07`, `I-7-01..I-7-04`, and `E-7-01..E-7-04` guide row before feature implementation.
2. Add intentional failing or no-op placeholders only long enough to make Phase 7 row selection visible. Every placeholder must name the future real assertion and must be replaced before Sprint 6 exit.
3. Declare support-only carryover files as `forbidden_id_files` so earlier phase evidence cannot accidentally claim Phase 7 IDs.
4. Generate the Phase 7 coverage ledger and schedules after the manifest exists.
5. Activate Phase 7 only when the implementation branch is ready for Phase 7 rows to participate in public phase wrappers.

Implementation tasks:
- Add `tools/phase7_test_map.json` with non-empty `claim` and `out_of_scope` text for every authoritative row.
- Include the execution dependency for each row explicitly: backend unit, backend store/integration, frontend unit, or browser functional.
- Record the Phase 3, Phase 4, Phase 5, and Phase 6 support boundaries in manifest notes rather than re-owning their existing IDs.
- Add test stubs in the final expected file locations so `make explain-phase PHASE=phase7` can discover symbols and row titles.
- Keep generated ledgers and schedules produced by canonical commands; do not hand-edit generated outputs.

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
- Planned: `tools/phase7_test_map.json`.
- Planned: `docs/testing/phase7_coverage_ledger.md`.
- Planned: generated service-backed and check schedule updates.
- Planned: selectable Phase 7 test symbols and browser rows.

Risks and open questions:
- Phase activation changes task-surface behavior. Keep Phase 7 as `planned` until the branch intentionally opts into active execution.
- Newly declared service-backed rows require explicit duration baselines after successful uncontaminated evidence; missing baselines must not be hidden by fallback weights.

Exit criteria:
- `make explain-phase PHASE=phase7` reports the manifest, ledger path, execution dependencies, service requirements, and target coverage.
- `make phase-ledger-drift` passes after ledger generation.
- `make phase-schedule-drift` passes after schedule generation.
- Phase 7 row IDs appear only in authoritative Phase 7 files or approved manifest support references.

## Sprint 1. Record-History Read Contract

Objective: Implement `GET /api/v1/records/{record_id}/history` as a record-scoped retained-history read route with deterministic ordering, pagination, tombstone concurrency, rollback metadata, and stable single-entry selectors.

Status: Planned.

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

Validation commands:
- `go test ./internal/modules/revisions -run 'TestPhase7_.*(U_7_01|U_7_02|U_7_07)'`
- `go test ./internal/modules/revisions ./internal/modules/workbook ./internal/modules/timeline -run 'TestPhase7_.*(I_7_02|I_7_04)'`
- `make phase-slice PHASE=phase7`
- `make service-backed-slice PHASE=phase7`
- `git diff --check`

Deliverables:
- Planned: OpenAPI history route and response schemas.
- Planned: history query store/service.
- Planned: durable `history_entry_ref` selector strategy.
- Planned: unit and integration coverage for `U-7-01`, `U-7-02`, `U-7-07`, `I-7-02`, and history portions of `I-7-04`.

Risks and open questions:
- `AC-184` and `AC-185` are listed on `U-7-01` but assigned to Phase 8 in the coverage index. Before implementation, either correct the guide, mark those ACs as support-only references in the Phase 7 manifest, or document an owner decision that Phase 7 owns the narrow history-query overlap.
- A transient `history_entry_ref` derived from result order or in-memory state would fail retained-history and restart requirements.

Exit criteria:
- History reads are stable across repeated reads, pagination, restart, delete/restore, rollback, and incident closure.
- Soft-deleted records expose the tombstone `row_version` needed by restore.
- The route does not require or accept `incident_id`, `view_schema_id`, or display labels as behavior inputs.

## Sprint 2. Soft-Delete and Restore

Objective: Implement first-class record soft-delete and restore with role gates, route-scoped idempotency, row-version preconditions, append-only history, projection invalidation, and collaboration events.

Status: Planned.

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

Test-first sequence:
1. `U-7-03` asserts delete request shape, role gates, stale row-version failure, idempotent replay, divergent `client_txn_id` reuse conflict, already-deleted failure, and patch-after-delete `record_deleted_use_restore`.
2. `U-7-04` asserts restore request shape, reviewer/admin role gate, tombstone `row_version` requirement, not-deleted failure, idempotent replay, and append-only prior history preservation.
3. `I-7-01` asserts delete and restore atomically update source rows, record envelope, projections, history rows, and emitted collaboration events.
4. `I-7-03` asserts stale restore fails closed and does not mutate current row state.
5. `E-7-03` later proves browser soft-delete/restore flow and two-client `remove`/`invalidate` observations.

Implementation tasks:
- Add delete and restore route contracts. Delete accepts `base_row_version`, `client_txn_id`, and optional normalized `reason`; restore accepts the same shape.
- Route delete through ordinary optimistic concurrency. The current profile says ordinary soft-delete is outside the destructive-operation lock family.
- Route restore through the shared destructive-operation lock helper because restore is in the destructive-operation family.
- Enforce delete roles: `editor`, `reviewer`, or `admin`. Enforce restore roles: `reviewer` or `admin`. A visible but underprivileged caller receives `403`; a caller without visibility receives `404`.
- Use idempotency key `(actor_user_id, record_id, client_txn_id)` and compare normalized reason, with omission, explicit `null`, and normalized-empty reason equivalent.
- On delete, set source and/or envelope soft-delete state, increment row version, create one attributed `change_set`, append reversible mutation entries, append a `record_revisions` row with `operation = soft_delete`, and remove the row from ordinary view queries.
- On restore, clear only the current soft-delete state, increment row version, create one attributed `change_set`, append reversible mutation entries, append a `record_revisions` row with `operation = restore`, and make the row eligible for ordinary view queries again.
- Preserve prior history in place. Do not hard-delete revisions, change sets, blobs, source rows, or link rows.
- Emit ordinary replayable `record_changed` events: delete uses `change_kind='remove'`; restore uses `change_kind='invalidate'`.
- Recompute or invalidate surviving derived rows whose chips, counts, or linked-record summaries change because of delete or restore.

Validation commands:
- `go test ./internal/modules/revisions -run 'TestPhase7_.*(U_7_03|U_7_04)'`
- `go test ./internal/modules/revisions ./internal/modules/workbook ./internal/modules/collaboration -run 'TestPhase7_.*(I_7_01|I_7_03)'`
- `make phase-slice PHASE=phase7`
- `make service-backed-slice PHASE=phase7`
- `git diff --check`

Deliverables:
- Planned: delete and restore OpenAPI contracts.
- Planned: route-level decoder and idempotency hashing.
- Planned: source-record dispatch registry for all current first-class record types.
- Planned: projection and collaboration invalidation for delete/restore.
- Planned: unit and integration coverage for `U-7-03`, `U-7-04`, delete/restore parts of `I-7-01`, and restore part of `I-7-03`.

Risks and open questions:
- First-class record dispatch must be explicit. The implementation must inventory every current `records.record_type` backed by an implemented source table and either add an adapter or record an owner decision for non-claiming that type. Do not silently support only Timeline if other current first-class records exist.
- Delete/restore must not be confused with deleting or restoring non-record mutation targets such as individual links, tags, mentions, observations, or evidence associations.

Exit criteria:
- Delete and restore preserve append-only history and source attribution.
- Ordinary queries hide soft-deleted rows and include restored rows only through current view contracts.
- Collaboration subscribers observe ordinary `record_changed` messages, not a Phase 7-specific event family.

## Sprint 3. Rollback Request and Single-Entry Reversal

Objective: Implement rollback route admission and single-entry rollback for history items that map to exactly one reversible mutation target.

Status: Planned.

Relevant IDs:
- `U-7-05`, `U-7-06`, `I-7-01`, `I-7-03`, `E-7-02`
- `REQ-01-089..REQ-01-111`
- `REQ-02-212..REQ-02-218`
- `REQ-03-141..REQ-03-144`
- `REQ-04-036..REQ-04-039`
- `AC-010`, `AC-216`, `AC-218`, `AC-353`

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
3. `U-7-06` asserts lock failure wins before stale-precondition evaluation for rollback when a protected-set lock is held.
4. `I-7-01` asserts single-entry rollback atomically updates source rows, projections, history rows, and collaboration events.
5. `I-7-03` asserts stale rollback fails closed and leaves current row state unchanged.
6. `E-7-02` later proves a reviewer can roll back one mistaken link, mention resolution, or evidence association without reverting later unrelated edits on the same row.

Implementation tasks:
- Add rollback route contract and decoder with a closed `target` union.
- Require current incident role `reviewer` or `admin`; caller without visibility receives `404`, visible underprivileged caller receives `403`.
- Reject rollback against currently soft-deleted records with `record_deleted_use_restore`.
- Resolve `history_entry_ref` only through the history selector mapping; do not accept storage primary keys or client-derived mutation-entry IDs.
- Compute the protected record set for the selected rollback target before evaluating stale row-version or target preconditions, then acquire locks through the shared destructive-operation helper.
- Use idempotency key `(actor_user_id, record_id, client_txn_id)` and normalized request comparison, including target and normalized reason.
- Implement single-entry inverse operations for mutation targets already implemented and carrying reversible before/after data. At minimum, cover row-field edits, current record-link mutations needed by existing supersede/evidence flows, existing mention resolution/dismiss/restore mutations, and existing evidence association mutations when the mutation substrate can prove a safe inverse.
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
- Planned: rollback OpenAPI request and response schemas.
- Planned: rollback error-family contract entries.
- Planned: durable single-entry rollback implementation for implemented reversible target families.
- Planned: backend coverage for `U-7-05`, `U-7-06`, `I-7-01`, and `I-7-03`.

Risks and open questions:
- Tag rollback appears in `E-7-02` and `AC-010`, but typed tags are Phase 8 scope. Before claiming tag rollback, obtain an owner decision to either pull minimal tag mutation support into Phase 7 or mark tag-specific evidence blocked/deferred while still covering link, mention, and evidence rollback.
- If earlier mutations did not persist reversible before/after values for a target family, do not infer rollback from projections or visible labels. Add the substrate or mark the target non-reversible with explicit current legality.

Exit criteria:
- Rollback route admits only documented selector shapes.
- Single-entry rollback uses server-provided selectors and never client-inferred labels or storage identifiers.
- Successful rollback is visible as a new attributed history item and through ordinary collaboration events.

## Sprint 4. Whole-Change-Set Rollback, Whole-Row Restore, Retained History, and Locks

Objective: Complete multi-entry rollback, row-backed snapshot restore, retained-history invariants, no-purge guarantees, and shared destructive-operation lock precedence.

Status: Planned.

Relevant IDs:
- `U-7-05`, `U-7-06`, `U-7-07`, `I-7-01`, `I-7-04`, `E-7-04`
- `REQ-01-089..REQ-01-111`
- `REQ-01-561..REQ-01-563`
- `REQ-02-205..REQ-02-220`
- `REQ-02-238..REQ-02-242`
- `REQ-03-101`, `REQ-03-141..REQ-03-144`
- `AC-011`, `AC-012`, `AC-187`, `AC-217`, `AC-218`, `AC-353`, `AC-383..AC-385`

Grep references:
- `target.kind = 'change_set'`
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
1. `U-7-05` asserts `change_set` rollback reverses every reversible mutation entry in reverse deterministic entry order and commits a new rollback change set.
2. `U-7-05` asserts `row_restore` restores only row-backed fields to the selected `record_revisions` snapshot and does not recreate or delete links, tags, mentions, observations, or evidence associations.
3. `U-7-06` asserts restore, rollback, and merge share identical destructive-lock precedence: active lock returns `record_locked` with `retryable=true` before stale row-version or route-specific preconditions; after lock release, stale inputs fall through to ordinary downstream errors.
4. `U-7-07` and `I-7-04` assert retained-history visibility and selector stability across incident closure, delete or restore cycles, rollback, restart, and ordinary background work.
5. `E-7-04` later proves whole-row restore creates a new attributed revision and moves the visible row back to the selected historical snapshot without erasing prior history.

Implementation tasks:
- Implement `target.kind='change_set'` rollback by loading the target change set through visible record history and reversing every currently reversible mutation entry in reverse deterministic entry order.
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
- Planned: whole-change-set rollback.
- Planned: whole-row restore.
- Planned: shared destructive-operation lock helper and merge integration.
- Planned: retained-history and no-purge evidence.
- Planned: service-backed coverage for lock precedence and restart stability.

Risks and open questions:
- Merge rollback may require additional mutation-entry detail beyond the current Phase 4 merge substrate. If the existing substrate cannot reconstruct pre-merge graph state, add that substrate before claiming `AC-217`; do not implement a partial unmerge based only on visible projections.
- Transaction lock implementation must not introduce a client-visible lock surface, lock-holder identity surface, manual unlock route, or queued destructive-operation behavior.

Exit criteria:
- Restore, rollback, and merge observe the same destructive-lock precedence.
- Retained history remains fully paginatable and stable through restart and incident closure.
- No public or operator-visible current-profile setting can narrow retained history for an extant incident.

## Sprint 5. Reviewer Workbook UI and Browser Evidence

Objective: Expose Phase 7 reviewer workflows from the workbook surface using server-provided history metadata and selectors.

Status: Planned.

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

Test-first sequence:
1. `E-7-01` opens history from a selected workbook row and asserts actor, timestamp, operation, diff summary, and legal rollback actions render from server metadata.
2. `E-7-02` rolls back one mistaken implemented target family, preferably a link, mention resolution, or evidence association, and proves later unrelated edits on the same row remain intact.
3. `E-7-03` soft-deletes and restores a row using tombstone concurrency and verifies another client observes `remove` on delete and `invalidate` on restore.
4. `E-7-04` performs whole-row restore and verifies a new attributed revision moves the visible row back to the selected historical row-backed snapshot without erasing prior history.

Implementation tasks:
- Add a workbook row-history entry point reachable in one click from the selected row. Add a shortcut only if it can be done without pulling Phase 9's broader keyboard contract forward.
- Render history as row-centric, not view-centric. The UI must not require the current visible surface to decide history identity.
- Display only actions returned by `available_rollback_actions[]`, `history_entry_ref`, and `revision_no`.
- Do not infer rollback legality from visible labels, diff-summary text, item order, SQL names, or projection storage names.
- Add confirmation or explicit action controls for destructive rollback, restore, and whole-row restore without adding a generalized approval workflow.
- On delete and restore, keep the workbook grid active and let ordinary collaboration updates remove, invalidate, or refresh rows.
- Preserve focus/selection where existing workbook state supports it; do not create a modal-only flow that forces users out of the workbook interaction model.

Validation commands:
- `make frontend-unit`
- `make frontend-typecheck`
- `make lint-biome`
- `FORCE_COLOR=0 NO_COLOR=1 tmp/node-runtime/bin/pnpm --dir apps/web exec playwright test e2e/phase7.history.spec.ts`
- `make browser-e2e-webserver-backed`
- `make phase-slice PHASE=phase7`
- `make service-backed-slice PHASE=phase7`
- `git diff --check`

Deliverables:
- Planned: reviewer history panel or inspector surface.
- Planned: rollback, soft-delete, restore, and whole-row restore browser evidence.
- Planned: frontend unit coverage for client-side history rendering and selector use.
- Planned: browser tests for `E-7-01..E-7-04`.

Risks and open questions:
- Tag rollback remains unresolved because typed tags are Phase 8. Browser `E-7-02` should use a Phase 7-supported target unless an owner decision explicitly pulls minimal tag support forward.
- The UI must not present actions optimistically from client logic if the server says the item is not currently reversible.

Exit criteria:
- Browser evidence proves the reviewer can execute Phase 7 workflows from the workbook surface.
- Client rollback requests use only server-supplied selectors and stable `record_id` values.
- Other clients observe delete, restore, and rollback through ordinary collaboration streams.

## Sprint 6. Phase Gate, Ledgers, Schedules, Baselines, and Handoff Cleanup

Objective: Prove Phase 7 is complete, resolve generated drift, update test ledgers and schedules, refresh baselines when needed, and hand off stable prerequisites to Phase 8.

Status: Planned.

Relevant IDs: all `U-7-*`, `I-7-*`, `E-7-*`; all Phase 7-owned ACs and unresolved support AC notes.

Files and areas:
- `PHASE7_IMPLEMENTATION_PLAN.md`
- `tools/phase7_test_map.json`
- `docs/testing/phase7_coverage_ledger.md`
- `tools/go_test_duration_baselines.json` if service-backed Go rows are added or promoted.
- Generated schedule manifests.
- Any generated OpenAPI or protocol outputs produced by `make generate`.

Test-first sequence:
1. Replace every Phase 7 placeholder with real assertions before claiming exit.
2. Run manifest and ledger validation before broader tests.
3. Run focused backend, frontend, browser, and public Phase 7 wrappers.
4. Run earlier-phase regression slices as needed for touched owner boundaries.
5. Run `make agent-finalize` as the required first end-of-run maintenance command before broader verification; supply `RESULTS_DIR=<successful run dir>` only when refreshing duration baselines from qualifying evidence.

Implementation tasks:
- Ensure `tools/phase7_test_map.json` has no stale placeholders, no duplicate row IDs, and correct support/forbidden file boundaries.
- Run `make phase-ledgers` and `make phase-ledger-drift`.
- Run `make phase-schedules` and `make phase-schedule-drift`.
- Run `make generate` only when source contracts or SQL changed, then prove `make generate-drift` passes.
- Refresh Go duration baselines from uncontaminated successful service-backed evidence if Phase 7 adds or promotes service-backed Go rows.
- Record artifact roots for public wrappers and gates in this plan as implementation progresses.
- Update this plan's Sprint Checklist statuses, validation results, blockers, and follow-up notes with actual evidence.

Validation commands:
- `make explain-phase PHASE=phase7`
- `make phase-ledgers`
- `make phase-ledger-drift`
- `make phase-schedules`
- `make phase-schedule-drift`
- Focused backend Phase 7 `go test` commands from Sprints 1-4
- `make frontend-unit`
- `make frontend-typecheck`
- `make lint-biome`
- Focused Phase 7 Playwright command
- `make phase-slice PHASE=phase7`
- `make service-backed-slice PHASE=phase7`
- `make test-fast`
- `make agent-finalize`
- `make check`
- `git diff --check`

Deliverables:
- Planned: all authoritative Phase 7 rows passing as real assertions.
- Planned: clean generated drift, phase-ledger drift, and phase-schedule drift.
- Planned: baseline updates where required.
- Planned: final artifact-root notes and Phase 8 handoff notes.

Risks and open questions:
- If `make check` fails outside Phase 7, record the blocker with exact target and artifact root without claiming it as Phase 7 evidence.
- If an owner decision remains unresolved, Phase 7 must either remove the corresponding claim from the manifest or mark it as an explicit non-claim. Do not silently treat unresolved guide ambiguity as completed behavior.

Exit criteria:
- All authoritative `U-7-*`, `I-7-*`, and `E-7-*` rows pass in their intended layers.
- No placeholder tests remain.
- Public `phase-slice` and `service-backed-slice` wrappers pass for Phase 7.
- Earlier touched phase slices still pass or any unrelated blocker is recorded with concrete artifacts.
- Generated drift is resolved through canonical commands.
- `make agent-finalize` has run and its outcome is recorded.
- `make check` passes or any failure is recorded as outside Phase 7 with exact evidence.

## Risks and Open Questions

1. Tag rollback is explicitly named in Phase 7 UI/AC language, but typed tags are Phase 8 scope. Required decision: either pull minimal typed-tag mutation and rollback support into Phase 7, or record tag-specific Phase 7 evidence as blocked/deferred and avoid claiming tag rollback until Phase 8.
2. `AC-184` and `AC-185` appear in `U-7-01`, but the coverage index assigns them to Phase 8. Required decision: correct the guide, mark them support-only for Phase 7, or explicitly split ownership.
3. First-class record dispatch must be explicit. Required decision: confirm the complete current `records.record_type` adapter matrix for history, delete, restore, and rollback before implementing behavior.
4. `history_entry_ref` must be stable and opaque. Default decision: persist a durable opaque selector or selector mapping. Do not use transient order, exposed database primary keys, or a key-rotation-sensitive derived token.
5. Merge rollback may need more reversible mutation-entry detail than the current Phase 4 merge path records. Required decision: inspect and, if needed, extend the merge substrate before claiming `AC-217`.
6. `record_locked` exists, but Phase 7 must add or confirm public errors for delete, restore, history, rollback, and rollback-precondition behavior. Do not rely on internal errors without contract coverage.
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
- Rollback creates new attributed `change_sets` with `source='rollback'` and never rewrites prior history.
- `history_entry_ref` values are stable for the retained-history lifetime of the record in the current deployment.
- Whole-row restore restores only row-backed fields and does not implicitly recreate or delete relationship-like mutation targets.
- Destructive-operation locks are shared by restore, rollback, and merge and fail fast before stale-precondition evaluation.
- Retained history for extant records is not narrowed by incident closure, delete or restore cycles, rollback, restart, or current-profile purge settings.

Phase 8 handoff must also state:
- Whether tag rollback was implemented in Phase 7 or remains Phase 8-owned.
- Whether `AC-184` and `AC-185` were treated as guide drift, support references, or shared ownership.
- The final first-class record adapter matrix for history, delete, restore, and rollback.
- Any mutation target families still visible in history but intentionally non-reversible, with the owner reason.
- The final artifact roots for Phase 7 public wrappers and broader verification.
