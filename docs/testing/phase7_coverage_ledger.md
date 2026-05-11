# Phase 7 Coverage Ledger

This ledger is generated from `tools/phase7_test_map.json`. Update the manifest row metadata first, then regenerate this file.

- Scope: Reviewer-facing history, delete or restore, and rollback.
- Normative owners: Core 01 §3.3.4.2 and §3.3.5; Core 02 §14–§15; Core 03 §10; Core 04 §2–§3.
- Authority: `tools/phase7_test_map.json` is the enforced Phase 7 traceability source. This ledger is a rendered companion and does not control the mechanical row inventory.
- Phase 7 is active so public phase wrappers execute the Phase 7 manifest. Sprint 1 completion is limited to `GET /api/v1/records/{record_id}/history`; later delete, restore, rollback, lock, and browser workflow rows execute only non-behavioral scope sentinels until their owning sprint.
- Sprint 1 accepts fixture-seeded delete, restore, and rollback-like state changes only as retained-history read-route invariants before the owning mutation routes exist.
- `AC-184` and `AC-185` appear in the Phase 7 guide row `U-7-01`, but the guide coverage index assigns those ACs to Phase 8. Phase 7 records them as support references only and does not claim them as completed Phase 7 evidence.
- Adjunct mutation-target-family history completeness is deferred to later Phase 7 rollback/UI work; Sprint 1 claims only row-backed record-history read evidence.
- Typed tag rollback language is treated as an owner-decision dependency; Phase 7 scope sentinels must not claim Phase 8 typed tag creation or mutation behavior.

## Authoritative Execution

- `backend-store` carries authoritative `U-7-*` route and store-domain rows through manifest-owned service-backed Go selection.
- `backend-integration` carries authoritative `I-7-*` real-runtime rows for source state, projections, history rows, collaboration events, pagination, restart, and incident-closure behavior.
- `browser-e2e-webserver-backed` carries authoritative `E-7-*` browser-functional rows through manifest-entry shards under `browser_functional`.
- `frontend-unit` may carry supplemental component scaffolding for the reviewer history surface, but frontend-unit rows do not replace authoritative browser `E-7-*` evidence.

## Support-Only Execution

- Phase 3 through Phase 6 mutation, projection, idempotency, conflict, evidence, merge, and collaboration tests remain carryover support only and must not claim Phase 7 identifiers.
- Missing duration baselines must be refreshed from clean successful service-backed and browser evidence rather than hidden behind fallback weights.

## Unit

| Row | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- |
| `U-7-01` | `internal/modules/revisions/phase7_history_test.go::TestPhase7_RecordHistoryEnvelope_U_7_01`, `TestPhase7_RecordHistoryOpenAPIContract_U_7_01` | `backend_store` | `GET /api/v1/records/{record_id}/history` returns the Phase 7 record-scoped history envelope with newest-first deterministic items, deleted-row tombstone `row_version`, and canonical `available_rollback_actions[]` ordering. | `AC-184` and `AC-185` remain ambiguous Phase 8-assigned support references until an owner decision resolves the guide drift. |
| `U-7-02` | `internal/modules/revisions/phase7_history_test.go::TestPhase7_HistoryEntryRefStability_U_7_02` | `backend_store` | `history_entry_ref` is present only for a logical history item that maps to exactly one reversible mutation target, stays opaque to clients, and remains stable across repeated reads. | Whole-change-set and whole-row rollback execution are covered by rollback rows. |
| `U-7-03` | `internal/modules/revisions/phase7_delete_restore_test.go::TestPhase7_SoftDeleteRoutePreconditions_U_7_03`, `TestPhase7_DeleteRestoreAdapterMatrix_U_7_03_U_7_04` | `backend_store` | Sprint 2 covers the first-class record soft-delete route contract, role gates, visibility masking, optimistic concurrency, route-scoped idempotency, deleted-record mutation failure, and explicit adapter matrix. | Rollback execution, whole-row restore, reviewer workbook UI, and non-record mutation target deletion remain outside Sprint 2. |
| `U-7-04` | `internal/modules/revisions/phase7_delete_restore_test.go::TestPhase7_RestoreTombstonePreconditions_U_7_04` | `backend_store` | Sprint 2 covers the restore route contract, reviewer/admin role gate, tombstone row-version requirement, active-state projection eligibility, route-scoped idempotency, and append-only delete/restore history. | Rollback execution, whole-row restore, reviewer workbook UI, and full browser reviewer workflow remain outside Sprint 2. |
| `U-7-05` | `internal/modules/revisions/phase7_rollback_test.go::TestPhase7_RollbackSelectorUnion_U_7_05` | `backend_store` | Sprint 1 records U-7-05 as a later-sprint scope sentinel only; rollback selector and execution behavior are not claimed by Sprint 1. | Rollback target validation, `source='rollback'` change sets, prior-history immutability, and typed tag mutation semantics remain later owner-controlled work. |
| `U-7-06` | `internal/modules/revisions/phase7_locks_test.go::TestPhase7_DestructiveOperationLocks_U_7_06` | `backend_store` | Sprint 1 records U-7-06 as a later-sprint scope sentinel only; destructive-operation lock behavior is not claimed by Sprint 1. | Restore, rollback, and merge lock precedence and downstream replay behavior remain later Phase 7 work. |
| `U-7-07` | `internal/modules/revisions/phase7_history_test.go::TestPhase7_RetainedHistoryInvariants_U_7_07` | `backend_store` | Retained history for extant records remains fully paginatable and preserves issued `history_entry_ref` values across incident closure, delete or restore cycles, rollback, and restart. | The current profile defines no public history-purge route or retention-horizon setting for extant incidents. |

