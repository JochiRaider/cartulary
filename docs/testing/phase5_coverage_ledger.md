# Phase 5 Coverage Ledger

This ledger is generated from `tools/phase5_test_map.json`. Update the manifest row metadata first, then regenerate this file.

- Scope: Evidence, blob lifecycle, evidence access, and object storage.
- Normative owners: Core 01 §3.3.8; Core 01 §16; Core 02 §4.5 and §13; Core 03 §8 and §16; Core 04 §4.3 and §4.5.
- Authority: `tools/phase5_test_map.json` is the enforced Phase 5 traceability source. This ledger is a rendered companion and does not control the mechanical row inventory.
- Sprint 0 activates the Phase 5 ownership map before feature work. Initial non-Sprint-1 rows are selectable placeholders until their owning behavior is implemented.

## Authoritative Execution

- `backend-unit` selects pure Phase 5 request, response, and route-contract rows through the Phase 5 manifest and `backend_unit` execution dependency.
- `backend-store` selects service-backed Phase 5 unit rows that need Postgres or object-store fixtures through manifest-owned store-domain selection.
- `frontend-unit` selects Phase 5 workbook component rows through the Phase 5 manifest for `frontend_unit`.
- `backend-integration` selects Phase 5 real-runtime HTTP and object-store rows through the Phase 5 manifest and `backend_integration` execution dependency.
- `browser-e2e-webserver-backed` carries authoritative `E-5-*` rows through duration-balanced Playwright manifest-entry shards under `browser_functional`.
- `browser-e2e-visual` carries authoritative `V-5-*` visual rows through `browser_visual`.

## Support-Only Execution

- Existing Phase 4 object-blob and handle route smoke tests remain Phase 4-owned carryover evidence and must not claim Phase 5 IDs.
- `backend-process` carries supplemental standalone-process Phase 5 smoke evidence only; store, integration, browser, and visual rows remain authoritative for Phase 5 behavior.

## Unit

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |
| `U-5-01` | `implemented` | `internal/modules/evidence/blob_create_test.go::TestObjectBlobCreate_Unit` | `backend_store` | `POST /api/v1/object-blobs` enforces required fields, malformed and unknown-member rejection, auth and membership denial, normalized success envelopes, expiry fields, upload targets, and no durable state for rejected requests. | Idempotent replay, object-store upload, evidence attachment, and workbook projection consequences remain separate Phase 5 rows. |
| `U-5-02` | `implemented` | `internal/modules/evidence/blob_create_test.go::TestBlobCreateIdempotency_Unit` | `backend_store` | Blob-create idempotency is keyed by `(actor_user_id, incident_id, client_txn_id)`, replays the same normalized request to the original slot, and rejects divergent replay with the route-owned conflict. | Expired-slot replay and fresh-target behavior are integration evidence in `I-5-02`. |
| `U-5-03` | `implemented` | `internal/modules/evidence/attach_test.go::TestAttachBlobValidation_Unit` | `backend_store` | `POST /api/v1/evidence-records/{record_id}/attach-blob` requires `object_blob_id`, `base_row_version`, and `client_txn_id`; the store rejects divergent attach replay and fails closed for pending, failed, missing, expired, incident-mismatched, or contract-mismatched blobs. | End-to-end object upload and workbook query readback are covered by integration and browser rows. |
| `U-5-04` | `implemented` | `internal/modules/evidence/lifecycle_test.go::TestEvidenceLifecycleSeparateFromBlob_Unit` | `backend_store` | Evidence lifecycle state remains separate from blob `upload_state`, allowing requested or pending-receipt evidence without a blob and later advancement without unrelated custody-history mutation. | Generalized custody workflow UX and scanner infrastructure are not completed by this row. |
| `U-5-05` | `implemented` | `internal/modules/evidence/handles_test.go::TestHandleIssueEmptyBodyNonIdempotent_Unit` | `backend_store` | Preview-handle and download-handle issuance routes accept only `{}`, reject idempotency members without side effects, return complete same-origin opaque handle metadata, set route-owned expiry and single-use flags, and issue fresh handles on every success. | Handle redemption, byte streaming, and current auth or membership loss after issuance are covered by `U-5-06`, `U-5-07`, and `I-5-03`. |
| `U-5-06` | `implemented` | `internal/modules/evidence/handles_test.go::TestHandleRedemptionRechecksCurrentState_Unit` | `backend_store` | Handle access revalidates current evidence/blob state at the store layer, and preview blocks explicitly for unsupported or unsafe preview conditions. | HTTP session and membership re-derivation during redemption is covered by `I-5-03`; browser rendering of preview outcomes is covered by `E-5-03`. |
| `U-5-07` | `implemented` | `internal/modules/evidence/handles_test.go::TestDownloadDispositionFallback_Unit`, `TestSanitizeFilenameRemovesNUL_Unit` | `backend_store` | Download responses use authoritative metadata for filename disposition and apply deterministic fallback naming when authoritative names are absent. | Authorization re-check and state invalidation during redemption are covered by `U-5-06` and `I-5-03`. |
| `U-5-08` | `implemented` | `apps/web/src/workbook/WorkbookShell.evidence.test.tsx::reflects attached evidence counts on the workbook surface without forcing navigation` | `frontend_unit` | A workbook invalidation refresh visibly updates Timeline evidence count and `has_evidence` cells without forcing navigation away from the current workbook surface. | Real object upload and backend projection rebuild behavior are integration evidence. |
| `U-5-09` | `implemented` | `internal/modules/evidence/blob_create_test.go::TestBlobCreateSizeCeiling_Unit`, `TestPreviewPayloadCeiling_Unit` | `backend_store` | Blob-create size ceilings reject oversize `byte_size` before slot creation, and configured image and text-inline preview ceilings fail with the route-owned oversized-preview reason while legal downloads remain available. | Expired-slot replay and full object upload completion remain separate integration evidence. |
| `U-5-10` | `implemented` | `internal/modules/evidence/conformance_test.go::TestReasonCodeRegistryConformance_Unit` | `backend_unit` | Evidence create, attach, preview issuance, and redemption use closed public error and reason-code registries, with attach failures owned by `evidence_attach_rejected` and no legacy unregistered attach reasons. | General non-evidence reason-code registry expansion is outside this Phase 5 row. |
| `U-5-11` | `implemented` | `internal/modules/evidence/conformance_test.go::TestEvidenceFieldKeyRegistryClosure_Unit` | `backend_unit` | Evidence and Timeline evidence-derived field keys are closed over the derived view-schema registry and reject label, storage-name, or non-canonical key fallbacks. | Whole-repository view-schema registry conformance beyond evidence-owned Phase 5 fields remains broader contract hygiene. |

