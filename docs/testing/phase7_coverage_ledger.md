# Phase 7 Coverage Ledger

This ledger is generated from `tools/phase7_test_map.json`. Update the manifest row metadata first, then regenerate this file.

- Scope: Reviewer-facing history, delete or restore, and rollback.
- Normative owners: Core 01 §3.3.4.2 and §3.3.5; Core 02 §14–§15; Core 03 §10; Core 04 §2–§3.
- Authority: `tools/phase7_test_map.json` is the enforced Phase 7 traceability source. This ledger is a rendered companion and does not control the mechanical row inventory.
- Phase 7 remains registered as planned. This manifest records ownership and placeholder evidence locations before feature work, but public phase wrappers must not execute Phase 7 rows until the registry status is intentionally activated.
- `AC-184` and `AC-185` appear in the Phase 7 guide row `U-7-01`, but the guide coverage index assigns those ACs to Phase 8. Sprint 0 records them as ambiguous support references, not completed Phase 7 claims.
- Typed tag rollback language is treated as an owner-decision dependency; Phase 7 placeholders must not claim Phase 8 typed tag creation or mutation behavior.

## Authoritative Execution

- `backend-store` will carry authoritative `U-7-*` route and store-domain rows through manifest-owned service-backed Go selection after Phase 7 activation.
- `backend-integration` will carry authoritative `I-7-*` real-runtime rows for source state, projections, history rows, collaboration events, pagination, restart, and incident-closure behavior after Phase 7 activation.
- `browser-e2e-webserver-backed` will carry authoritative `E-7-*` browser-functional rows through manifest-entry shards under `browser_functional` after Phase 7 activation.
- `frontend-unit` may carry supplemental component scaffolding for the reviewer history surface, but frontend-unit rows do not replace authoritative browser `E-7-*` evidence.

## Support-Only Execution

- Phase 3 through Phase 6 mutation, projection, idempotency, conflict, evidence, merge, and collaboration tests remain carryover support only and must not claim Phase 7 identifiers.
- When Phase 7 activates, missing duration baselines must be refreshed from clean successful service-backed and browser evidence rather than hidden behind fallback weights.

## Unit

| Row | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- |
| `U-7-01` | `internal/modules/revisions/phase7_history_test.go::TestPhase7_RecordHistoryEnvelope_U_7_01` | `backend_store` | `GET /api/v1/records/{record_id}/history` returns the Phase 7 record-scoped history envelope with newest-first deterministic items, deleted-row tombstone `row_version`, and canonical `available_rollback_actions[]` ordering. | `AC-184` and `AC-185` remain ambiguous Phase 8-assigned support references until an owner decision resolves the guide drift. |
| `U-7-02` | `internal/modules/revisions/phase7_history_test.go::TestPhase7_HistoryEntryRefStability_U_7_02` | `backend_store` | `history_entry_ref` is present only for a logical history item that maps to exactly one reversible mutation target, stays opaque to clients, and remains stable across repeated reads. | Whole-change-set and whole-row rollback execution are covered by rollback rows. |
| `U-7-03` | `internal/modules/revisions/phase7_delete_restore_test.go::TestPhase7_SoftDeleteRoutePreconditions_U_7_03` | `backend_store` | First-class record soft-delete requires current `row_version`, respects role gates, and returns route-owned failures including `record_deleted_use_restore`, `record_not_deleted`, and applicable `record_locked` behavior. | Restore append-only consequences and destructive lock precedence are separate Phase 7 rows. |
| `U-7-04` | `internal/modules/revisions/phase7_delete_restore_test.go::TestPhase7_RestoreTombstonePreconditions_U_7_04` | `backend_store` | Restore requires the current tombstone `row_version`, respects role gates, returns the record to active state, and never mutates prior history rows in place. | Browser-visible delete and restore collaboration behavior is covered by `E-7-03`. |
| `U-7-05` | `internal/modules/revisions/phase7_rollback_test.go::TestPhase7_RollbackSelectorUnion_U_7_05` | `backend_store` | Rollback accepts only `history_entry`, `change_set`, and `row_restore` targets and creates a new `change_set` with `source='rollback'` rather than mutating prior history. | Typed tag creation or mutation semantics remain outside Sprint 0 and must not be claimed until an owner decision pulls them into Phase 7. |
| `U-7-06` | `internal/modules/revisions/phase7_locks_test.go::TestPhase7_DestructiveOperationLocks_U_7_06` | `backend_store` | Restore, rollback, and merge use shared destructive-operation lock precedence, fail fast before stale-precondition evaluation, and replay ordinary downstream route errors after lock release. | Ordinary soft-delete and ordinary patch stay outside the destructive-operation family. |
| `U-7-07` | `internal/modules/revisions/phase7_history_test.go::TestPhase7_RetainedHistoryInvariants_U_7_07` | `backend_store` | Retained history for extant records remains fully paginatable and preserves issued `history_entry_ref` values across incident closure, delete or restore cycles, rollback, and restart. | The current profile defines no public history-purge route or retention-horizon setting for extant incidents. |

