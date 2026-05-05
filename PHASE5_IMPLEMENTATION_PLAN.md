# Phase 5 Implementation Plan

## Summary

This file is the execution roadmap and progress marker for Cartulary Phase 5: Evidence, blob lifecycle, evidence access, and object storage.

`docs/guides/cartulary_implementation_testing_guide.md` section 7.5 is the controlling implementation-scope reference for this phase. Normative behavior remains owned by the core documents, especially Core 01 section 3.3.8, Core 01 section 16, Core 02 sections 4.5 and 13, Core 03 sections 8 and 16, and Core 04 sections 4.3 and 4.5.

This planning artifact does not implement Phase 5 behavior. It is intentionally root-level so agents can find it quickly during handoff or interrupted implementation sessions. No README update is required for discoverability.

## Sprint Checklist

| Done | Sprint | Validation | Blockers | Follow-up Notes |
| --- | --- | --- | --- | --- |
| [x] | 0. Phase 5 ownership manifest and harness setup | [x] validated | None for Sprint 0. | Phase 5 is active and selectable; the initial Sprint 1 stubs have been replaced with real assertions. |
| [x] | 1. Blob create and upload-slot contract | [x] validated | Phase aggregate targets still fail on later Sprint 3+ placeholders. | Blob-create contract, idempotency, size ceiling, and expired-slot replay are implemented. Preview-size behavior remains Sprint 3. |
| [x] | 2. Attach finalization, lifecycle bridge, and no-blob evidence | [x] validated | Phase aggregate targets still fail on later Sprint 3/4 placeholders: `U-5-05`, `U-5-06`, `U-5-07`, `I-5-03`, and `I-5-04`. | Attach finalization, no-blob evidence, lifecycle bridge guards, workbook evidence-count projections, `U-5-08`, `I-5-01`, and browser `E-5-04` are implemented. |
| [ ] | 3. Handle issuance and redeem hardening | [ ] pending |  |  |
| [ ] | 4. Object-storage boundary and cleanup/quarantine behavior | [ ] pending |  |  |
| [ ] | 5. Workbook UI and browser evidence flows | [ ] pending |  |  |
| [ ] | 6. Phase gate, ledgers, baselines, and handoff cleanup | [ ] pending |  |  |

## Global References

- Controlling guide: `docs/guides/cartulary_implementation_testing_guide.md`, `Phase 5`, `U-5-01..U-5-09`, `I-5-01..I-5-04`, `E-5-01..E-5-04`.
- Phase-owned ACs: `AC-004`, `AC-015..AC-016`, `AC-053`, `AC-102..AC-103`, `AC-154..AC-155`, `AC-252..AC-255`, `AC-321..AC-322`, `AC-405`.
- Primary REQs: `REQ-01-243..REQ-01-247`, `REQ-01-458..REQ-01-465`, `REQ-02-186..REQ-02-201`, `REQ-03-116..REQ-03-128`, `REQ-03-242..REQ-03-246`, `REQ-04-023`, `REQ-04-048`, `REQ-04-053`, `REQ-04-079..REQ-04-080`.
- Existing backend call path: `internal/app/runtime.go` -> `evidence.RegisterRoutes()` -> `internal/modules/evidence/routes.go` -> `Service.handleCreateBlob`, `Service.handleAttachBlob`, `Service.handleIssueHandle`, `Service.handleRedeemHandle` -> `internal/modules/evidence/store.go` -> `objectstore.Store`.
- Existing frontend call path: `apps/web/src/WorkbookShell.tsx` -> `issueEvidenceHandle` -> `/api/v1/evidence-records/{record_id}/{preview|download}-handle`.
- Generated boundaries: do not hand-edit `internal/gen/**` or `packages/protocol-ts/src/generated/**`; if SQL query generation is required, edit `db/queries/**` then run `make generate`.

## Sprint 0. Phase 5 Ownership Manifest and Harness Setup

Objective: Establish Phase 5 test ownership before feature work so TDD rows can be selected by repo tooling.

