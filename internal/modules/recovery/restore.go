package recovery

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"
	"github.com/JochiRaider/cartulary/internal/platform/workbookprobe"
)

type RestoreStep string

const (
	RestoreStepPostgresRestore    RestoreStep = "postgres_restore"
	RestoreStepObjectStoreRestore RestoreStep = "object_store_restore"
	RestoreStepExtensionBindings  RestoreStep = "extension_binding_validation"
	RestoreStepProjectionRebuild  RestoreStep = "projection_rebuild"
	RestoreStepConsistencyCheck   RestoreStep = "consistency_check"
	RestoreStepReadiness          RestoreStep = "readiness"
)

type RestoreStageError struct {
	Stage RestoreStep
	Cause error
}

func (err *RestoreStageError) Error() string {
	if err == nil || err.Cause == nil {
		return "recovery: restore stage failed"
	}
	return err.Cause.Error()
}

func (err *RestoreStageError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Cause
}

func restoreStageFailure(stage RestoreStep, cause error) error {
	if cause == nil {
		return nil
	}
	return &RestoreStageError{Stage: stage, Cause: cause}
}

type RestoreRunner struct {
	store            backupRepository
	storage          BackupStorage
	extensionBackups *ExtensionBackupCatalog
	now              func() time.Time
	stateCatalog     *recoverystate.Catalog
}

type RestoreTarget struct {
	Postgres        postgres.DB
	ObjectStore     objectstore.Store
	EvidenceObjects EvidenceRecoveryProvider
	Projections     restorecontract.ProjectionRebuilder
	Readiness       RestoreReadinessGate
	Failure         RestoreFailureGate
	Observer        RestoreStepObserver
}

type RestoreReadinessGate interface {
	MarkRestoreReady(ctx context.Context, result RestoreResult) error
}

type RestoreFailureGate interface {
	MarkRestoreFailed(ctx context.Context, cause error)
}

type RestoreStepObserver interface {
	RecordRestoreStep(step RestoreStep)
}

type RestoreResult struct {
	BackupSet                 BackupSet
	ConsistencyReport         RestoreConsistencyReport
	ObjectStoreBackupManifest ObjectStoreBackupManifest
	ProjectionRebuildResult   restorecontract.ProjectionRebuildResult
	ExtensionBindings         []ExtensionBindingProof
	SelectedIncidentID        *string
	WorkbookProbe             *workbookprobe.Result
	IntegrityManifestSHA256   string
	RestoredObjectCount       int64
}

type RestoreConsistencyReport struct {
	AuthoritativeRowsSHA256 string
	AuthoritativeRowCount   int
	ChangeSetsSHA256        string
	ChangeSetRowCount       int
	BlobHashesSHA256        string
	BlobCount               int
}

type selectedRestoreArtifacts struct {
	PostgresSnapshot          PostgresSnapshotArtifact
	ObjectStoreSnapshot       ObjectStoreSnapshotArtifact
	ObjectStoreBackupManifest ObjectStoreBackupManifest
	ExtensionBindings         []ExtensionBindingProof
}

