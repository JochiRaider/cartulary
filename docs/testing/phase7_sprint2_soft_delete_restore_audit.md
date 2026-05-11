---
doc_id: PHASE7-SPRINT2-SOFT-DELETE-RESTORE-AUDIT-2026-05-11
title: Phase 7 Sprint 2 Soft-Delete And Restore Audit
status: complete
role: audit
---

# Phase 7 Sprint 2 Soft-Delete And Restore Audit

## Audit Objective

Audit whether Phase 7 Sprint 2 correctly implements first-class record
soft-delete and restore before Sprint 3 begins.

Source of truth:

- Phase 7 implementation/testing guide, especially `docs/guides/cartulary_implementation_testing_guide.md` section 9.7.
- `tools/phase7_test_map.json` and generated `docs/testing/phase7_coverage_ledger.md`.
- Core 01 section 3.3.5 and section 3.3.5.0 for delete/restore routes and destructive-operation locking.
- Core 01 section 3.3.6 and section 3.3.10.1 for public errors and `record_changed`.
- Core 02 section 3, section 14, and section 15 for record envelopes, soft-delete, and append-only history.
- Core 04 `AC-181`, `AC-182`, `AC-183`, `AC-353`, and related authorization/audit checks.
- `docs/domain.md` for first-class record vocabulary.

## Verdict

`product_pass_with_validation_blocker`

Sprint 2 delete/restore behavior is sufficiently implemented for the audited
product scope. The required routes, contracts, authorization gates,
idempotency, tombstone concurrency, append-only history, projection refresh, and
ordinary collaboration events are present and covered by passing Phase 7 public
evidence.

Sprint 3 progression still has one validation blocker outside the delete/restore
product behavior: `make agent-finalize RESULTS_DIR=<full-check-run>` failed
because scheduler timing drift was detected in a prior full `check` artifact.
That issue needs owner resolution or waiver before claiming all exit criteria
are clean.

## Scope

In scope:

- `DELETE /api/v1/records/{record_id}`.
- `POST /api/v1/records/{record_id}/restore`.
- OpenAPI shape, error registry usage, success envelopes, role gates, visibility
  masking, row-version/tombstone concurrency, route-scoped idempotency,
  normalized reason comparison, source/envelope tombstones, history, projections,
  collaboration events, first-class dispatch, generated-artifact drift, and
  required Phase 7 test evidence.

Non-scope honored:

- Rollback execution.
- Whole-row restore.
- Reviewer workbook UI and full browser reviewer workflows.
- Operational backup/restore.
- Public delete/restore targeting for links, tags, mentions, observations, blobs,
  evidence associations, or administrative state.

## Findings

| ID | Severity | Finding | Evidence | Required action |
| --- | --- | --- | --- | --- |
| S2-AUD-BLOCK-001 | blocker | `agent-finalize` with retained full `check` runs refreshed duration baselines but failed at scheduler timing drift. | `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260511T015430Z-p49153` failed with `check-service-backed` warm duration `60003ms` exceeding a `60000ms` budget and one backend-integration-support lane exceeding the peer-median threshold. Closeout rerun against `.cartulary/test-results/20260511T122448Z-p62472` failed with `check-service-backed` warm duration `65388ms` exceeding the same budget. | Scheduler/timing owner must either resolve the timing drift, provide a valid replacement full-run artifact, or explicitly waive it before Sprint 3 exit criteria are fully satisfied. |
| S2-AUD-GAP-001 | evidence_gap | The Phase 7-only run root is not suitable for duration-baseline refresh because it lacks full browser timing coverage. | `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260511T020451Z-p95976` failed with missing observed browser entry timings for earlier `E-*` rows. | Use full-run artifacts for duration refresh; do not treat the Phase 7 slice artifact as sufficient for full duration baseline maintenance. |
| S2-AUD-FU-001 | follow_up_accepted | Restore locking is implemented with a route-local transaction-scoped `FOR UPDATE NOWAIT` helper rather than a broader shared lock abstraction. | `internal/modules/revisions/delete_restore_store.go` uses `lockRecordEnvelopeNowaitTx` only for restore; Core 01 §3.3.5.0 defines restore's protected set as exactly the target `record_id`. Tests cover fail-fast `record_locked`. | Product/API owner accepts `lockRecordEnvelopeNowaitTx` as the shared destructive-operation helper for the current one-record restore protected set. Broader protected-set abstraction is deferred until rollback/merge locking expands. |

