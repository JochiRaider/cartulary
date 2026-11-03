package recovery

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	graphrestore "github.com/JochiRaider/cartulary/internal/modules/graphprojection/restore"
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
	RestoreStepProjectionRebuild  RestoreStep = "projection_rebuild"
	RestoreStepConsistencyCheck   RestoreStep = "consistency_check"
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
	fence              *FencedRestore
	RestoreOperationID uuid.UUID
	TargetGenerationID uuid.UUID
	ProofScratch       *os.Root // borrowed; caller closes after the operation
	Postgres           postgres.DB
	ObjectStore        objectstore.Store
	ReferencePacks     ReferencePackStorage
	ExportOutputs      RootObjectStorage
	EvidenceObjects    EvidenceRecoveryProvider
	Projections        restorecontract.ProjectionRebuilder
	Observer           RestoreStepObserver
}

type RestoreStepObserver interface {
	RecordRestoreStep(step RestoreStep)
}

type RestoreResult struct {
	BackupSet                  BackupSet
	ConsistencyReport          RestoreConsistencyReport
	ProjectionRebuildResult    restorecontract.ProjectionRebuildResult
	RestoreGeneration          *RestoreGenerationEvidence
	SelectedIncidentID         *string
	WorkbookProbe              *workbookprobe.Result
	IntegrityManifestSHA256    string
	RestoredObjectCount        int64
	RecoveryStateCatalogSHA256 string
	CodecRegistrySHA256        string
}

type RestoreConsistencyReport struct {
	AuthoritativeRowsSHA256 string
	AuthoritativeRowCount   int
	ChangeSetsSHA256        string
	ChangeSetRowCount       int
	BlobHashesSHA256        string
	BlobCount               int
}