func NewRestoreRunner(store backupRepository, storage BackupStorage, extensionBackups *ExtensionBackupCatalog) *RestoreRunner {
	return &RestoreRunner{
		store:            store,
		storage:          storage,
		extensionBackups: extensionBackups,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func NewVersionedRestoreRunner(
	store backupRepository,
	storage BackupStorage,
	extensionBackups *ExtensionBackupCatalog,
	stateCatalog *recoverystate.Catalog,
) *RestoreRunner {
	runner := NewRestoreRunner(store, storage, extensionBackups)
	runner.stateCatalog = stateCatalog
	return runner
}

func (runner *RestoreRunner) RestoreLatestSuccessfulRetained(ctx context.Context, target RestoreTarget, asOf time.Time) (RestoreResult, error) {
	if runner == nil || runner.store == nil || runner.storage == nil || runner.extensionBackups == nil {
		return RestoreResult{}, fmt.Errorf("%w: restore runner requires store and backup storage", ErrInvalidBackupMetadata)
	}
	if target.Postgres == nil {
		return RestoreResult{}, fmt.Errorf("%w: restore target postgres is required", ErrInvalidBackupArtifact)
	}
	if target.ObjectStore == nil {
		return RestoreResult{}, fmt.Errorf("%w: restore target object store is required", ErrInvalidBackupArtifact)
	}
	if target.Projections == nil {
		return RestoreResult{}, fmt.Errorf("%w: restore projection rebuilder is required", ErrInvalidBackupArtifact)
	}
	if target.EvidenceObjects == nil {
		return RestoreResult{}, fmt.Errorf("%w: Evidence recovery provider is required", ErrInvalidBackupArtifact)
	}
	if asOf.IsZero() {
		asOf = runner.now()
	}
	backupSet, err := NewBackupCatalog(runner.store, runner.storage, runner.extensionBackups).RestoreCandidateBackup(ctx, asOf)
	if err != nil {
		return RestoreResult{}, err
	}
	return runner.RestoreBackupSet(ctx, target, backupSet)
}

func (runner *RestoreRunner) RestoreBackupSet(ctx context.Context, target RestoreTarget, backupSet BackupSet) (result RestoreResult, restoreErr error) {
	defer func() {
		if restoreErr != nil && target.Failure != nil {
			target.Failure.MarkRestoreFailed(context.WithoutCancel(ctx), restoreErr)
		}
	}()
	if runner == nil || runner.store == nil || runner.storage == nil || runner.extensionBackups == nil {
		return RestoreResult{}, fmt.Errorf("%w: restore runner requires store and backup storage", ErrInvalidBackupMetadata)
	}
	if backupSet.BackupSetID == uuid.Nil {
		return RestoreResult{}, ErrBackupSetNotFound
	}
	if target.Postgres == nil {
		return RestoreResult{}, fmt.Errorf("%w: restore target postgres is required", ErrInvalidBackupArtifact)
	}
	if target.ObjectStore == nil {
		return RestoreResult{}, fmt.Errorf("%w: restore target object store is required", ErrInvalidBackupArtifact)
	}
	if target.Projections == nil {
		return RestoreResult{}, fmt.Errorf("%w: restore projection rebuilder is required", ErrInvalidBackupArtifact)
	}
	if target.EvidenceObjects == nil {
		return RestoreResult{}, fmt.Errorf("%w: Evidence recovery provider is required", ErrInvalidBackupArtifact)
	}
	if err := requireEmptyRestoreTarget(ctx, target, runner.extensionBackups); err != nil {
		return RestoreResult{}, err
	}
	if _, vNext := VNextLogicalRefFromMetadataKey(backupSet.IntegrityManifestKey); vNext {
		return runner.restoreVNextBackupSet(ctx, target, backupSet)
	}
	artifacts, err := runner.loadSelectedRestoreArtifacts(ctx, backupSet)
	if err != nil {
		return RestoreResult{}, err
	}
	partialResult := RestoreResult{
		BackupSet:                 backupSet,
		ObjectStoreBackupManifest: artifacts.ObjectStoreBackupManifest,
		ExtensionBindings:         append([]ExtensionBindingProof(nil), artifacts.ExtensionBindings...),
		IntegrityManifestSHA256:   backupSet.IntegrityManifestSHA256,
		RestoredObjectCount:       int64(artifacts.ObjectStoreBackupManifest.ObjectCount),
	}

	recordStep(target.Observer, RestoreStepPostgresRestore)
	if err := restorePostgresSnapshot(ctx, target.Postgres, artifacts.PostgresSnapshot); err != nil {
		return partialResult, restoreStageFailure(RestoreStepPostgresRestore, err)
	}

	recordStep(target.Observer, RestoreStepObjectStoreRestore)
	if err := restoreObjectStoreSnapshot(ctx, target.ObjectStore, artifacts.ObjectStoreSnapshot); err != nil {
		return partialResult, restoreStageFailure(RestoreStepObjectStoreRestore, err)
	}

	recordStep(target.Observer, RestoreStepExtensionBindings)
	if err := validateRestoredExtensionBindings(ctx, runner.extensionBackups, artifacts.ExtensionBindings, target.Postgres); err != nil {
		return partialResult, restoreStageFailure(RestoreStepExtensionBindings, err)
	}

	recordStep(target.Observer, RestoreStepProjectionRebuild)
	projectionResult, err := target.Projections.RebuildRestoreProjections(ctx, restoreProjectionRebuildRequest(backupSet))
	partialResult.ProjectionRebuildResult = projectionResult
	if err != nil {
		return partialResult, restoreStageFailure(RestoreStepProjectionRebuild, err)
	}
	if !projectionResult.ReadinessSatisfied() {
		return partialResult, restoreStageFailure(
			RestoreStepProjectionRebuild,
			fmt.Errorf("%w: projection rebuild did not produce ready restore state: status=%q readiness_outcome=%q", ErrInvalidBackupArtifact, projectionResult.Status, projectionResult.ReadinessOutcome),
		)
	}

	recordStep(target.Observer, RestoreStepConsistencyCheck)
	report, err := verifyRestoredConsistency(ctx, target, artifacts)
	if err != nil {
		return partialResult, restoreStageFailure(RestoreStepConsistencyCheck, err)
	}
	result = RestoreResult{
		BackupSet:                 backupSet,
		ConsistencyReport:         report,
		ObjectStoreBackupManifest: artifacts.ObjectStoreBackupManifest,
		ProjectionRebuildResult:   projectionResult,
		ExtensionBindings:         append([]ExtensionBindingProof(nil), artifacts.ExtensionBindings...),
		IntegrityManifestSHA256:   backupSet.IntegrityManifestSHA256,
		RestoredObjectCount:       int64(artifacts.ObjectStoreBackupManifest.ObjectCount),
	}

	if target.Readiness != nil {
		recordStep(target.Observer, RestoreStepReadiness)
		if err := target.Readiness.MarkRestoreReady(ctx, result); err != nil {
			return result, restoreStageFailure(RestoreStepReadiness, err)
		}
	}
	return result, nil
}

func (runner *RestoreRunner) restoreVNextBackupSet(
	ctx context.Context,
	target RestoreTarget,
	backupSet BackupSet,
) (RestoreResult, error) {
	if runner.stateCatalog == nil {
		return RestoreResult{}, fmt.Errorf("%w: current recovery-state catalog is required", ErrVNextBackup)
	}
	streaming, err := RequireStreamingBackupStorage(runner.storage)
	if err != nil {
		return RestoreResult{}, err
	}
	algorithms, err := NewVNextRestoreAlgorithmCatalog(
		runner.stateCatalog,
		RequiredVNextRestoreAlgorithmIDs(runner.stateCatalog)...,
	)
	if err != nil {
		return RestoreResult{}, err
	}
	restore, err := NewVNextRestoreService(streaming, runner.stateCatalog, algorithms)
	if err != nil {
		return RestoreResult{}, err
	}
	integrityProof, err := VNextProofFromMetadata(
		ctx,
		streaming,
		backupSet.IntegrityManifestKey,
		vNextJSONContentType,
		backupSet.IntegrityManifestSizeBytes,
		backupSet.IntegrityManifestSHA256,
	)
	if err != nil {
		return RestoreResult{}, err
	}
	recordStep(target.Observer, RestoreStepPostgresRestore)
	recordStep(target.Observer, RestoreStepObjectStoreRestore)
	if err := restore.Restore(ctx, &vNextRestoreTarget{
		target: target, stateCatalog: runner.stateCatalog,
	}, integrityProof); err != nil {
		return RestoreResult{BackupSet: backupSet}, restoreStageFailure(RestoreStepPostgresRestore, err)
	}
	verificationEvidence, err := restore.ReadVerificationEvidence(ctx, integrityProof)
	if err != nil {
		return RestoreResult{BackupSet: backupSet}, restoreStageFailure(RestoreStepConsistencyCheck, err)
	}
	recordStep(target.Observer, RestoreStepProjectionRebuild)
	projectionResult, err := target.Projections.RebuildRestoreProjections(
		ctx,
		restoreProjectionRebuildRequest(backupSet),
	)
	if err != nil {
		return RestoreResult{BackupSet: backupSet}, restoreStageFailure(RestoreStepProjectionRebuild, err)
	}
	if !projectionResult.ReadinessSatisfied() {
		return RestoreResult{BackupSet: backupSet}, restoreStageFailure(
			RestoreStepProjectionRebuild,
			fmt.Errorf("%w: projection rebuild did not produce ready restore state", ErrInvalidBackupArtifact),
		)
	}
	recordStep(target.Observer, RestoreStepConsistencyCheck)
	report, err := vNextConsistencyReport(ctx, target, runner.stateCatalog, backupSet)
	if err != nil {
		return RestoreResult{BackupSet: backupSet}, restoreStageFailure(RestoreStepConsistencyCheck, err)
	}
	result := RestoreResult{
		BackupSet: backupSet, ConsistencyReport: report,
		ProjectionRebuildResult: projectionResult,
		IntegrityManifestSHA256: verificationEvidence.ManifestSHA256,
		RestoredObjectCount:     verificationEvidence.RestoredObjectCount,
	}
	if target.Readiness != nil {
		recordStep(target.Observer, RestoreStepReadiness)
		if err := target.Readiness.MarkRestoreReady(ctx, result); err != nil {
			return result, restoreStageFailure(RestoreStepReadiness, err)
		}
	}
	return result, nil
}

type vNextRestoreTarget struct {
	target       RestoreTarget
	stateCatalog *recoverystate.Catalog
}

func (target *vNextRestoreTarget) WithAtomicRestore(
	ctx context.Context,
	run func(VNextRestoreMutation) error,
) error {
	tx, err := target.target.Postgres.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin vNext restore: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `SET LOCAL session_replication_role = replica`); err != nil {
		return fmt.Errorf("disable vNext restore referential triggers: %w", err)
	}
	mutation := &vNextRestoreMutation{
		tx: tx, objects: target.target.ObjectStore, stateCatalog: target.stateCatalog,
	}
	if err := run(mutation); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit vNext restore: %w", err)
	}
	return nil
}

