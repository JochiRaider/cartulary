# Phase 6 Coverage Ledger

This ledger is generated from `tools/phase6_test_map.json`. Update the manifest row metadata first, then regenerate this file.

- Scope: Collaboration, presence, and same-field conflict resolution.
- Normative owners: Core 01 §3.3.10; Core 01 §3.3.5; Core 03 §3–§4; Core 04 §1–§2 and §4.5.
- Authority: `tools/phase6_test_map.json` is the enforced Phase 6 traceability source. This ledger is a rendered companion and does not control the mechanical row inventory.
- Sprint 0 activates the Phase 6 ownership map before feature work. Initial later-sprint rows are selectable placeholders until their owning behavior is implemented.
- Phase 3 remains owner for server-side field-level patch rebase, same-field conflict transport, and collection conflict values. Phase 6 owns resolver, concurrent-client, and local pending-queue behavior.

## Authoritative Execution

- `backend-unit` selects pure Phase 6 conflict, resolver-contract, and transport-domain rows through the Phase 6 manifest and `backend_unit` execution dependency.
- `backend-store` selects service-backed Phase 6 workbook mutation and focused socket route-contract rows through manifest-owned service-backed selection.
- `frontend-unit` selects Phase 6 workbook resolver, save-state, and pending-queue component rows through the Phase 6 manifest for `frontend_unit`.
- `backend-integration` selects Phase 6 real-runtime WebSocket and concurrent-edit rows through the Phase 6 manifest and `backend_integration` execution dependency.
- `browser-e2e-webserver-backed` carries authoritative `E-6-*` rows through duration-balanced Playwright manifest-entry shards under `browser_functional`.
- `browser-e2e-visual` carries authoritative `V-6-*` visual rows through `browser_visual`.

## Support-Only Execution

- Existing Phase 1 and Phase 3 socket, revocation, autosave, and same-field conflict tests remain carryover support evidence and must not claim Phase 6 IDs.

## Unit

| Row | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- |
| `U-6-01` | `internal/modules/workbook/phase6_conflict_test.go::TestPhase6_GridWriteConcurrency_U_6_01` | `backend_store` | Every grid write includes `record_id`, `base_row_version`, and changed fields only; different-field concurrent edits auto-rebase, while same-field concurrent edits reject with an explicit conflict payload. | Same-surface resolver UX and local pending-queue replay behavior remain separate Phase 6 rows. |
| `U-6-02` | `internal/modules/workbook/phase6_conflict_test.go::TestPhase6_SameFieldConflictTransport_U_6_02` | `backend_store` | Same-field conflict transport uses `409`, `error.code='same_field_conflict'`, and an `error.conflict` object with the required base and current values. | Phase 3 remains owner for existing Timeline conflict transport support evidence; Phase 6 owns generalized resolver-consumable coverage. |
| `U-6-03` | `internal/modules/workbook/phase6_conflict_test.go::TestPhase6_TextCompareMerge_U_6_03` | `backend_unit` | `text_compare_merge` treats the field as plain text, normalizes line endings only for merge computation, and never silently auto-commits a clean merge suggestion. | Browser presentation of suggested merge choices remains resolver UI evidence. |
| `U-6-04` | `internal/modules/workbook/phase6_conflict_test.go::TestPhase6_CollectionReviewWireFamilies_U_6_04` | `backend_store` | `collection_review` fields use `collection_value_v1` in read or conflict payloads and `collection_actions_v1` in explicit resolution writes. | Interactive resolver choice rendering remains frontend evidence. |
| `U-6-05` | `apps/web/src/WorkbookShell.phase6.test.tsx::Phase 6 U-6-05 keeps the grid visible, conflict unresolved, and focus bound to the same cell` | `frontend_unit` | The resolver keeps the grid visible, leaves conflict state unresolved until explicit action, and returns focus to the same cell after resolution or clear. | Backend same-field conflict payload construction remains backend evidence. |
| `U-6-06` | `apps/web/src/WorkbookShell.phase6.test.tsx::Phase 6 U-6-06 applies explicit resolver outcomes to local conflict state and revisions` | `frontend_unit` | `keep_saved` clears the local conflict without creating a new revision, while `use_unsaved` and `merged_value` create a new attributed change set. | Low-level mutation persistence and audit attribution remain backend evidence. |
| `U-6-07` | `internal/modules/collaboration/phase6_socket_test.go::TestPhase6_IncidentSocketHandshakeResume_U_6_07` | `backend_store` | The first application message on the incident socket is exactly one of `hello` or `resume`, and resume outside the replay window yields reset behavior rather than partial replay. | Presence snapshot and revocation close behavior are owned by `U-6-08` and integration rows. |
| `U-6-08` | `internal/modules/collaboration/phase6_socket_test.go::TestPhase6_IncidentSocketHeartbeatIdleExpiry_U_6_08` | `backend_store` | Presence payloads are incident-scoped and ephemeral; heartbeats do not extend idle expiry; `session_revoked` closes the socket with the route-owned reason code set. | Pure replay filtering, presence sorting, and reason-code helpers remain support evidence in `internal/platform/ws`. |
| `U-6-09` | `apps/web/src/WorkbookShell.phase6.test.tsx::Phase 6 U-6-09 keeps save-state labels and pending queue replay bounded and explicit` | `frontend_unit` | Save-state labels use only `Syncing`, `Saved`, and `Conflict`; the local pending queue is FIFO, bounded, coalesces only contiguous same-record units, and halts on blocking failures without silent eviction or reload restore. | Browser-runtime re-authentication recovery remains `E-6-05` evidence. |

