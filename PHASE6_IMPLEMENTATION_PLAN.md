# Phase 6 Implementation Plan

## Summary

This file is the execution roadmap and progress marker for Cartulary Phase 6: Collaboration, presence, and same-field conflict resolution.

`docs/guides/cartulary_implementation_testing_guide.md` section 8 is the controlling implementation-scope reference for this phase. Normative behavior remains owned by the core documents, especially Core 01 section 3.3.10, Core 01 section 3.3.5, Core 03 sections 3 and 4, Core 04 sections 1 and 2, and Core 04 section 4.5.

This planning artifact does not implement Phase 6 behavior. It is intentionally root-level so agents can find it quickly during handoff or interrupted implementation sessions. No README update is required for discoverability.

## Sprint Checklist

| Done | Sprint                                                                | Validation                         | Blockers                                                                                                                                                                          | Follow-up Notes                                                                                                                                                  |
| ---- | --------------------------------------------------------------------- | ---------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [x]  | 0. Phase 6 ownership manifest and harness setup                       | [x] passed                         | None. Docker is reachable and both Phase 6 public slice wrappers pass.                                                                                                            | `phase-slice` artifact root: `.cartulary/test-results/20260505T222643Z-p41927/phase-slice`; `service-backed-slice` artifact root: `.cartulary/test-results/20260505T222657Z-p45629/service-backed-slice`. |
| [x]  | 1. Patch conflict contract and explicit resolver domain               | [x] passed | None. The workbook collection auto-rebase regression was fixed as a test-fixture issue, and the workbook package plus Phase 6 public slices pass. | Backend `U-6-01..U-6-04` and frontend `U-6-05..U-6-06` are implemented. Phase 3 Timeline hot-path coverage remains support evidence.                           |
| [x]  | 2. Incident socket handshake, resume, replay, and presence hardening  | [x] passed | None. Focused socket tests, backend integration, and both Phase 6 public slice wrappers pass. | Backend `U-6-07`, `U-6-08`, `I-6-01`, `I-6-02`, and `I-6-04` are implemented.                                                                                                    |
| [x]  | 3. Frontend collaboration client, save state, and local pending queue | [x] passed | None. `E-6-03` and `E-6-05` are now real browser-functional evidence.                    | Frontend runtime, `U-6-09`, same-runtime re-auth recovery, FIFO replay, non-retryable halt, and reload-loss boundaries are implemented.                                                                                                    |
| [x]  | 4. Browser presence indicators and live-cell anchoring                | [x] passed | None. Backend sparse `record_changed` patches, browser presence indicators, live-cell anchoring, and Phase 6 browser-functional rows pass.                                          | `E-6-01` and `E-6-04` are implemented, with focused frontend unit coverage for keyed presence state and sparse live patch anchoring.                                                                      |
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

Status: Complete for Sprint 0. `tools/phase_registry.json` now marks Phase 6 as `active`; `tools/phase6_test_map.json`, generated ledger, generated schedules, and selectable test rows exist.

Relevant IDs: all `U-6-*`, `I-6-*`, `E-6-*`; `tools/phase6_test_map.json`; `docs/testing/phase6_coverage_ledger.md`.

Files and areas:
- Added `tools/phase6_test_map.json` with authoritative rows from guide section 8.
- Generated `docs/testing/phase6_coverage_ledger.md` via `make phase-ledgers`.
- Regenerated `tools/service_backed_schedule_manifest.json` and `tools/check_schedule_manifest.json` via `make phase-schedules`.
- Added Phase 6 stubs/placeholders in `internal/modules/workbook/phase6_conflict_test.go`, `internal/modules/collaboration/phase6_socket_test.go`, `internal/modules/collaboration/phase6_integration_test.go`, `internal/modules/workbook/phase6_integration_test.go`, `internal/platform/ws/phase6_ws_test.go`, `apps/web/src/WorkbookShell.phase6.test.tsx`, and `apps/web/e2e/phase6.collaboration.spec.ts`.
- Did not extend reusable WebSocket fixtures under `internal/testutil/incidentwstest`; resume/replay helper work remains deferred to Sprint 2.

