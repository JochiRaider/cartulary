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

## Support-Only Execution

- Existing Phase 4 object-blob and handle route smoke tests remain Phase 4-owned carryover evidence and must not claim Phase 5 IDs.

## Unit

| Row | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- |
| `U-5-01` | `internal/modules/evidence/phase5_blob_create_test.go::TestPhase5_ObjectBlobCreate_U_5_01` | `backend_unit` | `POST /api/v1/object-blobs` requires `incident_id`, `client_txn_id`, and declared `byte_size`, rejects unknown top-level members, and echoes the normalized `accepted_contract` on success. | Idempotent slot replay, object-store upload, evidence attachment, and workbook projection consequences remain separate Phase 5 rows. |
| `U-5-02` | `internal/modules/evidence/phase5_blob_create_test.go::TestPhase5_BlobCreateIdempotency_U_5_02` | `backend_store` | Blob-create idempotency is keyed by `(actor_user_id, incident_id, client_txn_id)`, replays the same normalized request to the original slot, and rejects divergent replay with the route-owned conflict. | Expired-slot replay and fresh-target behavior are integration evidence in `I-5-02`. |
| `U-5-03` | `internal/modules/evidence/phase5_attach_test.go::TestPhase5_AttachBlobValidation_U_5_03` | `backend_store` | `POST /api/v1/evidence-records/{record_id}/attach-blob` requires `object_blob_id`, `base_row_version`, and `client_txn_id` and fails closed for pending, failed, missing, expired, incident-mismatched, or contract-mismatched blobs. | End-to-end object upload and workbook query readback are covered by integration and browser rows. |
| `U-5-04` | `internal/modules/evidence/phase5_lifecycle_test.go::TestPhase5_EvidenceLifecycleSeparateFromBlob_U_5_04` | `backend_store` | Evidence lifecycle state remains separate from blob `upload_state`, allowing requested or pending-receipt evidence without a blob and later advancement without unrelated custody-history mutation. | Generalized custody workflow UX and scanner infrastructure are not completed by this row. |
| `U-5-05` | `internal/modules/evidence/phase5_handles_test.go::TestPhase5_HandleIssueEmptyBodyNonIdempotent_U_5_05` | `backend_unit` | Preview-handle and download-handle issuance routes accept only `{}` and are intentionally non-idempotent, with each success yielding a fresh opaque same-origin handle. | Current-state redemption and object streaming behavior are covered by `U-5-06`, `U-5-07`, and `I-5-03`. |
| `U-5-06` | `internal/modules/evidence/phase5_handles_test.go::TestPhase5_HandleRedemptionRechecksCurrentState_U_5_06` | `backend_store` | Handle redemption re-derives current session validity, incident membership, and evidence/blob state, and preview blocks explicitly for unsupported or unsafe preview conditions. | Browser rendering of preview outcomes is covered by `E-5-03`. |
| `U-5-07` | `internal/modules/evidence/phase5_handles_test.go::TestPhase5_DownloadDispositionFallback_U_5_07` | `backend_unit` | Download responses use authoritative metadata for filename disposition and apply deterministic fallback naming when authoritative names are absent. | Authorization re-check and state invalidation during redemption are covered by `U-5-06` and `I-5-03`. |
| `U-5-08` | `apps/web/src/WorkbookShell.phase5.test.tsx::Phase 5 U-5-08 reflects attached evidence counts on the workbook surface without forcing navigation` | `frontend_unit` | Evidence attachment updates workbook-visible evidence counts and derived flags without forcing navigation away from the current workbook surface. | Real object upload and backend projection rebuild behavior are integration evidence. |
| `U-5-09` | `internal/modules/evidence/phase5_blob_create_test.go::TestPhase5_BlobCreateSizeCeiling_U_5_09` | `backend_unit` | Blob-create size ceilings reject oversize `byte_size` before slot creation, and preview-handle issuance fails with the route-owned oversized-preview reason while legal downloads remain available. | Expired-slot replay and full object upload completion remain separate integration evidence. |

## Integration