No delete/restore product blockers were found.

## Contract Checklist

| Check | Result | Evidence |
| --- | --- | --- |
| Required routes exist | pass | `internal/modules/revisions/routes.go` registers `DELETE /api/v1/records/{record_id}` and `POST /api/v1/records/{record_id}/restore`; OpenAPI has `deleteRecord` and `restoreRecord`. |
| Path-only record scope | pass | OpenAPI request body uses `RecordDeleteRestoreRequest`; no `incident_id` or `view_schema_id` member is admitted. |
| Request members | pass | `RecordDeleteRestoreRequest` has exactly required `base_row_version`, required `client_txn_id`, and optional `reason`; decoder rejects unknown members. |
| Success envelope | pass | `RecordDeleteRestoreEnvelope` wraps `RecordDeleteRestoreData` with required `record_id`, `incident_id`, `row_version`, `deleted`, `deleted_at`, `deleted_by_user_id`, and `change_set_id`. |
| Error envelope coverage | pass | Route maps invalid payload, authorization, not-found masking, `row_version_conflict`, `client_txn_conflict`, `record_already_deleted`, `record_not_deleted`, and `record_locked`. |
| `record_locked.retryable=true` | pass | Public error writer marks `record_locked` retryable; U-7-04 asserts retryable `true`. |

## Delete Checks

| Check | Result | Evidence |
| --- | --- | --- |
| Requires valid JSON, `base_row_version >= 1`, and non-empty `client_txn_id` | pass | `DecodeDeleteRestoreRequest` enforces all three and returns `invalid_mutation_payload`. |
| Role gate allows `editor`, `reviewer`, `admin` | pass | Route handler allows these roles; `U-7-03` covers `viewer` failure and `editor` success path. |
| Visibility masking returns `404` | pass | Route derives incident from authoritative `record_id`; missing/invisible record path returns current route convention `incident_not_found`. |
| Stale row version fails closed | pass | `U-7-03` asserts `409 row_version_conflict` with base/current details. |
| Success tombstones envelope and increments version | pass | `U-7-03` asserts `deleted=true`, row version `2`, non-null tombstone attribution. |
| Already deleted fails except replay | pass | `U-7-03` asserts `record_already_deleted`; replay returns original `change_set_id` and no second mutation. |
| Patch after delete fails | pass | `U-7-03` asserts `PATCH` returns `record_deleted_use_restore`. |
| Ordinary delete does not use destructive lock path | pass | Store calls the NOWAIT lock helper only when `deleting=false`; delete remains on optimistic row-version evaluation plus ordinary transactional update locking. |

## Restore Checks

| Check | Result | Evidence |
| --- | --- | --- |
| Requires current tombstone row version | pass | `U-7-04` deletes first, uses tombstone version for success, and stale tombstone version for `row_version_conflict`. |
| Role gate allows only `reviewer` and `admin` | pass | Route handler requires `reviewer|admin`; `U-7-04` asserts `editor` receives `403`. |
| Success clears current tombstone and increments version | pass | `U-7-04` asserts `deleted=false`, `deleted_at=null`, `deleted_by_user_id=null`, and next row version. |
| Active record restore fails except replay | pass | `U-7-04` asserts `record_not_deleted`; replay of the original restore returns the original `change_set_id`. |
| Fail-fast destructive lock | pass | Restore calls `lockRecordEnvelopeNowaitTx` before stale-row-version evaluation; `U-7-04` asserts stale request under held row lock returns retryable `record_locked`, then stale after release returns `row_version_conflict`. |
| Restore is not whole-row restore or rollback | pass | Request schema admits no snapshot, target, revision, or `change_set_id` selector. |

## Idempotency And Reason Normalization

| Check | Result | Evidence |
| --- | --- | --- |
| Route-scoped key | pass | Delete uses `RouteKey=records.delete`; restore uses `RouteKey=records.restore`; scope key is `record_id`; actor and `client_txn_id` are included. |
| Exact replay returns original success | pass | `U-7-03` and `U-7-04` assert replay returns original `change_set_id` and creates no second mutation. |
| Divergent same key fails before stale evaluation | pass | `U-7-03` asserts divergent reason under same delete key returns `client_txn_conflict`. |
| Omission/null/normalized-empty reason equivalence | pass | Normalization converts omission, `null`, and normalized-empty to `nil`; `U-7-03` verifies normalized-empty persists as `null`, and `U-7-04` verifies replay equivalence between explicit `null` and omission. |
| Non-empty reason persistence | pass | `U-7-03` asserts non-empty delete reason persistence after `reason_note_v1` normalization; `U-7-04` asserts non-empty restore reason persistence after `reason_note_v1` normalization. | Evidence added as Sprint 2 audit closeout coverage. |

