# Network Flow F-18–F-23 release handoff

This change completes S-09–S-16 in the
[controlling tracker](networkflow-module-refactor-tracker.md#16-f-18f-23-execution-ledger).
Network Flow public major **6**, durable state **4**, valid materialization payload
**v1**, identifiers, digest bytes and historical receipts remain compatible.
There is no SQL migration or persisted digest rewrite.

## Product and internal changes

- Safe digests require a valid explicit key ID, exactly 32 key bytes and a valid
  value class. Missing or invalid material returns an error; the fixed engineering
  key and key ID are removed. The key ring still owns rotation and active-key policy.
- Table and indicator-link mutations use private application components, typed
  commands/outcomes, semantic failures and separate receipt adapters. HTTP remains
  responsible for framing/presentation. Current admission and each command family's
  distinct replay policy remain intact.
- Workers, restore reconciliation and startup admission share strict payload
  decoding. Duplicate/unknown/missing/null/mistyped fields, trailing data, invalid
  identities, nonpositive generations and incident mismatches are rejected before
  a payload can identify a declaration. Malformed retained state remains unchanged
  and cannot become ready; no silent repair or compatibility decoder is added.
- Table-ID and cursor-nonce entropy are independent, construction-time readers
  defaulting to production entropy. Table-ID allocation retries only primary-key
  collisions using `ON CONFLICT ... DO NOTHING RETURNING`, with the existing
  eight-attempt bound. Other database constraints remain errors.

## Removed surfaces and caller migration

| Removed surface | Replacement / caller action |
| --- | --- |
| `sortRows`, `pageFlowRowsAfter`, `pageDiagnosticsAfter`, exclusive `compareRowFieldForSort` | Use the live bounded SQL queries; PostgreSQL assertions replace obsolete algorithm tests. Independently used comparators remain. |
| `parseTimestamp`, `decodeIndicatorLinkRequest`, `networkFlowLinkableIPField`, `composeGraphSourceFromSemanticHTTP`, `graphProjectionFailedForContextHTTP` | Tests use the live parser with record context, route/object admission and selector policy. Fixture request decoding is test-only. |
| Unused `store.GetTable` and unbounded `store.ListRejectedRowDiagnostics` | Existing authorized reads and bounded diagnostic queries remain. |
| Production `store.CreateTable`, `RenameTable`, `SoftDeleteTable`, `RetainedCounts` conveniences | Fixture operations live in `_test.go`; external integration bridges retain necessary fixture access. Transaction-bound owner methods remain production code. |
| Source-profile v1 schema/index root, generated `SourceProfileList` and exclusive `EffectiveLimits` exports | All callers consume `cartulary.network_flow.source_profile_list.v2` and its current limits. Shared `SourceProfile` remains supported. Generated Go/TypeScript artifacts and Import registry fingerprints ship together. |
| Four `cartulary.test.network_flow_*_control.v1` schemas and attachments | Use the matching v2 projections, fixtures and consumers. No dual-schema/token aliases are provided. |

## Harness v2 migration

The four families remain functional. Shared enablement, host/origin and token
checks run before strict control decoding. Controls are instance scoped; consumed
controls are removed before effects, mismatches remain pending and process
replacement clears state. Ordinary builds have no control routes or registry
dependencies. Auth/audit fixture references must bind to verified rows in the
owned runtime before arming; a bare control registry is not product evidence.

- **Faults:** retain only `network_flow.import.before_owner_apply`,
  `network_flow.import.after_owner_apply`,
  `network_flow.import.before_transaction_commit`,
  `network_flow.import.after_transaction_commit_before_reply`,
  `network_flow.worker.before_handler_start`,
  `network_flow.worker.before_cancellation_check`,
  `network_flow.worker.before_final_commit`, and
  `network_flow.worker.after_completed_publication`. Retire preparation,
  before-apply-start, between-commit/publication, acknowledgment and replay
  reconciliation tokens. Fixtures now decorate real Import/Jobs dependencies.
  Precommit errors, panic and cancellation prove rollback; postcommit errors
  preserve committed outcomes. Crash controls terminate only an owned child.
- **Randomness:** retain `network_flow.table_id` (`uuid`, 16 bytes) and
  `network_flow.cursor_nonce` (`hex_bytes`, exactly 12 bytes). Remove row-ID,
  diagnostic-ID, import-job-ID, import-source-ref, safe-digest-nonce and
  graph-invocation-ID streams, the `token` value kind and string consumer.
  Armed sequence exhaustion fails closed.
- **Authorization:** retain `network_flow.route.before_authorization` and
  `network_flow.cursor.before_authorization_recheck`. Remove later route,
  publication and synthetic fixture checkpoints, `extension_claim_removed`,
  `hidden_response_kind`, `must_not_disclose_resource` and hidden-response enums.
  Replace `incident_soft_deleted` with `incident_deleted`: Incidents has no soft
  delete state; only disposable owned fixture incidents without retained
  references can be removed. Membership revocation/restoration, session
  revocation, table rename and table soft deletion remain. Fixtures change real
  authority/state before an ordinary authenticated request and assert its actual
  owner-defined response.
- **Audit:** retain six event codes and their actual table/graph/binding resources.
  Remove `network_flow_import` and `expected_replay_increment`. Table creation
  through Imports is asserted against its resulting table. Consumers read a real
  committed baseline and final count scoped by actor, incident, operation, event,
  resource and correlation. Graph query rejects `no_audit_replay`; supported
  idempotent operations prove no extra occurrence. Failed or unconsumed required
  assertions fail the fixture.

## Validation, rollout and rollback

The tracker records per-slice assertions, exact commands, run roots, corrected
fixture failures and final gates. Earlier successful runs are historical only.
`make agent-finalize` ran before final gates; retained-run maintenance was skipped
because `RESULTS_DIR` was unset. Markdown remains outside executable evidence.

Release code, generated contracts and matching harness fixtures together.
Rollback requires compatible code that preserves the strict payload boundary,
explicit-key requirement, immutable receipts and existing supported recovery
history. Do not restore the engineering fallback, weaken retained-state admission,
or reactivate retired controls to make rollback appear compatible.

Repository validation is complete when the S-16 record says complete. Operational
deployment and the previous telemetry dashboard/alert cutover were not executed;
the operational owner must perform and evidence those separately.
