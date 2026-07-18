package operatorops

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/operatorcli"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const (
	restoreMinimumSchemaVersion       int64 = 22
	recoveryOperationAdvisoryLockKey  int64 = 401010
	restoreVerificationTargetSchemaID       = "cartulary.restore_verification_target.v1"
)

type PostgresPool interface {
	postgres.DB
	Close()
}

type ConfigLoader func(string) (config.Config, error)
type PostgresOpener func(context.Context, config.Config) (PostgresPool, error)
type ObjectStoreOpener func(context.Context, config.Config) (objectstore.Store, error)
type BackupStorageFactory func(config.Config) (recovery.BackupStorage, error)
type ProjectionRebuilderFactory func(postgres.DB) restorecontract.ProjectionRebuilder
type JournalKeyLoader func() (recovery.RecoveryEncryptionKey, error)

type Service struct {
	LoadConfig             ConfigLoader
	SetupPostgres          PostgresOpener
	SetupObjectStore       ObjectStoreOpener
	NewBackupStorage       BackupStorageFactory
	NewProjectionRebuilder ProjectionRebuilderFactory
	LoadJournalKey         JournalKeyLoader
	Now                    func() time.Time
}

var _ operatorcli.Operations = Service{}

func (service Service) BackupInspectLatest(ctx context.Context, parsed operatorcli.Command, progress operatorcli.ProgressEmitter) (operatorcli.Outcome, error) {
	progress.Emit("preflight", 0, nil)
	_, pool, backupStorage, closeFn, err := service.openSourceRuntime(ctx, parsed.SourceConfigPath)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	defer closeFn()
	progress.Emit("catalog_select", 0, nil)
	selection, err := recovery.NewBackupCatalog(recovery.NewStore(pool), backupStorage).RestoreCandidateBackupSelection(ctx, service.now())
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	progress.Emit("artifact_validate", 0, nil)
	progress.Emit("finalize", 1, operatorcli.IntPtr(1))
	return operatorcli.OutcomeForBackupSet(selection.BackupSet, "backup_attestation", "cartulary.backup_attestation.v1"), nil
}

func (service Service) BackupCreate(ctx context.Context, parsed operatorcli.Command, progress operatorcli.ProgressEmitter) (outcome operatorcli.Outcome, err error) {
	progress.Emit("preflight", 0, nil)
	cfg, pool, backupStorage, closeFn, err := service.openSourceRuntime(ctx, parsed.SourceConfigPath)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	defer closeFn()
	unlock, err := service.acquireOperationLock(ctx, pool)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	defer unlock()
	if err := service.recordRecoveryStart(ctx, pool, parsed); err != nil {
		return operatorcli.Outcome{}, err
	}
	defer func() { service.finishJournalAndAudit(ctx, pool, parsed, outcome, &err) }()

	consistencyPointAt := service.now()
	backupSetID := uuid.New()
	progress.Emit("postgres_backup", 0, nil)
	postgresArtifact, err := recovery.CapturePostgresSnapshotArtifact(ctx, pool)
	if err != nil {
		return operatorcli.OutcomeForCandidate(backupSetID, consistencyPointAt), fmt.Errorf("capture postgres artifact: %w", err)
	}
	blobIndex, err := loadBackupObjectBlobIndex(ctx, pool)
	if err != nil {
		return operatorcli.OutcomeForCandidate(backupSetID, consistencyPointAt), err
	}
	objectBucket, err := backupObjectStoreBucket(cfg)
	if err != nil {
		return operatorcli.OutcomeForCandidate(backupSetID, consistencyPointAt), err
	}
	objectStore, err := service.setupObjectStore(ctx, cfg)
	if err != nil {
		return operatorcli.OutcomeForCandidate(backupSetID, consistencyPointAt), fmt.Errorf("open object store: %w", err)
	}
	defer func() { _ = objectStore.Close() }()
	progress.Emit("object_backup", 0, nil)
	objectArtifacts, err := recovery.CaptureSeaweedFSS3ObjectStoreBackupArtifacts(ctx, objectStore, recovery.ObjectStoreBackupCaptureParams{
		BackupSetID:               backupSetID,
		ConsistencyPointAt:        consistencyPointAt,
		Bucket:                    objectBucket,
		BlobObjectIDsByStorageRef: blobIndex,
	})
	if err != nil {
		return operatorcli.OutcomeForCandidate(backupSetID, consistencyPointAt), fmt.Errorf("capture object-store artifacts: %w", err)
	}
	progress.Emit("attestation_write", 0, nil)
	backupSet, err := recovery.NewCaptureService(recovery.NewStore(pool), backupStorage).CaptureBackupSet(ctx, recovery.CaptureBackupSetParams{
		BackupSetID:                       backupSetID,
		ConsistencyPointAt:                consistencyPointAt,
		CreatedAt:                         consistencyPointAt,
		RetainedUntil:                     consistencyPointAt.Add(recovery.MinimumRetentionDuration),
		PostgresArtifact:                  recovery.BackupArtifact{Body: postgresArtifact, ContentType: "application/json"},
		ObjectStoreArtifact:               recovery.BackupArtifact{Body: objectArtifacts.SnapshotBody, ContentType: "application/json"},
		ObjectStoreBackupManifestArtifact: recovery.BackupArtifact{Body: objectArtifacts.ManifestBody, ContentType: "application/json"},
		ObjectStoreBackupSummaryArtifact:  recovery.BackupArtifact{Body: objectArtifacts.SummaryBody, ContentType: "application/json"},
	})
	if err != nil {
		return operatorcli.OutcomeForCandidate(backupSetID, consistencyPointAt), fmt.Errorf("capture backup set: %w", err)
	}
	if err := recovery.NewBackupCatalog(recovery.NewStore(pool), backupStorage).VerifyBackupSetDurability(ctx, backupSet); err != nil {
		return operatorcli.OutcomeForBackupSet(backupSet, "backup_attestation", "cartulary.backup_attestation.v1"), fmt.Errorf("verify captured backup durability: %w", err)
	}
	progress.Emit("journal_write", 0, nil)
	progress.Emit("finalize", 1, operatorcli.IntPtr(1))
	return operatorcli.OutcomeForBackupSet(backupSet, "backup_attestation", "cartulary.backup_attestation.v1"), nil
}