Status: Complete. Phase 5 is active in `tools/phase_registry.json`, `tools/phase5_test_map.json` owns all authoritative `U-5-*`, `I-5-*`, and `E-5-*` rows, and `docs/testing/phase5_coverage_ledger.md` is generated from the manifest.

Relevant IDs: all `U-5-*`, `I-5-*`, `E-5-*`; `tools/phase5_test_map.json`; `docs/testing/phase5_coverage_ledger.md`.

Files and areas:
- Added `tools/phase5_test_map.json`.
- Generated ledger via `make phase-ledgers`.
- Reused existing test harness ownership; no `internal/testutil/phase5test` package was introduced.
- Backend placeholders live in `internal/modules/evidence/phase5_*_test.go`.
- Browser placeholders live in `apps/web/e2e/phase5.evidence.spec.ts`; frontend unit placeholder lives in `apps/web/src/WorkbookShell.phase5.test.tsx`.

Test-first sequence:
1. Completed: manifest rows were added with expected symbols before behavior implementation.
2. Completed: Sprint 1 failing stubs were added for `TestPhase5_ObjectBlobCreate_U_5_01`, `TestPhase5_BlobCreateIdempotency_U_5_02`, `TestPhase5_BlobCreateSizeCeiling_U_5_09`, and `TestPhase5_ExpiredSlotReplay_I_5_02`.
3. Completed: future frontend and browser rows are selectable no-op placeholders so manifest selection can prove coverage without blocking Sprint 1 on later UI work.
4. Completed: manifest/name validation ran before implementation work.

Implementation tasks:
- Completed: authoritative rows are defined for `U-5-01..U-5-09`, `I-5-01..I-5-04`, and `E-5-01..E-5-04`.
- Completed: Phase 4 IDs in existing evidence tests remain unchanged.
- Completed: Phase 4 carryover/support files are listed in `forbidden_id_files` so they cannot claim Phase 5 IDs.

Validation commands:
- Passed: `make explain-phase PHASE=phase5`
- Passed: `make phase-ledgers`
- Passed: `make phase-ledger-drift`
- Passed: `make target-plan-json`
- Passed: `make phase-test-name-check`
- Passed: `tmp/node-runtime/bin/node scripts/test-task-guidance.mjs`
- Completed follow-up: Sprint 1 replaced the initial failing symbols with real assertions for `U-5-01`, `U-5-02`, blob-create `U-5-09`, and `I-5-02`.
- Current expected failure: `make phase-slice PHASE=phase5` selects Phase 5 rows and now fails only on later Sprint 3+ placeholders.
- Passed: `go test ./internal/modules/evidence -run 'TestPhase5_.*(U_5_01|U_5_02|U_5_09|I_5_02)'`

Deliverables:
- Delivered: `tools/phase5_test_map.json`
- Delivered: `docs/testing/phase5_coverage_ledger.md`
- Delivered: initial Phase 5 Sprint 1 symbols, now replaced by passing Sprint 1 assertions.

Risks and assumptions:
- Assumption retained: create `internal/testutil/phase5test` only when naming clarity or helper ownership is needed; otherwise reuse existing helpers.
- Follow-up risk retained for later sprints: service-backed timing baselines may need refresh after remaining Phase 5 placeholders become real tests.

Exit criteria:
- Met: `make explain-phase PHASE=phase5` discovers the active manifest and planned rows.
- Met: no Phase 5 authoritative ID appears in support-only files.

## Sprint 1. Blob Create and Upload-Slot Contract

Objective: Complete `POST /api/v1/object-blobs` request validation, normalization, idempotency, accepted contract echo, size ceiling, and expired-slot replay behavior.

Status: Complete. Sprint 1 covers the blob-slot portion only; preview payload size enforcement remains in Sprint 3 handle issuance/redeem work.

Relevant IDs:
- `U-5-01`, `U-5-02`, `U-5-09`, `I-5-02`
- `REQ-01-243..REQ-01-246`, `REQ-04-079..REQ-04-080`
- `AC-128`, `AC-154`, `AC-155`, `AC-321`

