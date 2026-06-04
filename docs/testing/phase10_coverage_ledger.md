# Phase 10 Coverage Ledger

This ledger is generated from `tools/phase10_test_map.json`. Update the manifest row metadata first, then regenerate this file.

- Scope: Operational backup, restore, restore verification, public route absence, deployment-local operator boundary, and backup-storage binding.
- Normative owners: Core 01 §12.1–§12.2; Core 04 §2, §6, §9.14, and §12.3.3.
- Authority: `tools/phase10_test_map.json` is the enforced Phase 10 traceability source. This ledger is a rendered companion and does not control the mechanical row inventory.
- Phase 10 is active with Sprint 9 remediation evidence for durable retained-backup classification, encrypted backup storage, and deployment-local due restore verification.
- Sprint 1 gives backup-storage binding groundwork for `U-10-05` and direct effective-config evidence for `E-10-04`; Sprint 9 carries the encrypted-storage completion for `U-10-05`.
- Sprint 2 gives `U-10-01`, `I-10-01`, and `E-10-01` direct artifact-backed backup metadata, retention-floor, latest-success, and deployment-local operator inspection evidence without claiming restore orchestration.
- Sprint 2 gives `U-10-04` and `E-10-03` direct public route-family absence evidence; operator authorization remains covered by `E-10-01`.
- Sprint 3 gives `U-10-02` and `I-10-02` direct process-harness restore readiness, same-point store restore, projection rebuild, workbook query, and row/change/blob consistency evidence.
- Sprint 4 remediation gives `U-10-03` and `I-10-03` direct broken-artifact fail-closed evidence and restore-verification state transition evidence.
- Sprint 5 remediation gives `E-10-02` direct browser-visible workbook recovery evidence through an isolated browser restore helper and ordinary workbook UI/API surfaces.
- Sprint 9 remediation gives `U-10-01`, `U-10-05`, `I-10-01`, and `E-10-01` direct evidence for durable artifact-backed classification, encrypted artifact envelopes, and the `operator restore-verify due` cadence runner.
- Operational backup and restore remain deployment-local recovery behavior and are distinct from the Incident Portability Extension Profile.
- Generated ledgers and schedules are downstream artifacts and must not be hand-edited.

## Authoritative Execution

- `backend-unit` selects static public route-family absence evidence for `U-10-04` and support-only backup-root configuration regression coverage.
- `backend-store` selects durable artifact-backed retained `backup_set` metadata, retention, and encrypted backup-storage boundary evidence for `U-10-01` and `U-10-05`.
- `backend-integration` selects real Postgres, object-store, encrypted backup-storage artifact, raw-ciphertext, and durable latest-lookup evidence for `I-10-01`.
- `backend-process` selects deployment-local operator inspection, restore, and due restore-verification evidence for `E-10-01`, process public-route absence evidence for `E-10-03`, direct restore readiness evidence for `U-10-02`, `U-10-03`, `I-10-02`, and `I-10-03`, and direct effective-config startup rejection evidence for `E-10-04`.
- `browser-e2e-webserver-backed` selects direct browser-visible restored workbook evidence for `E-10-02`.

## Support-Only Execution

- `internal/platform/config/config_phase10_test.go` runs through `backend-unit` with `TestSupportPhase10_` and is forbidden from claiming `U-10-*` identifiers.
- Existing runtime-root, object-store, projection, workbook, route, and browser evidence from earlier phases is support-only substrate and cannot satisfy Phase 10 rows.