## Integration

| Row | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- |
| `I-7-01` | `internal/modules/revisions/phase7_integration_test.go::TestPhase7_DeleteRestoreRollbackAtomicConsequences_I_7_01` | `backend_integration` | Sprint 2 covers delete and restore atomic consequences across envelope/source tombstones, projections, change sets, mutation rows, record revisions, retained history, and ordinary record_changed collaboration events. | Rollback atomicity and whole-row restore remain later Phase 7 work. |
| `I-7-02` | `internal/modules/revisions/phase7_integration_test.go::TestPhase7_HistoryPaginationRecordBinding_I_7_02` | `backend_integration` | History pagination remains bound to `record_id`, rejects cross-record cursor replay, and preserves deterministic item ordering across pages. | General pagination registry behavior remains existing platform support evidence. |
| `I-7-03` | `internal/modules/revisions/phase7_integration_test.go::TestPhase7_StaleRestoreRollbackFailsClosed_I_7_03` | `backend_integration` | Sprint 2 covers stale restore failure as a closed transaction: no envelope/source state, projection, history, idempotency, or collaboration event changes. | Stale rollback failure behavior remains later Phase 7 work. |
| `I-7-04` | `internal/modules/revisions/phase7_integration_test.go::TestPhase7_RetainedHistoryAcrossRestartAndClosure_I_7_04` | `backend_integration` | History for an extant record remains fully paginatable and stable across service restart, incident closure, delete or restore, and rollback, with prior `history_entry_ref` values preserved. | Incident portability selector preservation across deployments is outside Phase 7. |

## Browser E2E

| Row | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- |
| `E-7-01` | `apps/web/e2e/phase7.history.spec.ts::E-7-01 opens row history from the workbook surface with legal rollback actions` | `browser_functional` | Sprint 1 records E-7-01 as a later-sprint scope sentinel only; reviewer workbook history UI behavior is not claimed by Sprint 1. | Workbook row-history UI, legal rollback actions, and browser presentation remain later Phase 7 work. |
| `E-7-02` | `apps/web/e2e/phase7.history.spec.ts::E-7-02 rolls back one mistaken mutation without reverting later unrelated edits` | `browser_functional` | Sprint 1 records E-7-02 as a later-sprint scope sentinel only; browser rollback behavior is not claimed by Sprint 1. | Single-entry rollback UI behavior and typed tag mutation semantics remain later owner-controlled work. |
| `E-7-03` | `apps/web/e2e/phase7.history.spec.ts::E-7-03 soft-deletes and restores a row with tombstone concurrency` | `browser_functional` | Sprint 2 provides HTTP and WebSocket support fixtures for later browser soft-delete/restore validation without claiming reviewer workbook UI behavior. | The full browser reviewer workflow, workbook controls, and UI presentation remain later Phase 7 work. |
| `E-7-04` | `apps/web/e2e/phase7.history.spec.ts::E-7-04 whole-row restore appends a new attributed revision` | `browser_functional` | Sprint 1 records E-7-04 as a later-sprint scope sentinel only; browser whole-row restore behavior is not claimed by Sprint 1. | Whole-row restore UI behavior and whole-change-set rollback internals remain later Phase 7 work. |

## Shared Harness Coverage

| Harness | Phase 7 evidence |
| --- | --- |
| Active manifest ownership | `tools/phase7_test_map.json` records every authoritative guide row while later-sprint scope sentinels remain explicit. |
| Generated ledger | `docs/testing/phase7_coverage_ledger.md` is generated from this manifest and must not be hand-edited. |
| Schedule boundary | Phase 7 is active, so generated schedules include Phase 7 execution work selected by public phase wrappers. |

## Support-Only Evidence

- `internal/modules/timeline`, `internal/modules/entities`, `internal/modules/evidence`, `internal/modules/workbook`, and `internal/modules/collaboration` Phase 3 through Phase 6 tests provide substrate support only.
- `apps/web/src/WorkbookShell.phase7.test.tsx` is supplemental UI scaffolding for future reviewer-history controls and must not replace authoritative `E-7-*` browser rows.
- Existing browser specs under `apps/web/e2e/phase4*`, `phase5*`, and `phase6*` remain prior-phase evidence and are forbidden from claiming Phase 7 IDs.