| Row | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- |
| `I-5-01` | `internal/modules/evidence/phase5_integration_test.go::TestPhase5_ObjectUploadAttachWorkbookProjection_I_5_01` | `backend_integration` | A real object-store upload can be created through a blob slot, finalized onto an evidence record, and queried through the workbook projection without duplicate attachment rows. | Browser-level attach interaction remains `E-5-01` and `E-5-02`. |
| `I-5-02` | `internal/modules/evidence/phase5_integration_test.go::TestPhase5_ExpiredSlotReplay_I_5_02` | `backend_integration` | Expired-slot replay returns the same expired slot, while a fresh upload target requires a new `client_txn_id`. | General idempotency for non-expired slots is unit/store evidence in `U-5-02`. |
| `I-5-03` | `internal/modules/evidence/phase5_integration_test.go::TestPhase5_HandleRedeemInvalidatesOnCurrentStateLoss_I_5_03` | `backend_integration` | A redeemed preview or download handle fails closed after logout, membership removal, or evidence/blob state invalidation. | Frontend display of blocked preview outcomes is covered by `E-5-03`. |
| `I-5-04` | `internal/modules/evidence/phase5_integration_test.go::TestPhase5_QuarantineBoundaryPreservesTwoStepAttach_I_5_04` | `backend_integration` | When quarantine or scanning state is present, two-step attach semantics remain intact and preview never bypasses the quarantine boundary. | Implementing a generalized scanner adjunct is outside the base Phase 5 row. |

## Browser E2E

| Row | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- |
| `E-5-01` | `apps/web/e2e/phase5.evidence.spec.ts::E-5-01 attaches a screenshot to a selected Timeline row without leaving the workbook surface` | `browser_functional` | An analyst attaches a screenshot to a selected Timeline row without leaving the workbook surface, and the row reflects the new evidence count after commit. | Backend object-store finalization is covered by `I-5-01`. |
| `E-5-02` | `apps/web/e2e/phase5.evidence.spec.ts::E-5-02 persists a screenshot-only Timeline row through the two-step evidence path` | `browser_functional` | A screenshot-only Timeline row can be persisted through the two-step evidence path. | Generic Timeline row create behavior remains Phase 3 evidence. |
| `E-5-03` | `apps/web/e2e/phase5.evidence.spec.ts::E-5-03 redeems inline-safe previews and shows explicit blocked-preview outcomes` | `browser_functional` | An inline-safe type receives a preview handle and renders through the same-origin redeem path, while unsupported or unsafe types return an explicit blocked-preview outcome. | Low-level handle issuance and redemption checks remain backend evidence. |
| `E-5-04` | `apps/web/e2e/phase5.evidence.spec.ts::E-5-04 tracks requested evidence before a blob exists and later advances it` | `browser_functional` | Requested evidence can be tracked before the blob exists and later advanced to an available state without breaking workbook pivots or counts. | A generalized custody or collection workflow is outside this row. |

## Shared Harness Coverage

| Harness | Phase 5 evidence |
| --- | --- |
| Real runtime and route helpers | Phase 5 initially reuses `internal/testutil/phase4test` for Postgres, MinIO, bootstrap-admin login, incident creation, and HTTP route helpers until Phase 5-specific helper ownership materially diverges. |
| Evidence module helpers | Phase 5 evidence tests own object blob, evidence record, access-handle, and object-store assertions in package-local test files rather than promoting Phase 4 smoke tests. |
| Workbook and browser helpers | `apps/web/src/WorkbookShell.phase5.test.tsx` and `apps/web/e2e/phase5.evidence.spec.ts` hold Phase 5 workbook and browser rows for evidence attachment, counts, previews, and requested-evidence flows. |

## Support-Only Evidence

- `internal/modules/evidence/evidence_integration_test.go` keeps Phase 4 object-blob and evidence-handle route smoke coverage as carryover support for Phase 5 implementation, but authoritative Phase 5 completion claims must live in `phase5_*` rows.
- `apps/web/src/WorkbookShell.phase4.support.test.tsx` remains Phase 4 support-only workbook helper coverage and must not claim Phase 5 IDs.