type vNextRestoreMutation struct {
	tx           pgx.Tx
	objects      objectstore.Store
	stateCatalog *recoverystate.Catalog
}

func (mutation *vNextRestoreMutation) PreparePostgresTables(
	ctx context.Context,
	_ []string,
) error {
	tableNames := make([]string, 0)
	for _, table := range mutation.stateCatalog.Document().Tables {
		switch table.RestoreAction {
		case recoverystate.RestoreState, recoverystate.RebuildState, recoverystate.InvalidateState:
			tableNames = append(tableNames, table.TableName)
		}
	}
	if len(tableNames) == 0 {
		return fmt.Errorf("%w: restore catalog has no tables", ErrVNextBackup)
	}
	if _, err := mutation.tx.Exec(
		ctx,
		"TRUNCATE "+sanitizedTableList(tableNames)+" RESTART IDENTITY CASCADE",
	); err != nil {
		return fmt.Errorf("truncate vNext restore target: %w", err)
	}
	return nil
}

func (mutation *vNextRestoreMutation) InsertPostgresRow(
	ctx context.Context,
	tableName string,
	row json.RawMessage,
) error {
	identifier := pgx.Identifier{tableName}.Sanitize()
	query := fmt.Sprintf(
		"INSERT INTO %s SELECT * FROM jsonb_populate_record(NULL::%s, $1::jsonb)",
		identifier,
		identifier,
	)
	if _, err := mutation.tx.Exec(ctx, query, string(row)); err != nil {
		return fmt.Errorf("restore vNext table %s: %w", tableName, err)
	}
	return nil
}

func (mutation *vNextRestoreMutation) FinishPostgresTable(
	ctx context.Context,
	tableName string,
) error {
	return resetOwnedSequences(ctx, mutation.tx, tableName)
}

func (mutation *vNextRestoreMutation) RestoreObject(
	ctx context.Context,
	object VNextObjectManifestEntry,
	reader io.Reader,
) error {
	hasher := sha256.New()
	counted := &countedReader{reader: io.TeeReader(reader, hasher)}
	if err := mutation.objects.PutObject(
		ctx,
		object.StorageKey,
		counted,
		object.PlaintextBytes,
		object.ContentType,
	); err != nil {
		return fmt.Errorf("restore vNext object %s: %w", object.StorageKey, err)
	}
	if counted.count != object.PlaintextBytes ||
		hex.EncodeToString(hasher.Sum(nil)) != object.PlaintextSHA256 {
		return fmt.Errorf("%w: restored object stream proof mismatch", ErrVNextBackup)
	}
	return nil
}

func (*vNextRestoreMutation) RunCatalogAlgorithm(context.Context, string) error {
	// PreparePostgresTables invalidates every excluded table. Owner projection
	// rebuilders run once after authoritative state commits.
	return nil
}

func vNextConsistencyReport(
	ctx context.Context,
	target RestoreTarget,
	stateCatalog *recoverystate.Catalog,
	backupSet BackupSet,
) (RestoreConsistencyReport, error) {
	var authoritativeCount int
	for _, tableName := range stateCatalog.RequiredTableNames() {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", pgx.Identifier{tableName}.Sanitize())
		if err := target.Postgres.QueryRow(ctx, query).Scan(&count); err != nil {
			return RestoreConsistencyReport{}, fmt.Errorf("count restored vNext table %s: %w", tableName, err)
		}
		authoritativeCount += count
	}
	var changeSetCount int
	if err := target.Postgres.QueryRow(ctx, `
SELECT (SELECT COUNT(*) FROM change_sets)
     + (SELECT COUNT(*) FROM change_set_mutations)
     + (SELECT COUNT(*) FROM record_history_entry_refs)
     + (SELECT COUNT(*) FROM record_revision_conflict_facts)
     + (SELECT COUNT(*) FROM record_revisions)
`).Scan(&changeSetCount); err != nil {
		return RestoreConsistencyReport{}, fmt.Errorf("count restored vNext change sets: %w", err)
	}
	blobDigest, blobCount, err := verifyRestoredBlobRowsDetailed(
		ctx,
		target.EvidenceObjects,
		target.ObjectStore,
	)
	if err != nil {
		return RestoreConsistencyReport{}, err
	}
	return RestoreConsistencyReport{
		AuthoritativeRowsSHA256: backupSet.PostgresArtifactSHA256,
		AuthoritativeRowCount:   authoritativeCount,
		ChangeSetsSHA256:        backupSet.PostgresArtifactSHA256,
		ChangeSetRowCount:       changeSetCount,
		BlobHashesSHA256:        blobDigest,
		BlobCount:               blobCount,
	}, nil
}