func (service Service) RestoreLatest(ctx context.Context, parsed operatorcli.Command, progress operatorcli.ProgressEmitter) (outcome operatorcli.Outcome, err error) {
	progress.Emit("preflight", 0, nil)
	sourceCfg, targetCfg, sourcePool, targetPool, sourceObjectStore, targetObjectStore, backupStorage, err := service.openRestoreRuntime(ctx, parsed)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	defer sourcePool.Close()
	defer targetPool.Close()
	defer func() { _ = sourceObjectStore.Close() }()
	defer func() { _ = targetObjectStore.Close() }()
	unlock, err := service.acquireOperationLock(ctx, sourcePool)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	defer unlock()
	if err := service.recordRecoveryStart(ctx, sourcePool, parsed); err != nil {
		return operatorcli.Outcome{}, err
	}
	defer func() { service.finishJournalAndAudit(ctx, sourcePool, parsed, outcome, &err) }()

	sourceStore := recovery.NewStore(sourcePool)
	backupSet, err := recovery.NewBackupCatalog(sourceStore, backupStorage).RestoreCandidateBackup(ctx, service.now())
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	confirmed, _ := uuid.Parse(parsed.ConfirmBackupSetID)
	if backupSet.BackupSetID != confirmed {
		return operatorcli.OutcomeForStoredBackupSet(backupSet), fmt.Errorf("confirmed backup_set_id does not match latest retained backup: %w", operatorcli.ErrConfirmationMismatch)
	}
	if err := service.preflightRestoreTarget(ctx, parsed.SourceConfigPath, parsed.TargetConfigPath, sourceCfg, targetCfg, targetPool, targetObjectStore); err != nil {
		return operatorcli.OutcomeForStoredBackupSet(backupSet), err
	}
	target, err := service.restoreTarget(targetPool, targetObjectStore)
	if err != nil {
		return operatorcli.OutcomeForStoredBackupSet(backupSet), err
	}
	progress.Emit("postgres_restore", 0, nil)
	progress.Emit("object_restore", 0, nil)
	progress.Emit("projection_rebuild", 0, nil)
	progress.Emit("invariant_check", 0, nil)
	result, err := recovery.NewRestoreRunner(sourceStore, backupStorage).RestoreBackupSet(ctx, target, backupSet)
	if err != nil {
		return operatorcli.OutcomeForStoredBackupSet(backupSet), err
	}
	progress.Emit("journal_write", 0, nil)
	progress.Emit("finalize", 1, operatorcli.IntPtr(1))
	return operatorcli.OutcomeForBackupSet(result.BackupSet, "restore_operation", "cartulary.restore_operation.v1"), nil
}

