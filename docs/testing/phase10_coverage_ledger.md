# Phase 10 Coverage Ledger

This ledger is generated from `tools/phase10_test_map.json`. Update the manifest row metadata first, then regenerate this file.

- Scope: Operational backup, restore, restore verification, public route absence, deployment-local operator boundary, and backup-storage binding.
- Normative owners: Core 01 §12.1–§12.2; Core 04 §2, §6, §9.14, and §12.3.3.
- Authority: `tools/phase10_test_map.json` is the enforced Phase 10 traceability source. This ledger is a rendered companion and does not control the mechanical row inventory.
- Phase 10 is active while final traceability refresh and retained-run handoff remain pending.
- Sprint 1 gives `U-10-05` and `E-10-04` direct backup-storage binding evidence without claiming backup or restore implementation.
- Sprint 2 gives `U-10-01`, `I-10-01`, and `E-10-01` direct artifact-backed backup metadata, retention-floor, latest-success, and deployment-local operator inspection evidence without claiming restore orchestration.
- Sprint 2 gives `U-10-04` and `E-10-03` direct public route-family absence evidence; operator authorization remains covered by `E-10-01`.
- Sprint 3 gives `U-10-02` and `I-10-02` direct process-harness restore readiness, same-point store restore, projection rebuild, workbook query, and row/change/blob consistency evidence.
- Sprint 4 remediation gives `U-10-03` and `I-10-03` direct broken-artifact fail-closed evidence and restore-verification state transition evidence.
- Sprint 5 remediation gives `E-10-02` direct browser-visible workbook recovery evidence through an isolated browser restore helper and ordinary workbook UI/API surfaces.
- Operational backup and restore remain deployment-local recovery behavior and are distinct from the Incident Portability Extension Profile.
- Generated ledgers and schedules are downstream artifacts and must not be hand-edited.

## Authoritative Execution

- `backend-unit` selects direct backup-storage configuration evidence for `U-10-05` and static public route-family absence evidence for `U-10-04`.
- `backend-store` selects artifact-backed retained `backup_set` metadata and retention evidence for `U-10-01`.
- `backend-integration` selects real Postgres, object-store, and backup-storage artifact evidence for `I-10-01`.
- `backend-process` selects deployment-local operator inspection and restore evidence for `E-10-01`, process public-route absence evidence for `E-10-03`, direct restore readiness evidence for `U-10-02`, `U-10-03`, `I-10-02`, and `I-10-03`, and direct effective-config startup rejection evidence for `E-10-04`.
- `browser-e2e-webserver-backed` selects direct browser-visible restored workbook evidence for `E-10-02`.

## Support-Only Execution

- Existing runtime-root, object-store, projection, workbook, route, and browser evidence from earlier phases is support-only substrate and cannot satisfy Phase 10 rows.

## Unit

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |
| `U-10-01` | `implemented` | `internal/modules/recovery/phase10_backup_metadata_store_test.go::TestPhase10_U_10_01_BackupMetadataShapeAndRetentionFloors`, `TestPhase10_U_10_01_VerificationVocabularyAndTimestampRules`, `TestPhase10_U_10_01_LatestSuccessfulRetainedBackupRequiresTwentyFourHourFloor`, `TestPhase10_U_10_01_RetentionFloorRejectsShortMetadataAndArtifacts`, `TestPhase10_U_10_01_CaptureRequiresArtifactProofs` | `backend_store` | Durable backup_set creation requires artifact-backed capture with one declared consistency_point_at, stable backup_set_id, backup-storage restore anchors, Postgres and object-store artifact proofs, integrity-manifest proof, exact verification_state vocabulary, timestamp/nullability enforcement, latest-success 24-hour freshness, and 30-day metadata and artifact retention floors. | This evidence does not implement restore orchestration, projection rebuild readiness, isolated restore verification cadence, or arbitrary point-in-time restore. |
| `U-10-02` | `implemented` | `cmd/server/main_phase10_recovery_sentinel_test.go::TestPhase10_U_10_02_RestoreReadinessAndCoherentStoreOrder` | `backend_process` | Restore readiness selects the latest retained successful backup_set only when its latest consistency point is unambiguous, restores Postgres and object storage from that one backup set, rebuilds required projections, verifies authoritative row, change-set, and blob-hash coherence, and marks readiness only after those steps complete in deterministic order. | This evidence does not implement fail-closed missing/corrupt artifact behavior, restore-verification cadence, browser recovery, public route families, or arbitrary timestamp restore. |
| `U-10-03` | `implemented` | `cmd/server/main_phase10_recovery_sentinel_test.go::TestPhase10_U_10_03_FailClosedRestoreVerificationBlocked` | `backend_process` | Restore fails before readiness when required artifacts or integrity proof are missing or corrupt, and restore verification records verified or failed state only after isolated verification execution. | This evidence does not implement browser-visible restore recovery or arbitrary timestamp restore. |
| `U-10-04` | `implemented` | `internal/modules/recovery/phase10_recovery_sentinel_test.go::TestPhase10_U_10_04_PublicRouteAbsenceStaticInventory` | `backend_unit` | Authored public route inventory exposes no `/api/v1/backups*`, `/api/v1/restores*`, `/api/v1/restore-verifications*`, `/ws/v1/backups*`, `/ws/v1/restores*`, or `/ws/v1/restore-verifications*` families. | Deployment-local operator authorization is covered by `E-10-01`; live process route probing is covered by `E-10-03`. |
| `U-10-05` | `implemented` | `internal/platform/config/config_phase10_test.go::TestPhase10_BackupStorageRootBinding_U_10_05` | `backend_unit` | roots.backup_storage is required for supported deployment profiles, accepts only profile-allowed binding kinds, and rejects export-output or temporary-work roots as backup storage. | Encrypted-storage conformance for incident-bearing backup artifacts remains later Phase 10 evidence. |

