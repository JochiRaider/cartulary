# Recovery Refactor Changed-File Inventory

- Execution base: `d45f3fbf`.
- Completed implementation and RS-12 ledger base: `391f2dc4`.
- The final RS-99 ledger changes only
  `docs/handoffs/recovery-module-refactor-tracker.md` after the substantive
  handoff commit.
- Status uses Git's name-status vocabulary: `A` added, `D` deleted, `M`
  modified, and `Rnnn` renamed with the reported similarity score.

## Exact changed-file inventory

```text
R070	tools/seaweedfs_migration_occurrence_classifications.json	tools/seaweedfs_occurrence_classifications.json
M	contracts/index.json
M	contracts/object-store/release-threat-policy.v1.json
M	contracts/projection-providers/index.json
A	contracts/recovery/backup-artifact-envelope.v2.schema.json
A	contracts/recovery/backup-integrity-manifest.v3.schema.json
A	contracts/recovery/common.v1.schema.json
A	contracts/recovery/fixtures/backup-artifact-envelope.v2.json
A	contracts/recovery/fixtures/backup-integrity-manifest.v3.json
A	contracts/recovery/fixtures/object-store-backup-manifest.v2.json
A	contracts/recovery/fixtures/object-store-backup-summary.v2.json
A	contracts/recovery/fixtures/operator-recovery-audit-summary.v2.json
A	contracts/recovery/fixtures/operator-recovery-journal-payload.v2.json
A	contracts/recovery/fixtures/postgres-snapshot-artifact.v2.json
A	contracts/recovery/fixtures/postgres-snapshot-unit.v1.json
A	contracts/recovery/fixtures/recovery-state-catalog.v1.json
A	contracts/recovery/fixtures/recovery-state-contribution.v1.json
A	contracts/recovery/fixtures/restore-target-marker.v2.json
A	contracts/recovery/fixtures/restore-verification.v2.json
A	contracts/recovery/fixtures/restore-workbook-probe-registration.v1.json
A	contracts/recovery/index.json
A	contracts/recovery/object-store-backup-manifest.v2.schema.json
A	contracts/recovery/object-store-backup-summary.v2.schema.json
A	contracts/recovery/operator-recovery-audit-summary.v2.schema.json
A	contracts/recovery/operator-recovery-journal-payload.v2.schema.json
A	contracts/recovery/postgres-snapshot-artifact.v2.schema.json
A	contracts/recovery/postgres-snapshot-unit.v1.schema.json
A	contracts/recovery/recovery-state-catalog.v1.schema.json
A	contracts/recovery/recovery-state-contribution.v1.schema.json
A	contracts/recovery/restore-target-marker.v2.schema.json
A	contracts/recovery/restore-verification.v2.schema.json
A	contracts/recovery/restore-workbook-probe-registration.v1.schema.json
M	deploy/mvp/README.md
M	deploy/mvp/restore-verification-target.marker.json.example
M	deploy/mvp/scripts/restore-verify-due.sh
M	docs/extension-subsystem-nlspec.md
M	docs/graph_projection_nlspec.md
A	docs/handoffs/recovery-module-refactor-changed-files.md
M	docs/handoffs/recovery-module-refactor-tracker.md
M	docs/spec/00_document_set_status_and_precedence.md
M	docs/spec/01_architecture_storage_and_view_contracts.md
M	docs/spec/04_security_deployment_and_conformance.md
M	docs/testing-harness-nlspec.md
M	internal/app/operator/operator_recovery_test.go
M	internal/app/operator/operator_recovery.go
M	internal/app/operator/operator_test.go
A	internal/app/operator/recoverycli/cli_test.go
A	internal/app/operator/recoverycli/cli.go
A	internal/app/recoveryassembly/evidence_repository_test.go
A	internal/app/recoveryassembly/evidence_repository.go
A	internal/app/recoveryassembly/state_catalog_test.go
A	internal/app/recoveryassembly/state_catalog.go
A	internal/app/recoveryassembly/state_coverage_integration_test.go
A	internal/app/recoveryassembly/state_coverage.go
M	internal/app/recoveryassembly/storage_test.go
M	internal/app/recoveryassembly/storage.go
A	internal/app/recoveryassembly/target_admission_integration_test.go
A	internal/app/recoveryassembly/target_admission.go
A	internal/app/recoveryassembly/vnext.go
M	internal/app/server/runtime_integration_test.go
M	internal/app/server/runtime.go
M	internal/app/serverprocess/recovery_sentinel_test.go
M	internal/app/timelineassembly/assembly.go
M	internal/gen/contractextensions/artifacts_gen.go
A	internal/gen/contractrecovery/artifacts_gen.go
A	internal/modules/artifacts/recovery_state.go
A	internal/modules/assessments/recovery_state.go
A	internal/modules/auth/recovery_state.go
A	internal/modules/collaboration/recovery_state.go
A	internal/modules/entities/recovery_state.go
A	internal/modules/evidence/recovery_inventory.go
A	internal/modules/evidence/recovery_state.go
A	internal/modules/evidence/recoveryprovider/provider.go
M	internal/modules/extensions/contract_test.go
A	internal/modules/extensions/recoverycontribution/contribution.go
A	internal/modules/extensions/recoverycontribution/inventory.go
A	internal/modules/graphprojection/recovery_state.go
A	internal/modules/imports/recovery_inventory.go
A	internal/modules/imports/recovery_state.go
A	internal/modules/incidentbundles/sourceport/recovery_inventory.go
A	internal/modules/incidentbundles/sourceport/recovery_state.go
A	internal/modules/incidents/recovery_state.go
A	internal/modules/indicators/recovery_state.go
A	internal/modules/links/recovery_state.go
A	internal/modules/networkflow/recovery_state.go
A	internal/modules/parties/recovery_state.go
M	internal/modules/projections/provider_boundary.go
A	internal/modules/projections/recovery_state.go
A	internal/modules/recovery/application/due_batch_test.go
A	internal/modules/recovery/application/due_batch.go
A	internal/modules/recovery/application/evidence.go
A	internal/modules/recovery/application/runtime_lifetime_test.go
A	internal/modules/recovery/application/service.go
A	internal/modules/recovery/application/target_admission_test.go
A	internal/modules/recovery/application/target_admission.go
A	internal/modules/recovery/application/types_test.go
A	internal/modules/recovery/application/types.go
M	internal/modules/recovery/backup_metadata_integration_test.go
M	internal/modules/recovery/capture.go
M	internal/modules/recovery/catalog.go
M	internal/modules/recovery/encryption.go
A	internal/modules/recovery/evidence_objects.go
M	internal/modules/recovery/extension_recovery_contract_test.go
M	internal/modules/recovery/extensions.go
M	internal/modules/recovery/object_store_backup_artifacts.go
D	internal/modules/recovery/object_store_migration_preservation_integration_test.go
D	internal/modules/recovery/object_store_migration_test.go
D	internal/modules/recovery/object_store_migration.go
D	internal/modules/recovery/operatorcli/cli.go
D	internal/modules/recovery/operatorcli/journal_test.go
D	internal/modules/recovery/operatorcli/journal.go
D	internal/modules/recovery/operatorops/operations.go
M	internal/modules/recovery/operatortest/operator_process_test.go
A	internal/modules/recovery/recovery_state.go
A	internal/modules/recovery/repositories.go
M	internal/modules/recovery/restore_owner_integration_test.go
M	internal/modules/recovery/restore_projection_contract_test.go
A	internal/modules/recovery/restore_verification_artifact_test.go
A	internal/modules/recovery/restore_verification_artifact.go
M	internal/modules/recovery/restore.go
M	internal/modules/recovery/store.go
A	internal/modules/recovery/streaming_encryption_test.go
A	internal/modules/recovery/streaming_encryption.go
M	internal/modules/recovery/verification.go
A	internal/modules/recovery/vnext_codec_test.go
A	internal/modules/recovery/vnext_codec.go
M	internal/modules/recovery/workbook_probe_test.go
M	internal/modules/recovery/workbook_probe.go
A	internal/modules/reference_data/recoverycontribution/contribution.go
A	internal/modules/reference_data/recoverycontribution/inventory.go
A	internal/modules/reportcomposition/recovery_state.go
A	internal/modules/reporting/recovery_inventory.go
A	internal/modules/reporting/recovery_state.go
A	internal/modules/revisions/recovery_state.go
A	internal/modules/savedviews/recovery_state.go
A	internal/modules/tasksdecisions/recovery_state.go
A	internal/modules/timeline/recovery_probe_test.go
A	internal/modules/timeline/recovery_probe.go
A	internal/modules/timeline/recovery_state.go
A	internal/modules/workbook/restoreprobe/registry_test.go
A	internal/modules/workbook/restoreprobe/registry.go
A	internal/platform/administrativeaudit/recovery_state.go
M	internal/platform/httpapi/json_object.go
A	internal/platform/jobs/recovery_state.go
A	internal/platform/postgres/recovery_state.go
M	internal/platform/processlease/lease.go
M	internal/platform/processlease/postgres_integration_test.go
M	internal/platform/processlease/postgres.go
A	internal/platform/recoverystate/catalog.go
A	internal/platform/strictjson/object.go
A	internal/platform/workbookprobe/contract.go
M	tools/backend_module_boundaries.json
M	tools/contractgen/main.go
A	tools/contractgen/recovery_validation.go
M	tools/contractgen/validation_test.go
M	tools/contractgen/validation.go
M	tools/execution_topology_manifest.json
M	tools/execution_topology_render_index.json
M	tools/go_test_duration_baselines.json
M	tools/harness_public_target_duration_baselines.json
M	tools/harness/generated-artifacts/check-json-shapes.mjs
M	tools/harness/observability/performance-evidence.mjs
M	tools/harness/tests/test-harness-contracts.mjs
M	tools/harness/tests/test-run-make-sequence-fast.sh
M	tools/harness/tests/test-run-make-sequence.sh
M	tools/recoverybrowserrestore/main.go
M	tools/release-evidence/seaweedfs-release-evidence.mjs
M	tools/release-evidence/tests/test-seaweedfs-release-evidence.mjs
M	tools/scheduler_manifest.json
M	tools/task_surface_manifest.json
M	tools/task_surface_owner.json
M	tools/task_surface.generated.mk
M	tools/test_families/app.operator.json
M	tools/test_families/app.server.json
M	tools/test_families/module.recovery.json
M	tools/test_families/module.timeline.json
M	tools/test_families/module.workbook.json
```

