# Phase 5 Implementation Plan

## Summary

This file is the execution roadmap and progress marker for Cartulary Phase 5: Evidence, blob lifecycle, evidence access, and object storage.

`docs/guides/cartulary_implementation_testing_guide.md` section 7.5 is the controlling implementation-scope reference for this phase. Normative behavior remains owned by the core documents, especially Core 01 section 3.3.8, Core 01 section 16, Core 02 sections 4.5 and 13, Core 03 sections 8 and 16, and Core 04 sections 4.3 and 4.5.

This planning artifact does not implement Phase 5 behavior. It is intentionally root-level so agents can find it quickly during handoff or interrupted implementation sessions. No README update is required for discoverability.

## Sprint Checklist

| Done | Sprint | Validation | Blockers | Follow-up Notes |
| --- | --- | --- | --- | --- |
| [x] | 0. Phase 5 ownership manifest and harness setup | [x] validated | Intentional TDD failures remain for `U-5-01`, `U-5-02`, `U-5-09`, and `I-5-02`. | Phase 5 is active and selectable; Sprint 1 should replace the failing stubs with real assertions. |
| [ ] | 1. Blob create and upload-slot contract | [ ] pending |  |  |
| [ ] | 2. Attach finalization, lifecycle bridge, and no-blob evidence | [ ] pending |  |  |
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
- Expected failure: `make phase-slice PHASE=phase5` selects Phase 5 rows and fails on the intentional Sprint 1 backend stubs.
- Expected failure: `go test ./internal/modules/evidence -run 'TestPhase5_.*(U_5_01|U_5_02|U_5_09|I_5_02)'` fails on the four intentional Sprint 1 stubs.

Deliverables:
- Delivered: `tools/phase5_test_map.json`
- Delivered: `docs/testing/phase5_coverage_ledger.md`
- Delivered: initial failing Phase 5 Sprint 1 symbols.

Risks and assumptions:
- Assumption retained: create `internal/testutil/phase5test` only when naming clarity or helper ownership is needed; otherwise reuse existing helpers.
- Follow-up risk: service-backed timing baselines may need refresh after Sprint 1 replaces failing stubs with real passing evidence.

Exit criteria:
- Met: `make explain-phase PHASE=phase5` discovers the active manifest and planned rows.
- Met: no Phase 5 authoritative ID appears in support-only files.

## Sprint 1. Blob Create and Upload-Slot Contract

Objective: Complete `POST /api/v1/object-blobs` request validation, normalization, idempotency, accepted contract echo, size ceiling, and expired-slot replay behavior.

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
1. Add/rename unit decoder tests for `U-5-01`: required `incident_id`, `client_txn_id`, `byte_size`, unknown fields, server-managed fields, optional omission/null equivalence.
2. Add store/route idempotency tests for `U-5-02`: same `(actor_user_id, incident_id, client_txn_id)` returns original slot; divergent normalized request returns `client_txn_conflict`.
3. Add size-limit tests for `U-5-09`: reject before slot creation with `413 blob_create_rejected`.
4. Add service-backed `I-5-02`: replay of expired slot returns original expired `target_expires_at` and `pending_expires_at`; fresh target requires new `client_txn_id`.

Implementation tasks:
- Ensure `invalid_blob_create_request` reason codes use registry tokens: `request_not_object`, `field_not_nullable`, `field_empty_after_normalization`, `invalid_byte_size`, `invalid_sha256_hex`, `unknown_field`, `server_managed_field`.
- Ensure oversize response includes `requested_byte_size`, `configured_limit_bytes`, and no durable `object_blobs` row.
- Preserve 60-minute upload target expiry and 24-hour pending slot expiry.
- Do not add same-slot target refresh or resumable upload.

Validation commands:
- `go test ./internal/modules/evidence -run 'TestPhase5_.*U_5_0(1|2|9)|TestPhase5_.*I_5_02'`
- `make backend-unit`
- `make backend-integration`

Deliverables:
- Passing `U-5-01`, `U-5-02`, `U-5-09`, `I-5-02`.
- Blob create route remains route-only and does not mutate evidence rows.

Risks and assumptions:
- Existing `TestPhase4_ObjectBlobCreate_I_4_BLOB_01` overlaps; keep it as Phase 4 route-shape smoke or split assertions into new Phase 5 rows without deleting useful regression coverage.
- Safe assumption: upload target `href` remains opaque and may be app-local for filesystem object storage.

Exit criteria:
- Blob-create failures leave no `object_blobs`, no `upload_target`, and no idempotency success payload.
- Expired replay never refreshes the original slot.

## Sprint 2. Attach Finalization, Lifecycle Bridge, and No-Blob Evidence

Objective: Complete evidence attachment finalization, no-blob evidence records, lifecycle separation, idempotent attach replay, and workbook-visible projection consequences.

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

Files and areas:
- `internal/modules/evidence/store.go`
- `internal/modules/evidence/routes.go`
- `internal/modules/workbook/mutation_store.go`
- `internal/modules/projections/store.go`
- `db/migrations/00009_phase4_evidence_blob_routes.sql`
- `internal/modules/workbook/workbook_mutation_integration_test.go`

Test-first sequence:
1. Add `U-5-03` tests for attach request shape, `base_row_version`, pending/failed/missing/expired/incident-mismatched/contract-mismatched blob failures.
2. Add `U-5-04` tests proving `evidence.lifecycle_state` and `object_blobs.upload_state` remain separate, and requested/pending evidence can exist without `object_blob_id`.
3. Add `U-5-08` projection tests: attaching evidence updates `timeline.evidence_count` and `timeline.has_evidence` for linked source rows without inline extra row payloads.
4. Add `I-5-01`: real object upload -> attach -> workbook query returns one evidence row and no duplicate attachment rows.
5. Add `E-5-04` browser flow only after backend row behavior is stable.

Implementation tasks:
- Add any missing schema for `owner_user_id` or custody event persistence only if required for `REQ-02-199..REQ-02-201`; otherwise record ambiguity and avoid inventing broader custody UX.
- Enforce legal lifecycle transitions from `REQ-02-190`; reject illegal direct writes with `illegal_transition` where the existing mutation path supports it.
- On successful attach, update evidence row, revision/change set, and affected projections atomically.
- Ensure failed pending finalization records `terminal_reason`, `failed_at`, and cleanup timing without mutating evidence state.
- Add projection refresh for Timeline evidence counts through `record_links link_type='supported_by'` or the route-specific link model chosen by existing code; do not invent a second relationship type.

Validation commands:
- `go test ./internal/modules/evidence ./internal/modules/workbook ./internal/modules/projections -run 'TestPhase5_.*(U_5_03|U_5_04|U_5_08|I_5_01)'`
- `make backend-store`
- `make backend-integration`
- `make service-backed-slice PHASE=phase5`

Deliverables:
- Passing attach and lifecycle tests.
- Evidence rows can exist with no blob.
- Successful binary attach updates workbook-visible counts.

Risks and assumptions:
- Ambiguity: the guide requires custody history preservation but current schema may not have custody-event rows. Safest assumption is to preserve existing structured fields and add append-only custody events only if Core 02 section 13 assertions cannot be satisfied without them.
- Risk: `record_links` direction for evidence support must match existing query/projection conventions.

Exit criteria:
- Replaying attach with same normalized request creates no second `change_set`.
- Pending, failed, missing, quarantined, expired, or mismatched blobs cannot produce visible attached evidence.

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
5. Add browser `E-5-04`: create requested no-blob evidence, later attach blob, counts/pivots remain stable.

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
- Passing `E-5-01..E-5-04`.
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