func restoreProjectionRebuildRequest(backupSet BackupSet) restorecontract.ProjectionRebuildRequest {
	return restorecontract.ProjectionRebuildRequest{
		RestoreOperationID:     uuid.New(),
		RestoredSourceStateRef: restoreProjectionSourceStateRef(backupSet),
		RebuildScope:           restorecontract.ProjectionRebuildScopeAllActiveProviders,
		ProviderRegistryRef:    restorecontract.ProviderRegistryRefCodeBacked,
	}
}

func restoreProjectionSourceStateRef(backupSet BackupSet) string {
	return fmt.Sprintf("backup_set:%s/postgres_artifact:%s", backupSet.BackupSetID.String(), backupSet.PostgresArtifactSHA256)
}

func requireEmptyRestoreTarget(ctx context.Context, target RestoreTarget, extensionBackups *ExtensionBackupCatalog) error {
	rows, err := target.Postgres.Query(ctx, `
SELECT table_name
  FROM information_schema.tables
 WHERE table_schema = 'public'
   AND table_type = 'BASE TABLE'
 ORDER BY table_name ASC
`)
	if err != nil {
		return fmt.Errorf("inspect restore target tables: %w", err)
	}
	tableNames := make([]string, 0)
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			rows.Close()
			return fmt.Errorf("scan restore target table: %w", err)
		}
		if IsAuthoritativePostgresSnapshotTable(tableName) {
			tableNames = append(tableNames, tableName)
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("iterate restore target tables: %w", err)
	}
	rows.Close()
	for _, tableName := range tableNames {
		if tableName == "extension_state_metadata" {
			if err := requirePristineExtensionMetadata(ctx, target.Postgres, extensionBackups); err != nil {
				return err
			}
			continue
		}
		var count int64
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", pgx.Identifier{tableName}.Sanitize())
		if err := target.Postgres.QueryRow(ctx, query).Scan(&count); err != nil {
			return fmt.Errorf("inspect restore target table %s: %w", tableName, err)
		}
		if count != 0 {
			return fmt.Errorf("%w: table %s contains %d rows", ErrRestoreTargetNotEmpty, tableName, count)
		}
	}
	objects, err := target.ObjectStore.ListObjects(ctx, "")
	if err != nil {
		return fmt.Errorf("inspect restore target object store: %w", err)
	}
	if len(objects) != 0 {
		return fmt.Errorf("%w: object store contains %d objects", ErrRestoreTargetNotEmpty, len(objects))
	}
	return nil
}

func requirePristineExtensionMetadata(ctx context.Context, db postgres.DB, catalog *ExtensionBackupCatalog) error {
	if catalog == nil {
		return fmt.Errorf("%w: extension backup catalog is required", ErrRestoreTargetNotEmpty)
	}
	rows, err := db.Query(ctx, `
SELECT profile_id, migration_lineage_id, state_version,
       COALESCE(last_migration_id, ''), metadata_version
  FROM extension_state_metadata
 ORDER BY profile_id ASC
`)
	if err != nil {
		return fmt.Errorf("inspect restore target extension metadata: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var actual ExtensionPristineMetadata
		if err := rows.Scan(
			&actual.ProfileID,
			&actual.MigrationLineageID,
			&actual.StateVersion,
			&actual.LastMigrationID,
			&actual.MetadataVersion,
		); err != nil {
			return fmt.Errorf("scan restore target extension metadata: %w", err)
		}
		expected, admitted := catalog.pristineMetadataRows[actual.ProfileID]
		if !admitted || actual != expected {
			return fmt.Errorf("%w: extension metadata for %s is not a pristine migration seed", ErrRestoreTargetNotEmpty, actual.ProfileID)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate restore target extension metadata: %w", err)
	}
	return nil
}

// ResetRestoreVerificationTarget returns an exclusively admitted, disposable
// verification target to its migration-owned pristine state between successful
// restore proofs. It is not part of ordinary restore and must never be used to
// make a nonempty production target admissible.
func ResetRestoreVerificationTarget(ctx context.Context, target RestoreTarget, catalog *ExtensionBackupCatalog) error {
	if target.Postgres == nil || target.ObjectStore == nil || catalog == nil {
		return fmt.Errorf("%w: verification target and extension catalog are required", ErrInvalidBackupArtifact)
	}
	rows, err := target.Postgres.Query(ctx, `
SELECT table_name
  FROM information_schema.tables
 WHERE table_schema = 'public'
   AND table_type = 'BASE TABLE'
 ORDER BY table_name ASC
`)
	if err != nil {
		return fmt.Errorf("list restore verification target tables: %w", err)
	}
	tableNames := make([]string, 0)
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			rows.Close()
			return fmt.Errorf("scan restore verification target table: %w", err)
		}
		if IsAuthoritativePostgresSnapshotTable(tableName) {
			tableNames = append(tableNames, tableName)
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("iterate restore verification target tables: %w", err)
	}
	rows.Close()

	tx, err := target.Postgres.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin restore verification target reset: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	if len(tableNames) > 0 {
		query := "TRUNCATE " + sanitizedTableList(tableNames) + " RESTART IDENTITY CASCADE"
		if _, err := tx.Exec(ctx, query); err != nil {
			return fmt.Errorf("truncate restore verification target: %w", err)
		}
	}
	profileIDs := make([]string, 0, len(catalog.pristineMetadataRows))
	for profileID := range catalog.pristineMetadataRows {
		profileIDs = append(profileIDs, profileID)
	}
	sort.Strings(profileIDs)
	for _, profileID := range profileIDs {
		row := catalog.pristineMetadataRows[profileID]
		var lastMigrationID any
		if row.LastMigrationID != "" {
			lastMigrationID = row.LastMigrationID
		}
		if _, err := tx.Exec(ctx, `
INSERT INTO extension_state_metadata (
    profile_id, migration_lineage_id, state_version, last_migration_id,
    metadata_version, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
`, row.ProfileID, row.MigrationLineageID, row.StateVersion, lastMigrationID, row.MetadataVersion); err != nil {
			return fmt.Errorf("restore pristine extension metadata for %s: %w", profileID, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit restore verification target reset: %w", err)
	}

	objects, err := target.ObjectStore.ListObjects(ctx, "")
	if err != nil {
		return fmt.Errorf("list restore verification target objects for reset: %w", err)
	}
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].Key < objects[j].Key
	})
	for _, object := range objects {
		if err := target.ObjectStore.DeleteObject(ctx, object.Key); err != nil {
			return fmt.Errorf("delete restore verification target object %s: %w", object.Key, err)
		}
	}
	return nil
}

