# Phase 10 Coverage Ledger

This ledger is generated from `tools/phase10_test_map.json`. Update the manifest row metadata first, then regenerate this file.

- Scope: Operational backup, restore, restore verification, public route absence, deployment-local operator boundary, and backup-storage binding.
- Normative owners: Core 01 §12.1–§12.2; Core 04 §2, §6, §9.14, and §12.3.3.
- Authority: `tools/phase10_test_map.json` is the enforced Phase 10 traceability source. This ledger is a rendered companion and does not control the mechanical row inventory.
- Phase 10 is active with Sprint 10 canonical operator recovery evidence for durable retained-backup classification, encrypted backup storage, and deployment-local recovery CLI behavior.
- Sprint 1 gives backup-storage binding groundwork for `U-10-05` and direct effective-config evidence for `E-10-04`; Sprint 9 carries the encrypted-storage completion for `U-10-05`.
- Sprint 2 gives `U-10-01` and `I-10-01` direct artifact-backed backup metadata, retention-floor, and latest-success evidence without claiming restore orchestration.
- Sprint 2 gives `U-10-04` and `E-10-03` direct public route-family absence evidence; Sprint 10 gives `E-10-01` canonical local-operator recovery evidence.
- Sprint 3 gives recovery-owned `U-10-02` restore readiness and same-point store ordering evidence plus process-owned `I-10-02` workbook and row/change/blob consistency evidence.
- Sprint 4 remediation gives `U-10-03` and `I-10-03` direct broken-artifact fail-closed evidence and restore-verification state transition evidence.
- Sprint 5 remediation gives `E-10-02` direct browser-visible workbook recovery evidence through an isolated browser restore helper and ordinary workbook UI/API surfaces.
- Sprint 9 remediation gives `U-10-01`, `U-10-05`, and `I-10-01` direct evidence for durable artifact-backed classification and encrypted artifact envelopes. Legacy deployment-admin recovery aliases are negative-only and excluded from Phase 10 conformance accounting.
- Operational backup and restore remain deployment-local recovery behavior and are distinct from the Incident Portability Extension Profile.
- Generated ledgers and schedules are downstream artifacts and must not be hand-edited.

## Authoritative Execution

- `backend-unit` selects static public route-family absence evidence for `U-10-04` and support-only backup-root configuration regression coverage.
- `backend-store` selects durable artifact-backed retained `backup_set` metadata, retention, and encrypted backup-storage boundary evidence for `U-10-01` and `U-10-05`.
- `backend-integration` selects recovery-owned restore readiness, fail-closed artifact, verification-state, and durable latest-lookup evidence for `U-10-02`, `U-10-03`, `I-10-01`, and `I-10-03`.
- `backend-process` selects canonical operator recovery evidence for `E-10-01`, process public-route absence evidence for `E-10-03`, fresh-environment real-server restore evidence for `I-10-02`, and direct effective-config startup rejection evidence for `E-10-04`.
- `browser-e2e-webserver-backed` selects direct browser-visible restored workbook evidence for `E-10-02`.

## Support-Only Execution

- `internal/platform/config/config_phase10_test.go` runs through `backend-unit` with `TestSupportPhase10_` and is forbidden from claiming `U-10-*` identifiers.
- Existing runtime-root, object-store, projection, workbook, route, and browser evidence from earlier phases is support-only substrate and cannot satisfy Phase 10 rows.

## Unit

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |
| `U-10-01` | `implemented` | `internal/modules/recovery/phase10_backup_metadata_store_test.go::TestPhase10_U_10_01_BackupMetadataShapeAndRetentionFloors`, `TestPhase10_U_10_01_VerificationVocabularyAndTimestampRules`, `TestPhase10_U_10_01_LatestSuccessfulRetainedBackupRequiresTwentyFourHourFloor`, `TestPhase10_U_10_01_DurableCatalogSkipsMetadataWithMissingArtifacts`, `TestPhase10_U_10_01_RetentionFloorRejectsShortMetadataAndArtifacts`, `TestPhase10_U_10_01_CaptureRequiresArtifactProofs` | `backend_store` | Durable backup_set creation requires artifact-backed capture with one declared consistency_point_at, stable backup_set_id, backup-storage restore anchors, Postgres and object-store artifact proofs, integrity-manifest proof, exact verification_state vocabulary, timestamp/nullability enforcement, latest-success 24-hour freshness, required artifact durability plus redacted durability diagnostics before successful-retained classification, and 30-day metadata and artifact retention floors. | This evidence does not implement restore orchestration, projection rebuild readiness, restore-verification cadence execution, or arbitrary point-in-time restore. |
| `U-10-02` | `implemented` | `internal/modules/recovery/phase10_restore_owner_integration_test.go::TestPhase10_U_10_02_RestoreReadinessAndCoherentStoreOrder` | `backend_integration` | Restore readiness selects the latest retained successful backup_set only when its latest consistency point is unambiguous, restores Postgres and object storage from that one backup set, rebuilds required projections, verifies authoritative row, change-set, and blob-hash coherence, and marks readiness only after those steps complete in deterministic order. | This evidence does not implement fail-closed missing/corrupt artifact behavior, restore-verification cadence, browser recovery, public route families, or arbitrary timestamp restore. |
| `U-10-03` | `implemented` | `internal/modules/recovery/phase10_restore_owner_integration_test.go::TestPhase10_U_10_03_FailClosedRestoreVerificationBlocked` | `backend_integration` | Selected-backup restore fails before readiness when required artifacts or integrity proof are missing or corrupt, and restore verification records verified or failed state only after isolated verification execution. | This evidence does not implement browser-visible restore recovery, deployment scheduling of the operator due runner, or arbitrary timestamp restore. |
| `U-10-04` | `implemented` | `internal/modules/recovery/phase10_recovery_sentinel_test.go::TestPhase10_U_10_04_PublicRouteAbsenceStaticInventory` | `backend_unit` | Authored public route inventory exposes no `/api/v1/backups*`, `/api/v1/restores*`, `/api/v1/restore-verifications*`, `/ws/v1/backups*`, `/ws/v1/restores*`, or `/ws/v1/restore-verifications*` families. | Deployment-local operator authorization is covered by `E-10-01`; live process route probing is covered by `E-10-03`. |
| `U-10-05` | `implemented` | `internal/modules/recovery/phase10_backup_metadata_store_test.go::TestPhase10_U_10_05_CaptureRequiresEncryptedBackupStorage`, `TestPhase10_U_10_05_EncryptedBackupStorageFailsClosedWithWrongKey` | `backend_store` | Incident-bearing backup capture requires an encrypted backup-storage adapter, encrypted artifact envelopes decrypt and verify with the configured recovery key, and wrong-key reads fail closed before proof validation can pass. | Effective backup-root configuration startup rejection is covered by `E-10-04`; managed-service backup storage remains unsupported until a later adapter proof contract exists. |