Completed test-first sequence:
1. Manifest rows now cover every `U-6-01..U-6-09`, `I-6-01..I-6-04`, and `E-6-01..E-6-05` row.
2. Sprint 1 replaced the intentional failing behavior stubs for `U-6-01..U-6-06`.
3. Later-sprint rows are selectable no-op placeholders and must be replaced before Phase 6 exit.
4. Manifest/name validation passes before any Phase 6 feature behavior implementation.

Completed implementation tasks:
- Added Phase 6 manifest metadata and ledger notes clarifying the Phase 3 shared AC boundary.
- Kept existing Phase 1 and Phase 3 socket/conflict/autosave tests as support evidence only.
- Added `forbidden_id_files` for known support-only carryover files.
- Set `phase6` active after manifest rows and referenced test symbols/titles existed.

Validation commands:
- `make phase-map-check`
- `make explain-phase PHASE=phase6`
- `make phase-ledgers`
- `make phase-ledger-drift`
- `make phase-schedules`
- `make phase-schedule-drift`
- `make backend-unit CARTULARY_MANIFEST_PHASE=phase6 CARTULARY_MANIFEST_SECTION=unit CARTULARY_MANIFEST_COVERAGE=authoritative CARTULARY_MANIFEST_EXECUTION_DEPENDENCY=backend_unit`
- `make frontend-unit CARTULARY_MANIFEST_PHASE=phase6 CARTULARY_MANIFEST_COVERAGE=authoritative CARTULARY_MANIFEST_EXECUTION_DEPENDENCY=frontend_unit`
- `make phase-slice PHASE=phase6`
- `make service-backed-slice PHASE=phase6`
- `git diff --check`

Validation results:
- `make phase-map-check` passed.
- `make explain-phase PHASE=phase6` passed and reports 18 authoritative rows across `backend_unit`, `backend_store`, `backend_integration`, `frontend_unit`, and `browser_functional`.
- `make phase-ledger-drift` passed after ledger generation.
- `make phase-schedule-drift` passed after schedule generation.
- Direct Phase 6 `backend_unit` selection initially reached the intentional `U-6-03` behavior stub failure; Sprint 1 has since replaced `U-6-01..U-6-04` with real assertions.
- Direct Phase 6 `frontend_unit` selection initially reached the intentional `U-6-05` and `U-6-06` behavior stub failures; Sprint 1 has since replaced those rows with real assertions.
- Docker is reachable: `docker info --format '{{json .ServerVersion}} {{json .OSType}} {{json .OperatingSystem}}'` returned `"28.5.1" "linux" "Docker Desktop"`.
- `make phase-slice PHASE=phase6` passed with 5/5 work units, 18 tests, 0 failures, and artifact root `.cartulary/test-results/20260505T222643Z-p41927/phase-slice`.
- `make service-backed-slice PHASE=phase6` passed with 3/3 work units, 12 tests, 0 failures, and artifact root `.cartulary/test-results/20260505T222657Z-p45629/service-backed-slice`.
- `git diff --check` passed.

Deliverables:
- `tools/phase6_test_map.json` exists and is selectable.
- `docs/testing/phase6_coverage_ledger.md` exists and names every Phase 6 row.
- Generated schedule manifests include active Phase 6 service-backed and browser-functional rows.
- Sprint 1 tests are no longer behavior stubs after the Sprint 1 implementation.

Risks and assumptions:
- Existing Phase 3 and Phase 1 socket tests already cover parts of the route shape. Phase 6 rows must state whether they are upgrading that evidence or adding new owner coverage.
- Later placeholder rows are intentionally selectable no-ops only during Phase 6 buildout and must be replaced before Phase 6 exit.
- Docker-backed Postgres/MinIO service image warm-up is available in the local environment as of the Sprint 0/1 Docker validation pass.

