# Phase 6 Implementation Plan

## Summary

This file is the execution roadmap and progress marker for Cartulary Phase 6: Collaboration, presence, and same-field conflict resolution.

`docs/guides/cartulary_implementation_testing_guide.md` section 8 is the controlling implementation-scope reference for this phase. Normative behavior remains owned by the core documents, especially Core 01 section 3.3.10, Core 01 section 3.3.5, Core 03 sections 3 and 4, Core 04 sections 1 and 2, and Core 04 section 4.5.

This planning artifact does not implement Phase 6 behavior. It is intentionally root-level so agents can find it quickly during handoff or interrupted implementation sessions. No README update is required for discoverability.

## Sprint Checklist

| Done | Sprint                                                                | Validation  | Blockers                                                                                                                                                                          | Follow-up Notes                                                                                                                                                |
| ---- | --------------------------------------------------------------------- | ----------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [ ]  | 0. Phase 6 ownership manifest and harness setup                       | [ ] pending | `tools/phase6_test_map.json` and `docs/testing/phase6_coverage_ledger.md` do not exist yet.                                                                                       | Add the Phase 6 manifest before feature work, activate `phase6`, and create selectable failing or placeholder rows for every `U-6-*`, `I-6-*`, and `E-6-*` ID. |
| [ ]  | 1. Patch conflict contract and explicit resolver domain               | [ ] pending | Phase 3 covers server-side Timeline rebase/conflict transport, but Phase 6 still needs generalized resolver semantics and browser-facing conflict state.                          | Start with backend/domain rows for `U-6-01..U-6-06`; reuse Phase 3 store behavior where valid instead of duplicating owner claims.                             |
| [ ]  | 2. Incident socket handshake, resume, replay, and presence hardening  | [ ] pending | Current socket implementation exists but needs Phase 6-owned contract coverage for reset, replay-only event filtering, canonical presence arrays, and revocation close semantics. | Cover `U-6-07`, `U-6-08`, `I-6-01`, `I-6-02`, and `I-6-04`.                                                                                                    |
| [ ]  | 3. Frontend collaboration client, save state, and local pending queue | [ ] pending | Workbook UI currently has surface flows, but Phase 6 needs a durable browser-runtime queue, save labels, conflict queue state, and reconnect/re-auth behavior.                    | Cover `U-6-05`, `U-6-06`, `U-6-09`, `E-6-03`, and `E-6-05`.                                                                                                    |
| [ ]  | 4. Browser presence indicators and live-cell anchoring                | [ ] pending | Presence and live patches must remain bound to `record_id` plus `field_key` through sort/filter/group/virtual scrolling or invalidation.                                          | Cover `E-6-01` and `E-6-04`, with focused frontend unit tests for row/cell key stability.                                                                      |
| [ ]  | 5. Two-client concurrent edit E2E and resolver UX                     | [ ] pending | Requires stable multi-context browser fixture and deterministic conflict setup.                                                                                                   | Cover `I-6-03` and `E-6-02`; verify same-surface resolver actions commit or clear exactly as specified.                                                        |
| [ ]  | 6. Phase gate, ledgers, schedules, baselines, and handoff cleanup     | [ ] pending | Depends on prior sprints.                                                                                                                                                         | Regenerate ledgers/schedules/baselines after all authoritative rows are real, then run phase and check gates.                                                  |

## Global References