## Integration

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |
| `I-5-01` | `implemented` | `internal/modules/evidence/integration_test.go::TestObjectUploadAttachWorkbookProjection_Integration` | `backend_integration` | A real object-store upload can be created through a blob slot, finalized onto an evidence record, linked to Timeline through the workbook mutation route, and queried through the workbook projection without duplicate attachment rows or links. | Browser-level attach interaction remains `E-5-01` and `E-5-02`. |
| `I-5-02` | `implemented` | `internal/modules/evidence/integration_test.go::TestExpiredSlotReplay_Integration` | `backend_integration` | Expired-slot replay returns the same expired slot, while a fresh upload target requires a new `client_txn_id`. | General idempotency for non-expired slots is unit/store evidence in `U-5-02`. |
| `I-5-03` | `implemented` | `internal/modules/evidence/integration_test.go::TestHandleRedeemInvalidatesOnCurrentStateLoss_Integration` | `backend_integration` | A redeemed preview or download handle fails closed after logout, membership removal, or evidence/blob state invalidation. | Frontend display of blocked preview outcomes is covered by `E-5-03`. |
| `I-5-04` | `implemented` | `internal/modules/evidence/integration_test.go::TestQuarantineBoundaryPreservesTwoStepAttach_Integration` | `backend_integration` | When quarantine or scanning state is present, two-step attach semantics remain intact and preview never bypasses the quarantine boundary. | Implementing a generalized scanner adjunct is outside the base Phase 5 row. |
| `I-5-05` | `implemented` | `internal/modules/evidence/integration_test.go::TestAttachRouteContract_Integration` | `backend_integration` | Attach route validation, authorization, route-owned error envelopes, and divergent replay rejection occur through the real HTTP surface before object observation mutates state. | Store-domain blob lifecycle retry counters remain `U-5-03` evidence. |
| `I-5-06` | `implemented` | `internal/modules/evidence/integration_test.go::TestAttachedEvidenceProjectionRebuild_Integration` | `backend_integration` | Attached-evidence projection storage can be corrupted and deterministically rebuilt from source without mutating source rows, links, or history, while the public workbook surface preserves correct evidence state. | Introducing a dedicated `evidence_grid_projection` is tracked separately as broader Core 01 projection-substrate work. |
| `I-5-07` | `implemented` | `internal/modules/evidence/integration_test.go::TestAttachPublishesWorkbookWebSocketRefresh_Integration` | `backend_integration` | A successful evidence attach publishes canonical `record_changed` messages for the Evidence row and affected Timeline evidence projection through the real incident WebSocket emitter. | Phase 6 continues to own general WebSocket resume, replay windows, presence, and revocation semantics. |