Grep references:
- `DecodeBlobCreateRequest`
- `BlobCreateRequestHash`
- `Store.CreateBlobSlot`
- `Service.handleCreateBlob`
- `blobCreateRouteKey`
- `object_blobs`
- `accepted_contract`
- `target_expires_at`
- `pending_expires_at`
- `blob_create_rejected`
- `byte_size_exceeds_limit`

Files and areas:
- `internal/modules/evidence/api.go`
- `internal/modules/evidence/routes.go`
- `internal/modules/evidence/store.go`
- `internal/platform/config/config.go`
- `db/migrations/00009_phase4_evidence_blob_routes.sql`
- `contracts/openapi/cartulary.openapi.yaml` only if route schema is incomplete.

Test-first sequence:
1. Completed: unit decoder tests for `U-5-01` cover required `incident_id`, `client_txn_id`, `byte_size`, non-object bodies, missing fields, null required fields, empty normalized `client_txn_id`, unknown fields, server-managed fields, invalid `byte_size`, invalid `sha256_hex`, and optional omission/null/empty normalization.
2. Completed: store idempotency tests for `U-5-02` prove `(actor_user_id, incident_id, client_txn_id)` route scoping, exact replay payload stability, divergent normalized request conflict behavior, and independent slots for different incidents or actors.
3. Completed: blob-create size-limit tests for `U-5-09` prove `536870912` succeeds and `536870913` rejects before durable slot creation with `413 blob_create_rejected`.
4. Completed: service-backed `I-5-02` proves replay of an expired slot returns the original expired `target_expires_at`, `pending_expires_at`, upload target, and `accepted_contract`; a fresh target requires a new `client_txn_id`.

Implementation tasks:
- Completed: `invalid_blob_create_request` reason codes use the blob-create registry tokens: `request_not_object`, `missing_required_field`, `field_not_nullable`, `field_empty_after_normalization`, `invalid_byte_size`, `invalid_sha256_hex`, `unknown_field`, and `server_managed_field`.
- Completed: known server-owned members classify as `server_managed_field`, including blob identity, upload state, expiry fields, upload target, accepted contract, storage key, observed metadata, lifecycle fields, cleanup fields, and related server-managed members.
- Completed: oversize response includes `requested_byte_size`, `configured_limit_bytes`, and creates no durable `object_blobs` row or idempotency success payload.
- Completed: 60-minute upload target expiry and 24-hour pending slot expiry are preserved.
- Completed: same-slot target refresh and resumable upload were not added.
- Completed: blob creation remains route-only; it does not create or mutate evidence rows, record links, preview handles, release state, or workflow objects.

Validation commands:
- Passed: `go test ./internal/modules/evidence -run 'TestPhase5_.*(U_5_01|U_5_02|U_5_09|I_5_02)'`
- Sprint 1 rows passed inside `make phase-slice PHASE=phase5`: `U-5-01`, `U-5-02`, `U-5-09`, and `I-5-02`.
- Expected failure: `make backend-unit` still fails on later Sprint 3 placeholders `U-5-05` and `U-5-07`.
- Expected failure: `make backend-store` still fails on later Sprint 3 placeholder `U-5-06`.
- Expected failure: `make backend-integration` still fails on later Sprint 3/4 placeholders `I-5-03` and `I-5-04`.
- Expected failure: `make phase-slice PHASE=phase5` still fails only because the later placeholder rows above are intentionally skipped.

Deliverables:
- Delivered: passing `U-5-01`, `U-5-02`, blob-create `U-5-09`, and `I-5-02`.
- Delivered: blob create route remains route-only and does not mutate evidence rows.

Risks and assumptions:
- Existing `TestPhase4_ObjectBlobCreate_I_4_BLOB_01` overlaps; keep it as Phase 4 route-shape smoke or split assertions into new Phase 5 rows without deleting useful regression coverage.
- Safe assumption: upload target `href` remains opaque and may be app-local for filesystem object storage.
- Retained assumption: malformed non-empty `incident_id` has no dedicated blob-create reason token in the current registry; no new public reason code was introduced.