func (service Service) RestoreVerifyLatest(ctx context.Context, parsed operatorcli.Command, progress operatorcli.ProgressEmitter) (outcome operatorcli.Outcome, err error) {
	progress.Emit("preflight", 0, nil)
	sourceCfg, targetCfg, sourcePool, targetPool, sourceObjectStore, targetObjectStore, backupStorage, err := service.openRestoreRuntime(ctx, parsed)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	defer sourcePool.Close()
	defer targetPool.Close()
	defer func() { _ = sourceObjectStore.Close() }()
	defer func() { _ = targetObjectStore.Close() }()
	unlock, err := service.acquireOperationLock(ctx, sourcePool)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	defer unlock()
	if err := service.recordRecoveryStart(ctx, sourcePool, parsed); err != nil {
		return operatorcli.Outcome{}, err
	}
	defer func() { service.finishJournalAndAudit(ctx, sourcePool, parsed, outcome, &err) }()
	if err := service.preflightRestoreVerificationTarget(ctx, parsed.SourceConfigPath, parsed.TargetConfigPath, sourceCfg, targetCfg, targetPool, targetObjectStore); err != nil {
		return operatorcli.Outcome{}, err
	}
	basis, err := restoreVerificationBasisForConfig(sourceCfg)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	target, err := service.restoreVerificationTarget(targetPool, targetObjectStore)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	progress.Emit("postgres_restore", 0, nil)
	progress.Emit("object_restore", 0, nil)
	progress.Emit("projection_rebuild", 0, nil)
	progress.Emit("invariant_check", 0, nil)
	progress.Emit("workbook_probe", 0, nil)
	sourceStore := recovery.NewStore(sourcePool)
	verify := recovery.NewRestoreVerificationService(sourceStore, recovery.NewRestoreRunner(sourceStore, backupStorage))
	result, err := verify.VerifyLatestSuccessfulRetained(ctx, target, service.now(), basis)
	outcome = operatorcli.OutcomeForStoredBackupSet(result.BackupSet)
	if result.Run.RestoreVerificationRunID != uuid.Nil {
		outcome.ArtifactRefs = append(outcome.ArtifactRefs, operatorcli.ArtifactRefFor("restore_verification", recovery.RestoreVerificationArtifactSchemaID, "restore_verification:"+result.Run.RestoreVerificationRunID.String(), outcome.BackupSetID))
	}
	if err != nil {
		return outcome, err
	}
	progress.Emit("attestation_update", 0, nil)
	progress.Emit("journal_write", 0, nil)
	progress.Emit("finalize", 1, operatorcli.IntPtr(1))
	return outcome, nil
}