- Controlling guide: `docs/guides/cartulary_implementation_testing_guide.md`, `Phase 6`, `U-6-01..U-6-09`, `I-6-01..I-6-04`, `E-6-01..E-6-05`.
- Phase-owned ACs: `AC-008`, `AC-037..AC-042`, `AC-129`, `AC-131..AC-136`, `AC-156..AC-163`, `AC-204`, `AC-226..AC-230`, `AC-376..AC-382`.
- Shared ACs to avoid overclaiming: `AC-009`, `AC-126`, and `AC-203` are shared with Phase 3. Phase 3 owns server-side field-level patch rebase, same-field conflict transport, and collection conflict values; Phase 6 owns resolver, concurrent-client, and local pending-queue behavior.
- Primary REQs: `REQ-01-015..REQ-01-017`, `REQ-01-029`, `REQ-01-062..REQ-01-067`, `REQ-01-250..REQ-01-277`, `REQ-03-033..REQ-03-100`, `REQ-03-223..REQ-03-235`, `REQ-04-010`, `REQ-04-013..REQ-04-016`, and `REQ-04-053`.
- Existing WebSocket call path: `internal/app/runtime.go` -> `collaboration.RegisterRoutes()` -> `internal/modules/collaboration/routes.go` -> `platformws.Accept`, `Hub.IssueResumeToken`, `Hub.ReplayMessages`, `Hub.UpsertPresence`, `Hub.BroadcastPresenceDelta`, `Hub.RevokeSession`, and `Hub.RevokeIncidentSession` in `internal/platform/ws/ws.go`.
- Existing mutation/collaboration publishing path: workbook or Timeline mutation route -> module store mutation result -> `publishRecordChange` -> `platformws.Hub.PublishRecordChange` -> incident socket `record_changed` message.
- Existing frontend areas to extend: `apps/web/src/WorkbookShell.tsx`, workbook surface tests in `apps/web/src`, and browser specifications in `apps/web/e2e`.
- Generated boundaries: do not hand-edit `internal/gen/**` or `packages/protocol-ts/src/generated/**`; if SQL query generation is required, edit `db/queries/**` then run `make generate`.

## Sprint 0. Phase 6 Ownership Manifest and Harness Setup

Objective: Establish Phase 6 test ownership before feature work so TDD rows can be selected by repo tooling.

Status: Not started. `tools/phase_registry.json` already contains Phase 6 as `planned`, but the authoritative manifest and coverage ledger are not present.

Relevant IDs: all `U-6-*`, `I-6-*`, `E-6-*`; `tools/phase6_test_map.json`; `docs/testing/phase6_coverage_ledger.md`.

Files and areas:
- Add `tools/phase6_test_map.json` with the authoritative rows from guide section 8.
- Generate `docs/testing/phase6_coverage_ledger.md` via `make phase-ledgers` after the manifest is accepted.
- Add or extend reusable WebSocket fixtures under `internal/testutil/incidentwstest` only when existing helpers cannot express Phase 6 semantics.
- Add frontend unit coverage in `apps/web/src/WorkbookShell.phase6.test.tsx` or smaller Phase 6-specific files when the UI is split.
- Add browser coverage in `apps/web/e2e/phase6.collaboration.spec.ts`.

Test-first sequence:
1. Add manifest rows with expected symbols for every `U-6-01..U-6-09`, `I-6-01..I-6-04`, and `E-6-01..E-6-05` row.
2. Add failing backend stubs for the first sprint rows before implementing behavior.
3. Use explicit placeholder/no-op rows only for later frontend or browser work that is intentionally out of the current sprint, and replace them before Phase 6 exit.
4. Run manifest/name validation before implementing feature behavior.

Implementation tasks:
- Add Phase 6 manifest metadata and ledger notes clarifying the Phase 3 shared AC boundary.
- Keep existing Phase 1 and Phase 3 socket tests as support evidence only; do not let them claim Phase 6 IDs unless the test symbol and manifest row are renamed/rewritten for Phase 6.
- Add `forbidden_id_files` for known support-only files such as older Phase 1/Phase 3 socket tests.
- Set `phase6` to active only when the manifest and first authoritative rows are ready.

Validation commands:
- `make explain-phase PHASE=phase6`
- `make phase-ledgers`
- `make phase-ledger-drift`
- `git diff --check`