## Unit

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |
| `U-10-01` | `implemented` | `internal/modules/recovery/phase10_backup_metadata_store_test.go::TestPhase10_U_10_01_BackupMetadataShapeAndRetentionFloors`, `TestPhase10_U_10_01_VerificationVocabularyAndTimestampRules`, `TestPhase10_U_10_01_LatestSuccessfulRetainedBackupRequiresTwentyFourHourFloor`, `TestPhase10_U_10_01_DurableCatalogSkipsMetadataWithMissingArtifacts`, `TestPhase10_U_10_01_RetentionFloorRejectsShortMetadataAndArtifacts`, `TestPhase10_U_10_01_CaptureRequiresArtifactProofs` | `backend_store` | Durable backup_set creation requires artifact-backed capture with one declared consistency_point_at, stable backup_set_id, backup-storage restore anchors, Postgres and object-store artifact proofs, integrity-manifest proof, exact verification_state vocabulary, timestamp/nullability enforcement, latest-success 24-hour freshness, required artifact durability plus redacted durability diagnostics before successful-retained classification, and 30-day metadata and artifact retention floors. | This evidence does not implement restore orchestration, projection rebuild readiness, restore-verification cadence execution, or arbitrary point-in-time restore. |
| `U-10-02` | `implemented` | `cmd/server/main_phase10_recovery_sentinel_test.go::TestPhase10_U_10_02_RestoreReadinessAndCoherentStoreOrder` | `backend_process` | Restore readiness selects the latest retained successful backup_set only when its latest consistency point is unambiguous, restores Postgres and object storage from that one backup set, rebuilds required projections, verifies authoritative row, change-set, and blob-hash coherence, and marks readiness only after those steps complete in deterministic order. | This evidence does not implement fail-closed missing/corrupt artifact behavior, restore-verification cadence, browser recovery, public route families, or arbitrary timestamp restore. |
| `U-10-03` | `implemented` | `cmd/server/main_phase10_recovery_sentinel_test.go::TestPhase10_U_10_03_FailClosedRestoreVerificationBlocked` | `backend_process` | Selected-backup restore fails before readiness when required artifacts or integrity proof are missing or corrupt, and restore verification records verified or failed state only after isolated verification execution. | This evidence does not implement browser-visible restore recovery, deployment scheduling of the operator due runner, or arbitrary timestamp restore. |
| `U-10-04` | `implemented` | `internal/modules/recovery/phase10_recovery_sentinel_test.go::TestPhase10_U_10_04_PublicRouteAbsenceStaticInventory` | `backend_unit` | Authored public route inventory exposes no `/api/v1/backups*`, `/api/v1/restores*`, `/api/v1/restore-verifications*`, `/ws/v1/backups*`, `/ws/v1/restores*`, or `/ws/v1/restore-verifications*` families. | Deployment-local operator authorization is covered by `E-10-01`; live process route probing is covered by `E-10-03`. |
| `U-10-05` | `implemented` | `internal/modules/recovery/phase10_backup_metadata_store_test.go::TestPhase10_U_10_05_CaptureRequiresEncryptedBackupStorage`, `TestPhase10_U_10_05_EncryptedBackupStorageFailsClosedWithWrongKey` | `backend_store` | Incident-bearing backup capture requires an encrypted backup-storage adapter, encrypted artifact envelopes decrypt and verify with the configured recovery key, and wrong-key reads fail closed before proof validation can pass. | Effective backup-root configuration startup rejection is covered by `E-10-04`; managed-service backup storage remains unsupported until a later adapter proof contract exists. |

## Integration

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |
| `I-10-01` | `implemented` | `internal/modules/recovery/phase10_backup_metadata_integration_test.go::TestPhase10_I_10_01_RealBackingStorageMetadataPersistsAndLatestLookup` | `backend_integration` | Real Postgres-backed runtime metadata persistence is reached only after capture writes encrypted Postgres row-snapshot, object-store snapshot, and integrity-manifest artifacts under `roots.backup_storage`; latest durable lookup exposes same-point restore anchors, default unverified/null verification state, artifact proofs, raw-ciphertext storage, and AC-398 retention floors from a fresh store. | This evidence does not perform restore orchestration or prove workbook recovery. |
| `I-10-02` | `implemented` | `cmd/server/main_phase10_recovery_sentinel_test.go::TestPhase10_I_10_02_FreshEnvironmentRestoreWorkbookConsistency` | `backend_process` | Restoring the latest successful retained backup_set into a fresh isolated environment rebuilds projections, starts the server ready only after restore completion, opens the restored incident through the process, executes the built-in timeline workbook query, and preserves authoritative rows, change sets, and blob hashes. | This evidence does not implement browser recovery, fail-closed missing/corrupt artifact behavior, restore-verification cadence, public route families, or arbitrary timestamp restore. |
| `I-10-03` | `implemented` | `cmd/server/main_phase10_recovery_sentinel_test.go::TestPhase10_I_10_03_MissingArtifactFailsBeforeReadinessBlocked` | `backend_process` | Selecting a retained backup_set missing a required artifact or integrity proof must fail before target readiness. | This evidence does not implement browser-visible restore recovery or arbitrary timestamp restore. |