## Integration

| Row | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- |
| `I-6-01` | `internal/modules/collaboration/phase6_integration_test.go::TestPhase6_TwoClientsPresenceReplay_I_6_01` | `backend_integration` | Two real clients connect to the same incident socket, exchange presence snapshots and deltas, and observe deterministic replay ordering within the replay window. | Browser rendering of presence indicators remains `E-6-01` evidence. |
| `I-6-02` | `internal/modules/collaboration/phase6_integration_test.go::TestPhase6_ResumeReplaysReplayableMessagesOnly_I_6_02` | `backend_integration` | Resume with a valid replay token replays replayable messages only, while presence is rehydrated through the documented presence flow rather than replay. | Low-level replay filtering helper behavior remains unit evidence. |
| `I-6-03` | `internal/modules/workbook/phase6_integration_test.go::TestPhase6_ConcurrentEditsResolverPath_I_6_03` | `backend_integration` | Concurrent edits to different fields succeed without operator intervention, while concurrent edits to the same field produce the resolver path and preserve both saved and local drafts correctly. | Full browser resolver interaction remains `E-6-02` evidence. |
| `I-6-04` | `internal/modules/collaboration/phase6_integration_test.go::TestPhase6_CookieSocketRejectsUntrustedOrigin_I_6_04` | `backend_integration` | Cookie-authenticated browser socket upgrades reject untrusted `Origin` values before incident subscription is granted. | General host and origin policy helpers remain platform support evidence. |

## Browser E2E