Deliverables:
- `tools/phase6_test_map.json` exists and is selectable.
- `docs/testing/phase6_coverage_ledger.md` exists and names every Phase 6 row.
- The first Sprint 1 tests fail for behavior, not because of missing harness plumbing.

Risks and assumptions:
- Existing Phase 3 and Phase 1 socket tests already cover parts of the route shape. Phase 6 rows must state whether they are upgrading that evidence or adding new owner coverage.
- If helper schedules need service-backed browser rows, keep schedule updates separate from codegen or migration drift.

Exit criteria:
- `make explain-phase PHASE=phase6` reports the manifest, ledger path, execution dependencies, service requirements, and target coverage.
- `make phase-ledger-drift` passes after ledger generation.

## Sprint 1. Patch Conflict Contract and Explicit Resolver Domain

Objective: Finish Phase 6-owned optimistic concurrency and same-field resolution semantics without re-owning Phase 3's server-side Timeline hot path.

Status: Not started.

Relevant IDs:
- `U-6-01`, `U-6-02`, `U-6-03`, `U-6-04`, `U-6-05`, `U-6-06`
- Shared or related integration row: `I-6-03`
- `REQ-01-062..REQ-01-067`, `REQ-03-033..REQ-03-082`
- `AC-009`, `AC-013`, `AC-037..AC-042`, `AC-118`, `AC-126`, `AC-163`, `AC-203..AC-204`, `AC-226..AC-230`

Grep references:
- `same_field_conflict`
- `error.conflict`
- `collection_value_v1`
- `collection_actions_v1`
- `base_row_version`
- `changed_field_keys`
- `row_version`
- `RecordChangePayload`

Files and areas:
- Backend conflict/domain tests likely belong in `internal/modules/timeline/phase6_*_test.go` and, when generalized workbook behavior is added, `internal/modules/workbook/phase6_*_test.go`.
- Frontend resolver state tests belong in `apps/web/src/WorkbookShell.phase6.test.tsx` or a dedicated resolver component test if the resolver is extracted.
- Avoid hand-editing generated protocol artifacts.

Test-first sequence:
1. Assert mutation requests include `record_id`, `base_row_version`, and changed fields only for Phase 6-owned grid writes.
2. Assert different-field concurrent writes rebase without analyst action while same-field writes return `409 same_field_conflict` with the complete `error.conflict` object.
3. Assert `text_compare_merge` treats content as plain text, normalizes line endings only for suggestion computation, and never silently commits a suggested merge.
4. Assert `collection_review` reads and conflicts use `collection_value_v1`, while explicit resolution writes use `collection_actions_v1`.
5. Assert resolver state remains same-surface, keeps the grid visible, preserves focus, and cannot disappear without explicit `keep_saved`, `use_unsaved`, `merged_value`, or clear behavior.
6. Assert `keep_saved` creates no revision while `use_unsaved` and `merged_value` create attributed change sets.

Implementation tasks:
- Confirm current Phase 3 conflict payload fields match the Phase 6 resolver needs; add adapter code only where UI/domain state needs a stable shape.
- Build resolver state as a local conflict queue keyed by `record_id` and `field_key`, not by row index or visible order.
- Implement explicit resolution writes with row-version preconditions and idempotency where the owner routes require it.
- Ensure conflict clear and resolution are auditable through normal mutation/change-set paths when they commit.

Validation commands:
- `go test ./internal/modules/timeline ./internal/modules/workbook -run 'TestPhase6_.*(U_6_01|U_6_02|U_6_03|U_6_04|U_6_05|U_6_06)'`
- `make frontend-unit`
- `make frontend-typecheck`
- `git diff --check`

Deliverables:
- Same-field conflict payloads are directly consumable by the same-surface resolver.
- Text and collection conflict modes expose explicit resolution choices and never silently auto-commit.
- Resolver local state is keyed by durable row and field identity.