func (service Service) RestoreVerifyDue(ctx context.Context, parsed operatorcli.Command, progress operatorcli.ProgressEmitter) (outcome operatorcli.Outcome, err error) {
	progress.Emit("preflight", 0, nil)
	sourceCfg, targetCfg, sourcePool, targetPool, sourceObjectStore, targetObjectStore, backupStorage, err := service.openRestoreRuntime(ctx, parsed)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	defer sourcePool.Close()
	defer targetPool.Close()
	defer func() { _ = sourceObjectStore.Close() }()
	defer func() { _ = targetObjectStore.Close() }()
	unlock, err := service.acquireOperationLock(ctx, sourcePool)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	defer unlock()
	if err := service.recordRecoveryStart(ctx, sourcePool, parsed); err != nil {
		return operatorcli.Outcome{}, err
	}
	defer func() { service.finishJournalAndAudit(ctx, sourcePool, parsed, outcome, &err) }()
	if err := service.preflightRestoreVerificationTarget(ctx, parsed.SourceConfigPath, parsed.TargetConfigPath, sourceCfg, targetCfg, targetPool, targetObjectStore); err != nil {
		return operatorcli.Outcome{}, err
	}
	basis, err := restoreVerificationBasisForConfig(sourceCfg)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	sourceStore := recovery.NewStore(sourcePool)
	due, err := recovery.NewBackupCatalog(sourceStore, backupStorage).ListBackupsDueForRestoreVerification(ctx, service.now(), basis)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	if len(due) == 0 {
		return operatorcli.Outcome{ArtifactRefs: []operatorcli.ArtifactRef{}, Result: "no_op"}, nil
	}
	verify := recovery.NewRestoreVerificationService(sourceStore, recovery.NewRestoreRunner(sourceStore, backupStorage))
	for _, backupSet := range due {
		target, err := service.restoreVerificationTarget(targetPool, targetObjectStore)
		if err != nil {
			return outcome, err
		}
		progress.Emit("postgres_restore", 0, nil)
		progress.Emit("object_restore", 0, nil)
		progress.Emit("projection_rebuild", 0, nil)
		progress.Emit("invariant_check", 0, nil)
		progress.Emit("workbook_probe", 0, nil)
		result, verifyErr := verify.VerifyBackupSet(ctx, target, backupSet, basis)
		if outcome.BackupSetID == nil {
			outcome = operatorcli.OutcomeForStoredBackupSet(result.BackupSet)
		}
		if result.Run.RestoreVerificationRunID != uuid.Nil {
			backupSetID := result.BackupSet.BackupSetID.String()
			outcome.ArtifactRefs = append(outcome.ArtifactRefs, operatorcli.ArtifactRefFor("restore_verification", recovery.RestoreVerificationArtifactSchemaID, "restore_verification:"+result.Run.RestoreVerificationRunID.String(), &backupSetID))
		}
		if verifyErr != nil {
			return outcome, verifyErr
		}
		progress.Emit("attestation_update", 0, nil)
	}
	progress.Emit("journal_write", 0, nil)
	progress.Emit("finalize", len(due), operatorcli.IntPtr(len(due)))
	return outcome, nil
}

func (service Service) openSourceRuntime(ctx context.Context, sourceConfigPath string) (config.Config, PostgresPool, recovery.BackupStorage, func(), error) {
	cfg, err := service.loadConfig(sourceConfigPath)
	if err != nil {
		return config.Config{}, nil, nil, nil, fmt.Errorf("load config: %w", err)
	}
	pool, err := service.setupPostgres(ctx, cfg)
	if err != nil {
		return config.Config{}, nil, nil, nil, fmt.Errorf("open postgres: %w", err)
	}
	backupStorage, err := service.newBackupStorage(cfg)
	if err != nil {
		pool.Close()
		return config.Config{}, nil, nil, nil, fmt.Errorf("open backup storage: %w", err)
	}
	return cfg, pool, backupStorage, func() { pool.Close() }, nil
}

func (service Service) openRestoreRuntime(ctx context.Context, parsed operatorcli.Command) (config.Config, config.Config, PostgresPool, PostgresPool, objectstore.Store, objectstore.Store, recovery.BackupStorage, error) {
	sourceCfg, err := service.loadConfig(parsed.SourceConfigPath)
	if err != nil {
		return config.Config{}, config.Config{}, nil, nil, nil, nil, nil, fmt.Errorf("load source config: %w", err)
	}
	targetCfg, err := service.loadConfig(parsed.TargetConfigPath)
	if err != nil {
		return config.Config{}, config.Config{}, nil, nil, nil, nil, nil, fmt.Errorf("load target config: %w", err)
	}
	sourcePool, err := service.setupPostgres(ctx, sourceCfg)
	if err != nil {
		return config.Config{}, config.Config{}, nil, nil, nil, nil, nil, fmt.Errorf("open source postgres: %w", err)
	}
	targetPool, err := service.setupPostgres(ctx, targetCfg)
	if err != nil {
		sourcePool.Close()
		return config.Config{}, config.Config{}, nil, nil, nil, nil, nil, fmt.Errorf("open target postgres: %w", err)
	}
	sourceObjectStore, err := service.setupObjectStore(ctx, sourceCfg)
	if err != nil {
		sourcePool.Close()
		targetPool.Close()
		return config.Config{}, config.Config{}, nil, nil, nil, nil, nil, fmt.Errorf("open source object store: %w", err)
	}
	targetObjectStore, err := service.setupObjectStore(ctx, targetCfg)
	if err != nil {
		sourcePool.Close()
		targetPool.Close()
		_ = sourceObjectStore.Close()
		return config.Config{}, config.Config{}, nil, nil, nil, nil, nil, fmt.Errorf("open target object store: %w", err)
	}
	backupStorage, err := service.newBackupStorage(sourceCfg)
	if err != nil {
		sourcePool.Close()
		targetPool.Close()
		_ = sourceObjectStore.Close()
		_ = targetObjectStore.Close()
		return config.Config{}, config.Config{}, nil, nil, nil, nil, nil, fmt.Errorf("open source backup storage: %w", err)
	}
	return sourceCfg, targetCfg, sourcePool, targetPool, sourceObjectStore, targetObjectStore, backupStorage, nil
}