| Row | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- |
| `E-6-01` | `apps/web/e2e/phase6.collaboration.spec.ts::E-6-01 shows two analysts each other's workbook presence within the expected interaction window` | `browser_functional` | Two analysts on the same incident see each other's presence on the workbook surface within the expected interaction window. | Socket-level presence canonicalization remains backend evidence. |
| `E-6-02` | `apps/web/e2e/phase6.collaboration.spec.ts::E-6-02 auto-merges different-field concurrent edits and requires explicit same-field resolution` | `browser_functional` | Concurrent edits to different fields on the same row auto-merge, while concurrent edits to the same field open the same-surface resolver and require explicit analyst resolution. | Resolver state unit behavior remains frontend-unit evidence. |
| `E-6-03` | `apps/web/e2e/phase6.collaboration.spec.ts::E-6-03 preserves unsaved local work after socket revocation and re-authentication` | `browser_functional` | Logout, expiry, or concurrency-limit revocation emits `session_revoked`, closes the socket, and preserves unsaved local work for later explicit recovery after re-authentication. | Low-level session revocation socket delivery remains backend evidence. |
| `E-6-04` | `apps/web/e2e/phase6.collaboration.spec.ts::E-6-04 keeps live updates conflict markers and presence markers anchored to record_id and field_key` | `browser_functional` | Live updates never retarget a pending local edit, same-field conflict marker, row-gutter presence marker, or same-cell presence marker away from the bound `record_id` and `field_key` during sorting, filtering, grouping, virtual scrolling, or live row patch. | Low-level grid adapter identity helpers may receive separate frontend unit coverage. |
| `E-6-05` | `apps/web/e2e/phase6.collaboration.spec.ts::E-6-05 replays queued unsent writes after re-authentication without silent reload restore` | `browser_functional` | Within one browser runtime, queued unsent writes survive transient disconnect or session revocation, replay in FIFO order after re-authentication, halt on the first blocking non-retryable failure, and are never silently restored after a full reload. | Queue size and coalescing unit semantics remain frontend-unit evidence. |

## Visual Regression

| Row | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- |
| `V-6-GRID-01` | `apps/web/e2e/workbook.visual.spec.ts::V-6-GRID-01 captures Phase 6 row-gutter and same-cell presence markers` | `browser_visual` | The visual harness captures deterministic Timeline row-gutter and same-cell presence markers keyed by `record_id` and `field_key`. | Socket payload canonicalization and timing remain backend and functional browser evidence. |
| `V-6-GRID-02` | `apps/web/e2e/workbook.visual.spec.ts::V-6-GRID-02 captures Phase 6 same-field conflict marker resolver and Conflict strip` | `browser_visual` | The visual harness captures the same-field conflict marker, same-surface resolver, and `Conflict` save-state strip on the Timeline grid. | Resolver mutation semantics remain frontend-unit and functional browser evidence. |
| `V-6-GRID-03` | `apps/web/e2e/workbook.visual.spec.ts::V-6-GRID-03 captures Phase 6 pending-queue save-state transitions` | `browser_visual` | The visual harness captures pending-queue `Syncing`, blocked `Conflict`, and recovered `Saved` save-state presentations. | Pending-queue FIFO, boundedness, and replay semantics remain frontend-unit and functional browser evidence. |

## Shared Harness Coverage

| Harness | Phase 6 evidence |
| --- | --- |
| Incident WebSocket helpers | `internal/testutil/incidentwstest` already owns canonical incident WebSocket handshakes and revocation assertions. Sprint 0 does not extend it; resume and replay-specific helpers belong with the real Sprint 2 tests. |
| Workbook resolver helpers | `apps/web/src/WorkbookShell.phase6.test.tsx` holds Phase 6 resolver, save-state, and pending-queue component rows until the UI is split into smaller owner components. |
| Browser collaboration helpers | `apps/web/e2e/phase6.collaboration.spec.ts` holds Phase 6 multi-client browser rows for presence, resolver UX, live-cell anchoring, and pending-queue recovery. |

## Support-Only Evidence

- `internal/modules/auth/phase1_integration_test.go` keeps Phase 1 session revocation socket coverage as carryover support for Phase 6 implementation, but authoritative Phase 6 completion claims must live in `phase6_*` rows.
- `internal/modules/timeline/phase3_integration_test.go` keeps Phase 3 canonical socket and same-field conflict transport coverage as carryover support for Phase 6 implementation, but authoritative Phase 6 completion claims must live in `phase6_*` rows.
- `internal/modules/workbook/workbook_mutation_integration_test.go` keeps existing workbook mutation conflict coverage as carryover support for Phase 6 implementation, but authoritative Phase 6 completion claims must live in `phase6_*` rows.
- `apps/web/src/WorkbookShell.phase3.autosave.test.tsx` keeps Phase 3 autosave and conflict UI coverage as carryover support for Phase 6 implementation, but authoritative Phase 6 completion claims must live in `phase6_*` rows.