Risks and assumptions:
- Some conflict transport may already pass through Phase 3 tests. Do not weaken those tests; add Phase 6 tests around client/resolver behavior and generalized grid contracts.
- If non-Timeline workbook routes are not yet generalized, keep the sprint scoped to surfaces currently exposed while documenting out-of-scope gaps in the manifest row.

Exit criteria:
- Backend and frontend tests prove all `U-6-01..U-6-06` claims without relying on browser-only assertions.
- Same-field conflicts remain visible and unresolved until explicit analyst action.

## Sprint 2. Incident Socket Handshake, Resume, Replay, and Presence Hardening

Objective: Make the incident WebSocket route fully Phase 6 compliant for handshake, resume, replay, presence, origin rejection, heartbeats, and revocation.

Status: Not started.

Relevant IDs:
- `U-6-07`, `U-6-08`, `I-6-01`, `I-6-02`, `I-6-04`
- Related browser row: `E-6-03`
- `REQ-01-250..REQ-01-277`, `REQ-03-090..REQ-03-100`, `REQ-04-010`, `REQ-04-053`
- `AC-008`, `AC-129`, `AC-131..AC-136`, `AC-156..AC-163`, `AC-255`

Grep references:
- `handleIncidentSocket`
- `establishSession`
- `readFirstMessage`
- `hello_ack`
- `resume_ack`
- `ResumeStatusResetNeeded`
- `PresenceSnapshotMessage`
- `BroadcastPresenceDelta`
- `RejectUntrustedBrowserOrigin`
- `session_revoked`
- `heartbeat_timeout`

Files and areas:
- `internal/modules/collaboration/routes.go`
- `internal/platform/ws/ws.go`
- `internal/testutil/incidentwstest`
- `internal/modules/collaboration/phase6_*_test.go`
- `internal/modules/timeline/phase6_integration_test.go` only if mutation-generated replay ordering is easier to prove through Timeline writes.

Test-first sequence:
1. Assert the first application message must be exactly one `hello` or `resume`, and later `hello` or `resume` messages close with route-owned invalid-message behavior.
2. Assert expired, mismatched, or replay-window-too-old resumes produce reset behavior instead of partial replay.
3. Assert valid resumes replay replayable messages only; `presence_snapshot` and `presence_delta` are rehydrated through the presence flow, not replayed from the stream.
4. Assert presence payloads are incident-scoped, ephemeral, duplicate-free, and canonically ordered in snapshots.
5. Assert heartbeats do not slide idle expiry, while logout, expiry, incident access revocation, and concurrency-limit revocation produce `session_revoked` and close the socket with the route-owned reason.
6. Assert cookie-authenticated browser upgrades reject untrusted `Origin` before incident subscription.

Implementation tasks:
- Tighten `Hub.ReplayMessages` filtering so only replayable message types can be replayed.
- Add canonicalization helpers for presence arrays if current ordering or duplicate handling is insufficient.
- Ensure socket revocation paths use consistent public `reason_code` values and a stable close reason.
- Add deterministic fixture controls for clock, replay-token expiry, high-water sequence, and multi-client presence assertions.

Validation commands:
- `go test ./internal/modules/collaboration ./internal/platform/ws ./internal/testutil/incidentwstest -run 'TestPhase6_.*(U_6_07|U_6_08)'`
- `go test ./internal/modules/collaboration ./internal/modules/timeline -run 'TestPhase6_.*(I_6_01|I_6_02|I_6_04)'`
- `make backend-integration`
- `git diff --check`

Deliverables:
- Socket handshake and resume behavior are Phase 6-owned and deterministic.
- Presence remains ephemeral and incident-scoped.
- Session and authorization revocations close sockets with observable `session_revoked` behavior.

Risks and assumptions:
- Heartbeat and expiry tests can be slow if they depend on wall-clock intervals. Prefer injected clock or small test-only route wiring over sleeping for production defaults.
- Origin rejection behavior may intentionally be status-only before WebSocket upgrade; tests should not require a JSON envelope unless the owner spec requires one.