func (service Service) restoreTarget(targetPool postgres.DB, targetObjectStore objectstore.Store) (recovery.RestoreTarget, error) {
	rebuilder, err := service.newProjectionRebuilder(targetPool)
	if err != nil {
		return recovery.RestoreTarget{}, err
	}
	return recovery.RestoreTarget{
		Postgres:    targetPool,
		ObjectStore: targetObjectStore,
		Projections: rebuilder,
	}, nil
}

func (service Service) restoreVerificationTarget(targetPool postgres.DB, targetObjectStore objectstore.Store) (recovery.RestoreVerificationTarget, error) {
	target, err := service.restoreTarget(targetPool, targetObjectStore)
	if err != nil {
		return recovery.RestoreVerificationTarget{}, err
	}
	return recovery.RestoreVerificationTarget{
		RestoreTarget: target,
		Probe:         recovery.RestoreVerificationWorkbookProbe{Postgres: targetPool},
	}, nil
}

func (service Service) preflightRestoreTarget(ctx context.Context, sourceConfigPath string, targetConfigPath string, sourceCfg config.Config, targetCfg config.Config, targetPool postgres.DB, targetObjectStore objectstore.Store) error {
	if sameConfigPath(sourceConfigPath, targetConfigPath) {
		return errors.New("restore target preflight failed: source-config and target-config must be different files")
	}
	sourcePostgres, err := postgres.ResolveSettings(sourceCfg, nil)
	if err != nil {
		return fmt.Errorf("resolve source postgres settings: %w", err)
	}
	targetPostgres, err := postgres.ResolveSettings(targetCfg, nil)
	if err != nil {
		return fmt.Errorf("resolve target postgres settings: %w", err)
	}
	if strings.TrimSpace(sourcePostgres.DSN) == strings.TrimSpace(targetPostgres.DSN) {
		return errors.New("restore target preflight failed: source and target postgres DSNs must differ")
	}
	sourceObject, err := objectstore.ResolveSettings(sourceCfg, nil)
	if err != nil {
		return fmt.Errorf("resolve source object-store settings: %w", err)
	}
	targetObject, err := objectstore.ResolveSettings(targetCfg, nil)
	if err != nil {
		return fmt.Errorf("resolve target object-store settings: %w", err)
	}
	if objectStoreBindingID(sourceObject) == objectStoreBindingID(targetObject) {
		return errors.New("restore target preflight failed: source and target object stores must differ")
	}
	var schemaVersion int64
	if err := targetPool.QueryRow(ctx, `SELECT COALESCE(MAX(version_id), 0)::bigint FROM goose_db_version WHERE is_applied = true`).Scan(&schemaVersion); err != nil {
		return fmt.Errorf("restore target preflight failed: inspect target schema version: %w", err)
	}
	if schemaVersion < restoreMinimumSchemaVersion {
		return fmt.Errorf("restore target preflight failed: target schema version %d is below required %d", schemaVersion, restoreMinimumSchemaVersion)
	}
	var rowCount int64
	if err := targetPool.QueryRow(ctx, `
SELECT
    (SELECT COUNT(*) FROM incidents)
  + (SELECT COUNT(*) FROM records)
  + (SELECT COUNT(*) FROM object_blobs)
`).Scan(&rowCount); err != nil {
		return fmt.Errorf("restore target preflight failed: inspect target data rows: %w", err)
	}
	if rowCount != 0 {
		return fmt.Errorf("restore target preflight failed: target database is not empty (%d incident/record/blob rows)", rowCount)
	}
	objects, err := targetObjectStore.ListObjects(ctx, "")
	if err != nil {
		return fmt.Errorf("restore target preflight failed: inspect target object store: %w", err)
	}
	if len(objects) != 0 {
		return fmt.Errorf("restore target preflight failed: target object store is not empty (%d objects)", len(objects))
	}
	return nil
}