## Integration

| Row | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- |
| `I-7-01` | `internal/modules/revisions/phase7_integration_test.go::TestPhase7_DeleteRestoreRollbackAtomicConsequences_I_7_01` | `backend_integration` | Delete, restore, and rollback update source rows, projections, history rows, and emitted collaboration events atomically. | Single-route request-shape validation remains unit or store-domain evidence. |
| `I-7-02` | `internal/modules/revisions/phase7_integration_test.go::TestPhase7_HistoryPaginationRecordBinding_I_7_02` | `backend_integration` | History pagination remains bound to `record_id`, rejects cross-record cursor replay, and preserves deterministic item ordering across pages. | General pagination registry behavior remains existing platform support evidence. |
| `I-7-03` | `internal/modules/revisions/phase7_integration_test.go::TestPhase7_StaleRestoreRollbackFailsClosed_I_7_03` | `backend_integration` | A stale restore or rollback precondition fails closed and never mutates current row state. | Lock acquisition precedence is covered by `U-7-06`. |
| `I-7-04` | `internal/modules/revisions/phase7_integration_test.go::TestPhase7_RetainedHistoryAcrossRestartAndClosure_I_7_04` | `backend_integration` | History for an extant record remains fully paginatable and stable across service restart, incident closure, delete or restore, and rollback, with prior `history_entry_ref` values preserved. | Incident portability selector preservation across deployments is outside Phase 7. |

## Browser E2E

| Row | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- |
| `E-7-01` | `apps/web/e2e/phase7.history.spec.ts::E-7-01 opens row history from the workbook surface with legal rollback actions` | `browser_functional` | A reviewer opens row history from the workbook surface and sees actor, timestamp, operation, diff summary, and server-provided legal rollback actions. | History route envelope construction remains backend evidence. |
| `E-7-02` | `apps/web/e2e/phase7.history.spec.ts::E-7-02 rolls back one mistaken mutation without reverting later unrelated edits` | `browser_functional` | A reviewer rolls back one mistaken link, tag, mention resolution, or evidence association without reverting later unrelated edits on the same row. | Typed tag mutation semantics remain owner-decision dependent and must not be inferred from this placeholder. |
| `E-7-03` | `apps/web/e2e/phase7.history.spec.ts::E-7-03 soft-deletes and restores a row with tombstone concurrency` | `browser_functional` | A reviewer soft-deletes and restores a row using tombstone concurrency, while other clients observe `remove` on delete and `invalidate` on restore. | Low-level destructive lock precedence remains backend evidence. |
| `E-7-04` | `apps/web/e2e/phase7.history.spec.ts::E-7-04 whole-row restore appends a new attributed revision` | `browser_functional` | Whole-row restore creates a new attributed revision and moves the visible row back to the selected historical snapshot without erasing prior history. | Whole-change-set rollback internals remain backend evidence. |

## Shared Harness Coverage

| Harness | Phase 7 evidence |
| --- | --- |
| Planned manifest ownership | `tools/phase7_test_map.json` records every authoritative guide row before product behavior exists. |
| Generated ledger | `docs/testing/phase7_coverage_ledger.md` is generated from this manifest and must not be hand-edited. |
| Schedule boundary | Phase 7 remains planned, so generated schedules must not add Phase 7 execution work until `tools/phase_registry.json` explicitly activates the phase. |

## Support-Only Evidence

- `internal/modules/timeline`, `internal/modules/entities`, `internal/modules/evidence`, `internal/modules/workbook`, and `internal/modules/collaboration` Phase 3 through Phase 6 tests provide substrate support only.
- `apps/web/src/WorkbookShell.phase7.test.tsx` is supplemental UI scaffolding for future reviewer-history controls and must not replace authoritative `E-7-*` browser rows.
- Existing browser specs under `apps/web/e2e/phase4*`, `phase5*`, and `phase6*` remain prior-phase evidence and are forbidden from claiming Phase 7 IDs.