Exit criteria:
- Valid resume never replays presence messages from retained replay storage.
- Invalid or stale resume never provides partial replay as if it were complete.

## Sprint 3. Frontend Collaboration Client, Save State, and Local Pending Queue

Objective: Add the browser-runtime client machinery for save-state labels, unsent-write queueing, replay, conflict halt, and re-auth recovery boundaries.

Status: Not started.

Relevant IDs:
- `U-6-09`, support for `U-6-05` and `U-6-06`
- Browser rows: `E-6-03`, `E-6-05`
- `REQ-01-029`, `REQ-01-250..REQ-01-277`, `REQ-03-072`, `REQ-03-089`, `REQ-03-095`, `REQ-03-099..REQ-03-100`, `REQ-04-013..REQ-04-016`
- `AC-131`, `AC-136`, `AC-156..AC-163`, `AC-376..AC-382`

Grep references:
- `Syncing`
- `Saved`
- `Conflict`
- `client_txn_id`
- `row_version`
- `base_row_version`
- `session_revoked`
- `logout`

Files and areas:
- `apps/web/src/WorkbookShell.tsx`
- New frontend helpers such as `apps/web/src/collaborationClient.ts`, `apps/web/src/pendingQueue.ts`, or equivalent if extraction improves testability.
- `apps/web/src/WorkbookShell.phase6.test.tsx`
- Browser specs in `apps/web/e2e/phase6.collaboration.spec.ts`

Test-first sequence:
1. Add unit tests for save-state labels: only `Syncing`, `Saved`, and `Conflict` are emitted.
2. Add unit tests for a FIFO queue bounded to 64 non-coalescible units.
3. Add unit tests for contiguous same-record coalescing and no cross-record reorder.
4. Add unit tests proving retryable transient disconnect and `session_revoked` preserve queued unsent writes in the same browser runtime.
5. Add unit tests proving non-retryable failures and same-field conflicts halt replay on the blocking item without silent eviction.
6. Add browser tests proving re-auth can explicitly recover unsaved local work, while a full reload does not silently restore old unsent writes.

Implementation tasks:
- Implement queue storage in browser runtime memory, not durable local storage, unless a later owner document changes the persistence boundary.
- Bind every queued write to `record_id`, `field_key`, `base_row_version`, normalized payload, and `client_txn_id`.
- Ensure queued writes replay FIFO after re-auth and stop at first blocking non-retryable failure.
- Surface conflict state through the same resolver path rather than an alert-only or route-changing workflow.

Validation commands:
- `make frontend-unit`
- `make frontend-typecheck`
- `make browser-e2e-webserver-backed`
- `git diff --check`

Deliverables:
- Local pending queue behavior is deterministic and bounded.
- Save-state labels conform exactly to the normative vocabulary.
- Browser runtime can recover unsent writes after re-auth without silently restoring them after full reload.

Risks and assumptions:
- Browser runtime recovery may require Playwright multi-context plus controlled route failure hooks; add test-only hooks under existing test boundaries, not production-only shortcuts.
- Queue coalescing must not obscure audit attribution or idempotency semantics.

Exit criteria:
- `U-6-09`, `E-6-03`, and `E-6-05` are real assertions, not placeholders.
- A same-field conflict leaves later queued items blocked until explicit resolution or clear.

## Sprint 4. Browser Presence Indicators and Live-Cell Anchoring

Objective: Make workbook-visible presence and live update state attach to durable row/cell identity through resorting, filtering, grouping, virtual scrolling, and row patches.

Status: Not started.

Relevant IDs:
- `E-6-01`, `E-6-04`
- Supporting unit coverage for `U-6-08` and `U-6-09`
- `REQ-01-015..REQ-01-017`, `REQ-03-086`, `REQ-03-090..REQ-03-098`, `REQ-03-223..REQ-03-235`
- `AC-008`, `AC-047`, `AC-132`, `AC-133`