Exit criteria:
- Met: blob-create failures leave no `object_blobs`, no `upload_target`, and no idempotency success payload.
- Met: expired replay never refreshes the original slot.

## Sprint 2. Attach Finalization, Lifecycle Bridge, and No-Blob Evidence

Objective: Complete evidence attachment finalization, no-blob evidence records, lifecycle separation, idempotent attach replay, and workbook-visible projection consequences.

Status: Complete. Sprint 2 covers the existing attach route and schema only. It does not add broad custody UX, scanner/quarantine workflows, cleanup workers, or the full workbook attach UI planned for later sprints.

Relevant IDs:
- `U-5-03`, `U-5-04`, `U-5-08`, `I-5-01`, `E-5-04`
- `REQ-01-245..REQ-01-247`, `REQ-02-186..REQ-02-201`, `REQ-03-116..REQ-03-126`, `REQ-03-242..REQ-03-246`
- `AC-004`, `AC-015`, `AC-016`, `AC-102`, `AC-103`, `AC-154`, `AC-155`

Grep references:
- `DecodeAttachBlobRequest`
- `AttachBlobRequestHash`
- `Store.AttachBlob`
- `ErrBlobNotAttachable`
- `ErrRowVersionConflict`
- `markBlobAvailableTx`
- `failBlobTx`
- `finalize_attempt_count`
- `terminal_reason`
- `declared_size_mismatch`
- `expected_sha256_mismatch`
- `pending_timeout`
- `evidence.lifecycle_state`
- `evidence.upload_state`
- `timeline.evidence_count`
- `timeline.has_evidence`
- `host.evidence_count`
- `identity.evidence_count`

Files and areas:
- `internal/modules/evidence/store.go`
- `internal/modules/evidence/routes.go`
- `internal/modules/workbook/mutation_store.go`
- `internal/modules/projections/entity_grids.go`
- `db/queries/timeline_phase3.sql`
- Generated from `make generate`: `internal/gen/sql/timeline_phase3.sql.go`
- `internal/modules/evidence/phase5_attach_test.go`
- `internal/modules/evidence/phase5_lifecycle_test.go`
- `internal/modules/evidence/phase5_integration_test.go`
- `apps/web/src/WorkbookShell.phase5.test.tsx`
- `apps/web/e2e/phase5.evidence.spec.ts`

Test-first sequence:
1. Completed: `U-5-03` covers attach request shape, row-version conflict, idempotent replay, missing/failed/expired/incident-mismatched/quarantined/size-mismatched/SHA-mismatched blob failures, terminal reason persistence, retry exhaustion, and no evidence mutation on failure.
2. Completed: `U-5-04` proves `evidence.lifecycle_state` and `object_blobs.upload_state` remain separate, requested/pending/received evidence can exist without `object_blob_id`, illegal lifecycle transitions are rejected, attach advances no-blob evidence, and structured no-blob fields are preserved until attach replaces the object locator.
3. Completed: `U-5-08` proves workbook record-change invalidation refreshes evidence-count fields without changing the active surface.
4. Completed: `I-5-01` proves real object upload -> attach -> workbook query readback returns one evidence row, no duplicate attachment rows, and stable idempotent replay.
5. Completed: `E-5-04` proves the browser can create requested evidence through the Evidence surface, advance it later through the two-step backend attach path, and keep workbook rows stable.

Implementation tasks:
- Completed: no new migration was added; existing blob/evidence fields are sufficient for this sprint, and custody-event/owner-user workflow remains later owner-driven work.
- Completed: legal evidence lifecycle transitions are enforced for ordinary workbook patches, with illegal transitions and bridge guard violations rejected as `409 illegal_transition`.
- Completed: successful attach updates evidence row, revision/change set, idempotency payload, and affected workbook projections atomically.
- Completed: failed pending finalization records `terminal_reason`, `failed_at`, and `cleanup_due_at` without mutating evidence state.
- Completed: non-terminal finalization failures increment `finalize_attempt_count`; the 4th failure marks `finalize_retry_exhausted`.
- Completed: projection refresh uses active `record_links` with `link_type='supported_by'`, `src_record_id=<workbook row>`, and `dst_record_id=<evidence record>`. Counted evidence must be `available` or `released` and linked to an `object_blobs.upload_state='available'` blob.
- Completed: attach publishes ordinary `record_changed` invalidations for affected source rows and does not add dependent workbook rows inline to the attach response.