## Integration

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |
| `I-10-01` | `implemented` | `internal/modules/recovery/phase10_backup_metadata_integration_test.go::TestPhase10_I_10_01_RealBackingStorageMetadataPersistsAndLatestLookup` | `backend_integration` | Real Postgres-backed runtime metadata persistence is reached only after capture writes encrypted Postgres row-snapshot, object-store snapshot, and integrity-manifest artifacts under `roots.backup_storage`; latest durable lookup exposes same-point restore anchors, default unverified/null verification state, artifact proofs, raw-ciphertext storage, and AC-398 retention floors from a fresh store. | This evidence does not perform restore orchestration or prove workbook recovery. |
| `I-10-02` | `implemented` | `internal/app/serverprocess/phase10_recovery_sentinel_test.go::TestPhase10_I_10_02_FreshEnvironmentRestoreWorkbookConsistency` | `backend_process` | Restoring the latest successful retained backup_set into a fresh isolated environment rebuilds projections, starts the server ready only after restore completion, opens the restored incident through the process, executes the built-in timeline workbook query, and preserves authoritative rows, change sets, and blob hashes. | This evidence does not implement browser recovery, fail-closed missing/corrupt artifact behavior, restore-verification cadence, public route families, or arbitrary timestamp restore. |
| `I-10-03` | `implemented` | `internal/modules/recovery/phase10_restore_owner_integration_test.go::TestPhase10_I_10_03_MissingArtifactFailsBeforeReadinessBlocked` | `backend_integration` | Selecting a retained backup_set missing a required artifact or integrity proof must fail before target readiness. | This evidence does not implement browser-visible restore recovery or arbitrary timestamp restore. |

## Browser E2E

| Row | Claim status | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- |
| `E-10-01` | `implemented` | `internal/modules/recovery/operatortest/operator_process_test.go::TestPhase10_E_10_01_CanonicalOperatorBackupInspectLatest`, `TestPhase10_E_10_01_CanonicalOperatorBackupCreate`, `TestPhase10_E_10_01_CanonicalOperatorRestoreLatest`, `TestPhase10_E_10_01_CanonicalOperatorRestoreVerifyLatest`, `TestPhase10_E_10_01_CanonicalOperatorRestoreVerifyDue` | `backend_process` | Canonical operator process evidence maps the implementation binary to the five Core logical recovery commands, proves local-operator invocation without runtime `deployment_admin` authorization, emits exactly one `cartulary.operator_recovery_result.v1` stdout envelope, validates `cartulary.operator_recovery_progress.v1` stderr JSONL when requested, checks closed invalid/preflight/key error mapping, proves target preflight before mutation, verifies due-order/no-op behavior, records encrypted recovery journal rows and safe administrative-audit summaries for admitted mutating operations, and keeps outputs free of credentials, raw DSNs, endpoints, buckets, object keys, raw paths, recovery keys, and incident content. | Legacy `backup capture`, `backup-metadata latest`, legacy recovery flags, and deployment-admin-gated recovery aliases are negative-only and are not conformance evidence. Object-store migration tooling is release-support evidence only and is not Phase 10 recovery conformance. |
| `E-10-02` | `implemented` | `apps/web/e2e/phase10.restore.spec.ts::Phase 10 E-10-02 restore recovers workbook surface and executes a built-in workbook query` | `browser_functional` | A harness-owned browser fixture captures a retained backup set from the source deployment, restores it into a separately migrated target deployment through the production restore runner, rebuilds projections, opens a restored incident through the target workbook UI, and executes a built-in workbook query through ordinary product routes. | This browser evidence does not expose public backup or restore route families and does not implement arbitrary timestamp restore. |
| `E-10-03` | `implemented` | `internal/app/serverprocess/phase10_recovery_sentinel_test.go::TestPhase10_E_10_03_PublicRouteInventoryAbsence` | `backend_process` | A running server exposes no public HTTP or WebSocket backup, restore, or restore-verification route families. | Deployment-local operator authorization is covered by `E-10-01`; this evidence does not implement restore orchestration or restore-verification cadence. |
| `E-10-04` | `implemented` | `internal/app/serverprocess/phase10_config_test.go::TestPhase10_EffectiveConfigBackupRoot_E_10_04` | `backend_process` | Effective deployment configuration fails the real server process before readiness when roots.backup_storage is missing, uses a profile-incompatible binding, or is satisfied by export or temporary-work roots. | Encrypted incident-bearing artifact conformance is covered by `U-10-05` and `I-10-01`; this process evidence does not implement backup metadata, restore orchestration, restore verification, or operator controls. |

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