Grep references:
- `presence_snapshot`
- `presence_delta`
- `record_changed`
- `affected_views`
- `patch_cells`
- `record_id`
- `field_key`
- `view_schema_id`

Files and areas:
- `apps/web/src/WorkbookShell.tsx`
- Any extracted row-grid, virtualization, or collaboration indicator components.
- `apps/web/src/WorkbookShell.phase6.test.tsx`
- `apps/web/e2e/phase6.collaboration.spec.ts`

Test-first sequence:
1. Add browser test with two analysts connected to the same incident and same workbook surface.
2. Assert presence appears within the expected interaction window with row-gutter and same-cell editing hints where available.
3. Assert live row patches, invalidations, sort changes, filters, grouping transitions, and virtual-scroll movement do not retarget pending local edits.
4. Assert conflict markers and presence markers stay attached to the intended `record_id` plus `field_key`, not row index.

Implementation tasks:
- Normalize all collaboration UI state around `record_id`, `field_key`, and `view_schema_id`.
- Make visible row rendering derive indicators from durable keys after every query refresh or live update.
- Add data-testid or accessibility labels only as stable test affordances, not as alternate behavior.
- Ensure presence expiration/removal clears markers without disturbing local drafts.

Validation commands:
- `make frontend-unit`
- `make browser-e2e-webserver-backed`
- `make frontend-typecheck`
- `git diff --check`

Deliverables:
- Two analysts see workbook-native presence indicators.
- Pending edits, conflict markers, and presence markers remain bound to the intended cell through view changes.

Risks and assumptions:
- If virtual scrolling is not fully implemented yet, the row should document the strongest current equivalent and leave a manifest `out_of_scope` note only when owner text permits phased evidence.
- Avoid implementing new saved-view semantics here; Phase 8 owns deeper sort/filter/group/query behavior.

Exit criteria:
- Browser evidence proves the user-visible collaboration markers satisfy Phase 6 without relying on internal state inspection alone.

## Sprint 5. Two-Client Concurrent Edit E2E and Resolver UX

Objective: Prove the full multi-client concurrent-edit path, including auto-merge for different fields and same-surface explicit resolver behavior for same-field conflicts.

Status: Not started.

Relevant IDs:
- `I-6-03`, `E-6-02`
- Supports final confidence for `U-6-01..U-6-06`
- `REQ-03-033..REQ-03-082`
- `AC-009`, `AC-037..AC-042`, `AC-203..AC-204`, `AC-226..AC-230`

Grep references:
- `same_field_conflict`
- `base_value`
- `current_value`
- `unsaved_value`
- `merged_value`
- `keep_saved`
- `use_unsaved`
- `record_changed`

Files and areas:
- `internal/modules/collaboration/phase6_integration_test.go` or `internal/modules/timeline/phase6_integration_test.go` for real-client API/socket assertions.
- `apps/web/e2e/phase6.collaboration.spec.ts` for two-context browser behavior.
- Frontend resolver tests if UX behavior needs more deterministic unit coverage than browser tests can provide.

Test-first sequence:
1. Create one incident, one workbook row, and two authenticated analysts in separate clients.
2. Different-field edit path: both clients edit distinct fields from the same base; both commits succeed; both clients converge after live updates or refresh.
3. Same-field edit path: second commit receives conflict; resolver opens on the same surface; saved and local values are both visible; local draft is preserved.
4. `keep_saved` path clears conflict without a new revision.
5. `use_unsaved` and `merged_value` paths create new attributed changes and publish record changes to the other client.
6. Verify focus returns to the same cell and the grid remains visible through the whole resolver flow.

Implementation tasks:
- Add deterministic multi-client Playwright helpers for two analysts on the same incident.
- Add server-side integration helper to force same-base concurrent writes without race-prone sleeps.
- Ensure conflict resolution updates row state, local queue state, and collaboration messages atomically from the user's perspective.
- Ensure no duplicate change sets are created on idempotent replay of resolution writes.

