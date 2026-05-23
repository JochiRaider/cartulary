# Phase 11 Coverage Ledger

This ledger is generated from `tools/phase11_test_map.json`. Update the manifest row metadata first, then regenerate this file.

- Scope: Selected Import Extension Profile route family and common job substrate only; Snapshot and Reporting, Reference Pack, Incident Portability, and Enterprise Authentication remain reserved and unclaimed.
- Normative owners: Core 00 §4.2 and §5.1; Core 01 §17 and §20; Core 04 extension-profile claim criteria; Core 05 only for claim-bearing timed or fixture-sensitive publication.
- Authority: `tools/phase11_test_map.json` is the enforced Phase 11 traceability source. This ledger is a rendered companion and does not control the mechanical row inventory.
- Phase 11 is active only for the selected Import Extension Profile and the common job substrate required by selected extension routes.
- Import evidence covers upload-envelope early failure, CSV and XLSX discovery, session and unit reads, preview, deterministic mapping approval, select, apply, terminal job summaries, durable state separation, and request-time authorization for common jobs.
- Snapshot and Reporting, Reference Pack, Incident Portability, and Enterprise Authentication remain unselected and unclaimed; reserved-family behavior continues to be their only default runtime exposure.
- Generated ledgers and schedules are downstream artifacts and must not be hand-edited.

## Authoritative Execution

- `backend-integration` selects request-time authorization re-derivation for incident-scoped and deployment-scoped common jobs.
- `backend-integration` selects Import upload-envelope no-durable-state behavior, CSV mapping/select/apply, and XLSX used-range discovery evidence.

## Support-Only Execution

- Upload-envelope helper unit tests and lower-level durable job manager tests remain support-only substrate outside the direct Phase 11 rows.
- Reserved-family extension discovery tests remain support-only evidence for unselected profiles.

## Unit

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |
| `U-11-JOBS-01` | `implemented` | `internal/modules/jobapi/routes_integration_test.go::TestPhase11_U_11_JOBS_01_IncidentJobAuthorizationReDerivedAtRequestTime` | `backend_integration` | Incident-scoped common job read and cancel authorization is re-derived at request time from current incident membership and role, including denied cancel for a non-admin member and hidden job access after membership revocation. | This evidence does not claim Snapshot and Reporting, Reference Pack, Incident Portability, or Enterprise Authentication route behavior. |
| `U-11-JOBS-02` | `implemented` | `internal/modules/jobapi/routes_integration_test.go::TestPhase11_U_11_JOBS_02_DeploymentJobAuthorizationReDerivedAtRequestTime` | `backend_integration` | Deployment-scoped common job reads and cancel requests are authorized from current deployment-admin or submitter state, while unrelated non-admin callers cannot observe or cancel the job. | This evidence does not claim any unselected extension profile route family. |

## Integration

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |
| `I-11-IMPORT-01` | `implemented` | `internal/modules/imports/imports_integration_test.go::TestPhase11_I_11_IMPORT_01_UploadMetadataNonObjectCreatesNoDurableRows` | `backend_integration` | Import upload metadata that is syntactically valid but not a JSON object fails with `invalid_import_request` and `reason_code=request_not_object` before durable sessions, units, jobs, or idempotency rows are created. | This evidence does not cover successful discovery, mapping, or apply. |
| `I-11-IMPORT-02` | `implemented` | `internal/modules/imports/imports_integration_test.go::TestPhase11_I_11_IMPORT_02_MappingSelectApplyCreatesTimelineRows` | `backend_integration` | CSV Import discovery, exhaustive mapping approval, deterministic mapping fingerprint exposure, selection, apply job completion, Timeline row creation, and duplicate-apply blocking operate through the selected Import route family. | This evidence does not claim XLSX discovery or unselected extension profiles. |
| `I-11-IMPORT-03` | `implemented` | `internal/modules/imports/imports_integration_test.go::TestPhase11_I_11_IMPORT_03_XLSXDiscoveryUsesBoundedUsedRange` | `backend_integration` | XLSX Import discovery accepts a bounded OpenXML workbook, records XLSX provenance, creates an `xlsx_used_range` unit, and exposes source headers plus preview rows without applying incident state. | This evidence does not claim table, named-range, or manual-region XLSX locators beyond the selected used-range discovery behavior. |

## Browser E2E

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |

## Shared Harness Coverage

| Harness | Phase 11 evidence |
| --- | --- |
| Active manifest ownership | `tools/phase11_test_map.json` records selected Import and common-job evidence only. |
| Generated ledger | `docs/testing/phase11_coverage_ledger.md` is generated from this manifest and must not be hand-edited. |
| Schedule boundary | Generated schedules must execute implemented Import/common-job evidence without implying claims for unselected extension profiles. |

## Support-Only Evidence

- Helper substrate alone does not claim an extension profile.
- OpenAPI and generated contract artifacts are derived evidence and do not replace direct route tests.
- Unselected extension family roots must continue to return `extension_profile_not_claimed`.