## Data And History Checks

| Check | Result | Evidence |
| --- | --- | --- |
| Delete creates attributed `change_set` | pass | `I-7-01` asserts `source='records.delete'` and actor attribution. |
| Restore creates attributed `change_set` | pass | `I-7-01` asserts `source='records.restore'` and actor attribution. |
| Reversible mutation entries | pass | `I-7-01` asserts `operation_kind='soft_delete'` and `operation_kind='restore'` entries with `target_kind='record'`. |
| Record revisions append | pass | `U-7-04` and `I-7-01` assert delete and restore revisions are appended and history is newest-first append-only. |
| Prior history is preserved | pass | `I-7-01` asserts `restore` then prior `soft_delete`; `U-7-04` asserts one delete and one restore mutation remain. |
| Envelope tombstone state | pass | `U-7-03` and `U-7-04` query `records` tombstone fields directly. |
| Source tombstone state where applicable | pass | `I-7-01` verifies indicator source tombstone set and cleared. |
| No hard deletion of history/source/link/blob rows | pass_by_evidence | Tests verify append-only counts for change-set mutations and revisions and do not observe hard deletion. Static inspection shows delete/restore uses updates plus inserts, not deletes. |

## Projection And Collaboration Checks

| Check | Result | Evidence |
| --- | --- | --- |
| Deleted record removed from ordinary view queries | pass | `U-7-03` checks host projection; `I-7-01` checks ordinary host view query returns no rows. |
| Restored record returns to ordinary view eligibility | pass | `U-7-04` checks projection row; `I-7-01` checks ordinary host view query returns the row. |
| Projection rebuild/invalidation occurs | pass | Store rebuilds current incident Timeline, Hosts, Identities, Indicators, and Assessments projections; tests cover Hosts and Indicators. |
| Delete emits ordinary `record_changed` remove | pass | `I-7-01` asserts delete event `change_kind='remove'`. |
| Restore emits ordinary `record_changed` invalidate | pass | `I-7-01` asserts restore event `change_kind='invalidate'`. |
| No Phase 7-specific event family | pass | WebSocket contract still lists `record_changed`; no delete/restore-specific event type was introduced. |
| Required collaboration arrays | pass | `RecordChangePayload` always includes `changed_field_keys` and `affected_views`; `I-7-01` asserts empty `changed_field_keys` and expected affected view. |
| Affected view duplicate/canonical ordering | pass_by_shape | Current delete/restore emits exactly one affected base view per event, so duplicates and ordering cannot occur in Sprint 2 emission. |

## First-Class Record Dispatch

Current first-class record types from Core 02/domain:

| Record type | Adapter result |
| --- | --- |
| `timeline_event` | present |
| `host` | present |
| `identity` | present |
| `party` | present |
| `indicator` | present |
| `artifact` | present, with artifact-type view dispatch |
| `task_request` | present |
| `decision` | present |
| `evidence` | present |
| `assessment` | present |

`TestPhase7_DeleteRestoreAdapterMatrix_U_7_03_U_7_04` covers the matrix.
Non-record mutation target misuse is covered by `U-7-03` with a seeded
`record_tags` target returning not-found masking; the route only loads from
`records`, so links, tags, mentions, observations, blobs, and administrative
rows cannot be first-class targets without a record envelope.

## Required Coverage

| Required evidence | Result | Notes |
| --- | --- | --- |
| `U-7-03` | pass | Included in `backend-store` Phase 7 public evidence. |
| `U-7-04` | pass | Included in `backend-store` Phase 7 public evidence. |
| Delete/restore portions of `I-7-01` | pass | Included in `backend-integration` Phase 7 public evidence. |
| Restore portion of `I-7-03` | pass | Included in `backend-integration` Phase 7 public evidence. |
| Later browser support fixture for `E-7-03` | pass | `phase-slice` and `service-backed-slice` both included `browser-e2e-webserver-backed` and passed; full reviewer workflow remains out of scope. |

## Validation Commands