## Retained historical recovery readers

The following exact historical schema identities remain read-only. They are
not aliases, writer formats, fallback guesses, or token-normalization paths:

- `cartulary.backup_artifact_envelope.v1`, read by the legacy encrypted
  `BackupStorage` path in `internal/modules/recovery/encryption.go`;
- `cartulary.backup_integrity_manifest.v2`, read by
  `DecodeIntegrityManifest` in `internal/modules/recovery/capture.go`;
- `cartulary.object_store_backup_manifest.v1`, read by
  `DecodeObjectStoreBackupManifestArtifact` in
  `internal/modules/recovery/object_store_backup_artifacts.go`;
- `cartulary.object_store_backup_summary.v1`, read by
  `DecodeObjectStoreBackupSummaryArtifact` in
  `internal/modules/recovery/object_store_backup_artifacts.go`;
- `cartulary.object_store_snapshot_artifact.v2`, read by
  `DecodeObjectStoreSnapshotArtifact` in
  `internal/modules/recovery/restore.go`;
- `cartulary.postgres_snapshot_artifact.v1`, read by
  `DecodePostgresSnapshotArtifact` in
  `internal/modules/recovery/restore.go`; and
- `cartulary.restore_verification.v1`, read by the strict
  `DecodeRestoreVerificationArtifactV1` in
  `internal/modules/recovery/object_store_backup_artifacts.go`.

Removal remains prohibited until every referencing `retained_until` has
passed and no retained backup metadata remains.