Exit criteria:
- `make explain-phase PHASE=phase6` reports the manifest, ledger path, execution dependencies, service requirements, and target coverage. Done.
- `make phase-ledger-drift` passes after ledger generation. Done.
- `make phase-schedule-drift` passes after schedule generation. Done.
- `make phase-slice PHASE=phase6` passes after Docker availability was restored. Done.
- `make service-backed-slice PHASE=phase6` passes after Docker availability was restored. Done.

## Sprint 1. Patch Conflict Contract and Explicit Resolver Domain

Objective: Finish Phase 6-owned optimistic concurrency and same-field resolution semantics without re-owning Phase 3's server-side Timeline hot path.

Status: Complete for Sprint 1. The resolver route, contract shape, backend generic workbook conflict domain, deterministic text merge suggestion logic, collection resolver wire-family validation, and client-local same-surface resolver state are implemented. Phase 3's Timeline patch hot path remains owned by Phase 3 and is reused as support evidence rather than re-owned here.

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
- `contracts/openapi/cartulary.openapi.yaml` now declares `POST /api/v1/records/{record_id}/conflicts/{conflict_token}/resolve`, `RecordConflictResolveRequest`, and the clear-response envelope used by `keep_saved`.
- `internal/modules/workbook/mutation_api.go`, `internal/modules/workbook/mutation_store.go`, `internal/modules/workbook/routes.go`, and `internal/modules/workbook/conflict_merge.go` implement the generic workbook resolver path, token claims, idempotent clear/commit behavior, non-overlap auto-rebase, and text merge suggestions.
- `internal/modules/workbook/phase6_conflict_test.go` contains the real `U-6-01..U-6-04` backend assertions.
- `internal/modules/workbook/workbook_mutation_integration_test.go` was adjusted so the existing service-backed workbook mutation evidence expects stale non-overlap auto-rebase instead of `row_version_conflict`. The follow-up regression fix uses a distinct decision fixture for the stale non-overlap collection add, carries the committed row version forward, and locks duplicate collection adds as `no_effective_change` with no durable side effects.
- `apps/web/src/WorkbookShell.tsx` implements the local conflict queue and same-surface resolver UI for Timeline workbook cells.
- `apps/web/src/WorkbookShell.phase6.test.tsx` contains the real `U-6-05` and `U-6-06` frontend assertions.
- Generated protocol artifacts under `internal/gen/**` and `packages/protocol-ts/src/generated/**` were not hand-edited.

Completed test-first sequence:
1. `U-6-01` asserts stale non-overlap and same-field overlap are distinguished by `field_key`.
2. `U-6-02` asserts same-field conflict payloads include stable resolver fields and an opaque content-free token with record/view/field/class/version/request-hash claims.
3. `U-6-03` asserts `text_compare_merge` normalizes line endings only for merge computation, suggests clean disjoint line merges, suppresses overlapping suggestions, and never auto-commits a suggestion.
4. `U-6-04` asserts `collection_review` conflict values use `collection_value_v1` while resolver commit payloads accept `collection_actions_v1`.
5. `U-6-05` asserts the grid remains visible, save state remains exactly `Conflict`, the saved value and local draft remain visible, Enter on open does not resolve, close leaves the conflict unresolved, and focus returns to the cell.
6. `U-6-06` asserts explicit resolver request bodies for `keep_saved` and `use_unsaved`, including no `resolved_value` for `keep_saved` and the local draft as `resolved_value` for `use_unsaved`.