func (runner *RestoreRunner) loadSelectedRestoreArtifacts(ctx context.Context, backupSet BackupSet) (selectedRestoreArtifacts, error) {
	manifestProof := BackupArtifactProof{
		Key:       backupSet.IntegrityManifestKey,
		SHA256:    backupSet.IntegrityManifestSHA256,
		SizeBytes: backupSet.IntegrityManifestSizeBytes,
	}
	manifestBody, err := VerifyArtifactProof(ctx, runner.storage, manifestProof)
	if err != nil {
		return selectedRestoreArtifacts{}, fmt.Errorf("verify selected backup integrity manifest: %w", err)
	}
	manifest, err := DecodeIntegrityManifest(manifestBody)
	if err != nil {
		return selectedRestoreArtifacts{}, fmt.Errorf("%w: decode selected backup integrity manifest: %v", ErrInvalidBackupArtifact, err)
	}
	if err := validateSelectedRestoreManifest(backupSet, manifest); err != nil {
		return selectedRestoreArtifacts{}, err
	}
	postgresBody, err := VerifyArtifactProof(ctx, runner.storage, manifest.PostgresArtifact)
	if err != nil {
		return selectedRestoreArtifacts{}, fmt.Errorf("verify selected postgres artifact: %w", err)
	}
	objectBody, err := VerifyArtifactProof(ctx, runner.storage, manifest.ObjectStoreArtifact)
	if err != nil {
		return selectedRestoreArtifacts{}, fmt.Errorf("verify selected object-store artifact: %w", err)
	}
	postgresSnapshot, err := DecodePostgresSnapshotArtifact(postgresBody)
	if err != nil {
		return selectedRestoreArtifacts{}, err
	}
	objectSnapshot, err := DecodeObjectStoreSnapshotArtifact(objectBody)
	if err != nil {
		return selectedRestoreArtifacts{}, err
	}
	if manifest.ObjectStoreBackupManifestArtifact == nil {
		return selectedRestoreArtifacts{}, fmt.Errorf("%w: selected backup is missing object-store backup manifest artifact", ErrInvalidBackupArtifact)
	}
	objectManifestBody, err := VerifyArtifactProof(ctx, runner.storage, *manifest.ObjectStoreBackupManifestArtifact)
	if err != nil {
		return selectedRestoreArtifacts{}, fmt.Errorf("verify selected object-store backup manifest artifact: %w", err)
	}
	objectManifest, err := DecodeObjectStoreBackupManifestArtifact(objectManifestBody)
	if err != nil {
		return selectedRestoreArtifacts{}, err
	}
	if err := ValidateObjectStoreBackupManifestForBackup(backupSet, objectManifest); err != nil {
		return selectedRestoreArtifacts{}, err
	}
	if err := ValidateObjectStoreManifestAgainstSnapshot(objectManifest, objectSnapshot); err != nil {
		return selectedRestoreArtifacts{}, err
	}
	if err := validateExtensionBindingProofs(runner.extensionBackups, manifest.ExtensionBindings, postgresSnapshot); err != nil {
		return selectedRestoreArtifacts{}, err
	}
	if manifest.ObjectStoreBackupSummaryArtifact != nil {
		summaryBody, err := VerifyArtifactProof(ctx, runner.storage, *manifest.ObjectStoreBackupSummaryArtifact)
		if err != nil {
			return selectedRestoreArtifacts{}, fmt.Errorf("verify selected object-store backup summary artifact: %w", err)
		}
		summary, err := DecodeObjectStoreBackupSummaryArtifact(summaryBody)
		if err != nil {
			return selectedRestoreArtifacts{}, err
		}
		if summary.BackupSetID != objectManifest.BackupSetID ||
			!summary.ConsistencyPointAt.Equal(objectManifest.ConsistencyPointAt) ||
			summary.ManifestSHA256 != objectManifest.ManifestSHA256 ||
			summary.ObjectCount != objectManifest.ObjectCount ||
			summary.TotalSizeBytes != objectManifest.TotalSizeBytes {
			return selectedRestoreArtifacts{}, fmt.Errorf("%w: object-store backup summary does not match private manifest", ErrInvalidBackupArtifact)
		}
	}
	return selectedRestoreArtifacts{
		PostgresSnapshot:          postgresSnapshot,
		ObjectStoreSnapshot:       objectSnapshot,
		ObjectStoreBackupManifest: objectManifest,
		ExtensionBindings:         append([]ExtensionBindingProof(nil), manifest.ExtensionBindings...),
	}, nil
}