## Browser E2E

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |
| `E-10-01` | `implemented` | `cmd/operator/operator_phase10_test.go::TestPhase10_E_10_01_DeploymentLocalOperatorInspectLatestBackupMetadata`, `TestPhase10_E_10_01_DeploymentLocalOperatorRestoreLatestBackup`, `TestPhase10_E_10_01_DeploymentLocalOperatorRestoreVerifyDueRunner`, `TestPhase10_E_10_01_ObjectStoreMigrationRunEmitsPassAndMismatchEvidence` | `backend_process` | Deployment-local operator commands expose latest durable successful retained backup_set metadata with redacted durability diagnostics, restore the latest durable retained backup, run due restore verification, and run object-store migration tooling only for an active deployment_admin; inactive deployment admins, non-admins, incident-admin-only users, unsafe same-config restore targets, unmarked verification targets, and target-side migration mismatches are rejected. | This evidence does not implement browser-visible restore recovery, public backup, restore, or migration APIs, arbitrary timestamp restore, or SWFS-OWNER-STORAGEREF-001 resolution. |
| `E-10-02` | `implemented` | `apps/web/e2e/phase10.restore.spec.ts::Phase 10 E-10-02 restore recovers workbook surface and executes a built-in workbook query` | `browser_functional` | A harness-owned browser fixture captures a retained backup set from the source deployment, restores it into a separately migrated target deployment through the production restore runner, rebuilds projections, opens a restored incident through the target workbook UI, and executes a built-in workbook query through ordinary product routes. | This browser evidence does not expose public backup or restore route families and does not implement arbitrary timestamp restore. |
| `E-10-03` | `implemented` | `cmd/server/main_phase10_recovery_sentinel_test.go::TestPhase10_E_10_03_PublicRouteInventoryAbsence` | `backend_process` | A running server exposes no public HTTP or WebSocket backup, restore, or restore-verification route families. | Deployment-local operator authorization is covered by `E-10-01`; this evidence does not implement restore orchestration or restore-verification cadence. |
| `E-10-04` | `implemented` | `cmd/server/main_phase10_config_test.go::TestPhase10_EffectiveConfigBackupRoot_E_10_04` | `backend_process` | Effective deployment configuration fails the real server process before readiness when roots.backup_storage is missing, uses a profile-incompatible binding, or is satisfied by export or temporary-work roots. | Encrypted incident-bearing artifact conformance is covered by `U-10-05` and `I-10-01`; this process evidence does not implement backup metadata, restore orchestration, restore verification, or operator controls. |

## Shared Harness Coverage

| Harness | Phase 10 evidence |
| --- | --- |
| Active manifest ownership | `tools/phase10_test_map.json` records Sprint 1 through Sprint 9 recovery-remediation direct evidence. |
| Generated ledger | `docs/testing/phase10_coverage_ledger.md` is generated from this manifest and must not be hand-edited. |
| Schedule boundary | Generated schedules must execute implemented Sprint 1 through Sprint 9 evidence without blocker sentinels. |

## Support-Only Evidence

- Earlier phase runtime-root and object-store evidence is support-only substrate.
- Existing projection rebuild tests remain support-only substrate except for `U-10-02` restore readiness evidence.
- Existing browser workbook query tests remain support-only substrate outside the direct Phase 10 recovery workflow.