Validation commands:
- Passed: `make generate`
- Passed: `make format`
- Passed: `git diff --check`
- Passed: `go test ./internal/modules/evidence ./internal/modules/workbook ./internal/modules/projections -run 'TestPhase5_.*(U_5_03|U_5_04|I_5_01)'`
- Passed: `make frontend-unit`
- Passed: `make frontend-typecheck`
- Passed: `make browser-e2e-webserver-backed`
- Expected failure: `make backend-store` still fails because later Sprint 3 row `U-5-06` is skipped.
- Expected failure: `make backend-integration` still fails because later Sprint 3/4 rows `I-5-03` and `I-5-04` are skipped.
- Expected failure: `make service-backed-slice PHASE=phase5` now passes the Sprint 2 browser/backend rows and fails only because of `U-5-06`, `I-5-03`, and `I-5-04`.

Deliverables:
- Delivered: passing attach validation/finalization coverage for `U-5-03`.
- Delivered: passing no-blob lifecycle bridge coverage for `U-5-04`.
- Delivered: passing frontend workbook invalidation coverage for `U-5-08`.
- Delivered: passing real object upload/attach/workbook readback coverage for `I-5-01`.
- Delivered: passing browser requested-evidence advancement coverage for `E-5-04`.
- Delivered: evidence rows can exist with no blob until successful attach.
- Delivered: successful binary attach updates workbook-visible Timeline, Host, and Identity evidence counts through projection invalidation.

Risks and assumptions:
- Retained ambiguity: the guide requires custody history preservation but current schema has no append-only custody-event owner path. Sprint 2 preserves structured no-blob fields and does not invent custody UX or event tables.
- Resolved: evidence support uses the active `record_links` direction `src_record_id=<workbook row>`, `dst_record_id=<evidence record>`, `link_type='supported_by'`.
- Retained follow-up: `E-5-01..E-5-03` remain in Sprint 5 with the full workbook attach/preview UI.

Exit criteria:
- Met: replaying attach with the same normalized request creates no second `change_set`.
- Met: pending unreadable, failed, missing, quarantined, expired, incident-mismatched, size-mismatched, or SHA-mismatched blobs cannot produce visible attached evidence.
- Met: requested/no-blob evidence can be created, legally advanced, later attached, and counted only after blob availability.

## Sprint 3. Handle Issuance and Redeem Hardening

Objective: Complete safe preview/download issuance and redemption with current authorization, session binding, state invalidation, expiry, single-use download, filename sanitation, range behavior, and preview-size limits.

Relevant IDs:
- `U-5-05`, `U-5-06`, `U-5-07`, `U-5-09`, `I-5-03`, `E-5-03`
- `REQ-01-458..REQ-01-465`, `REQ-04-023`, `REQ-04-053`
- `AC-251`, `AC-252..AC-255`, `AC-321`, `AC-322`

Grep references:
- `DecodeHandleIssueRequest`
- `Store.LoadEvidenceAccess`
- `classifyEvidenceAccess`
- `Store.InsertHandle`
- `Store.LoadHandle`
- `Store.CheckHandleAccess`
- `Store.ConsumeDownloadHandle`
- `Service.handleIssueHandle`
- `Service.handleRedeemHandle`
- `sanitizeFilename`
- `classifyMedia`
- `evidence_access_unavailable`
- `unsupported_preview`
- `preview_payload_too_large`
- `handle_not_found_or_revoked`
- `handle_expired`
- `handle_consumed`

Files and areas:
- `internal/modules/evidence/api.go`
- `internal/modules/evidence/routes.go`
- `internal/modules/evidence/store.go`
- `contracts/errors/index.json`
- `contracts/openapi/cartulary.openapi.yaml`