func NewVersionedRestoreRunner(store backupRepository, storage BackupStorage, extensionBackups *ExtensionBackupCatalog, stateCatalog *recoverystate.Catalog) *RestoreRunner {
	return &RestoreRunner{
		store:            store,
		storage:          storage,
		extensionBackups: extensionBackups,
		stateCatalog:     stateCatalog,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// NewSelectedRestoreRunner consumes an already authenticated selection without a
// live catalog repository. Both live and portable selection use RestoreBackupSet.
func NewSelectedRestoreRunner(storage BackupStorage, extensions *ExtensionBackupCatalog, catalog *recoverystate.Catalog) *RestoreRunner {
	return NewVersionedRestoreRunner(nil, storage, extensions, catalog)
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
	backupSet, err := NewBackupCatalog(runner.store, runner.storage, runner.extensionBackups, runner.stateCatalog).RestoreCandidateBackup(ctx, asOf)
	if err != nil {
		return RestoreResult{}, err
	}
	return runner.RestoreBackupSet(ctx, target, backupSet)
}

func (runner *RestoreRunner) RestoreBackupSet(ctx context.Context, target RestoreTarget, backupSet BackupSet) (result RestoreResult, restoreErr error) {
	fence := target.fence
	if fence == nil || fence.runner != runner || fence.backupSet.BackupSetID != backupSet.BackupSetID {
		return RestoreResult{}, ErrTargetIncomplete
	}
	fence.operation.Lock()
	defer fence.operation.Unlock()
	if fence.closed || fence.used {
		return RestoreResult{}, ErrTargetIncomplete
	}
	if err := ctx.Err(); err != nil {
		return RestoreResult{}, err
	}
	if err := fence.ctx.Err(); err != nil {
		return RestoreResult{}, err
	}
	if err := fence.capability.AssertHeld(); err != nil {
		return RestoreResult{}, err
	}
	fence.used = true
	defer func() {
		if !fence.committed {
			restoreErr = errors.Join(restoreErr, rollbackRestoreTransaction(fence.tx))
		}
	}()
	target = fence.target
	backupSet = fence.backupSet
	work, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(fence.ctx, cancel)
	defer cancel()
	defer stop()
	ctx = work
	return runner.restoreVNextBackupSet(ctx, target, backupSet)
}

func (runner *RestoreRunner) restoreVNextBackupSet(
	ctx context.Context,
	target RestoreTarget,
	backupSet BackupSet,
) (RestoreResult, error) {
	restore, integrityProof, verificationEvidence := target.fence.service, target.fence.proof, target.fence.evidence

	recordStep(target.Observer, RestoreStepPostgresRestore)
	recordStep(target.Observer, RestoreStepObjectStoreRestore)
	if err := restore.Restore(ctx, &vNextRestoreTarget{
		target:           target,
		extensionBackups: runner.extensionBackups,
	}, integrityProof); err != nil {
		return RestoreResult{BackupSet: backupSet}, restoreStageFailure(RestoreStepPostgresRestore, err)
	}
	recordStep(target.Observer, RestoreStepProjectionRebuild)
	if target.ReferencePacks != nil {
		if err := target.ReferencePacks.ValidateHistoricalState(ctx, target.Postgres); err != nil {
			return RestoreResult{BackupSet: backupSet}, restoreStageFailure(RestoreStepProjectionRebuild, err)
		}
	} else {
		var present bool
		if err := target.Postgres.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reference_pack_candidates)`).Scan(&present); err != nil || present {
			return RestoreResult{BackupSet: backupSet}, restoreStageFailure(RestoreStepProjectionRebuild, errors.New("reference pack restore participant is required"))
		}
	}
	projectionResult, err := runner.runVNextProjectionRebuilds(ctx, target, backupSet, verificationEvidence)
	if err != nil {
		return RestoreResult{BackupSet: backupSet, ProjectionRebuildResult: projectionResult}, restoreStageFailure(RestoreStepProjectionRebuild, err)
	}

	recordStep(target.Observer, RestoreStepConsistencyCheck)
	report, err := vNextConsistencyReport(ctx, target, verificationEvidence.stateCatalog, backupSet)
	if err != nil {
		return RestoreResult{BackupSet: backupSet}, restoreStageFailure(RestoreStepConsistencyCheck, err)
	}
	result := RestoreResult{
		BackupSet: backupSet, ConsistencyReport: report,
		ProjectionRebuildResult:    projectionResult,
		RestoreGeneration:          &RestoreGenerationEvidence{TargetGenerationID: target.TargetGenerationID, RecoveryStateCatalogSHA256: verificationEvidence.GraphRestoreArtifacts.RecoveryStateCatalogSHA256, CodecRegistrySHA256: verificationEvidence.codecRegistrySHA256},
		IntegrityManifestSHA256:    verificationEvidence.ManifestSHA256,
		RestoredObjectCount:        verificationEvidence.RestoredObjectCount,
		RecoveryStateCatalogSHA256: verificationEvidence.GraphRestoreArtifacts.RecoveryStateCatalogSHA256,
		CodecRegistrySHA256:        verificationEvidence.codecRegistrySHA256,
	}
	if err := target.fence.capability.AssertHeld(); err != nil {
		return result, err
	}
	target.fence.validated = true
	return result, nil
}

// Frozen Graph identities are admitted as metadata. Extension-derived tables
// remain absent; only independent Core providers execute during structural restore.
func (runner *RestoreRunner) runVNextProjectionRebuilds(ctx context.Context, target RestoreTarget, backupSet BackupSet, verification VNextRestoreVerificationEvidence) (restorecontract.ProjectionRebuildResult, error) {
	graphTables := make([]string, 0, len(graphrestore.RestoreGraphTableIDs()))
	for _, table := range verification.stateCatalog.Document().Tables {
		if table.OwnerID == "module.graphprojection" && table.RestoreAction == recoverystate.RebuildState {
			graphTables = append(graphTables, table.TableName)
		}
	}
	sort.Strings(graphTables)
	if !equalStringSlices(graphTables, graphrestore.RestoreGraphTableIDs()) || verification.GraphRestoreArtifacts.AlgorithmID != graphrestore.RestoreAlgorithmID {
		return restorecontract.ProjectionRebuildResult{}, fmt.Errorf("%w: Graph Projection frozen catalog mismatch", ErrInvalidBackupArtifact)
	}
	required := RequiredVNextRestoreAlgorithmIDs(verification.stateCatalog)
	for _, id := range []string{graphrestore.RestoreAlgorithmID, "workbook.restore_projections.v1"} {
		index := sort.SearchStrings(required, id)
		if index == len(required) || required[index] != id {
			return restorecontract.ProjectionRebuildResult{}, fmt.Errorf("%w: required projection identity is unavailable", ErrInvalidBackupArtifact)
		}
	}
	result, err := target.Projections.RebuildRestoreProjections(ctx, restoreProjectionRebuildRequest(target, backupSet))
	if err != nil {
		return result, err
	}
	if !result.ReadinessSatisfied() {
		return result, fmt.Errorf("%w: independent Core rebuild did not complete", ErrInvalidBackupArtifact)
	}
	for _, table := range graphTables {
		var populated bool
		if err := target.Postgres.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM "+pgx.Identifier{"public", table}.Sanitize()+")").Scan(&populated); err != nil {
			return result, err
		}
		if populated {
			return result, fmt.Errorf("%w: structural restore populated extension-derived state", ErrInvalidBackupArtifact)
		}
	}
	return result, nil
}

func equalStringSlices(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

type vNextRestoreTarget struct {
	extensionBackups *ExtensionBackupCatalog
	target           RestoreTarget
}

func (target *vNextRestoreTarget) WithAtomicRestore(
	ctx context.Context,
	stateCatalog *recoverystate.Catalog,
	run func(VNextRestoreMutation) error,
) error {
	fence := target.target.fence
	if fence == nil || fence.tx == nil || fence.closed || fence.committed || !fence.used {
		return ErrTargetIncomplete
	}
	tx := fence.tx
	if _, err := tx.Exec(ctx, `SET LOCAL session_replication_role = replica`); err != nil {
		return fmt.Errorf("disable vNext restore referential triggers: %w", err)
	}
	mutation := &vNextRestoreMutation{
		tx: tx, proofScratch: target.target.ProofScratch, objects: target.target.ObjectStore, referencePacks: target.target.ReferencePacks, exportOutputs: target.target.ExportOutputs, stateCatalog: stateCatalog,
	}
	if err := run(mutation); err != nil {
		return err
	}
	if err := validateRestoredPostgresRows(ctx, tx); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `SET LOCAL session_replication_role = origin`); err != nil {
		return fmt.Errorf("restore referential enforcement: %w", err)
	}
	if err := fence.capability.AssertHeld(); err != nil {
		return err
	}
	if err := fence.ctx.Err(); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit vNext restore: %w", err)
	}
	fence.committed = true
	return nil
}

type vNextRestoreMutation struct {
	tx             pgx.Tx
	proofScratch   *os.Root
	objects        objectstore.Store
	referencePacks ReferencePackStorage
	exportOutputs  RootObjectStorage
	stateCatalog   *recoverystate.Catalog
	insertPlans    map[string]restoreInsertPlan
}

type restoreInsertPlan struct {
	query     string
	writable  []string
	fields    []string
	generated bool
}

// Snapshot rows include stored generated values as integrity evidence. Restore
// supplies only writable columns and verifies PostgreSQL's recomputation of
// generated columns against that evidence. The plan is resolved once per table.
func (mutation *vNextRestoreMutation) insertPlan(ctx context.Context, tableName string) (restoreInsertPlan, error) {
	if plan, ok := mutation.insertPlans[tableName]; ok {
		return plan, nil
	}
	identifier := pgx.Identifier{"public", tableName}.Sanitize()
	rows, err := mutation.tx.Query(ctx, `SELECT attname,attgenerated<>'' FROM pg_catalog.pg_attribute WHERE attrelid=$1::regclass AND attnum>0 AND NOT attisdropped ORDER BY attnum`, identifier)
	if err != nil {
		return restoreInsertPlan{}, err
	}
	var plan restoreInsertPlan
	var columns []string
	for rows.Next() {
		var name string
		var generated bool
		if err := rows.Scan(&name, &generated); err != nil {
			rows.Close()
			return plan, err
		}
		plan.fields = append(plan.fields, name)
		if generated {
			plan.generated = true
			continue
		}
		plan.writable = append(plan.writable, name)
		columns = append(columns, pgx.Identifier{name}.Sanitize())
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return plan, err
	}
	if len(columns) == 0 {
		return plan, fmt.Errorf("%w: restore table has no writable columns", ErrInvalidBackupArtifact)
	}
	list := strings.Join(columns, ",")
	plan.query = fmt.Sprintf("INSERT INTO %s (%s) SELECT %s FROM jsonb_populate_record(NULL::%s, $1::jsonb)", identifier, list, list, identifier)
	if plan.generated {
		plan.query = fmt.Sprintf("WITH restored AS (INSERT INTO %s (%s) SELECT %s FROM jsonb_populate_record(NULL::%s, $1::jsonb) RETURNING *) SELECT (to_jsonb(restored)-$2::text[]) IS NOT DISTINCT FROM ($1::jsonb-$2::text[]) FROM restored", identifier, list, list, identifier)
	}
	if mutation.insertPlans == nil {
		mutation.insertPlans = map[string]restoreInsertPlan{}
	}
	mutation.insertPlans[tableName] = plan
	return plan, nil
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
		"TRUNCATE "+sanitizedTableList(tableNames)+" CASCADE",
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
	plan, err := mutation.insertPlan(ctx, tableName)
	if err != nil {
		return fmt.Errorf("resolve restore columns: %w", err)
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(row, &fields) != nil || len(fields) != len(plan.fields) {
		return fmt.Errorf("%w: restore row does not match its retained table shape", ErrInvalidBackupArtifact)
	}
	for _, name := range plan.fields {
		if _, ok := fields[name]; !ok {
			return fmt.Errorf("%w: restore row omits a required column", ErrInvalidBackupArtifact)
		}
	}
	if plan.generated {
		var valid bool
		if err := mutation.tx.QueryRow(ctx, plan.query, string(row), plan.writable).Scan(&valid); err != nil {
			return fmt.Errorf("restore vNext table %s: %w", tableName, err)
		}
		if !valid {
			return fmt.Errorf("%w: restored generated column mismatch", ErrInvalidBackupArtifact)
		}
		return nil
	}
	if _, err := mutation.tx.Exec(ctx, plan.query, string(row)); err != nil {
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
	if object.OwnerID == "module.reference_data" || object.ObjectFamilyID == "reference_packs.members" {
		if object.OwnerID != "module.reference_data" || object.ObjectFamilyID != "reference_packs.members" || mutation.referencePacks == nil {
			return errors.New("reference pack restore dispatch is unavailable")
		}
		return mutation.referencePacks.RestoreMember(ctx, object.StorageKey, object.PlaintextSHA256, object.PlaintextBytes, reader)
	}
	if object.OwnerID == "module.incidentbundles" || object.ObjectFamilyID == "incident_bundles.files" {
		if object.OwnerID != "module.incidentbundles" || object.ObjectFamilyID != "incident_bundles.files" || mutation.exportOutputs == nil {
			return errors.New("incident bundle restore dispatch is unavailable")
		}
		return mutation.exportOutputs.RestoreMember(ctx, object.StorageKey, object.PlaintextSHA256, object.PlaintextBytes, reader)
	}
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

func (mutation *vNextRestoreMutation) RunCatalogAlgorithm(ctx context.Context, algorithmID string) error {
	err := RebuildPostgresDerivedState(ctx, mutation.tx, algorithmID)
	if errors.Is(err, ErrUnsupportedPostgresRebuild) {
		// Other excluded projections were invalidated by PreparePostgresTables;
		// their owner rebuilders execute after authoritative state commits.
		return nil
	}
	return err
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

func restoreProjectionRebuildRequest(target RestoreTarget, backupSet BackupSet) restorecontract.ProjectionRebuildRequest {
	return restorecontract.ProjectionRebuildRequest{
		RestoreOperationID:     target.RestoreOperationID,
		RestoredSourceStateRef: restoreProjectionSourceStateRef(backupSet),
		RebuildScope:           restorecontract.ProjectionRebuildScopeAllActiveProviders,
		ProviderRegistryRef:    restorecontract.ProviderRegistryRefCodeBacked,
	}
}

func restoreProjectionSourceStateRef(backupSet BackupSet) string {
	return fmt.Sprintf("backup_set:%s/postgres_artifact:%s", backupSet.BackupSetID.String(), backupSet.PostgresArtifactSHA256)
}

type restoreTargetReader interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func requireEmptyRestoreTarget(ctx context.Context, target RestoreTarget, extensionBackups *ExtensionBackupCatalog, stateCatalog *recoverystate.Catalog) error {
	return requireEmptyRestoreTargetWithReader(ctx, target, target.Postgres, extensionBackups, stateCatalog)
}

func requireEmptyRestoreTargetWithReader(ctx context.Context, target RestoreTarget, db restoreTargetReader, extensionBackups *ExtensionBackupCatalog, stateCatalog *recoverystate.Catalog) error {
	if err := inspectPostgresBindingMappings(ctx, db, extensionBackups); err != nil {
		return err
	}
	tableNames, err := mutableRestoreTargetTables(ctx, db, stateCatalog)
	if err != nil {
		return err
	}

	for _, tableName := range tableNames {
		if tableName == "extension_state_metadata" {
			if err := requirePristineExtensionMetadata(ctx, db, extensionBackups); err != nil {
				return err
			}
			continue
		}
		var count int64
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", pgx.Identifier{tableName}.Sanitize())
		if err := db.QueryRow(ctx, query).Scan(&count); err != nil {
			return fmt.Errorf("inspect restore target table %s: %w", tableName, err)
		}
		if count != 0 {
			return fmt.Errorf("%w: table %s contains %d rows", ErrRestoreTargetNotEmpty, tableName, count)
		}
	}
	objects, err := target.ObjectStore.ListObjects(ctx, "")
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRestoreObjectNamespaceNotFresh, err)
	}
	if len(objects) != 0 {
		return fmt.Errorf("%w: %w", ErrRestoreObjectNamespaceNotFresh, ErrRestoreTargetNotEmpty)
	}
	if target.ExportOutputs != nil {
		if err := target.ExportOutputs.RequireEmpty(ctx); err != nil {
			return fmt.Errorf("%w: export output storage is not empty", ErrRestoreTargetNotEmpty)
		}
	}
	if target.ReferencePacks != nil {
		if err := target.ReferencePacks.RequireEmpty(ctx); err != nil {
			return fmt.Errorf("%w: Reference Pack storage is not empty", ErrRestoreTargetNotEmpty)
		}
	}
	return nil
}

func requirePristineExtensionMetadata(ctx context.Context, db restoreTargetReader, catalog *ExtensionBackupCatalog) error {
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

// Schema metadata belongs to admitted initialization, never to a restore or
// disposable reset. Validate the full catalog before selecting mutable tables.
func mutableRestoreTargetTables(ctx context.Context, db restoreTargetReader, catalog *recoverystate.Catalog) ([]string, error) {
	if err := catalog.ValidateFrozen(); err != nil {
		return nil, err
	}
	// information_schema hides tables without privileges. Coverage must also
	// reject unknown tables the Recovery role cannot access.
	rows, err := db.Query(ctx, `SELECT c.relname FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname = 'public' AND c.relkind IN ('r','p','f') ORDER BY c.relname`)
	if err != nil {
		return nil, fmt.Errorf("inspect restore target table coverage: %w", err)
	}
	defer rows.Close()
	var actual []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		actual = append(actual, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := catalog.ValidateDatabaseTableNames(actual); err != nil {
		return nil, err
	}
	var mutable []string
	for _, table := range catalog.Document().Tables {
		if table.BackupInclusion != recoverystate.InclusionSchemaMetadata && table.TableName != TargetStateTableName {
			mutable = append(mutable, table.TableName)
		}
	}
	return mutable, nil
}

// ResetRestoreVerificationTarget returns an exclusively admitted, disposable
// verification target to its migration-owned pristine state between successful
// restore proofs. It is not part of ordinary restore and must never be used to
// make a nonempty production target admissible.
func ResetRestoreVerificationTarget(ctx context.Context, capability TargetControlCapability, target RestoreTarget, generationID uuid.UUID, bindings TargetBindingDigests, catalog *ExtensionBackupCatalog, stateCatalog *recoverystate.Catalog) error {
	if capability == nil || capability.TargetDatabase() == nil || generationID == uuid.Nil || !bindings.valid() || target.ObjectStore == nil || catalog == nil {
		return ErrTargetIncomplete
	}
	if err := capability.AssertHeld(); err != nil {
		return err
	}
	work, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(capability.Context(), cancel)
	defer cancel()
	defer stop()
	ctx = work
	target.Postgres = capability.TargetDatabase()
	tx, err := beginRestoreTargetTransaction(ctx, target.Postgres, stateCatalog)
	if err != nil {
		return err
	}
	defer func() { _ = rollbackRestoreTransaction(tx) }()
	current, err := ReadTargetControl(ctx, tx)
	if err != nil {
		return err
	}
	next := TargetControl{SchemaID: TargetStateSchemaID, State: TargetResetting, Revision: 1, TargetGenerationID: targetControlText(generationID.String()), Purpose: targetControlText("restore_verification_target"), BindingDigests: &bindings}
	if current.State == TargetInitialized {
		if err := requireEmptyRestoreTargetWithReader(ctx, target, tx, catalog, stateCatalog); err != nil {
			return err
		}
	} else if !sameTargetBinding(current, next) || current.Purpose == nil || *current.Purpose != "restore_verification_target" {
		return ErrTargetIncomplete
	}
	if current.State != TargetResetting {
		current, err = advanceTargetControl(ctx, capability, current, next)
		if err != nil {
			return err
		}
	}
	tableNames, err := mutableRestoreTargetTables(ctx, tx, stateCatalog)
	if err != nil {
		return err
	}
	if len(tableNames) > 0 {
		query := "TRUNCATE " + sanitizedTableList(tableNames) + " CASCADE"
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
	if err := capability.AssertHeld(); err != nil {
		return err
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
	if target.ReferencePacks != nil {
		if err := target.ReferencePacks.ResetVerificationTarget(ctx); err != nil {
			return err
		}
	}
	if target.ExportOutputs != nil {
		if err := target.ExportOutputs.ResetVerificationTarget(ctx); err != nil {
			return err
		}
	}
	if err := requireEmptyRestoreTarget(ctx, target, catalog, stateCatalog); err != nil {
		return err
	}
	next = current
	next.State = TargetVerificationReady
	_, err = advanceTargetControl(ctx, capability, current, next)
	return err
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

// InspectPristineRestoreTarget shares the restore engine's catalog-driven admission
// with target proof issuance; migration seeds are the only permitted rows.
func InspectPristineRestoreTarget(ctx context.Context, target RestoreTarget, extensions *ExtensionBackupCatalog, catalog *recoverystate.Catalog) error {
	if target.Postgres == nil {
		return ErrTargetIncomplete
	}
	control, err := ReadTargetControl(ctx, target.Postgres)
	if err != nil {
		return err
	}
	if control.State != TargetInitialized && control.State != TargetVerificationReady {
		return ErrTargetIncomplete
	}
	return requireEmptyRestoreTarget(ctx, target, extensions, catalog)
}