Completed implementation tasks:
- Added a route-scoped resolver request decoder with closed top-level fields and existing error codes: `same_field_conflict`, `row_version_conflict`, `client_txn_conflict`, and `invalid_mutation_payload`.
- Added an opaque conflict token claim envelope for generic workbook same-field conflicts. It contains record, view schema, field, conflict class, base/current row versions, and request hash, but not field content.
- Made generic workbook stale non-overlap patches auto-rebase when revision history can prove no requested field changed since the client base row version.
- Added resolver store flow for generic workbook surfaces: `keep_saved` refreshes the current row without a change set, while `use_unsaved` and `merged_value` commit through the ordinary field write path with current row-version preconditions.
- Added deterministic text merge suggestion logic for `text_compare_merge`.
- Added client-local conflict queue state keyed by `record_id:field_key`, separate from the pending save queue.
- Added same-surface resolver UI, conflict marker, explicit resolver actions, and focus/scroll restoration for the Timeline workbook surface.

Validation commands:
- `go test ./internal/modules/workbook -run 'TestPhase6_.*(U_6_01|U_6_02|U_6_03|U_6_04)'`
- `go test ./internal/modules/workbook -run '^TestPhase4_CoordinationCollections_I_4_COORD_02$' -count=1 -v`
- `go test ./internal/modules/workbook`
- `make phase-slice PHASE=phase6`
- `make service-backed-slice PHASE=phase6`
- `make frontend-unit`
- `make frontend-typecheck`
- `make lint-biome`
- `git diff --check`
- OpenAPI file parse with `tmp/node-runtime/bin/node -e 'const fs=require("fs"); JSON.parse(fs.readFileSync("contracts/openapi/cartulary.openapi.yaml","utf8")); console.log("openapi json ok")'`
- Broader check attempted: `go test ./internal/modules/workbook`

Validation results:
- Focused backend Phase 6 unit test command passed.
- Docker is reachable: `docker info --format '{{json .ServerVersion}} {{json .OSType}} {{json .OperatingSystem}}'` returned `"28.5.1" "linux" "Docker Desktop"`.
- `make frontend-unit` passed with artifact root `.cartulary/test-results/20260505T222733Z-p48309/frontend-unit`.
- `make frontend-typecheck` passed with artifact root `.cartulary/test-results/20260505T222733Z-p48407/frontend-typecheck`.
- `make lint-biome` passed with artifact root `.cartulary/test-results/20260505T222746Z-p49278/lint-biome`.
- Regression remediation passed: `go test ./internal/modules/workbook -run '^TestPhase4_CoordinationCollections_I_4_COORD_02$' -count=1 -v`.
- `go test ./internal/modules/workbook` passed after the fixture correction.
- `go test ./internal/modules/workbook -run 'TestPhase6_.*(U_6_01|U_6_02|U_6_03|U_6_04)'` passed after the fixture correction.
- `make phase-slice PHASE=phase6` passed with 5/5 work units, 18 tests, 0 failures, and artifact root `.cartulary/test-results/20260505T223937Z-p53180/phase-slice`.
- `make service-backed-slice PHASE=phase6` passed with 3/3 work units, 12 tests, 0 failures, and artifact root `.cartulary/test-results/20260505T223955Z-p56214/service-backed-slice`.
- `git diff --check` passed after the documentation edit.
- OpenAPI JSON parse passed after the documentation edit.

Deliverables:
- Same-field conflict payloads for generic workbook surfaces are directly consumable by the resolver route.
- Timeline workbook conflicts are consumed by the browser resolver through the same stable `error.conflict` shape.
- Text and collection conflict modes expose explicit resolution choices and never silently auto-commit.
- Resolver local state is keyed by durable row and field identity.

Risks and assumptions:
- Phase 3 Timeline store conflict transport remains support evidence. Sprint 1 does not re-own or refactor that hot path.
- Sprint 1 adds generic workbook resolver route semantics; Timeline backend resolver route wiring should be treated as later integration/E2E hardening unless a future sprint explicitly pulls Phase 3 Timeline commits into the shared resolver service.
- No durable server-side conflict draft table was added.
- The stale non-overlap collection regression was a test-fixture issue, not a product behavior change: the original stale add targeted an already-linked decision and correctly produced `no_effective_change`. The fixed fixture proves auto-rebase with a real collection change while preserving duplicate/no-op rejection.

