# Phase 11 Coverage Ledger

This ledger is generated from `tools/phase11_test_map.json`. Update the manifest row metadata first, then regenerate this file.

- Scope: Selected Import Extension Profile route family, selected Snapshot and Reporting Extension Profile route family, and common job substrate; Reference Pack, Incident Portability, and Enterprise Authentication remain reserved and unclaimed.
- Normative owners: Core 00 §4.2 and §5.1; Core 01 §17 and §20; Core 04 extension-profile claim criteria; Core 05 only for claim-bearing timed or fixture-sensitive publication.
- Authority: `tools/phase11_test_map.json` is the enforced Phase 11 traceability source. This ledger is a rendered companion and does not control the mechanical row inventory.
- Phase 11 is active for the selected Import Extension Profile, the selected Snapshot and Reporting Extension Profile, and the common job substrate required by selected extension routes.
- Import evidence covers upload-envelope early failure, CSV and XLSX discovery, session and unit reads, preview, deterministic mapping approval, select, apply, terminal job summaries, durable state separation, and request-time authorization for common jobs.
- Snapshot and Reporting evidence covers immutable snapshot replay, redaction profile validation, deterministic redaction precedence, manifest provenance, external-release binary and working-material boundaries, distinct approvals, publication, and state conflicts.
- Reference Pack, Incident Portability, and Enterprise Authentication remain unselected and unclaimed; reserved-family behavior continues to be their only default runtime exposure.
- Generated ledgers and schedules are downstream artifacts and must not be hand-edited.

## Authoritative Execution

- `backend-integration` selects request-time authorization re-derivation for incident-scoped and deployment-scoped common jobs.
- `backend-integration` selects Import upload-envelope no-durable-state behavior, CSV mapping/select/apply, XLSX used-range discovery evidence, and Snapshot/Reporting snapshot-release lifecycle evidence.
- `backend-unit` selects Snapshot/Reporting redaction profile and post-redaction validation fixtures.

## Support-Only Execution

- Upload-envelope helper unit tests and lower-level durable job manager tests remain support-only substrate outside the direct Phase 11 rows.
- Reserved-family extension discovery tests remain support-only evidence for unselected profiles.

## Unit

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |
| `U-11-JOBS-01` | `implemented` | `internal/modules/jobapi/routes_integration_test.go::TestPhase11_U_11_JOBS_01_IncidentJobAuthorizationReDerivedAtRequestTime` | `backend_integration` | Incident-scoped common job read and cancel authorization is re-derived at request time from current incident membership and role, including denied cancel for a non-admin member and hidden job access after membership revocation. | This evidence does not claim Snapshot and Reporting, Reference Pack, Incident Portability, or Enterprise Authentication route behavior. |
| `U-11-JOBS-02` | `implemented` | `internal/modules/jobapi/routes_integration_test.go::TestPhase11_U_11_JOBS_02_DeploymentJobAuthorizationReDerivedAtRequestTime` | `backend_integration` | Deployment-scoped common job reads and cancel requests are authorized from current deployment-admin or submitter state, while unrelated non-admin callers cannot observe or cancel the job. | This evidence does not claim any unselected extension profile route family. |
| `U-11-REPORTING-01` | `implemented` | `internal/modules/reporting/redaction_test.go::TestPhase11_U_11_REPORTING_01_RedactionProfilePrecedenceActionsAndManifest` | `backend_unit` | Snapshot/Reporting redaction applies deterministic path-over-class-over-default precedence, executes current-profile action semantics, and emits manifest entries bound to rule ids, outcomes, and immutable profile digest. | This unit fixture does not exercise route authorization or release approval lifecycle. |
| `U-11-REPORTING-02` | `implemented` | `internal/modules/reporting/redaction_test.go::TestPhase11_U_11_REPORTING_02_RedactionProfileRejectsConflictsHashAndUnsafeBounds` | `backend_unit` | Snapshot/Reporting rejects duplicate same-precedence redaction rules, the reserved hash action, and unsafe truncate bounds before render admission. | This unit fixture does not claim future keyed pseudonymization behavior for hash. |
| `U-11-REPORTING-03` | `implemented` | `internal/modules/reporting/redaction_test.go::TestPhase11_U_11_REPORTING_03_ExternalValidationRejectsOpaqueBytesAndWorkingMaterial` | `backend_unit` | Snapshot/Reporting post-redaction validation fails closed when an external release would include opaque source bytes or working material. | This unit fixture uses an in-memory export model and does not inspect packaged browser output. |
| `U-11-REPORTING-04` | `implemented` | `internal/modules/reporting/redaction_test.go::TestPhase11_U_11_REPORTING_04_DisclosurePartitionsAndCuratedSupportRefsFailClosed` | `backend_unit` | Snapshot/Reporting recipient redaction profiles enforce disclosure partition allowlists, keep dropped partition fields in the redaction manifest, and fail closed when external curated narrative lacks support references. | This unit fixture uses in-memory export models and does not inspect rendered route output. |
| `U-11-REPORTING-05` | `implemented` | `internal/modules/reporting/redaction_test.go::TestPhase11_U_11_REPORTING_05_DecoderNormalizationAndRegisteredReasons` | `backend_unit` | Snapshot/Reporting route decoders reject legacy alias fields with registered reason codes and normalize omitted, null, and empty release-action reasons to the same idempotency payload. | This unit fixture does not exercise session authorization or database idempotency persistence. |
| `U-11-REPORTING-06` | `implemented` | `internal/modules/reporting/openapi_contract_test.go::TestPhase11_U_11_REPORTING_06_OpenAPIReleaseEnumsAndExactResources` | `backend_unit` | Snapshot/Reporting OpenAPI uses reusable closed enum schemas for release output kind and release scope in both release create and durable release resources, and documents exact snapshot/release resource envelopes for route responses. | This contract-shape evidence is derived from generated OpenAPI artifacts and does not replace direct route behavior tests. |