Validation commands:
- `go test ./internal/modules/collaboration ./internal/modules/timeline -run 'TestPhase6_.*I_6_03'`
- `make browser-e2e-webserver-backed`
- `make frontend-unit`
- `git diff --check`

Deliverables:
- Real clients prove different-field auto-merge.
- Real clients prove same-field resolver behavior and explicit resolution semantics.
- Browser E2E shows analysts can resolve conflicts without leaving the workbook surface.

Risks and assumptions:
- E2E concurrency can be flaky if it depends on simultaneous clicks. Prefer API-assisted setup for base versions, then assert browser-visible consequences.
- Resolver UI should not invent domain terminology outside `docs/domain.md` vocabulary.

Exit criteria:
- `I-6-03` and `E-6-02` pass repeatedly with deterministic setup.
- All resolver actions have matching backend and browser evidence.

## Sprint 6. Phase Gate, Ledgers, Schedules, Baselines, and Handoff Cleanup

Objective: Close Phase 6 by replacing placeholders, refreshing generated planning artifacts, and running the authoritative phase and repository gates.

Status: Not started.

Relevant IDs: all `U-6-*`, `I-6-*`, `E-6-*`; `tools/phase6_test_map.json`; `docs/testing/phase6_coverage_ledger.md`; service-backed schedule and duration baseline artifacts as applicable.

Files and areas:
- `tools/phase6_test_map.json`
- `docs/testing/phase6_coverage_ledger.md`
- `tools/service_backed_schedule_manifest.json` and other schedule outputs if phase schedule generation changes them.
- Go duration baseline files and Playwright duration baseline files if service-backed/browser timing artifacts change.
- `PHASE6_IMPLEMENTATION_PLAN.md` sprint checklist and follow-up notes.

Test-first sequence:
1. Remove all placeholder Phase 6 rows or convert them into real assertions.
2. Run the narrow phase slice and service-backed slice before broad gates.
3. Refresh ledgers, schedules, and baselines only from successful current artifacts.
4. Run final repository gates and record artifact paths in this plan.

Implementation tasks:
- Update this plan's sprint checklist with validated statuses and any final caveats.
- Run `make phase-ledgers` and `make phase-ledger-drift`.
- Run `make phase-schedules` and `make phase-schedule-drift` when the manifest or schedule ownership changes.
- Run `make go-test-duration-baselines RESULTS_DIR=<dir>` and/or `make browser-e2e-duration-baseline-drift RESULTS_DIR=<dir>` only when successful artifacts justify baseline updates.
- Ensure generated-code drift and migration drift are separate and explained.

Validation commands:
- `make explain-phase PHASE=phase6`
- `make phase-slice PHASE=phase6`
- `make service-backed-slice PHASE=phase6`
- `make phase-ledger-drift`
- `make phase-schedule-drift`
- `make test-fast`
- `make check`

Deliverables:
- Phase 6 manifest, ledger, schedules, and baselines are current.
- Phase 6 authoritative rows pass through public phase wrappers.
- Final gate artifact paths are recorded in the Sprint Checklist follow-up notes.

Risks and assumptions:
- `make check` may require local Postgres, MinIO, browser, Node, and pinned Go toolchain availability. If environment limitations block it, record exact failures and artifact paths.
- If any Phase 6 row depends on Phase 8 sort/filter/group behavior not yet implemented, document the owner-boundary reason and keep the row's assertion faithful to current Phase 6 obligations.

Exit criteria:
- `make phase-slice PHASE=phase6` and `make service-backed-slice PHASE=phase6` pass or report legitimate no-op only where the manifest declares no work.
- `make check` passes, or every remaining failure is documented as an environmental limitation or a separate owner-phase blocker.