Exit criteria:
- Backend and frontend tests prove all `U-6-01..U-6-06` claims without relying on browser-only assertions. Done.
- Same-field conflicts remain visible and unresolved until explicit analyst action. Done for the Timeline workbook client surface.

## Sprint 2. Incident Socket Handshake, Resume, Replay, and Presence Hardening

Objective: Make the incident WebSocket route fully Phase 6 compliant for handshake, resume, replay, presence, origin rejection, heartbeats, and revocation.

Status: Complete for Sprint 2. Phase 6 now owns deterministic socket handshake, resume reset/replay behavior, replayable-only filtering, canonical incident-scoped presence snapshots, origin rejection, and revocation close semantics.

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
- `internal/platform/ws/ws.go` now exposes replayable message classification and filters retained replay output to `record_changed` and `job_progress`.
- `internal/testutil/incidentwstest` now has resume helpers that capture `resume_ack`, replayed messages, and the fresh post-resume `presence_snapshot`.
- `internal/modules/collaboration/phase6_socket_test.go`, `internal/modules/collaboration/phase6_integration_test.go`, and `internal/platform/ws/phase6_ws_test.go` contain the real Sprint 2 Phase 6 assertions.

Completed test-first sequence:
1. `U-6-07` asserts first-message `hello`/`resume` rules, later establishment-message rejection, and reset behavior for mismatched, future, and expired resumes.
2. `U-6-08` asserts replay filtering, reset edge cases, incident-scoped canonical presence snapshots, expiry pruning, and revocation reason delivery.
3. `I-6-01` asserts two-client presence snapshots/deltas, replay ordering within the window, and logout, expiry, concurrency-limit, and incident-access revocations.
4. `I-6-02` asserts valid resume replays `record_changed` and `job_progress` only, while presence is rehydrated through the fresh snapshot.
5. `I-6-04` asserts cookie-authenticated untrusted `Origin` rejection before incident subscription.

Completed implementation tasks:
- Tightened `Hub.ReplayMessages` so unknown, expired, mismatched, future, and too-old resumes return `reset_required` with no missed messages.
- Added an explicit replayable-message classifier and applied it during replay so retained non-replayable messages cannot be returned.
- Preserved existing route revocation behavior: `session_revoked` is sent with the public `reason_code` and the socket closes with policy violation reason `session_revoked`.
- Added deterministic fixture coverage for resume token expiry, high-water replay sequence, multi-client presence, origin rejection, and revocation delivery.

Validation commands:
- `go test ./internal/modules/collaboration ./internal/platform/ws ./internal/testutil/incidentwstest -run 'TestPhase6_.*(U_6_07|U_6_08)'`
- `go test ./internal/modules/collaboration ./internal/modules/timeline -run 'TestPhase6_.*(I_6_01|I_6_02|I_6_04)'`
- `make backend-integration`
- `make phase-slice PHASE=phase6`
- `make service-backed-slice PHASE=phase6`
- `git diff --check`

Validation results:
- `go test ./internal/modules/collaboration ./internal/platform/ws ./internal/testutil/incidentwstest -run 'TestPhase6_.*(U_6_07|U_6_08)' -count=1` passed.
- `go test ./internal/modules/collaboration ./internal/modules/timeline -run 'TestPhase6_.*(I_6_01|I_6_02|I_6_04)' -count=1` passed.
- `make backend-integration` passed with 101 tests, 0 failures, and artifact root `.cartulary/test-results/20260505T230546Z-p79001/backend-integration`.
- `make phase-slice PHASE=phase6` passed with 5/5 work units, 18 tests, 0 failures, and artifact root `.cartulary/test-results/20260505T230546Z-p78996/phase-slice`.
- `make service-backed-slice PHASE=phase6` passed with 3/3 work units, 12 tests, 0 failures, and artifact root `.cartulary/test-results/20260505T230546Z-p78995/service-backed-slice`.