## Browser E2E

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |
| `E-5-01` | `implemented` | `apps/web/e2e/evidence.spec.ts::attaches a screenshot to a selected Timeline row without leaving the workbook surface` | `browser_functional` | An analyst attaches a screenshot to a selected Timeline row without leaving the workbook surface, and the row reflects the new evidence count after commit. | Backend object-store finalization is covered by `I-5-01`. |
| `E-5-02` | `implemented` | `apps/web/e2e/evidence.spec.ts::persists a screenshot-only Timeline row through the two-step evidence path` | `browser_functional` | A screenshot-only Timeline row can be persisted through the two-step evidence path. | Generic Timeline row create behavior remains Phase 3 evidence. |
| `E-5-03` | `implemented` | `apps/web/e2e/evidence.spec.ts::redeems inline-safe previews and shows explicit blocked-preview outcomes` | `browser_functional` | An inline-safe type receives a preview handle and renders through the same-origin redeem path, while unsupported or unsafe types return an explicit blocked-preview outcome. | Low-level handle issuance and redemption checks remain backend evidence. |
| `E-5-04` | `implemented` | `apps/web/e2e/evidence.spec.ts::tracks requested evidence before a blob exists and later advances it` | `browser_functional` | Requested evidence can be tracked before the blob exists, linked to a Timeline row without counting unavailable content, advanced to an available counted state, and preserved across Evidence and Timeline workbook pivots. | A generalized custody or collection workflow is outside this row. |
| `E-5-05` | `implemented` | `apps/web/e2e/evidence.spec.ts::refreshes a second live workbook from the real evidence attach stream` | `browser_functional` | A second live workbook session receives the real `/ws/v1/incidents/{incident_id}` evidence attach event and refreshes Timeline evidence count and flags without navigation. | General collaboration socket lifecycle semantics remain Phase 6 ownership. |

## Visual Regression

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |
| `V-5-GRID-01` | `implemented` | `apps/web/e2e/workbook.visual.spec.ts::captures requested and available Evidence states on the required Evidence surface` | `browser_visual` | The visual harness captures requested evidence and the same row after workbook-surface attach advances it to available. | Low-level object upload and attach validation remain backend evidence. |
| `V-5-GRID-02` | `implemented` | `apps/web/e2e/workbook.visual.spec.ts::captures blocked preview feedback and Timeline evidence badges` | `browser_visual` | The visual harness captures blocked evidence access feedback plus Timeline evidence-count and has-evidence badge presentation after browser attach. | Preview-handle issuance and redemption contracts remain backend and functional browser evidence. |

## Shared Harness Coverage

| Harness | Phase 5 evidence |
| --- | --- |
| Real runtime and route helpers | Phase 5 initially reuses `internal/modules/workbook/testsupport/scenariotest` for Postgres, object-store services, bootstrap-admin login, incident creation, and HTTP route helpers until Phase 5-specific helper ownership materially diverges. |
| Evidence module helpers | Phase 5 evidence tests own object blob, evidence record, access-handle, and object-store assertions in package-local test files rather than promoting Phase 4 smoke tests. |
| Workbook and browser helpers | `apps/web/src/workbook/WorkbookShell.evidence.test.tsx` and `apps/web/e2e/evidence.spec.ts` hold Phase 5 workbook and browser rows for evidence attachment, counts, previews, and requested-evidence flows. |

## Support-Only Evidence

- `internal/modules/evidence/evidence_integration_test.go` keeps Phase 4 object-blob and evidence-handle route smoke coverage as carryover support for Phase 5 implementation, but authoritative Phase 5 completion claims must live in `phase5_*` rows.
- `apps/web/src/workbook/WorkbookShell.support.test.tsx` remains Phase 4 support-only workbook helper coverage and must not claim Phase 5 IDs.