## Integration

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |
| `I-10-01` | `implemented` | `internal/modules/recovery/phase10_backup_metadata_integration_test.go::TestPhase10_I_10_01_RealBackingStorageMetadataPersistsAndLatestLookup` | `backend_integration` | Real Postgres-backed runtime metadata persistence is reached only after capture writes and verifies Postgres row-snapshot, object-store snapshot, and integrity-manifest artifacts under `roots.backup_storage`; latest lookup exposes same-point restore anchors, default unverified/null verification state, artifact proofs, and AC-398 retention floors from a fresh store. | This evidence does not perform restore orchestration or prove workbook recovery. |
| `I-10-02` | `implemented` | `cmd/server/main_phase10_recovery_sentinel_test.go::TestPhase10_I_10_02_FreshEnvironmentRestoreWorkbookConsistency` | `backend_process` | Restoring the latest successful retained backup_set into a fresh isolated environment rebuilds projections, starts the server ready only after restore completion, opens the restored incident through the process, executes the built-in timeline workbook query, and preserves authoritative rows, change sets, and blob hashes. | This evidence does not implement browser recovery, fail-closed missing/corrupt artifact behavior, restore-verification cadence, public route families, or arbitrary timestamp restore. |
| `I-10-03` | `implemented` | `cmd/server/main_phase10_recovery_sentinel_test.go::TestPhase10_I_10_03_MissingArtifactFailsBeforeReadinessBlocked` | `backend_process` | Selecting a retained backup_set missing a required artifact or integrity proof must fail before target readiness. | This evidence does not implement browser-visible restore recovery or arbitrary timestamp restore. |

## Browser E2E

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |
| `E-10-01` | `implemented` | `cmd/operator/operator_phase10_test.go::TestPhase10_E_10_01_DeploymentLocalOperatorInspectLatestBackupMetadata`, `TestPhase10_E_10_01_DeploymentLocalOperatorRestoreLatestBackup` | `backend_process` | Deployment-local operator commands expose latest successful retained backup_set metadata and restore the latest retained backup only for an active deployment_admin; inactive deployment admins, non-admins, incident-admin-only users, and unsafe same-config restore targets are rejected. | This evidence does not implement browser-visible restore recovery or arbitrary timestamp restore. |
| `E-10-02` | `implemented` | `apps/web/e2e/phase10.restore.spec.ts::Phase 10 E-10-02 restore recovers workbook surface and executes a built-in workbook query` | `browser_functional` | A harness-owned browser fixture captures a retained backup set from the source deployment, restores it into a separately migrated target deployment through the production restore runner, rebuilds projections, opens a restored incident through the target workbook UI, and executes a built-in workbook query through ordinary product routes. | This browser evidence does not expose public backup or restore route families and does not implement arbitrary timestamp restore. |
| `E-10-03` | `implemented` | `cmd/server/main_phase10_recovery_sentinel_test.go::TestPhase10_E_10_03_PublicRouteInventoryAbsence` | `backend_process` | A running server exposes no public HTTP or WebSocket backup, restore, or restore-verification route families. | Deployment-local operator authorization is covered by `E-10-01`; this evidence does not implement restore orchestration or restore-verification cadence. |
| `E-10-04` | `implemented` | `cmd/server/main_phase10_config_test.go::TestPhase10_EffectiveConfigBackupRoot_E_10_04` | `backend_process` | Effective deployment configuration fails the real server process before readiness when roots.backup_storage is missing, uses a profile-incompatible binding, or is satisfied by export or temporary-work roots. | This process evidence does not implement backup metadata, restore orchestration, restore verification, or operator controls. |

## Shared Harness Coverage

| Harness | Phase 10 evidence |
| --- | --- |
| Active manifest ownership | `tools/phase10_test_map.json` records Sprint 1 through browser-visible restore-remediation direct evidence. |
| Generated ledger | `docs/testing/phase10_coverage_ledger.md` is generated from this manifest and must not be hand-edited. |
| Schedule boundary | Generated schedules must execute implemented Sprint 1 through Sprint 5 evidence without blocker sentinels. |

## Support-Only Evidence

- Earlier phase runtime-root and object-store evidence is support-only substrate.
- Existing projection rebuild tests remain support-only substrate except for `U-10-02` restore readiness evidence.
- Existing browser workbook query tests remain support-only substrate outside the direct Phase 10 recovery workflow.