Deliverables:
- Socket handshake and resume behavior are Phase 6-owned and deterministic.
- Presence remains ephemeral and incident-scoped.
- Session and authorization revocations close sockets with observable `session_revoked` behavior.

Risks and assumptions:
- Heartbeat and expiry tests can be slow if they depend on wall-clock intervals. Prefer injected clock or small test-only route wiring over sleeping for production defaults.
- Origin rejection behavior may intentionally be status-only before WebSocket upgrade; tests should not require a JSON envelope unless the owner spec requires one.

Exit criteria:
- Valid resume never replays presence messages from retained replay storage. Done.
- Invalid or stale resume never provides partial replay as if it were complete. Done.

## Sprint 3. Frontend Collaboration Client, Save State, and Local Pending Queue

Objective: Add the browser-runtime client machinery for save-state labels, unsent-write queueing, replay, conflict halt, and re-auth recovery boundaries.

Status: Complete for Sprint 3. The Timeline workbook frontend runtime, `U-6-09` unit evidence, and browser E2E rows `E-6-03` and `E-6-05` are implemented and validated.

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
- `apps/web/src/WorkbookShell.tsx` now owns the in-memory pending replay runtime for Timeline workbook hot-path autosave writes.
- `apps/web/src/WorkbookShell.phase6.test.tsx` contains real `U-6-09` assertions alongside the existing `U-6-05` and `U-6-06` resolver assertions.
- `apps/web/src/timelineWorkbookTestSupport.ts` now provides a controllable WebSocket mock for frontend collaboration tests.
- `apps/web/e2e/phase6.collaboration.spec.ts` contains real `E-6-03` and `E-6-05` browser assertions plus local route and socket helpers for the pending replay scenarios.

Completed test-first sequence:
1. Replaced the `U-6-09` placeholder with frontend unit assertions for exact save-state vocabulary, FIFO replay, contiguous same-record coalescing, fixed 64-unit capacity, session-revocation pause/resume, and same-field conflict halt.
2. Extended the frontend test WebSocket harness so unit tests can emit `hello_ack`, `resume_ack`, `record_changed`, `ping`, and `session_revoked` messages.
3. Preserved existing `U-6-05` and `U-6-06` resolver tests while routing hot-path autosave writes through the new queue.
4. Replaced the `E-6-03` and `E-6-05` browser placeholders with real Playwright flows for socket revocation, same-runtime re-authentication replay, FIFO queue order, non-retryable replay halt, and full-reload queue loss.

Completed implementation tasks:
- Added `PendingReplayUnit`, pending queue runtime state, and collaboration replay state inside the Timeline workbook client.
- Routed autosave-originated Timeline row creates and patches through an incident/client-instance memory-local queue.
- Kept queue capacity fixed at exactly `64` replay units in production.
- Implemented FIFO replay, contiguous same-record coalescing, retryable transport retry, `401`/`403` auth pause, `session_revoked` pause, reset-required HTTP re-query pause, and same-field conflict halt.
- Kept later queued items blocked behind a same-field conflict until explicit conflict resolution or clear.
- Derived `Syncing`, `Saved`, and `Conflict` from queue/conflict/runtime state and added a same-surface queued-edit notice.
- Added `beforeunload` warning while queued writes or unresolved local conflicts exist.
- Did not persist replay units to `localStorage`, `sessionStorage`, IndexedDB, or server state.
- Added browser-only route controls in `phase6.collaboration.spec.ts` for held patches, synthetic auth failures, non-retryable failures, auth-session recovery gates, and WebSocket frame observation.