func (service Service) preflightRestoreVerificationTarget(ctx context.Context, sourceConfigPath string, targetConfigPath string, sourceCfg config.Config, targetCfg config.Config, targetPool postgres.DB, targetObjectStore objectstore.Store) error {
	if sameConfigPath(sourceConfigPath, targetConfigPath) {
		return errors.New("restore verification target preflight failed: source-config and target-config must be different files")
	}
	sourcePostgres, err := postgres.ResolveSettings(sourceCfg, nil)
	if err != nil {
		return fmt.Errorf("resolve source postgres settings: %w", err)
	}
	targetPostgres, err := postgres.ResolveSettings(targetCfg, nil)
	if err != nil {
		return fmt.Errorf("resolve target postgres settings: %w", err)
	}
	if strings.TrimSpace(sourcePostgres.DSN) == strings.TrimSpace(targetPostgres.DSN) {
		return errors.New("restore verification target preflight failed: source and target postgres DSNs must differ")
	}
	sourceObject, err := objectstore.ResolveSettings(sourceCfg, nil)
	if err != nil {
		return fmt.Errorf("resolve source object-store settings: %w", err)
	}
	targetObject, err := objectstore.ResolveSettings(targetCfg, nil)
	if err != nil {
		return fmt.Errorf("resolve target object-store settings: %w", err)
	}
	if objectStoreBindingID(sourceObject) == objectStoreBindingID(targetObject) {
		return errors.New("restore verification target preflight failed: source and target object stores must differ")
	}
	if err := requireRestoreVerificationTargetMarker(targetCfg); err != nil {
		return err
	}
	var schemaVersion int64
	if err := targetPool.QueryRow(ctx, `SELECT COALESCE(MAX(version_id), 0)::bigint FROM goose_db_version WHERE is_applied = true`).Scan(&schemaVersion); err != nil {
		return fmt.Errorf("restore verification target preflight failed: inspect target schema version: %w", err)
	}
	if schemaVersion < restoreMinimumSchemaVersion {
		return fmt.Errorf("restore verification target preflight failed: target schema version %d is below required %d", schemaVersion, restoreMinimumSchemaVersion)
	}
	if _, err := targetObjectStore.ListObjects(ctx, ""); err != nil {
		return fmt.Errorf("restore verification target preflight failed: inspect target object store: %w", err)
	}
	return nil
}

func RestoreVerificationTargetMarkerPath(cfg config.Config) (string, error) {
	if cfg.Roots.BackupStorage.BindingKind != "filesystem_root" {
		return "", errors.New("restore verification target marker requires filesystem backup storage")
	}
	if strings.TrimSpace(cfg.Roots.BackupStorage.Path) == "" {
		return "", errors.New("restore verification target marker requires roots.backup_storage.path")
	}
	return filepath.Join(cfg.Roots.BackupStorage.Path, "restore-verification-target.json"), nil
}

func requireRestoreVerificationTargetMarker(cfg config.Config) error {
	markerPath, err := RestoreVerificationTargetMarkerPath(cfg)
	if err != nil {
		return fmt.Errorf("restore verification target preflight failed: %w", err)
	}
	body, err := os.ReadFile(markerPath) // #nosec G304 -- marker path is derived from validated target backup storage root.
	if err != nil {
		return fmt.Errorf("restore verification target preflight failed: read target marker: %w", err)
	}
	var marker struct {
		SchemaID string `json:"schema_id"`
		Purpose  string `json:"purpose"`
	}
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&marker); err != nil {
		return fmt.Errorf("restore verification target preflight failed: decode target marker: %w", err)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return errors.New("restore verification target preflight failed: decode target marker: trailing JSON content")
	}
	if marker.SchemaID != restoreVerificationTargetSchemaID || marker.Purpose != "restore_verification_target" {
		return errors.New("restore verification target preflight failed: target marker is not a restore-verification target")
	}
	return nil
}

