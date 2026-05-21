# Phase 10 Coverage Ledger

This ledger is generated from `tools/phase10_test_map.json`. Update the manifest row metadata first, then regenerate this file.

- Scope: Operational backup, restore, restore verification, public route absence, deployment-local operator boundary, and backup-storage binding.
- Normative owners: Core 01 §12.1–§12.2; Core 04 §2, §6, §9.14, and §12.3.3.
- Authority: `tools/phase10_test_map.json` is the enforced Phase 10 traceability source. This ledger is a rendered companion and does not control the mechanical row inventory.
- Phase 10 is planned, not active, while all authoritative rows are blocker sentinels.
- Every authoritative Phase 10 guide row has exactly one blocked owner so traceability exists without claiming implementation.
- Operational backup and restore remain deployment-local recovery behavior and are distinct from the Incident Portability Extension Profile.
- Generated ledgers and schedules are downstream artifacts and must not be hand-edited.

## Authoritative Execution

- `backend-unit` will select direct configuration and route-inventory evidence once `U-10-04` and `U-10-05` are implemented.
- `backend-store` will select retained `backup_set` metadata and retention evidence once `U-10-01` is implemented.
- `backend-integration` will select real backing-storage metadata evidence once `I-10-01` is implemented.
- `backend-process` will select restore, fail-closed, operator, route-inventory, and effective-config evidence once process-bound rows are implemented.
- `browser-e2e-webserver-backed` will select workbook recovery evidence once `E-10-02` is implemented.

## Support-Only Execution

- Existing runtime-root, object-store, projection, workbook, route, and browser evidence from earlier phases is support-only substrate and cannot satisfy Phase 10 rows.

## Unit

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |
| `U-10-01` | `blocked` | `internal/modules/recovery/phase10_recovery_sentinel_test.go::TestPhase10_U_10_01_BackupSetMetadataRetentionBlocked` | `backend_store` | Retained backup_set metadata, exact verification_state vocabulary, one coherent consistency_point_at, and latest-success plus 30-day retention floors require future direct persistent-state evidence. | This sentinel does not implement backup metadata, retention, restore anchors, or verification-state transitions. |
| `U-10-02` | `blocked` | `cmd/server/main_phase10_recovery_sentinel_test.go::TestPhase10_U_10_02_RestoreReadinessBlocked` | `backend_process` | Restore readiness requires exactly one retained backup_set, same-point Postgres and object-store restore, and projection rebuild before readiness; future process-bound evidence must replace this sentinel. | This sentinel does not implement restore orchestration or readiness gating. |
| `U-10-03` | `blocked` | `cmd/server/main_phase10_recovery_sentinel_test.go::TestPhase10_U_10_03_FailClosedRestoreVerificationBlocked` | `backend_process` | Restore must fail before readiness when required artifacts or integrity proof are missing, and restore verification must update state only according to isolated verification results. | This sentinel does not implement artifact proof checks or restore-verification state transitions. |
| `U-10-04` | `blocked` | `internal/modules/recovery/phase10_recovery_sentinel_test.go::TestPhase10_U_10_04_PublicRouteAbsenceDeploymentAdminBlocked` | `backend_unit` | Public route inventories must expose no backup, restore, or restore-verification families, and any built-in operator control must be deployment-local and deployment_admin-gated. | This sentinel does not provide direct route inventory or deployment-admin control evidence. |
| `U-10-05` | `blocked` | `internal/modules/recovery/phase10_recovery_sentinel_test.go::TestPhase10_U_10_05_BackupStorageRootBindingBlocked` | `backend_unit` | roots.backup_storage must be a required persistent runtime root with allowed binding kinds and must remain distinct from export and temporary-work roots. | This sentinel does not replace direct Phase 10 configuration-contract evidence. |

## Integration

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |
| `I-10-01` | `blocked` | `internal/modules/recovery/phase10_recovery_sentinel_test.go::TestPhase10_I_10_01_RealBackingStorageMetadataBlocked` | `backend_integration` | The most recent successful retained backup_set must expose required metadata, restore anchors, retention timestamps, and verification transitions against real backing storage. | This sentinel does not implement or verify real backup storage metadata. |
| `I-10-02` | `blocked` | `cmd/server/main_phase10_recovery_sentinel_test.go::TestPhase10_I_10_02_FreshEnvironmentRestoreWorkbookConsistencyBlocked` | `backend_process` | Restoring the latest successful retained backup_set into a fresh environment must rebuild projections, open an incident when data exists, execute a built-in workbook query, and preserve row, change-set, and blob consistency. | This sentinel does not perform fresh-environment restore or workbook recovery. |
| `I-10-03` | `blocked` | `cmd/server/main_phase10_recovery_sentinel_test.go::TestPhase10_I_10_03_MissingArtifactFailsBeforeReadinessBlocked` | `backend_process` | Selecting a retained backup_set missing a required artifact or integrity proof must fail before target readiness. | This sentinel does not create broken backup fixtures or readiness-failure evidence. |

## Browser E2E

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |
| `E-10-01` | `blocked` | `cmd/server/main_phase10_recovery_sentinel_test.go::TestPhase10_E_10_01_DeploymentLocalOperatorInspectBlocked` | `backend_process` | A deployment-local operator must inspect the latest successful retained backup_set and see exact verification, retention, and verification timestamps. | This sentinel does not implement operator inspection. |
| `E-10-02` | `blocked` | `apps/web/e2e/phase10.restore.spec.ts::Phase 10 E-10-02 restore recovers workbook surface and executes a built-in workbook query` | `browser_functional` | Restoring the latest successful retained backup_set into a fresh deployment must recover the workbook surface and execute at least one built-in workbook query successfully. | This sentinel does not run a browser-visible restore workflow. |
| `E-10-03` | `blocked` | `cmd/server/main_phase10_recovery_sentinel_test.go::TestPhase10_E_10_03_PublicRouteInventoryAbsenceBlocked` | `backend_process` | Public HTTP and WebSocket route inventories must expose no backup, restore, or restore-verification families, and any built-in operator control must remain deployment-local and deployment_admin-gated. | This sentinel does not provide live route-inventory evidence. |
| `E-10-04` | `blocked` | `cmd/server/main_phase10_recovery_sentinel_test.go::TestPhase10_E_10_04_EffectiveConfigBackupRootBlocked` | `backend_process` | Effective deployment configuration must require roots.backup_storage with allowed binding kinds and reject attempts to satisfy backup storage with export or temporary-work roots. | This sentinel does not provide effective startup rejection evidence. |

## Shared Harness Coverage

| Harness | Phase 10 evidence |
| --- | --- |
| Planned manifest ownership | `tools/phase10_test_map.json` records every authoritative guide row as blocked until direct evidence exists. |
| Generated ledger | `docs/testing/phase10_coverage_ledger.md` is generated from this manifest and must not be hand-edited. |
| Schedule boundary | Phase 10 remains planned, so generated schedules must not treat Phase 10 rows as active executable evidence. |

## Support-Only Evidence

- Earlier phase runtime-root and object-store evidence is support-only substrate.
- Existing projection rebuild tests are support-only substrate until Phase 10 restore readiness owns projection rebuild evidence.
- Existing browser workbook query tests are support-only substrate until Phase 10 restore recovery evidence owns the row.