## Integration

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |
| `I-11-IMPORT-01` | `implemented` | `internal/modules/imports/imports_integration_test.go::TestPhase11_I_11_IMPORT_01_UploadMetadataNonObjectCreatesNoDurableRows` | `backend_integration` | Import upload metadata that is syntactically valid but not a JSON object fails with `invalid_import_request` and `reason_code=request_not_object` before durable sessions, units, jobs, or idempotency rows are created. | This evidence does not cover successful discovery, mapping, or apply. |
| `I-11-IMPORT-02` | `implemented` | `internal/modules/imports/imports_integration_test.go::TestPhase11_I_11_IMPORT_02_MappingSelectApplyCreatesTimelineRows` | `backend_integration` | CSV Import discovery, exhaustive mapping approval, deterministic mapping fingerprint exposure, selection, apply job completion, Timeline row creation, and duplicate-apply blocking operate through the selected Import route family. | This evidence does not claim XLSX discovery or unselected extension profiles. |
| `I-11-IMPORT-03` | `implemented` | `internal/modules/imports/imports_integration_test.go::TestPhase11_I_11_IMPORT_03_XLSXDiscoveryUsesBoundedUsedRange` | `backend_integration` | XLSX Import discovery accepts a bounded OpenXML workbook, records XLSX provenance, creates an `xlsx_used_range` unit, and exposes source headers plus preview rows without applying incident state. | This evidence does not claim table, named-range, or manual-region XLSX locators beyond the selected used-range discovery behavior. |
| `I-11-REPORTING-01` | `implemented` | `internal/modules/reporting/reporting_integration_test.go::TestPhase11_I_11_REPORTING_01_SnapshotReplayAndReleaseProvenanceAreStable` | `backend_integration` | Snapshot creation resolves and persists an immutable source boundary, exact replay returns the original snapshot instead of later live state, and external release creation binds redaction profile digest, manifest digest, output hash, and dropped/truncated manifest outcomes without packaging working material. | This evidence does not claim browser UI exposure or recipient-specific multi-profile output variants. |
| `I-11-REPORTING-02` | `implemented` | `internal/modules/reporting/reporting_integration_test.go::TestPhase11_I_11_REPORTING_02_ExternalReleaseApprovalPublishAndStateConflicts` | `backend_integration` | External releases require distinct reviewer and admin approvals, transition to approved only after the complete approval set, publish synchronously from approved state, and reject duplicate publication with a stable release_state_conflict reason. | This evidence does not claim deployment-wide approval policies beyond the Snapshot and Reporting release gate. |
| `I-11-REPORTING-03` | `implemented` | `internal/modules/reporting/reporting_integration_test.go::TestPhase11_I_11_REPORTING_03_BoundaryReplayDefaultsAndActionIdempotency` | `backend_integration` | Snapshot/Reporting route idempotency treats an omitted snapshot source boundary and the originally resolved explicit boundary as the same request, defaults omitted release_scope to approved internal_draft, and replays release actions with omitted/null/empty reason before fresh state checks. | This evidence does not claim browser report inspection. |
| `I-11-REPORTING-04` | `implemented` | `internal/modules/reporting/reporting_integration_test.go::TestPhase11_I_11_REPORTING_04_ExactShapesAndRouteScopedVisibility` | `backend_integration` | Snapshot/Reporting singleton reads and release actions enforce exact durable resource shapes, route-scoped hidden-resource errors, singleton pagination rejection, admin-only publish/invalidate authorization, and action response parity with release reads. | This evidence does not claim browser report inspection or network-level renderer instrumentation. |

## Browser E2E

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |

## Shared Harness Coverage

| Harness | Phase 11 evidence |
| --- | --- |
| Active manifest ownership | `tools/phase11_test_map.json` records selected Import, Snapshot/Reporting, and common-job evidence. |
| Generated ledger | `docs/testing/phase11_coverage_ledger.md` is generated from this manifest and must not be hand-edited. |
| Schedule boundary | Generated schedules must execute implemented Import/common-job evidence without implying claims for unselected extension profiles. |

## Support-Only Evidence

- Helper substrate alone does not claim an extension profile.
- OpenAPI and generated contract artifacts are derived evidence and do not replace direct route tests.
- Unselected extension family roots must continue to return `extension_profile_not_claimed`.