Test-first sequence:
1. Add `U-5-05` request-shape and non-idempotency tests: `{}` only; `client_txn_id` rejected; repeated issuance creates fresh `href`.
2. Add `U-5-06` for unsupported preview, pending/failed/missing/quarantined/inconsistent states, and no silent fallback to download.
3. Add `U-5-07` for `Content-Disposition` and filename fallback: sanitize `/`, `\`, NUL, CR, LF; fallback `evidence-<record_id><canonical_extension_if_known>`.
4. Add `U-5-09` preview ceiling test with `limits.previews.max_previewable_payload_bytes` and `limits.previews.max_text_inline_bytes`; download remains legal.
5. Add `I-5-03`: redeem fails after logout/session revoke, membership removal, blob replacement/detach, evidence delete, quarantine, or failed/pending transition.

Implementation tasks:
- Ensure standard auth/session failures happen before handle lookup.
- Bind handle to issuing session, incident, `record_id`, `object_blob_id`, handle kind, filename, disposition, and preview kind.
- Keep preview handles reusable until 5-minute expiry; keep download handles single-use after first byte delivery or validated redirect.
- Return `Content-Disposition: inline` for preview and `attachment` for download.
- Add `filename*=` if practical through existing header helper; at minimum do not expose raw storage keys or caller-provided overrides.

Validation commands:
- `go test ./internal/modules/evidence -run 'TestPhase5_.*(U_5_05|U_5_06|U_5_07|U_5_09|I_5_03)'`
- `make backend-integration`
- `make go-gosec-targeted`

Deliverables:
- Passing handle and redemption tests.
- Handles never expose bucket names, object keys, or long-lived object-store credentials.

Risks and assumptions:
- Existing `mime.FormatMediaType` may not emit both `filename=` and `filename*=`; if adding dual-parameter support is invasive, document the residual gap and keep sanitation/fallback exact.
- Browser range requests must not consume preview handles and must consume download handles only after successful delivery starts.

Exit criteria:
- All redemption invalidation cases fail closed with registered errors.
- Oversized preview fails with `preview_payload_too_large`; download still succeeds.

## Sprint 4. Object-Storage Boundary and Cleanup/Quarantine Behavior

Objective: Prove binary payloads remain outside Postgres, object-store loss fails preview/download without corrupting structured rows, and cleanup/quarantine boundaries do not bypass the two-step attach model.

Relevant IDs:
- `I-5-04`, `AC-053`, `AC-405`
- `REQ-01-002`, `REQ-01-278..REQ-01-280`, `REQ-02-187`, `REQ-02-194..REQ-02-196`, `REQ-03-123..REQ-03-126`, `REQ-04-048`, `REQ-04-053`

Grep references:
- `objectstore.Store`
- `ReadObject`
- `StatObject`
- `DeleteObject`
- `object_blobs.storage_key`
- `evidence_access_handles`
- `terminal_reason`
- `cleanup_due_at`
- `cleaned_up_at`
- `evidence_quarantined`
- `evidence_inconsistent`

Files and areas:
- `internal/platform/objectstore/objectstore.go`
- `internal/modules/evidence/store.go`
- `internal/modules/evidence/routes.go`
- `internal/testutil/s3test`
- `internal/testutil/suiteservices`
- New cleanup worker only if existing background-job shell can host it without widening scope.

Test-first sequence:
1. Add `AC-405` integration test: attach unique binary marker bytes, confirm structured DB dump/query does not contain payload bytes, delete object bytes, verify row remains and preview/download fail `blob_missing`.
2. Add cleanup test for failed unattached blobs: `cleanup_due_at`, object deletion, `cleaned_up_at`.
3. Add quarantine test for available blob -> quarantined: evidence row becomes or surfaces `quarantined`, preview/download fail `evidence_quarantined`.
4. Add `I-5-04` scanner-adjunct test as conditional: if no scanner adjunct exists, assert no bypass path exists and quarantine state blocks access.

Implementation tasks:
- Add narrow cleanup function for expired pending slots and failed unattached object bytes if missing.
- Implement legal quarantine transitions only if Phase 5 tests require them; do not build a generalized malware scanner.
- Keep object-store loss as access failure, not structured-row deletion.
- Avoid storing binary evidence inline in Postgres, JSON fields, revisions, or idempotency payloads.

Validation commands:
- `go test ./internal/modules/evidence -run 'TestPhase5_.*(I_5_04|AC_405|Cleanup|Quarantine)'`
- `make backend-integration`
- `make go-gosec-targeted`
- `make service-backed-slice PHASE=phase5`

Deliverables:
- Passing object-boundary and quarantine/cleanup tests.
- Documented no-op interpretation for scanner adjunct if none exists.

Risks and assumptions:
- Ambiguity: guide references an upload-scanning adjunct "if the deployment uses" one. Safest assumption is no scanner adjunct in base unless already present; Phase 5 should prove quarantine blocks access, not implement scanner infrastructure.
- Risk: idempotency payloads or revision snapshots might accidentally include too much object metadata; tests should search for exact payload marker bytes.

Exit criteria:
- Postgres structured state contains metadata and object references only.
- Losing object bytes does not delete or corrupt committed evidence rows.

## Sprint 5. Workbook UI and Browser Evidence Flows

Objective: Complete analyst-visible evidence upload, attach, preview, download, blocked-preview, and requested-evidence flows without leaving the workbook surface.

Relevant IDs:
- `E-5-01`, `E-5-02`, `E-5-03`, `E-5-04`
- `REQ-03-102`, `REQ-03-116..REQ-03-128`, `REQ-03-242..REQ-03-246`, `REQ-01-458..REQ-01-465`
- `AC-002`, `AC-004`, `AC-015`, `AC-016`, `AC-102`, `AC-103`, `AC-154`, `AC-155`, `AC-252..AC-255`

Grep references:
- `apps/web/src/WorkbookShell.tsx`
- `issueEvidenceHandle`
- `evidence-preview-panel`
- `evidence-preview-frame-`
- `evidence-access-message-`
- `timeline.evidence_count`
- `timeline.has_evidence`
- `object-blobs`
- `attach-blob`
- `preview-handle`
- `download-handle`
- `apps/web/e2e/phase5.evidence.spec.ts`

Files and areas:
- `apps/web/src/WorkbookShell.tsx`
- `apps/web/src/WorkbookShell.phase5.test.tsx`
- `apps/web/e2e/phase5.evidence.spec.ts`
- `apps/web/e2e/phase4Helpers.ts`
- `apps/web/e2e/helpers.ts`

Test-first sequence:
1. Add frontend unit tests for evidence attach payload flow: create blob slot, upload target, attach blob with `base_row_version`, update selected row state, do not navigate away.
2. Add browser `E-5-01`: attach screenshot to selected Timeline row; verify grid remains and `timeline.evidence_count` increments.
3. Add browser `E-5-02`: screenshot-only Timeline row persists through two-step evidence path.
4. Add browser `E-5-03`: safe inline type previews through same-origin `href`; unsupported/unsafe type surfaces blocked state inline.
5. Completed in Sprint 2: browser `E-5-04` creates requested no-blob evidence, later attaches a blob through the backend two-step path, and verifies the Evidence surface row remains stable.

Implementation tasks:
- Add UI affordance for file/clipboard/drop upload only as needed for Phase 5 tests; avoid broad file manager or scanner UI.
- Use `POST /api/v1/object-blobs`, upload to `upload_target.href`, then `POST /api/v1/evidence-records/{record_id}/attach-blob`.
- Treat `accepted_contract` as server-approved; do not rebuild it from local state.
- Keep `href` opaque and same-origin for preview/download.
- Surface blocked preview reason inline with the grid/inspector still present.

Validation commands:
- `make frontend-unit`
- `make browser-e2e-webserver-backed`
- `make browser-e2e-visual` only if screenshots are added/changed.
- `make phase-slice PHASE=phase5`
- `make service-backed-slice PHASE=phase5`

Deliverables:
- Passing `E-5-01..E-5-03`; `E-5-04` is already delivered by Sprint 2.
- Browser evidence flows stay workbook-native and do not open object-store URLs.

Risks and assumptions:
- Browser clipboard APIs can be flaky in Playwright. Safest implementation is to support both file input/drop helper and paste, with E2E using the most deterministic path that still satisfies "screenshot attachment".
- Avoid marketing-style or separate evidence module UI; keep controls inside workbook surface/inspector.

Exit criteria:
- User-visible grid remains in place during attach and blocked preview.
- Evidence count and requested/available lifecycle state are visible after commit.

## Sprint 6. Phase Gate, Ledgers, Baselines, and Handoff Cleanup

Objective: Make Phase 5 selectable, validated, and handoff-ready without hiding residual gaps.

Relevant IDs:
- All `U-5-*`, `I-5-*`, `E-5-*`
- Phase-owned AC set from guide section 16.2 row for Phase 5.

Files and areas:
- `tools/phase5_test_map.json`
- `docs/testing/phase5_coverage_ledger.md`
- `tools/go_test_duration_baselines.json`
- Browser duration baselines if new Playwright specs are authoritative.
- `PHASE5_IMPLEMENTATION_PLAN.md` checklist updates.

Test-first sequence:
1. Confirm every Phase 5 row has a passing authoritative symbol.
2. Run drift and ledgers before broad gates.
3. Refresh baselines only from successful uncontaminated service-backed results.

Implementation tasks:
- Update the checklist in `PHASE5_IMPLEMENTATION_PLAN.md`.
- Mark known unsupported scanner adjunct behavior explicitly if no scanner exists.
- Keep Phase 4 support rows support-only.
- Avoid spec owner edits unless the implementation revealed a real normative ambiguity.

Validation commands:
- `make phase-ledgers`
- `make phase-ledger-drift`
- `make phase-slice PHASE=phase5`
- `make service-backed-slice PHASE=phase5`
- `make test-fast`
- `make check`
- If service-backed timing artifacts changed: `make go-test-duration-baselines RESULTS_DIR=<successful-run-dir>` and `make go-test-duration-baseline-drift RESULTS_DIR=<successful-run-dir>`.
- If browser timings changed: `make browser-e2e-duration-baseline-drift RESULTS_DIR=<successful-run-dir>`.

Deliverables:
- Phase 5 manifest and ledger are current.
- Checklist records completion, validation status, blockers, and follow-up notes.
- No generated drift, migration drift, or unmapped test IDs remain.

Risks and assumptions:
- `make check` may be costly; run narrower phase commands first.
- If Phase 5 browser rows require new batch metadata, update only the manifest-owned scheduler artifacts generated by existing commands.

Exit criteria:
- `make phase-slice PHASE=phase5` and `make service-backed-slice PHASE=phase5` pass.
- `make check` passes or has a documented unrelated failure with artifact paths and rerun command.
- Every Phase 5 AC in the guide has either passing evidence or an explicit, owner-backed `N/A` rationale.

## Ambiguities and Safe Defaults

- Scanner adjunct: The guide says `I-5-04` applies if deployed. Default to no scanner adjunct unless existing code/config proves one exists; still test quarantine and fail-closed access.
- Custody events: Core 02 requires append-only custody history where custody narrative/handoff commentary is preserved. Do not invent a broad custody subsystem unless a failing Phase 5 test proves existing structured fields are insufficient.
- Filename header: Core 01 says the header should include both ASCII `filename=` and Unicode `filename*=`. Treat sanitation and deterministic fallback as required; implement dual parameter if feasible without custom brittle header code.
- Screenshot-only Timeline create: Keep scope to the two-step evidence path required by `E-5-02`; do not build broad import, file management, or report attachment features.
- Discoverability update: Root-level `PHASE5_IMPLEMENTATION_PLAN.md` is sufficient. Do not touch `README.md` or normative docs for discoverability unless explicitly requested.