Validation commands:
- `tmp/node-runtime/bin/pnpm --dir apps/web exec vitest run src/WorkbookShell.phase6.test.tsx --reporter=verbose --testTimeout=10000`
- `make frontend-unit CARTULARY_MANIFEST_PHASE=phase6`
- `make frontend-typecheck`
- `make lint-biome`
- `FORCE_COLOR=0 NO_COLOR=1 CARTULARY_WEB_E2E_BACKEND_PORT=8080 CARTULARY_WEB_E2E_FRONTEND_PORT=4173 CARTULARY_WEB_E2E_API_ORIGIN=http://127.0.0.1:8080 CARTULARY_WEB_E2E_PUBLIC_ORIGIN=http://127.0.0.1:4173 tmp/node-runtime/bin/pnpm --dir apps/web exec playwright test e2e/phase6.collaboration.spec.ts -g "E-6-0(3|5)"`
- `make phase-slice PHASE=phase6`
- `git diff --check`

Validation results:
- Focused Phase 6 frontend unit file passed with 6 tests.
- `make frontend-unit CARTULARY_MANIFEST_PHASE=phase6` passed with artifact root `.cartulary/test-results/20260505T233137Z-p2374/frontend-unit`.
- `make frontend-typecheck` passed with artifact root `.cartulary/test-results/20260505T235838Z-p26805/frontend-typecheck`.
- `make lint-biome` passed with artifact root `.cartulary/test-results/20260505T235838Z-p26767/lint-biome`.
- Focused `E-6-03` and `E-6-05` Playwright run passed with 2 tests.
- `make phase-slice PHASE=phase6` passed with 5/5 work units, 18 tests, 0 failures, and artifact root `.cartulary/test-results/20260505T235715Z-p22491/phase-slice`.
- `git diff --check` passed.

Deliverables:
- Timeline hot-path local pending queue behavior is deterministic, bounded, and memory-local.
- Save-state labels conform exactly to the normative vocabulary for the covered frontend runtime paths.
- Session revocation preserves unsent queued writes in the same browser runtime and resumes after an accepted socket/session establishment message.
- Same-field conflicts move the blocking unit out of the pending queue and into the existing client-local conflict queue.
- Browser evidence proves real socket `session_revoked` handling, re-authenticated FIFO replay, first non-retryable failure halt, and no silent full-reload queue restore.

Risks and assumptions:
- Evidence upload, mention resolution, lifecycle actions, and other non-hot-path operations remain outside the local pending queue.
- The frontend unit test proves the production capacity constant is `64`; it does not enqueue 64 rendered rows because that made the component test impractically heavy.
- Queue replay updates stale patch `base_row_version` from the current visible row before replay; backend row-version and same-field conflict checks remain authoritative.

Exit criteria:
- `U-6-09` is real frontend unit evidence. Done.
- A same-field conflict leaves later queued items blocked until explicit resolution or clear. Done in frontend unit evidence.
- `E-6-03` and `E-6-05` are real browser-functional assertions. Done.

## Sprint 4. Browser Presence Indicators and Live-Cell Anchoring

Objective: Make workbook-visible presence and live update state attach to durable row/cell identity through resorting, filtering, grouping, virtual scrolling, and row patches.

Status: Complete for Sprint 4. Browser presence indicators and live-cell anchoring are implemented on the Timeline workbook surface, backed by server sparse `record_changed` patch payloads and authoritative Phase 6 browser-functional coverage.

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
- `internal/platform/ws/ws.go` now supports `RecordChange.PatchCells`, `RecordChangePayload` emits `change_kind='patch'` when safe, and `BuildViewRowPatch` constructs sparse `view_row_patch_v1` payloads with changed cells and grouping scalars.
- `internal/modules/timeline/routes.go` and `internal/modules/workbook/routes.go` publish sparse patch payloads from mutation results while retaining the existing invalidate fallback when a patch cannot be safely built.
- `internal/modules/collaboration/routes.go` includes `connection_id` on `resume_ack` so the browser can keep self-presence filtering stable after reconnect.
- `apps/web/src/WorkbookShell.tsx` maintains browser presence by exact `connection_id`, filters by exact `sheet_ref`, ignores the current socket connection, renders header/row/same-cell indicators, sends presence updates from durable row/field state, and applies live sparse patches by `record_id`.
- `apps/web/src/workbookShellPhase4.ts` and `apps/web/src/timelineWorkbookTestSupport.ts` now understand `affected_views[].patch_cells` for frontend unit tests.
- `apps/web/src/WorkbookShell.phase6.test.tsx` contains focused unit coverage for keyed presence snapshot/delta handling, save-state independence, and sparse live patch anchoring.
- `apps/web/e2e/phase6.collaboration.spec.ts` replaces the `E-6-01` and `E-6-04` placeholders with real browser-functional tests.