func validateSelectedRestoreManifest(backupSet BackupSet, manifest BackupIntegrityManifest) error {
	if manifest.SchemaID != BackupIntegrityManifestSchemaID {
		return fmt.Errorf("%w: unsupported integrity manifest schema %q", ErrInvalidBackupArtifact, manifest.SchemaID)
	}
	if manifest.StorageEncryption.Mode != BackupStorageEncryptionModeAESGCM ||
		manifest.StorageEncryption.EnvelopeSchemaID != BackupArtifactEnvelopeSchemaID ||
		!validSHA256Hex(manifest.StorageEncryption.KeyFingerprintSHA256) {
		return fmt.Errorf("%w: integrity manifest does not prove encrypted backup storage", ErrInvalidBackupArtifact)
	}
	if manifest.BackupSetID != backupSet.BackupSetID.String() {
		return fmt.Errorf("%w: manifest backup_set_id does not match selected backup", ErrInvalidBackupArtifact)
	}
	if !manifest.ConsistencyPointAt.Equal(backupSet.ConsistencyPointAt) {
		return fmt.Errorf("%w: manifest consistency_point_at does not match selected backup", ErrInvalidBackupArtifact)
	}
	if manifest.PostgresRestoreAnchor != backupSet.PostgresRestoreAnchor ||
		manifest.ObjectStoreRestoreAnchor != backupSet.ObjectStoreRestoreAnchor {
		return fmt.Errorf("%w: manifest restore anchors do not match selected backup", ErrInvalidBackupArtifact)
	}
	if manifest.PostgresRestoreAnchor != backupStorageAnchorScheme+manifest.PostgresArtifact.Key ||
		manifest.ObjectStoreRestoreAnchor != backupStorageAnchorScheme+manifest.ObjectStoreArtifact.Key {
		return fmt.Errorf("%w: manifest restore anchors do not match artifact keys", ErrInvalidBackupArtifact)
	}
	if !backupProofMatches(manifest.PostgresArtifact, backupSet.PostgresArtifactKey, backupSet.PostgresArtifactSHA256, backupSet.PostgresArtifactSizeBytes) ||
		!backupProofMatches(manifest.ObjectStoreArtifact, backupSet.ObjectStoreArtifactKey, backupSet.ObjectStoreArtifactSHA256, backupSet.ObjectStoreArtifactSizeBytes) {
		return fmt.Errorf("%w: manifest artifact proofs do not match selected backup", ErrInvalidBackupArtifact)
	}
	return nil
}

func backupProofMatches(proof BackupArtifactProof, key string, sha256 string, sizeBytes int64) bool {
	return proof.Key == key && proof.SHA256 == sha256 && proof.SizeBytes == sizeBytes
}

func DecodePostgresSnapshotArtifact(body []byte) (PostgresSnapshotArtifact, error) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var artifact PostgresSnapshotArtifact
	if err := decoder.Decode(&artifact); err != nil {
		return PostgresSnapshotArtifact{}, fmt.Errorf("%w: decode postgres snapshot artifact: %v", ErrInvalidBackupArtifact, err)
	}
	if artifact.SchemaID != PostgresSnapshotArtifactSchemaID {
		return PostgresSnapshotArtifact{}, fmt.Errorf("%w: unsupported postgres snapshot schema %q", ErrInvalidBackupArtifact, artifact.SchemaID)
	}
	seen := make(map[string]struct{}, len(artifact.Tables))
	for _, table := range artifact.Tables {
		if !IsAuthoritativePostgresSnapshotTable(table.TableName) {
			return PostgresSnapshotArtifact{}, fmt.Errorf("%w: postgres snapshot contains non-authoritative table %q", ErrInvalidBackupArtifact, table.TableName)
		}
		if _, exists := seen[table.TableName]; exists {
			return PostgresSnapshotArtifact{}, fmt.Errorf("%w: postgres snapshot contains duplicate table %q", ErrInvalidBackupArtifact, table.TableName)
		}
		seen[table.TableName] = struct{}{}
		if table.RowCount != int64(len(table.Rows)) {
			return PostgresSnapshotArtifact{}, fmt.Errorf("%w: postgres snapshot row_count mismatch for %q", ErrInvalidBackupArtifact, table.TableName)
		}
	}
	sort.Slice(artifact.Tables, func(i, j int) bool {
		return artifact.Tables[i].TableName < artifact.Tables[j].TableName
	})
	return artifact, nil
}

func DecodeObjectStoreSnapshotArtifact(body []byte) (ObjectStoreSnapshotArtifact, error) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var artifact ObjectStoreSnapshotArtifact
	if err := decoder.Decode(&artifact); err != nil {
		return ObjectStoreSnapshotArtifact{}, fmt.Errorf("%w: decode object-store snapshot artifact: %v", ErrInvalidBackupArtifact, err)
	}
	if artifact.SchemaID != ObjectStoreSnapshotArtifactSchemaID {
		return ObjectStoreSnapshotArtifact{}, fmt.Errorf("%w: unsupported object-store snapshot schema %q", ErrInvalidBackupArtifact, artifact.SchemaID)
	}
	seen := make(map[string]struct{}, len(artifact.Objects))
	for index, item := range artifact.Objects {
		if strings.TrimSpace(item.Key) == "" {
			return ObjectStoreSnapshotArtifact{}, fmt.Errorf("%w: object-store snapshot object key is required", ErrInvalidBackupArtifact)
		}
		if _, exists := seen[item.Key]; exists {
			return ObjectStoreSnapshotArtifact{}, fmt.Errorf("%w: object-store snapshot contains duplicate key %q", ErrInvalidBackupArtifact, item.Key)
		}
		seen[item.Key] = struct{}{}
		body, err := base64.StdEncoding.DecodeString(item.BodyBase64)
		if err != nil {
			return ObjectStoreSnapshotArtifact{}, fmt.Errorf("%w: decode object-store body for %s: %v", ErrInvalidBackupArtifact, item.Key, err)
		}
		if int64(len(body)) != item.SizeBytes {
			return ObjectStoreSnapshotArtifact{}, fmt.Errorf("%w: object-store snapshot size mismatch for %s", ErrInvalidBackupArtifact, item.Key)
		}
		if sha256Hex(body) != item.SHA256 {
			return ObjectStoreSnapshotArtifact{}, fmt.Errorf("%w: object-store snapshot sha256 mismatch for %s", ErrInvalidBackupArtifact, item.Key)
		}
		artifact.Objects[index].ContentType = artifactContentType(BackupArtifact{ContentType: item.ContentType})
	}
	sort.Slice(artifact.Objects, func(i, j int) bool {
		return artifact.Objects[i].Key < artifact.Objects[j].Key
	})
	return artifact, nil
}