func (service Service) acquireOperationLock(ctx context.Context, pool PostgresPool) (func(), error) {
	locked, err := tryRecoveryOperationAdvisoryLock(ctx, pool)
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, operatorcli.ErrOperationLockUnavailable
	}
	return func() { _ = unlockRecoveryOperationAdvisoryLock(context.Background(), pool) }, nil
}

func tryRecoveryOperationAdvisoryLock(ctx context.Context, pool postgres.DB) (bool, error) {
	var locked bool
	if err := pool.QueryRow(ctx, `SELECT pg_try_advisory_lock($1)`, recoveryOperationAdvisoryLockKey).Scan(&locked); err != nil {
		return false, fmt.Errorf("acquire restore verification advisory lock: %w", err)
	}
	return locked, nil
}

func unlockRecoveryOperationAdvisoryLock(ctx context.Context, pool postgres.DB) error {
	var unlocked bool
	if err := pool.QueryRow(ctx, `SELECT pg_advisory_unlock($1)`, recoveryOperationAdvisoryLockKey).Scan(&unlocked); err != nil {
		return fmt.Errorf("release restore verification advisory lock: %w", err)
	}
	return nil
}

func (service Service) recordRecoveryStart(ctx context.Context, pool PostgresPool, parsed operatorcli.Command) error {
	return service.journalStore(pool).Append(ctx, operatorcli.JournalRecord{
		OperationID: parsed.OperationID,
		Operation:   parsed.Operation,
		Result:      "started",
		Summary: map[string]any{
			"source_config_supplied": parsed.SourceConfigPath != "",
			"target_config_supplied": parsed.TargetConfigPath != "",
		},
	})
}

func (service Service) finishJournalAndAudit(ctx context.Context, pool PostgresPool, parsed operatorcli.Command, outcome operatorcli.Outcome, operationErr *error) {
	result := outcome.Result
	if result == "" {
		result = "succeeded"
	}
	var errorCode string
	var reasonCode string
	if operationErr != nil && *operationErr != nil {
		result = "failed"
		mapped, _ := operatorcli.MapError(parsed.Operation, *operationErr)
		errorCode = mapped.Code
		reasonCode = mapped.ReasonCode
	}
	record := operatorcli.JournalRecord{
		OperationID: parsed.OperationID,
		Operation:   parsed.Operation,
		Result:      result,
		BackupSetID: outcome.BackupSetID,
		ErrorCode:   errorCode,
		ReasonCode:  reasonCode,
		Summary: map[string]any{
			"artifact_ref_count": len(outcome.ArtifactRefs),
			"has_backup_set_id":  outcome.BackupSetID != nil,
		},
	}
	store := service.journalStore(pool)
	if err := store.Append(ctx, record); err != nil {
		mergeOperationError(operationErr, err)
		return
	}
	if err := store.AppendAuditSummary(ctx, record); err != nil {
		mergeOperationError(operationErr, err)
	}
}

func (service Service) journalStore(pool PostgresPool) operatorcli.JournalStore {
	return operatorcli.JournalStore{
		DB:      pool,
		LoadKey: service.LoadJournalKey,
		Now:     service.now,
	}
}

func loadBackupObjectBlobIndex(ctx context.Context, db postgres.DB) (map[string]uuid.UUID, error) {
	rows, err := db.Query(ctx, `
SELECT storage_key, object_blob_id
FROM object_blobs
WHERE storage_key IS NOT NULL AND storage_key <> ''
ORDER BY storage_key ASC
`)
	if err != nil {
		return nil, fmt.Errorf("list object blobs for backup manifest: %w", err)
	}
	defer rows.Close()
	index := make(map[string]uuid.UUID)
	for rows.Next() {
		var storageKey string
		var objectBlobID uuid.UUID
		if err := rows.Scan(&storageKey, &objectBlobID); err != nil {
			return nil, fmt.Errorf("scan object blob for backup manifest: %w", err)
		}
		index[storageKey] = objectBlobID
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate object blobs for backup manifest: %w", err)
	}
	return index, nil
}