Completed test-first sequence:
1. `E-6-01` now creates two analysts on one incident, verifies header presence, row presence, and same-cell editing presence through the real incident socket path, and confirms ambient presence does not change save state.
2. `E-6-04` now holds a local pending edit, exercises sort/filter/group state, applies remote live changes, and verifies the local draft plus conflict marker stay anchored to the original `record_id` and `field_key`.
3. Frontend unit coverage proves `presence_snapshot` and `presence_delta` are treated as keyed state, self-presence and saved-view mismatches are excluded, and removing a presence clears indicators without touching save state.
4. Frontend unit coverage proves sparse live row patches update the intended row version/cells by `record_id` without forcing an HTTP re-query or moving the active local draft.

Completed implementation tasks:
- Normalized collaboration UI state around `sheet_ref`, `record_id`, `field_key`, and `connection_id`.
- Added workbook-header presence, row presence, and same-cell editing indicators with bounded visible counts and overflow behavior.
- Added immediate and coalesced presence publication from selection/editing state while keeping presence display-only.
- Added sparse live patch application for `record_changed.payload.affected_views[]` entries keyed by exact `view_schema_id`; existing HTTP query refresh remains the fallback for invalidate, stream gaps, reset-required resumes, and unsafe patch cases.
- Preserved pending edit, conflict queue, and save-state semantics during presence updates and live patches.
- Kept generated protocol artifacts untouched.

Validation commands:
- `go test ./internal/platform/ws ./internal/modules/collaboration`
- `go test ./internal/modules/timeline ./internal/modules/workbook`
- `make frontend-unit`
- `make frontend-typecheck`
- `make service-backed-slice PHASE=phase6`
- `git diff --check`

Validation results:
- `go test ./internal/platform/ws ./internal/modules/collaboration` passed.
- `go test ./internal/modules/timeline ./internal/modules/workbook` passed.
- `make frontend-typecheck` passed with artifact root `.cartulary/test-results/20260506T002436Z-p45580/frontend-typecheck`.
- `make frontend-unit` passed with artifact root `.cartulary/test-results/20260506T002436Z-p45629/frontend-unit`.
- `make service-backed-slice PHASE=phase6` passed with 3/3 work units, 12 tests, 0 failures, and artifact root `.cartulary/test-results/20260506T002452Z-p46572/service-backed-slice`.
- `git diff --check` passed.

Deliverables:
- Two analysts see workbook-native presence indicators on the same Timeline workbook surface.
- Header, row, and same-cell presence indicators are keyed from socket presence records rather than visible row or column positions.
- Backend live updates can publish sparse row patches with changed field-keyed cells and grouping scalars.
- Pending edits, conflict markers, and presence markers remain bound to the intended cell through view changes covered by Phase 6 evidence.

Risks and assumptions:
- The current grid adapter has virtualization disabled in the existing implementation; Sprint 4 covers the strongest current equivalent through stable row/cell anchors across live patch, refresh fallback, sort, filter, and grouping controls.
- Saved-view semantics were not expanded. The browser preserves the required distinction by exact `sheet_ref.kind` and `sheet_ref.id` when filtering presence state.
- Presence remains ambient display state only. It does not affect mutation authorization, pending queue membership, conflict state, or save-state labels.
- Phase 8 still owns deeper saved-view/query behavior.

Exit criteria:
- Browser evidence proves the user-visible collaboration markers satisfy Phase 6 without relying on internal state inspection alone. Done.
- Sparse live patches and invalidate fallback remain covered by backend and frontend evidence. Done.

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