func restorePostgresSnapshot(ctx context.Context, db postgres.DB, artifact PostgresSnapshotArtifact) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin postgres restore: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, `SET LOCAL session_replication_role = replica`); err != nil {
		return fmt.Errorf("disable postgres restore referential triggers: %w", err)
	}
	tableNames := make([]string, 0, len(artifact.Tables))
	for _, table := range artifact.Tables {
		tableNames = append(tableNames, table.TableName)
	}
	if len(tableNames) > 0 {
		truncateSQL := "TRUNCATE " + sanitizedTableList(tableNames) + " RESTART IDENTITY CASCADE"
		if _, err := tx.Exec(ctx, truncateSQL); err != nil {
			return fmt.Errorf("truncate postgres restore target tables: %w", err)
		}
	}
	for _, table := range artifact.Tables {
		identifier := pgx.Identifier{table.TableName}.Sanitize()
		insertSQL := fmt.Sprintf("INSERT INTO %s SELECT * FROM jsonb_populate_record(NULL::%s, $1::jsonb)", identifier, identifier)
		for _, rawRow := range table.Rows {
			if _, err := tx.Exec(ctx, insertSQL, string(rawRow)); err != nil {
				return fmt.Errorf("restore postgres table %s: %w", table.TableName, err)
			}
		}
		if err := resetOwnedSequences(ctx, tx, table.TableName); err != nil {
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit postgres restore: %w", err)
	}
	return nil
}

func sanitizedTableList(tableNames []string) string {
	parts := make([]string, 0, len(tableNames))
	for _, tableName := range tableNames {
		parts = append(parts, pgx.Identifier{tableName}.Sanitize())
	}
	return strings.Join(parts, ", ")
}

func resetOwnedSequences(ctx context.Context, tx pgx.Tx, tableName string) error {
	rows, err := tx.Query(ctx, `
SELECT column_name
FROM information_schema.columns
WHERE table_schema = 'public'
  AND table_name = $1
  AND column_default LIKE 'nextval(%'
ORDER BY ordinal_position ASC
`, tableName)
	if err != nil {
		return fmt.Errorf("list postgres restore sequences for %s: %w", tableName, err)
	}

	columnNames := make([]string, 0)
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			rows.Close()
			return fmt.Errorf("scan postgres restore sequence for %s: %w", tableName, err)
		}
		columnNames = append(columnNames, columnName)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("iterate postgres restore sequences for %s: %w", tableName, err)
	}
	rows.Close()

	for _, columnName := range columnNames {
		var sequenceName pgtype.Text
		if err := tx.QueryRow(ctx, `SELECT pg_get_serial_sequence($1, $2)`, "public."+tableName, columnName).Scan(&sequenceName); err != nil {
			return fmt.Errorf("resolve postgres restore sequence for %s.%s: %w", tableName, columnName, err)
		}
		if !sequenceName.Valid || sequenceName.String == "" {
			continue
		}
		identifier := pgx.Identifier{tableName}.Sanitize()
		columnIdentifier := pgx.Identifier{columnName}.Sanitize()
		var nextValue int64
		query := fmt.Sprintf("SELECT COALESCE(MAX(%s), 0)::bigint + 1 FROM %s", columnIdentifier, identifier)
		if err := tx.QueryRow(ctx, query).Scan(&nextValue); err != nil {
			return fmt.Errorf("compute postgres restore sequence value for %s.%s: %w", tableName, columnName, err)
		}
		if _, err := tx.Exec(ctx, `SELECT setval($1, $2, false)`, sequenceName.String, nextValue); err != nil {
			return fmt.Errorf("reset postgres restore sequence for %s.%s: %w", tableName, columnName, err)
		}
	}
	return nil
}

func restoreObjectStoreSnapshot(ctx context.Context, store objectstore.Store, artifact ObjectStoreSnapshotArtifact) error {
	existing, err := store.ListObjects(ctx, "")
	if err != nil {
		return fmt.Errorf("list object-store restore target: %w", err)
	}
	sort.Slice(existing, func(i, j int) bool {
		return existing[i].Key < existing[j].Key
	})
	for _, object := range existing {
		if err := store.DeleteObject(ctx, object.Key); err != nil {
			return fmt.Errorf("delete object-store restore target object %s: %w", object.Key, err)
		}
	}
	for _, item := range artifact.Objects {
		body, err := base64.StdEncoding.DecodeString(item.BodyBase64)
		if err != nil {
			return fmt.Errorf("%w: decode object-store restore body for %s: %v", ErrInvalidBackupArtifact, item.Key, err)
		}
		if err := store.PutObject(ctx, item.Key, bytes.NewReader(body), int64(len(body)), item.ContentType); err != nil {
			return fmt.Errorf("restore object-store object %s: %w", item.Key, err)
		}
	}
	return nil
}

func verifyRestoredConsistency(ctx context.Context, target RestoreTarget, artifacts selectedRestoreArtifacts) (RestoreConsistencyReport, error) {
	authoritativeDigest, authoritativeCount, err := postgresSnapshotDigest(artifacts.PostgresSnapshot, nil)
	if err != nil {
		return RestoreConsistencyReport{}, err
	}
	targetSnapshotBody, err := CapturePostgresSnapshotArtifact(ctx, target.Postgres)
	if err != nil {
		return RestoreConsistencyReport{}, fmt.Errorf("capture restored postgres consistency snapshot: %w", err)
	}
	targetSnapshot, err := DecodePostgresSnapshotArtifact(targetSnapshotBody)
	if err != nil {
		return RestoreConsistencyReport{}, err
	}
	targetDigest, _, err := postgresSnapshotDigest(targetSnapshot, nil)
	if err != nil {
		return RestoreConsistencyReport{}, err
	}
	if targetDigest != authoritativeDigest {
		return RestoreConsistencyReport{}, fmt.Errorf("%w: restored authoritative row digest mismatch", ErrInvalidBackupArtifact)
	}

	changeSetDigest, changeSetCount, err := postgresSnapshotDigest(artifacts.PostgresSnapshot, isChangeSetTable)
	if err != nil {
		return RestoreConsistencyReport{}, err
	}
	targetChangeSetDigest, _, err := postgresSnapshotDigest(targetSnapshot, isChangeSetTable)
	if err != nil {
		return RestoreConsistencyReport{}, err
	}
	if targetChangeSetDigest != changeSetDigest {
		return RestoreConsistencyReport{}, fmt.Errorf("%w: restored change-set digest mismatch", ErrInvalidBackupArtifact)
	}

	blobDigest, blobCount, err := verifyRestoredBlobHashes(ctx, target, artifacts.ObjectStoreSnapshot)
	if err != nil {
		return RestoreConsistencyReport{}, err
	}
	return RestoreConsistencyReport{
		AuthoritativeRowsSHA256: authoritativeDigest,
		AuthoritativeRowCount:   authoritativeCount,
		ChangeSetsSHA256:        changeSetDigest,
		ChangeSetRowCount:       changeSetCount,
		BlobHashesSHA256:        blobDigest,
		BlobCount:               blobCount,
	}, nil
}