func backupObjectStoreBucket(cfg config.Config) (string, error) {
	settings, err := objectstore.ResolveSettings(cfg, nil)
	if err != nil {
		return "", fmt.Errorf("resolve object-store settings for backup create: %w", err)
	}
	switch settings.BindingKind {
	case "managed_service":
		if strings.TrimSpace(settings.Bucket) == "" {
			return "", errors.New("backup create requires configured object-store bucket")
		}
		return settings.Bucket, nil
	case "filesystem_root":
		if strings.TrimSpace(settings.RootPath) == "" {
			return "", errors.New("backup create requires configured filesystem object-store root")
		}
		return "filesystem-root:" + filepath.Clean(settings.RootPath), nil
	default:
		return "", fmt.Errorf("backup create unsupported object-store binding kind %q", settings.BindingKind)
	}
}

func restoreVerificationBasisForConfig(cfg config.Config) (string, error) {
	return recovery.RestoreVerificationBasisSHA256(map[string]string{
		"backup_mechanism":         "cartulary.backup.filesystem_snapshot.v1",
		"database_storage_binding": rootBindingBasis(cfg.Roots.DatabaseStorage),
		"object_storage_binding":   rootBindingBasis(cfg.Roots.ObjectStorage),
		"backup_storage_binding":   rootBindingBasis(cfg.Roots.BackupStorage),
	})
}

func rootBindingBasis(binding config.RootBinding) string {
	switch binding.BindingKind {
	case "filesystem_root":
		return "filesystem_root:" + filepath.Clean(binding.Path)
	case "managed_service":
		return "managed_service:" + strings.TrimSpace(binding.ServiceRef)
	default:
		return strings.TrimSpace(binding.BindingKind)
	}
}

func sameConfigPath(sourcePath string, targetPath string) bool {
	sourceAbs, sourceErr := filepath.Abs(strings.TrimSpace(sourcePath))
	targetAbs, targetErr := filepath.Abs(strings.TrimSpace(targetPath))
	if sourceErr == nil && targetErr == nil {
		return filepath.Clean(sourceAbs) == filepath.Clean(targetAbs)
	}
	return filepath.Clean(strings.TrimSpace(sourcePath)) == filepath.Clean(strings.TrimSpace(targetPath))
}

func objectStoreBindingID(settings objectstore.Settings) string {
	switch settings.BindingKind {
	case "filesystem_root":
		return "filesystem_root:" + filepath.Clean(settings.RootPath)
	case "managed_service":
		return fmt.Sprintf("managed_service:%s:%t:%s", settings.Endpoint, settings.Secure, settings.Bucket)
	default:
		return settings.BindingKind
	}
}

func mergeOperationError(operationErr *error, err error) {
	if operationErr == nil || err == nil {
		return
	}
	if *operationErr == nil {
		*operationErr = err
		return
	}
	*operationErr = fmt.Errorf("%w; additionally %v", *operationErr, err)
}

func (service Service) loadConfig(path string) (config.Config, error) {
	if service.LoadConfig == nil {
		return config.Config{}, errors.New("operator recovery requires config loader")
	}
	return service.LoadConfig(path)
}

func (service Service) setupPostgres(ctx context.Context, cfg config.Config) (PostgresPool, error) {
	if service.SetupPostgres == nil {
		return nil, errors.New("operator recovery requires postgres opener")
	}
	return service.SetupPostgres(ctx, cfg)
}

func (service Service) setupObjectStore(ctx context.Context, cfg config.Config) (objectstore.Store, error) {
	if service.SetupObjectStore == nil {
		return nil, errors.New("operator recovery requires object-store opener")
	}
	return service.SetupObjectStore(ctx, cfg)
}

func (service Service) newBackupStorage(cfg config.Config) (recovery.BackupStorage, error) {
	if service.NewBackupStorage == nil {
		return nil, errors.New("operator recovery requires backup-storage opener")
	}
	return service.NewBackupStorage(cfg)
}

func (service Service) newProjectionRebuilder(db postgres.DB) (restorecontract.ProjectionRebuilder, error) {
	if service.NewProjectionRebuilder == nil {
		return nil, errors.New("operator recovery requires projection rebuilder")
	}
	rebuilder := service.NewProjectionRebuilder(db)
	if rebuilder == nil {
		return nil, errors.New("operator recovery requires projection rebuilder")
	}
	return rebuilder, nil
}

func (service Service) now() time.Time {
	if service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}