| Command | Result |
| --- | --- |
| `make task-guide ROLE=feature-dev PHASE=phase7` | pass; reported Phase 7 has 15 authoritative rows across `backend_store`, `backend_integration`, and `browser_functional`. |
| `make explain-phase PHASE=phase7` | pass; confirmed manifest, ledger, owners, dependencies, and latest artifact paths. |
| `make explain-target TARGET=backend-store DETAIL=rows` | pass; listed `U-7-03`, `U-7-04`, and related Phase 7 rows. |
| `make explain-target TARGET=backend-integration DETAIL=rows` | pass; listed `I-7-01`, `I-7-03`, and related Phase 7 rows. |
| `make phase-slice PHASE=phase7` | pass; 3/3 work units, 44 tests, 0 failed, 0 missing, 0 unmapped; run root `.cartulary/test-results/20260511T020427Z-p93092`. |
| `make service-backed-slice PHASE=phase7` | pass; 3/3 work units, 44 tests, 0 failed, 0 missing, 0 unmapped; run root `.cartulary/test-results/20260511T020451Z-p95976`. |
| `make test-fast` | pass; 450 tests, 0 failed, 0 missing; run root `.cartulary/test-results/20260511T020539Z-p1848`. |
| `make lint` | pass; run root `.cartulary/test-results/20260511T020631Z-p10289`. |
| `make generate-drift` | pass; latest run root `.cartulary/test-results/20260511T021040Z-p16146`. |
| `make phase-ledger-drift` | pass; latest run root `.cartulary/test-results/20260511T021040Z-p16340`. |
| `make phase-schedule-drift` | pass before and after finalization attempts; latest run root `.cartulary/test-results/20260511T021040Z-p16383`. |
| `make agent-finalize` | pass without `RESULTS_DIR`; duration baseline refresh skipped; run root `.cartulary/test-results/20260511T020515Z-p99304`. |
| `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260511T020451Z-p95976` | fail; Phase 7-only run root lacks full browser timing entries. |
| `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260511T015430Z-p49153` | fail; duration baselines refreshed, but scheduler timing drift was detected. |
| `make go-test-duration-baseline-drift RESULTS_DIR=.cartulary/test-results/20260511T015430Z-p49153` | pass; run root `.cartulary/test-results/20260511T020806Z-p15002`. |
| `make browser-e2e-duration-baseline-drift RESULTS_DIR=.cartulary/test-results/20260511T015430Z-p49153` | pass; run root `.cartulary/test-results/20260511T020806Z-p15077`. |

Focused `go test` diagnostics were not needed because the public Phase 7
wrappers passed.

## Drift And Generated Artifacts

- `generate-drift`, `phase-ledger-drift`, and `phase-schedule-drift` passed.
- `agent-finalize` without `RESULTS_DIR` passed and completed schedule/json-shape
  maintenance with duration refresh skipped; its stdout reported updates to
  `tools/scheduler_manifest.json` and `tools/execution_topology_render_index.json`.
- The failed full-run `agent-finalize RESULTS_DIR` attempt refreshed duration
  baselines before failing at scheduler timing drift; follow-up drift checks for
  Go and browser duration baselines passed against that same full-run root.
- Closeout reran full `make check` and `agent-finalize
  RESULTS_DIR=.cartulary/test-results/20260511T122448Z-p62472`; finalization
  again failed only at scheduler warm-duration drift. Standalone Go, browser,
  service-backed target, and harness-smoke duration-baseline drift checks passed
  against that fresh full-run root. The failed-refresh unstaged generated timing
  outputs were discarded; the existing staged audit timing artifacts remain
  pending scheduler/timing owner waiver or a future passing finalization.

## Exit Criteria

| Criterion | Result |
| --- | --- |
| Required public routes conform to Core/OpenAPI | pass |
| Delete and restore pass authorization, visibility, row-version, tombstone, lock, and idempotency checks | pass |
| Append-only history, mutation entries, change sets, and revisions are correct | pass |
| Ordinary queries remove deleted records and re-include restored records | pass |
| Collaboration emits ordinary `record_changed` with `remove` and `invalidate` semantics | pass |
| Every first-class record type has adapter coverage or owner decision | pass |
| Required Sprint 2 test rows are evidenced | pass |
| Required validation and drift commands pass or failures are owner-routed | block pending `S2-AUD-BLOCK-001` |

Sprint 2 delete/restore implementation is ready on product behavior. Sprint 3
should not claim fully clean exit criteria until the scheduler timing-drift
validation blocker is resolved or explicitly waived.