func postgresSnapshotDigest(artifact PostgresSnapshotArtifact, include func(string) bool) (string, int, error) {
	digest := sha256.New()
	rowCount := 0
	tables := append([]PostgresSnapshotTable(nil), artifact.Tables...)
	sort.Slice(tables, func(i, j int) bool {
		return tables[i].TableName < tables[j].TableName
	})
	for _, table := range tables {
		if include != nil && !include(table.TableName) {
			continue
		}
		_, _ = digest.Write([]byte("table:" + table.TableName + "\n"))
		rows := make([]string, 0, len(table.Rows))
		for _, rawRow := range table.Rows {
			normalized, err := normalizeJSONForDigest(rawRow)
			if err != nil {
				return "", 0, fmt.Errorf("normalize postgres snapshot row for %s: %w", table.TableName, err)
			}
			rows = append(rows, normalized)
		}
		sort.Strings(rows)
		for _, row := range rows {
			_, _ = digest.Write([]byte(row))
			_, _ = digest.Write([]byte("\n"))
			rowCount++
		}
	}
	return hex.EncodeToString(digest.Sum(nil)), rowCount, nil
}

func normalizeJSONForDigest(raw json.RawMessage) (string, error) {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", err
	}
	normalized, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(normalized), nil
}

func isChangeSetTable(tableName string) bool {
	switch tableName {
	case "change_sets", "change_set_mutations", "record_history_entry_refs", "record_revision_conflict_facts", "record_revisions":
		return true
	default:
		return false
	}
}

func verifyRestoredBlobHashes(ctx context.Context, target RestoreTarget, artifact ObjectStoreSnapshotArtifact) (string, int, error) {
	digest := sha256.New()
	objects := append([]ObjectStoreSnapshotItem(nil), artifact.Objects...)
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].Key < objects[j].Key
	})
	for _, item := range objects {
		reader, info, err := target.ObjectStore.ReadObject(ctx, item.Key, objectstore.ReadOptions{})
		if err != nil {
			return "", 0, fmt.Errorf("read restored object %s: %w", item.Key, err)
		}
		body, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil {
			return "", 0, fmt.Errorf("read restored object body %s: %w", item.Key, readErr)
		}
		if closeErr != nil {
			return "", 0, fmt.Errorf("close restored object %s: %w", item.Key, closeErr)
		}
		if info.Size != item.SizeBytes || int64(len(body)) != item.SizeBytes || sha256Hex(body) != item.SHA256 {
			return "", 0, fmt.Errorf("%w: restored object proof mismatch for %s", ErrInvalidBackupArtifact, item.Key)
		}
		_, _ = digest.Write([]byte("object:" + item.Key + ":" + item.SHA256 + "\n"))
	}

	rowDigest, _, err := verifyRestoredBlobRowsDetailed(ctx, target.EvidenceObjects, target.ObjectStore)
	if err != nil {
		return "", 0, err
	}
	_, _ = digest.Write([]byte("blob_rows:" + rowDigest + "\n"))
	return hex.EncodeToString(digest.Sum(nil)), len(objects), nil
}

func verifyRestoredBlobRowsDetailed(ctx context.Context, provider EvidenceRecoveryProvider, store objectstore.Store) (string, int, error) {
	if provider == nil {
		return "", 0, fmt.Errorf("%w: Evidence recovery provider is required", ErrInvalidBackupArtifact)
	}
	objects, err := provider.ListAvailableRecoveryObjects(ctx)
	if err != nil {
		return "", 0, fmt.Errorf("list restored Evidence objects: %w", err)
	}

	digest := sha256.New()
	count := 0
	for _, object := range objects {
		reader, _, err := store.ReadObject(ctx, object.StorageKey, objectstore.ReadOptions{})
		if err != nil {
			return "", count, fmt.Errorf("read restored Evidence object %s: %w", object.StorageKey, err)
		}
		body, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil {
			return "", count, fmt.Errorf("read restored Evidence object body %s: %w", object.StorageKey, readErr)
		}
		if closeErr != nil {
			return "", count, fmt.Errorf("close restored Evidence object %s: %w", object.StorageKey, closeErr)
		}
		sha := sha256Hex(body)
		if int64(len(body)) != object.ByteSize {
			return "", count, fmt.Errorf("%w: restored Evidence object byte_size mismatch for %s", ErrInvalidBackupArtifact, object.StorageKey)
		}
		if object.ObservedSize != nil && *object.ObservedSize != int64(len(body)) {
			return "", count, fmt.Errorf("%w: restored Evidence object observed_size mismatch for %s", ErrInvalidBackupArtifact, object.StorageKey)
		}
		for label, value := range map[string]*string{
			"expected_sha256_hex": object.ExpectedSHA256Hex,
			"observed_sha256_hex": object.ObservedSHA256Hex,
			"blob_hash":           object.BlobHash,
		} {
			if value == nil || strings.TrimSpace(*value) == "" {
				continue
			}
			if !blobHashMatches(*value, sha) {
				return "", count, fmt.Errorf("%w: restored Evidence object %s mismatch for %s", ErrInvalidBackupArtifact, label, object.StorageKey)
			}
		}
		_, _ = digest.Write([]byte(object.StorageKey + ":" + sha + "\n"))
		count++
	}
	return hex.EncodeToString(digest.Sum(nil)), count, nil
}

func blobHashMatches(value string, sha string) bool {
	normalized := strings.TrimSpace(strings.ToLower(value))
	return normalized == sha || normalized == "sha256:"+sha
}

func recordStep(observer RestoreStepObserver, step RestoreStep) {
	if observer != nil {
		observer.RecordRestoreStep(step)
	}
}
