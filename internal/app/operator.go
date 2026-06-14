package app

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const (
	OperatorBackupCaptureResultSchemaID              = "cartulary.operator.backup_capture_result.v1"
	BackupCaptureQuiescenceProofSchemaID             = "cartulary.backup_capture_quiescence_proof.v1"
	BackupMetadataInspectionSchemaID                 = "cartulary.operator.backup_metadata.v1"
	OperatorRestoreResultSchemaID                    = "cartulary.operator.restore_result.v1"
	OperatorRestoreVerificationSchemaID              = "cartulary.operator.restore_verification_result.v1"
	OperatorRestoreVerificationDueSchemaID           = "cartulary.operator.restore_verification_due_result.v1"
	OperatorObjectStoreInitResultSchemaID            = "cartulary.operator.object_store_init_result.v1"
	OperatorObjectStoreMigrationResultSchemaID       = "cartulary.operator.object_store_migration_result.v1"
	RestoreVerificationTargetMarkerSchemaID          = "cartulary.restore_verification_target.v1"
	phase10RestoreMinimumSchemaVersion         int64 = 22
	restoreVerificationAdvisoryLockKey         int64 = 401010
)

type operatorPostgresPool interface {
	postgres.DB
	Close()
}

type operatorRunner struct {
	stdout                  io.Writer
	stderr                  io.Writer
	loadConfig              func(string) (config.Config, error)
	setupPostgres           func(context.Context, config.Config) (operatorPostgresPool, error)
	setupObjectStore        func(context.Context, config.Config) (objectstore.Store, error)
	ensureObjectStoreBucket func(context.Context, config.Config) (objectstore.EnsureBucketResult, error)
	newBackupStorage        func(config.Config) (recovery.BackupStorage, error)
	now                     func() time.Time
}

type operatorCLIResult struct {
	stop               bool
	exitCode           int
	command            string
	email              string
	asOf               time.Time
	sourceConfigPath   string
	targetConfigPath   string
	captureBackupSetID uuid.UUID
	confirmBackupSetID uuid.UUID
	runID              uuid.UUID
	artifactsDir       string
	quiescenceProof    string
	sourceBackend      string
	targetBackend      string
}

type BackupMetadataInspection struct {
	SchemaID                              string                       `json:"schema_id"`
	BackupSetID                           string                       `json:"backup_set_id"`
	ConsistencyPointAt                    time.Time                    `json:"consistency_point_at"`
	PostgresRestoreAnchor                 string                       `json:"postgres_restore_anchor"`
	ObjectStoreRestoreAnchor              string                       `json:"object_store_restore_anchor"`
	PostgresArtifactKey                   string                       `json:"postgres_artifact_key"`
	PostgresArtifactSHA256                string                       `json:"postgres_artifact_sha256"`
	PostgresArtifactSizeBytes             int64                        `json:"postgres_artifact_size_bytes"`
	ObjectStoreArtifactKey                string                       `json:"object_store_artifact_key"`
	ObjectStoreArtifactSHA256             string                       `json:"object_store_artifact_sha256"`
	ObjectStoreArtifactSizeBytes          int64                        `json:"object_store_artifact_size_bytes"`
	IntegrityManifestKey                  string                       `json:"integrity_manifest_key"`
	IntegrityManifestSHA256               string                       `json:"integrity_manifest_sha256"`
	IntegrityManifestSizeBytes            int64                        `json:"integrity_manifest_size_bytes"`
	CreatedAt                             time.Time                    `json:"created_at"`
	RetainedUntil                         time.Time                    `json:"retained_until"`
	PostgresRestoreAnchorRetainedUntil    time.Time                    `json:"postgres_restore_anchor_retained_until"`
	ObjectStoreRestoreAnchorRetainedUntil time.Time                    `json:"object_store_restore_anchor_retained_until"`
	VerificationState                     string                       `json:"verification_state"`
	LastVerifiedRestoreAt                 *time.Time                   `json:"last_verified_restore_at"`
	LastVerificationBasisSHA256           string                       `json:"last_verification_basis_sha256,omitempty"`
	DurabilityDiagnostics                 []BackupDurabilityDiagnostic `json:"durability_diagnostics,omitempty"`
}

type OperatorBackupCaptureResult struct {
	SchemaID                              string                                `json:"schema_id"`
	BackupSetID                           string                                `json:"backup_set_id"`
	ConsistencyPointAt                    time.Time                             `json:"consistency_point_at"`
	CreatedAt                             time.Time                             `json:"created_at"`
	RetainedUntil                         time.Time                             `json:"retained_until"`
	PostgresRestoreAnchor                 string                                `json:"postgres_restore_anchor"`
	ObjectStoreRestoreAnchor              string                                `json:"object_store_restore_anchor"`
	PostgresArtifactKey                   string                                `json:"postgres_artifact_key"`
	PostgresArtifactSHA256                string                                `json:"postgres_artifact_sha256"`
	PostgresArtifactSizeBytes             int64                                 `json:"postgres_artifact_size_bytes"`
	ObjectStoreArtifactKey                string                                `json:"object_store_artifact_key"`
	ObjectStoreArtifactSHA256             string                                `json:"object_store_artifact_sha256"`
	ObjectStoreArtifactSizeBytes          int64                                 `json:"object_store_artifact_size_bytes"`
	IntegrityManifestKey                  string                                `json:"integrity_manifest_key"`
	IntegrityManifestSHA256               string                                `json:"integrity_manifest_sha256"`
	IntegrityManifestSizeBytes            int64                                 `json:"integrity_manifest_size_bytes"`
	PostgresRestoreAnchorRetainedUntil    time.Time                             `json:"postgres_restore_anchor_retained_until"`
	ObjectStoreRestoreAnchorRetainedUntil time.Time                             `json:"object_store_restore_anchor_retained_until"`
	VerificationState                     string                                `json:"verification_state"`
	LastVerifiedRestoreAt                 *time.Time                            `json:"last_verified_restore_at"`
	ObjectStoreBackupSummary              recovery.ObjectStoreBackupSummary     `json:"object_store_backup_summary"`
	StorageEncryption                     recovery.BackupStorageEncryptionProof `json:"storage_encryption"`
	QuiescenceProof                       OperatorBackupCaptureQuiescenceProof  `json:"quiescence_proof"`
}

type OperatorBackupCaptureQuiescenceProof struct {
	SchemaID                string    `json:"schema_id"`
	ProofKind               string    `json:"proof_kind"`
	CheckedAt               time.Time `json:"checked_at"`
	ProcessState            string    `json:"process_state"`
	HTTPListenerClosed      bool      `json:"http_listener_closed"`
	WebSocketListenerClosed bool      `json:"websocket_listener_closed"`
}

type BackupDurabilityDiagnostic struct {
	BackupSetID        string    `json:"backup_set_id"`
	ConsistencyPointAt time.Time `json:"consistency_point_at"`
	Code               string    `json:"code"`
}

type OperatorRestoreResult struct {
	SchemaID           string                            `json:"schema_id"`
	BackupSetID        string                            `json:"backup_set_id"`
	ConsistencyPointAt time.Time                         `json:"consistency_point_at"`
	VerificationState  string                            `json:"verification_state"`
	ConsistencyReport  recovery.RestoreConsistencyReport `json:"consistency_report"`
}

type OperatorRestoreVerificationResult struct {
	SchemaID                             string                            `json:"schema_id"`
	BackupSetID                          string                            `json:"backup_set_id"`
	RestoreVerificationRunID             string                            `json:"restore_verification_run_id"`
	VerificationState                    string                            `json:"verification_state"`
	VerificationBasisSHA256              string                            `json:"verification_basis_sha256"`
	CompletedAt                          time.Time                         `json:"completed_at"`
	ConsistencyReport                    recovery.RestoreConsistencyReport `json:"consistency_report"`
	RestoreVerificationArtifactKey       string                            `json:"restore_verification_artifact_key,omitempty"`
	RestoreVerificationArtifactSHA256    string                            `json:"restore_verification_artifact_sha256,omitempty"`
	RestoreVerificationArtifactSizeBytes int64                             `json:"restore_verification_artifact_size_bytes,omitempty"`
}

type OperatorRestoreVerificationDueResult struct {
	SchemaID                string                               `json:"schema_id"`
	VerificationBasisSHA256 string                               `json:"verification_basis_sha256"`
	AsOf                    time.Time                            `json:"as_of"`
	DueCount                int                                  `json:"due_count"`
	VerifiedCount           int                                  `json:"verified_count"`
	FailedCount             int                                  `json:"failed_count"`
	Results                 []OperatorRestoreVerificationDueItem `json:"results"`
}

type OperatorRestoreVerificationDueItem struct {
	BackupSetID                          string     `json:"backup_set_id"`
	RestoreVerificationRunID             string     `json:"restore_verification_run_id,omitempty"`
	VerificationState                    string     `json:"verification_state"`
	CompletedAt                          *time.Time `json:"completed_at,omitempty"`
	FailureReason                        string     `json:"failure_reason,omitempty"`
	RestoreVerificationArtifactKey       string     `json:"restore_verification_artifact_key,omitempty"`
	RestoreVerificationArtifactSHA256    string     `json:"restore_verification_artifact_sha256,omitempty"`
	RestoreVerificationArtifactSizeBytes int64      `json:"restore_verification_artifact_size_bytes,omitempty"`
}

type OperatorObjectStoreInitResult struct {
	SchemaID      string `json:"schema_id"`
	Result        string `json:"result"`
	Created       bool   `json:"created"`
	AlreadyExists bool   `json:"already_exists"`
}

type OperatorObjectStoreMigrationResult struct {
	SchemaID             string                                `json:"schema_id"`
	RunID                string                                `json:"run_id"`
	CurrentState         string                                `json:"current_state"`
	TerminalResult       *string                               `json:"terminal_result"`
	CutoverReady         bool                                  `json:"cutover_ready"`
	BlockingFailure      bool                                  `json:"blocking_failure"`
	MigrationRunArtifact operatorMigrationArtifactPathAndProof `json:"migration_run_artifact"`
	CopyLedgerArtifact   operatorMigrationArtifactPathAndProof `json:"copy_ledger_artifact"`
	ValidationArtifact   operatorMigrationArtifactPathAndProof `json:"validation_artifact"`
	ProbeArtifact        operatorMigrationArtifactPathAndProof `json:"probe_artifact"`
	RollbackArtifact     operatorMigrationArtifactPathAndProof `json:"rollback_artifact"`
}

type operatorMigrationArtifactPathAndProof struct {
	Path        string `json:"path"`
	Key         string `json:"key"`
	SHA256      string `json:"sha256"`
	SizeBytes   int64  `json:"size_bytes"`
	ContentType string `json:"content_type"`
}

func RunOperatorCLIContext(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) int {
	return newOperatorRunnerForCLI(stdout, stderr).runCLI(ctx, args)
}

func RunOperatorCLI(args []string, stdout io.Writer, stderr io.Writer) int {
	return RunOperatorCLIContext(context.Background(), args, stdout, stderr)
}

var newOperatorRunnerForCLI = newOperatorRunner

func newOperatorRunner(stdout io.Writer, stderr io.Writer) operatorRunner {
	return operatorRunner{
		stdout: normalizeOperatorWriter(stdout),
		stderr: normalizeOperatorWriter(stderr),
		loadConfig: func(path string) (config.Config, error) {
			if strings.TrimSpace(path) == "" {
				return config.Load()
			}
			return config.LoadWithOptions(config.LoadOptions{Path: path})
		},
		setupPostgres: func(ctx context.Context, cfg config.Config) (operatorPostgresPool, error) {
			return postgres.Setup(ctx, cfg)
		},
		setupObjectStore: func(ctx context.Context, cfg config.Config) (objectstore.Store, error) {
			return objectstore.Setup(ctx, cfg)
		},
		ensureObjectStoreBucket: func(ctx context.Context, cfg config.Config) (objectstore.EnsureBucketResult, error) {
			return objectstore.EnsureConfiguredBucket(ctx, cfg, nil)
		},
		newBackupStorage: func(cfg config.Config) (recovery.BackupStorage, error) {
			return recovery.NewBackupStorageFromConfig(cfg)
		},
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (runner operatorRunner) runCLI(ctx context.Context, args []string) int {
	parsed := parseOperatorCLIArgs(args, runner.stderr)
	if parsed.stop {
		return parsed.exitCode
	}

	if err := runner.run(ctx, parsed); err != nil {
		runner.logger().Error("operator command failed", "error", err)
		return 1
	}
	return 0
}

func (runner operatorRunner) run(ctx context.Context, parsed operatorCLIResult) error {
	switch parsed.command {
	case "backup capture":
		return runner.runBackupCapture(ctx, parsed)
	case "backup-metadata latest":
		return runner.runBackupMetadataLatest(ctx, parsed)
	case "restore latest":
		return runner.runRestoreLatest(ctx, parsed)
	case "restore-verify latest":
		return runner.runRestoreVerifyLatest(ctx, parsed)
	case "restore-verify due":
		return runner.runRestoreVerifyDue(ctx, parsed)
	case "object-store init":
		return runner.runObjectStoreInit(ctx, parsed)
	case "object-store-migration run":
		return runner.runObjectStoreMigration(ctx, parsed)
	default:
		return fmt.Errorf("unsupported operator command %q", parsed.command)
	}
}

func (runner operatorRunner) runBackupCapture(ctx context.Context, parsed operatorCLIResult) error {
	cfg, err := runner.loadConfig(parsed.sourceConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	pool, err := runner.setupPostgres(ctx, cfg)
	if err != nil {
		return fmt.Errorf("open postgres: %w", err)
	}
	defer pool.Close()
	objectStore, err := runner.setupObjectStore(ctx, cfg)
	if err != nil {
		return fmt.Errorf("open object store: %w", err)
	}
	defer func() { _ = objectStore.Close() }()
	backupStorage, err := runner.newBackupStorage(cfg)
	if err != nil {
		return fmt.Errorf("open backup storage: %w", err)
	}
	if err := authorizeDeploymentAdmin(ctx, pool, parsed.email); err != nil {
		return err
	}
	proof, err := loadBackupCaptureQuiescenceProof(parsed.quiescenceProof)
	if err != nil {
		return err
	}
	if err := validateBackupCaptureQuiescenceProof(proof); err != nil {
		return err
	}

	consistencyPointAt := parsed.asOf
	if consistencyPointAt.IsZero() {
		consistencyPointAt = runner.now()
	}
	consistencyPointAt = consistencyPointAt.UTC()
	backupSetID := parsed.captureBackupSetID
	if backupSetID == uuid.Nil {
		backupSetID = uuid.New()
	}
	postgresArtifact, err := recovery.CapturePostgresSnapshotArtifact(ctx, pool)
	if err != nil {
		return fmt.Errorf("capture postgres artifact: %w", err)
	}
	blobIndex, err := loadBackupCaptureObjectBlobIndex(ctx, pool)
	if err != nil {
		return err
	}
	objectBucket, err := backupCaptureObjectStoreBucket(cfg)
	if err != nil {
		return err
	}
	objectArtifacts, err := recovery.CaptureSeaweedFSS3ObjectStoreBackupArtifacts(ctx, objectStore, recovery.ObjectStoreBackupCaptureParams{
		BackupSetID:               backupSetID,
		ConsistencyPointAt:        consistencyPointAt,
		Bucket:                    objectBucket,
		BlobObjectIDsByStorageRef: blobIndex,
	})
	if err != nil {
		return fmt.Errorf("capture object-store artifacts: %w", err)
	}
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
		return fmt.Errorf("capture backup set: %w", err)
	}
	if err := recovery.NewBackupCatalog(recovery.NewStore(pool), backupStorage).VerifyBackupSetDurability(ctx, backupSet); err != nil {
		return fmt.Errorf("verify captured backup durability: %w", err)
	}
	storageProof, err := backupCaptureStorageEncryptionProof(backupStorage)
	if err != nil {
		return err
	}
	return runner.encodeJSON(backupCaptureResultFromStore(backupSet, objectArtifacts.Summary, storageProof, proof))
}

func (runner operatorRunner) runBackupMetadataLatest(ctx context.Context, parsed operatorCLIResult) error {
	cfg, err := runner.loadConfig(parsed.sourceConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	pool, err := runner.setupPostgres(ctx, cfg)
	if err != nil {
		return fmt.Errorf("open postgres: %w", err)
	}
	defer pool.Close()

	user, err := authn.NewStore(pool).GetUserByNormalizedEmail(ctx, parsed.email)
	if err != nil || !user.IsActive || !user.IsDeploymentAdmin {
		return errors.New("deployment admin authorization failed")
	}

	asOf := parsed.asOf
	if asOf.IsZero() {
		asOf = runner.now()
	}
	backupStorage, err := runner.newBackupStorage(cfg)
	if err != nil {
		return fmt.Errorf("open backup storage: %w", err)
	}
	selection, err := recovery.NewBackupCatalog(recovery.NewStore(pool), backupStorage).RestoreCandidateBackupSelection(ctx, asOf)
	if err != nil {
		return err
	}
	payload := backupMetadataInspectionFromSelection(selection)
	encoder := json.NewEncoder(runner.stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(payload); err != nil {
		return fmt.Errorf("encode backup metadata inspection: %w", err)
	}
	return nil
}

func (runner operatorRunner) runRestoreLatest(ctx context.Context, parsed operatorCLIResult) error {
	sourceCfg, targetCfg, sourcePool, targetPool, sourceObjectStore, targetObjectStore, backupStorage, err := runner.openRestoreRuntime(ctx, parsed)
	if err != nil {
		return err
	}
	defer sourcePool.Close()
	defer targetPool.Close()
	defer func() { _ = sourceObjectStore.Close() }()
	defer func() { _ = targetObjectStore.Close() }()

	if err := authorizeDeploymentAdmin(ctx, sourcePool, parsed.email); err != nil {
		return err
	}
	asOf := parsed.asOf
	if asOf.IsZero() {
		asOf = runner.now()
	}
	sourceStore := recovery.NewStore(sourcePool)
	backupSet, err := recovery.NewBackupCatalog(sourceStore, backupStorage).RestoreCandidateBackup(ctx, asOf)
	if err != nil {
		return err
	}
	if parsed.confirmBackupSetID == uuid.Nil {
		return errors.New("confirm-backup-set-id is required")
	}
	if backupSet.BackupSetID != parsed.confirmBackupSetID {
		return fmt.Errorf("confirmed backup_set_id %s does not match latest retained backup %s", parsed.confirmBackupSetID, backupSet.BackupSetID)
	}
	if err := runner.preflightRestoreTarget(ctx, parsed.sourceConfigPath, parsed.targetConfigPath, sourceCfg, targetCfg, targetPool, targetObjectStore); err != nil {
		return err
	}

	result, err := recovery.NewRestoreRunner(sourceStore, backupStorage).RestoreLatestSuccessfulRetained(ctx, recovery.RestoreTarget{
		Postgres:    targetPool,
		ObjectStore: targetObjectStore,
		Projections: projections.NewStore(targetPool),
	}, asOf)
	if err != nil {
		return err
	}
	payload := OperatorRestoreResult{
		SchemaID:           OperatorRestoreResultSchemaID,
		BackupSetID:        result.BackupSet.BackupSetID.String(),
		ConsistencyPointAt: result.BackupSet.ConsistencyPointAt,
		VerificationState:  string(result.BackupSet.VerificationState),
		ConsistencyReport:  result.ConsistencyReport,
	}
	return runner.encodeJSON(payload)
}

func (runner operatorRunner) runRestoreVerifyLatest(ctx context.Context, parsed operatorCLIResult) error {
	sourceCfg, targetCfg, sourcePool, targetPool, sourceObjectStore, targetObjectStore, backupStorage, err := runner.openRestoreRuntime(ctx, parsed)
	if err != nil {
		return err
	}
	defer sourcePool.Close()
	defer targetPool.Close()
	defer func() { _ = sourceObjectStore.Close() }()
	defer func() { _ = targetObjectStore.Close() }()

	if err := authorizeDeploymentAdmin(ctx, sourcePool, parsed.email); err != nil {
		return err
	}
	asOf := parsed.asOf
	if asOf.IsZero() {
		asOf = runner.now()
	}
	if err := runner.preflightRestoreTarget(ctx, parsed.sourceConfigPath, parsed.targetConfigPath, sourceCfg, targetCfg, targetPool, targetObjectStore); err != nil {
		return err
	}
	basis, err := restoreVerificationBasisForConfig(sourceCfg)
	if err != nil {
		return err
	}
	sourceStore := recovery.NewStore(sourcePool)
	verify := recovery.NewRestoreVerificationService(sourceStore, recovery.NewRestoreRunner(sourceStore, backupStorage))
	result, err := verify.VerifyLatestSuccessfulRetained(ctx, recovery.RestoreVerificationTarget{
		RestoreTarget: recovery.RestoreTarget{
			Postgres:    targetPool,
			ObjectStore: targetObjectStore,
			Projections: projections.NewStore(targetPool),
		},
		Probe: RestoreVerificationWorkbookProbe{Postgres: targetPool},
	}, asOf, basis)
	if err != nil {
		return err
	}
	payload := OperatorRestoreVerificationResult{
		SchemaID:                             OperatorRestoreVerificationSchemaID,
		BackupSetID:                          result.BackupSet.BackupSetID.String(),
		RestoreVerificationRunID:             result.Run.RestoreVerificationRunID.String(),
		VerificationState:                    string(result.Run.VerificationState),
		VerificationBasisSHA256:              result.Run.VerificationBasisSHA256,
		CompletedAt:                          result.Run.CompletedAt,
		ConsistencyReport:                    result.Run.ConsistencyReport,
		RestoreVerificationArtifactKey:       result.ArtifactProof.Key,
		RestoreVerificationArtifactSHA256:    result.ArtifactProof.SHA256,
		RestoreVerificationArtifactSizeBytes: result.ArtifactProof.SizeBytes,
	}
	return runner.encodeJSON(payload)
}

func (runner operatorRunner) runRestoreVerifyDue(ctx context.Context, parsed operatorCLIResult) error {
	sourceCfg, targetCfg, sourcePool, targetPool, sourceObjectStore, targetObjectStore, backupStorage, err := runner.openRestoreRuntime(ctx, parsed)
	if err != nil {
		return err
	}
	defer sourcePool.Close()
	defer targetPool.Close()
	defer func() { _ = sourceObjectStore.Close() }()
	defer func() { _ = targetObjectStore.Close() }()

	if err := authorizeDeploymentAdmin(ctx, sourcePool, parsed.email); err != nil {
		return err
	}
	asOf := parsed.asOf
	if asOf.IsZero() {
		asOf = runner.now()
	}
	if err := runner.preflightRestoreVerificationTarget(ctx, parsed.sourceConfigPath, parsed.targetConfigPath, sourceCfg, targetCfg, targetPool, targetObjectStore); err != nil {
		return err
	}
	locked, err := tryRestoreVerificationAdvisoryLock(ctx, sourcePool)
	if err != nil {
		return err
	}
	if !locked {
		return errors.New("restore verification due runner is already active")
	}
	defer func() {
		_ = unlockRestoreVerificationAdvisoryLock(context.Background(), sourcePool)
	}()

	basis, err := restoreVerificationBasisForConfig(sourceCfg)
	if err != nil {
		return err
	}
	sourceStore := recovery.NewStore(sourcePool)
	catalog := recovery.NewBackupCatalog(sourceStore, backupStorage)
	due, err := catalog.ListBackupsDueForRestoreVerification(ctx, asOf, basis)
	if err != nil {
		return err
	}
	verify := recovery.NewRestoreVerificationService(sourceStore, recovery.NewRestoreRunner(sourceStore, backupStorage))
	summary := OperatorRestoreVerificationDueResult{
		SchemaID:                OperatorRestoreVerificationDueSchemaID,
		VerificationBasisSHA256: basis,
		AsOf:                    asOf.UTC(),
		DueCount:                len(due),
		Results:                 make([]OperatorRestoreVerificationDueItem, 0, len(due)),
	}
	for _, backupSet := range due {
		result, verifyErr := verify.VerifyBackupSet(ctx, recovery.RestoreVerificationTarget{
			RestoreTarget: recovery.RestoreTarget{
				Postgres:    targetPool,
				ObjectStore: targetObjectStore,
				Projections: projections.NewStore(targetPool),
			},
			Probe: RestoreVerificationWorkbookProbe{Postgres: targetPool},
		}, backupSet, basis)
		item := OperatorRestoreVerificationDueItem{
			BackupSetID:       backupSet.BackupSetID.String(),
			VerificationState: string(recovery.VerificationFailed),
		}
		if result.Run.RestoreVerificationRunID != uuid.Nil {
			item.RestoreVerificationRunID = result.Run.RestoreVerificationRunID.String()
			item.VerificationState = string(result.Run.VerificationState)
			completedAt := result.Run.CompletedAt
			item.CompletedAt = &completedAt
			item.FailureReason = result.Run.FailureReason
			item.RestoreVerificationArtifactKey = result.ArtifactProof.Key
			item.RestoreVerificationArtifactSHA256 = result.ArtifactProof.SHA256
			item.RestoreVerificationArtifactSizeBytes = result.ArtifactProof.SizeBytes
		}
		if verifyErr != nil {
			summary.FailedCount++
			if item.FailureReason == "" {
				item.FailureReason = "restore_verification_failed"
			}
			summary.Results = append(summary.Results, item)
			continue
		}
		summary.VerifiedCount++
		summary.Results = append(summary.Results, item)
	}
	if err := runner.encodeJSON(summary); err != nil {
		return err
	}
	if summary.FailedCount > 0 {
		return fmt.Errorf("restore verification due runner recorded %d failed verification(s)", summary.FailedCount)
	}
	return nil
}

func (runner operatorRunner) runObjectStoreInit(ctx context.Context, parsed operatorCLIResult) error {
	cfg, err := runner.loadConfig(parsed.sourceConfigPath)
	if err != nil {
		return sanitizeObjectStoreInitError(err)
	}
	result, err := runner.ensureObjectStoreBucket(ctx, cfg)
	if err != nil {
		return sanitizeObjectStoreInitError(err)
	}
	payload := OperatorObjectStoreInitResult{
		SchemaID:      OperatorObjectStoreInitResultSchemaID,
		Result:        operatorObjectStoreInitResultCode(result),
		Created:       result.Created,
		AlreadyExists: result.AlreadyExists,
	}
	return runner.encodeJSON(payload)
}

func (runner operatorRunner) runObjectStoreMigration(ctx context.Context, parsed operatorCLIResult) error {
	sourceCfg, targetCfg, sourcePool, sourceObjectStore, targetObjectStore, backupStorage, err := runner.openObjectStoreMigrationRuntime(ctx, parsed)
	if err != nil {
		return err
	}
	defer sourcePool.Close()
	defer func() { _ = sourceObjectStore.Close() }()
	defer func() { _ = targetObjectStore.Close() }()

	if err := authorizeDeploymentAdmin(ctx, sourcePool, parsed.email); err != nil {
		return err
	}
	sourceSettings, targetSettings, err := runner.preflightObjectStoreMigrationConfig(parsed, sourceCfg, targetCfg)
	if err != nil {
		return err
	}
	proof, err := loadObjectStoreMigrationQuiescenceProof(parsed.quiescenceProof)
	if err != nil {
		return err
	}
	if err := recovery.ValidateObjectStoreMigrationWriteQuiescenceProof(proof); err != nil {
		return err
	}
	asOf := parsed.asOf
	if asOf.IsZero() {
		asOf = runner.now()
	}
	sourceStore := recovery.NewStore(sourcePool)
	backupSet, err := recovery.NewBackupCatalog(sourceStore, backupStorage).RestoreCandidateBackup(ctx, asOf)
	if err != nil {
		return err
	}
	if backupSet.BackupSetID != parsed.confirmBackupSetID {
		return fmt.Errorf("confirmed backup_set_id %s does not match latest retained backup %s", parsed.confirmBackupSetID, backupSet.BackupSetID)
	}
	backupRef, err := recovery.LoadObjectStoreMigrationBackupRefs(ctx, backupStorage, backupSet)
	if err != nil {
		return fmt.Errorf("load migration backup refs: %w", err)
	}
	runID := parsed.runID
	if runID == uuid.Nil {
		runID = uuid.New()
	}
	now := runner.now()
	run, err := recovery.NewObjectStoreMigrationRun(runID, now, parsed.email, sourceSettings.Endpoint, targetSettings.Endpoint, sourceSettings.Bucket, targetSettings.Bucket)
	if err != nil {
		return err
	}
	if err := recovery.ApplyObjectStoreMigrationEvent(&run, recovery.ObjectStoreMigrationEventPreflightPassed, now.Add(time.Millisecond), map[string]string{
		"source_backend": parsed.sourceBackend,
		"target_backend": parsed.targetBackend,
	}); err != nil {
		return err
	}
	if err := recovery.ApplyObjectStoreMigrationEvent(&run, recovery.ObjectStoreMigrationEventWriteQuiescenceVerified, proof.CheckedAt, map[string]string{
		"proof_kind": proof.ProofKind,
	}); err != nil {
		return err
	}
	run.BackupRefs = append(run.BackupRefs, backupRef)
	if err := recovery.ApplyObjectStoreMigrationEvent(&run, recovery.ObjectStoreMigrationEventBackupCaptured, backupSet.ConsistencyPointAt, map[string]string{
		"backup_set_id": backupSet.BackupSetID.String(),
	}); err != nil {
		return err
	}

	artifactDir, err := filepath.Abs(parsed.artifactsDir)
	if err != nil {
		return fmt.Errorf("resolve artifacts-dir: %w", err)
	}
	probe, probeBody, err := recovery.ProbeObjectStoreMigrationTarget(ctx, runID, targetSettings.Bucket, targetObjectStore, runner.now())
	if err != nil {
		return err
	}
	probeArtifact, err := writeOperatorMigrationArtifact(artifactDir, "target-probe.json", probeBody)
	if err != nil {
		return err
	}
	run.ProbeRef = &probeArtifact.ref
	if err := recovery.ApplyObjectStoreMigrationEvent(&run, recovery.ObjectStoreMigrationEventTargetPrepared, probe.CompletedAt, map[string]string{
		"probe_ref": probeArtifact.path,
	}); err != nil {
		return err
	}

	rollback, rollbackBody, err := recovery.BuildObjectStoreMigrationRollbackEvidence(runID, runner.now())
	if err != nil {
		return err
	}
	rollbackArtifact, err := writeOperatorMigrationArtifact(artifactDir, "rollback-evidence.json", rollbackBody)
	if err != nil {
		return err
	}
	run.RollbackRef = &rollbackArtifact.ref

	objects, err := recovery.ListObjectStoreMigrationBlobs(ctx, sourcePool)
	if err != nil {
		return err
	}
	if err := recovery.ApplyObjectStoreMigrationEvent(&run, recovery.ObjectStoreMigrationEventCopyStarted, runner.now(), nil); err != nil {
		return err
	}
	copyLedger, copyLedgerBody, err := recovery.CopyObjectStoreMigrationObjects(ctx, recovery.ObjectStoreMigrationCopyParams{
		RunID:         runID,
		SourceBackend: parsed.sourceBackend,
		TargetBackend: parsed.targetBackend,
		SourceBucket:  sourceSettings.Bucket,
		TargetBucket:  targetSettings.Bucket,
		SourceStore:   sourceObjectStore,
		TargetStore:   targetObjectStore,
		Objects:       objects,
	})
	if err != nil {
		return err
	}
	copyArtifact, err := writeOperatorMigrationArtifact(artifactDir, "copy-ledger.json", copyLedgerBody)
	if err != nil {
		return err
	}
	run.CopyLedgerRef = &copyArtifact.ref

	validationStartedAt := runner.now()
	validation, validationBody, err := recovery.ValidateObjectStoreMigration(ctx, recovery.ObjectStoreMigrationValidationParams{
		RunID:         runID,
		StartedAt:     validationStartedAt,
		CompletedAt:   validationStartedAt.Add(time.Millisecond),
		SourceBackend: parsed.sourceBackend,
		TargetBackend: parsed.targetBackend,
		SourceBucket:  sourceSettings.Bucket,
		TargetBucket:  targetSettings.Bucket,
		SourceStore:   sourceObjectStore,
		TargetStore:   targetObjectStore,
		Objects:       objects,
	})
	if err != nil {
		return err
	}
	validationArtifact, err := writeOperatorMigrationArtifact(artifactDir, "validation.json", validationBody)
	if err != nil {
		return err
	}
	run.ValidationRef = &validationArtifact.ref

	blockingFailure := copyLedger.Result != "pass" || validation.Result != "pass"
	if !blockingFailure {
		if err := recovery.ApplyObjectStoreMigrationEvent(&run, recovery.ObjectStoreMigrationEventCopyCompleted, runner.now(), map[string]string{
			"copy_ledger_ref": copyArtifact.path,
		}); err != nil {
			return err
		}
		if err := recovery.ApplyObjectStoreMigrationEvent(&run, recovery.ObjectStoreMigrationEventValidationStarted, validationStartedAt, map[string]string{
			"validation_ref": validationArtifact.path,
		}); err != nil {
			return err
		}
		if err := recovery.ApplyObjectStoreMigrationEvent(&run, recovery.ObjectStoreMigrationEventValidationPassed, *validation.CompletedAt, map[string]string{
			"validation_result": validation.Result,
		}); err != nil {
			return err
		}
	} else {
		if err := recovery.ApplyObjectStoreMigrationEvent(&run, recovery.ObjectStoreMigrationEventBlockingFailure, runner.now(), map[string]string{
			"copy_ledger_result": copyLedger.Result,
			"validation_result":  validation.Result,
		}); err != nil {
			return err
		}
	}

	runBody, err := recovery.EncodeObjectStoreMigrationRun(run)
	if err != nil {
		return err
	}
	runArtifact, err := writeOperatorMigrationArtifact(artifactDir, "migration-run.json", runBody)
	if err != nil {
		return err
	}
	var terminal *string
	if run.TerminalResult != nil {
		value := string(*run.TerminalResult)
		terminal = &value
	}
	payload := OperatorObjectStoreMigrationResult{
		SchemaID:             OperatorObjectStoreMigrationResultSchemaID,
		RunID:                run.RunID,
		CurrentState:         string(run.CurrentState),
		TerminalResult:       terminal,
		CutoverReady:         run.CurrentState == recovery.ObjectStoreMigrationStateCutoverReady,
		BlockingFailure:      blockingFailure,
		MigrationRunArtifact: runArtifact.payload(),
		CopyLedgerArtifact:   copyArtifact.payload(),
		ValidationArtifact:   validationArtifact.payload(),
		ProbeArtifact:        probeArtifact.payload(),
		RollbackArtifact:     rollbackArtifact.payload(),
	}
	if err := runner.encodeJSON(payload); err != nil {
		return err
	}
	if blockingFailure {
		return errors.New("object-store migration blocked before cutover")
	}
	_ = rollback
	return nil
}

func parseOperatorCLIArgs(args []string, stderr io.Writer) operatorCLIResult {
	if len(args) < 2 {
		_, _ = fmt.Fprintln(normalizeOperatorWriter(stderr), operatorUsage())
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	if len(args) >= 2 && args[0] == "backup" && args[1] == "capture" {
		return parseBackupCaptureArgs(args[2:], stderr)
	}
	if len(args) >= 2 && args[0] == "object-store" && args[1] == "init" {
		return parseObjectStoreInitArgs(args[2:], stderr)
	}
	if len(args) >= 2 && args[0] == "object-store-migration" && args[1] == "run" {
		return parseObjectStoreMigrationRunArgs(args[2:], stderr)
	}
	switch args[0] + " " + args[1] {
	case "backup-metadata latest":
		return parseBackupMetadataLatestArgs(args[2:], stderr)
	case "restore latest":
		return parseRestoreLatestArgs(args[2:], stderr)
	case "restore-verify latest":
		return parseRestoreVerifyLatestArgs(args[2:], stderr)
	case "restore-verify due":
		return parseRestoreVerifyDueArgs(args[2:], stderr)
	default:
		_, _ = fmt.Fprintln(normalizeOperatorWriter(stderr), operatorUsage())
		return operatorCLIResult{stop: true, exitCode: 2}
	}
}

func parseBackupCaptureArgs(args []string, stderr io.Writer) operatorCLIResult {
	flags := flag.NewFlagSet("operator backup capture", flag.ContinueOnError)
	flags.SetOutput(normalizeOperatorWriter(stderr))
	email := flags.String("deployment-admin-email", "", "active deployment-admin email authorized to capture a backup")
	sourceConfig := flags.String("source-config", "", "optional source deployment config path")
	backupSetIDRaw := flags.String("backup-set-id", "", "optional stable backup_set_id UUID")
	asOfRaw := flags.String("as-of", "", "RFC3339 consistency point timestamp")
	quiescenceProof := flags.String("quiescence-proof", "", "path to backup capture write-quiescence proof JSON")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return operatorCLIResult{stop: true, exitCode: 0}
		}
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	normalizedEmail, asOf, ok := parseOperatorCommonFlags(stderr, *email, *asOfRaw)
	if !ok {
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	var backupSetID uuid.UUID
	if strings.TrimSpace(*backupSetIDRaw) != "" {
		parsed, err := uuid.Parse(strings.TrimSpace(*backupSetIDRaw))
		if err != nil {
			_, _ = fmt.Fprintf(normalizeOperatorWriter(stderr), "backup-set-id must be a UUID: %v\n", err)
			return operatorCLIResult{stop: true, exitCode: 2}
		}
		backupSetID = parsed
	}
	proofPath := strings.TrimSpace(*quiescenceProof)
	if proofPath == "" {
		_, _ = fmt.Fprintln(normalizeOperatorWriter(stderr), "quiescence-proof is required")
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	return operatorCLIResult{
		command:            "backup capture",
		email:              normalizedEmail,
		asOf:               asOf,
		sourceConfigPath:   strings.TrimSpace(*sourceConfig),
		captureBackupSetID: backupSetID,
		quiescenceProof:    proofPath,
	}
}

func parseObjectStoreInitArgs(args []string, stderr io.Writer) operatorCLIResult {
	flags := flag.NewFlagSet("operator object-store init", flag.ContinueOnError)
	flags.SetOutput(normalizeOperatorWriter(stderr))
	sourceConfig := flags.String("config", "", "optional deployment config path")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return operatorCLIResult{stop: true, exitCode: 0}
		}
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	return operatorCLIResult{
		command:          "object-store init",
		sourceConfigPath: strings.TrimSpace(*sourceConfig),
	}
}

func parseObjectStoreMigrationRunArgs(args []string, stderr io.Writer) operatorCLIResult {
	flags := flag.NewFlagSet("operator object-store-migration run", flag.ContinueOnError)
	flags.SetOutput(normalizeOperatorWriter(stderr))
	email := flags.String("deployment-admin-email", "", "active deployment-admin email authorized to migrate object storage")
	sourceConfig := flags.String("source-config", "", "source deployment config path")
	targetConfig := flags.String("target-config", "", "target deployment config path")
	confirmRaw := flags.String("confirm-backup-set-id", "", "latest backup_set_id confirmation")
	asOfRaw := flags.String("as-of", "", "RFC3339 timestamp for latest-success freshness evaluation")
	runIDRaw := flags.String("run-id", "", "optional stable migration run UUID")
	artifactsDir := flags.String("artifacts-dir", "", "directory for retained migration artifacts")
	quiescenceProof := flags.String("quiescence-proof", "", "path to process_stopped write-quiescence proof JSON")
	sourceBackend := flags.String("source-backend", recovery.ObjectStoreBackendMinIOS3, "source backend label for migration evidence")
	targetBackend := flags.String("target-backend", recovery.ObjectStoreBackendSeaweedFSS3, "target backend label for migration evidence")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return operatorCLIResult{stop: true, exitCode: 0}
		}
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	normalizedEmail, asOf, ok := parseOperatorCommonFlags(stderr, *email, *asOfRaw)
	if !ok {
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	confirm, ok := parseRequiredUUIDFlag(stderr, "confirm-backup-set-id", *confirmRaw)
	if !ok {
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	sourcePath, targetPath, ok := parseRequiredRestoreConfigFlags(stderr, *sourceConfig, *targetConfig)
	if !ok {
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	artifactPath := strings.TrimSpace(*artifactsDir)
	if artifactPath == "" {
		_, _ = fmt.Fprintln(normalizeOperatorWriter(stderr), "artifacts-dir is required")
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	proofPath := strings.TrimSpace(*quiescenceProof)
	if proofPath == "" {
		_, _ = fmt.Fprintln(normalizeOperatorWriter(stderr), "quiescence-proof is required")
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	var runID uuid.UUID
	if strings.TrimSpace(*runIDRaw) != "" {
		parsed, err := uuid.Parse(strings.TrimSpace(*runIDRaw))
		if err != nil {
			_, _ = fmt.Fprintf(normalizeOperatorWriter(stderr), "run-id must be a UUID: %v\n", err)
			return operatorCLIResult{stop: true, exitCode: 2}
		}
		runID = parsed
	}
	return operatorCLIResult{
		command:            "object-store-migration run",
		email:              normalizedEmail,
		asOf:               asOf,
		sourceConfigPath:   sourcePath,
		targetConfigPath:   targetPath,
		confirmBackupSetID: confirm,
		runID:              runID,
		artifactsDir:       artifactPath,
		quiescenceProof:    proofPath,
		sourceBackend:      strings.TrimSpace(*sourceBackend),
		targetBackend:      strings.TrimSpace(*targetBackend),
	}
}

func parseBackupMetadataLatestArgs(args []string, stderr io.Writer) operatorCLIResult {
	flags := flag.NewFlagSet("operator backup-metadata latest", flag.ContinueOnError)
	flags.SetOutput(normalizeOperatorWriter(stderr))
	email := flags.String("deployment-admin-email", "", "active deployment-admin email authorized to inspect backup metadata")
	sourceConfig := flags.String("source-config", "", "optional source deployment config path")
	asOfRaw := flags.String("as-of", "", "RFC3339 timestamp for latest-success freshness evaluation")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return operatorCLIResult{stop: true, exitCode: 0}
		}
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	normalizedEmail, asOf, ok := parseOperatorCommonFlags(stderr, *email, *asOfRaw)
	if !ok {
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	return operatorCLIResult{
		command:          "backup-metadata latest",
		email:            normalizedEmail,
		asOf:             asOf,
		sourceConfigPath: strings.TrimSpace(*sourceConfig),
	}
}

func parseRestoreLatestArgs(args []string, stderr io.Writer) operatorCLIResult {
	flags := flag.NewFlagSet("operator restore latest", flag.ContinueOnError)
	flags.SetOutput(normalizeOperatorWriter(stderr))
	email := flags.String("deployment-admin-email", "", "active deployment-admin email authorized to restore")
	sourceConfig := flags.String("source-config", "", "source deployment config path")
	targetConfig := flags.String("target-config", "", "fresh target deployment config path")
	confirmRaw := flags.String("confirm-backup-set-id", "", "latest backup_set_id confirmation")
	asOfRaw := flags.String("as-of", "", "RFC3339 timestamp for latest-success freshness evaluation")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return operatorCLIResult{stop: true, exitCode: 0}
		}
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	normalizedEmail, asOf, ok := parseOperatorCommonFlags(stderr, *email, *asOfRaw)
	if !ok {
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	confirm, ok := parseRequiredUUIDFlag(stderr, "confirm-backup-set-id", *confirmRaw)
	if !ok {
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	sourcePath, targetPath, ok := parseRequiredRestoreConfigFlags(stderr, *sourceConfig, *targetConfig)
	if !ok {
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	return operatorCLIResult{
		command:            "restore latest",
		email:              normalizedEmail,
		asOf:               asOf,
		sourceConfigPath:   sourcePath,
		targetConfigPath:   targetPath,
		confirmBackupSetID: confirm,
	}
}

func parseRestoreVerifyLatestArgs(args []string, stderr io.Writer) operatorCLIResult {
	flags := flag.NewFlagSet("operator restore-verify latest", flag.ContinueOnError)
	flags.SetOutput(normalizeOperatorWriter(stderr))
	email := flags.String("deployment-admin-email", "", "active deployment-admin email authorized to verify restores")
	sourceConfig := flags.String("source-config", "", "source deployment config path")
	targetConfig := flags.String("target-config", "", "fresh target deployment config path")
	asOfRaw := flags.String("as-of", "", "RFC3339 timestamp for latest-success freshness evaluation")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return operatorCLIResult{stop: true, exitCode: 0}
		}
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	normalizedEmail, asOf, ok := parseOperatorCommonFlags(stderr, *email, *asOfRaw)
	if !ok {
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	sourcePath, targetPath, ok := parseRequiredRestoreConfigFlags(stderr, *sourceConfig, *targetConfig)
	if !ok {
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	return operatorCLIResult{
		command:          "restore-verify latest",
		email:            normalizedEmail,
		asOf:             asOf,
		sourceConfigPath: sourcePath,
		targetConfigPath: targetPath,
	}
}

func parseRestoreVerifyDueArgs(args []string, stderr io.Writer) operatorCLIResult {
	result := parseRestoreVerifyLatestArgs(args, stderr)
	if result.stop {
		return result
	}
	result.command = "restore-verify due"
	return result
}

func parseOperatorCommonFlags(stderr io.Writer, emailRaw string, asOfRaw string) (string, time.Time, bool) {
	normalizedEmail := strings.TrimSpace(emailRaw)
	if normalizedEmail == "" {
		_, _ = fmt.Fprintln(normalizeOperatorWriter(stderr), "deployment-admin-email is required")
		return "", time.Time{}, false
	}
	var asOf time.Time
	if strings.TrimSpace(asOfRaw) != "" {
		parsed, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(asOfRaw))
		if err != nil {
			_, _ = fmt.Fprintf(normalizeOperatorWriter(stderr), "as-of must be RFC3339: %v\n", err)
			return "", time.Time{}, false
		}
		asOf = parsed.UTC()
	}
	return normalizedEmail, asOf, true
}

func parseRequiredUUIDFlag(stderr io.Writer, name string, raw string) (uuid.UUID, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		_, _ = fmt.Fprintf(normalizeOperatorWriter(stderr), "%s is required\n", name)
		return uuid.Nil, false
	}
	parsed, err := uuid.Parse(value)
	if err != nil {
		_, _ = fmt.Fprintf(normalizeOperatorWriter(stderr), "%s must be a UUID: %v\n", name, err)
		return uuid.Nil, false
	}
	return parsed, true
}

func parseRequiredRestoreConfigFlags(stderr io.Writer, sourceConfig string, targetConfig string) (string, string, bool) {
	sourcePath := strings.TrimSpace(sourceConfig)
	targetPath := strings.TrimSpace(targetConfig)
	if sourcePath == "" {
		_, _ = fmt.Fprintln(normalizeOperatorWriter(stderr), "source-config is required")
		return "", "", false
	}
	if targetPath == "" {
		_, _ = fmt.Fprintln(normalizeOperatorWriter(stderr), "target-config is required")
		return "", "", false
	}
	return sourcePath, targetPath, true
}

func operatorUsage() string {
	return strings.Join([]string{
		"usage:",
		"  operator backup capture -deployment-admin-email <email> -quiescence-proof <path> [-source-config <path>] [-backup-set-id <uuid>] [-as-of <RFC3339>]",
		"  operator backup-metadata latest -deployment-admin-email <email> [-source-config <path>] [-as-of <RFC3339>]",
		"  operator restore latest -source-config <path> -target-config <path> -deployment-admin-email <email> -confirm-backup-set-id <uuid> [-as-of <RFC3339>]",
		"  operator restore-verify latest -source-config <path> -target-config <path> -deployment-admin-email <email> [-as-of <RFC3339>]",
		"  operator restore-verify due -source-config <path> -target-config <path> -deployment-admin-email <email> [-as-of <RFC3339>]",
		"  operator object-store init [-config <path>]",
		"  operator object-store-migration run -source-config <path> -target-config <path> -deployment-admin-email <email> -confirm-backup-set-id <uuid> -quiescence-proof <path> -artifacts-dir <dir> [-as-of <RFC3339>] [-run-id <uuid>]",
	}, "\n")
}

func backupMetadataInspectionFromSelection(selection recovery.BackupCatalogSelection) BackupMetadataInspection {
	payload := backupMetadataInspectionFromStore(selection.BackupSet)
	for _, diagnostic := range selection.DurabilityDiagnostics {
		payload.DurabilityDiagnostics = append(payload.DurabilityDiagnostics, BackupDurabilityDiagnostic{
			BackupSetID:        diagnostic.BackupSetID.String(),
			ConsistencyPointAt: diagnostic.ConsistencyPointAt,
			Code:               diagnostic.Code,
		})
	}
	return payload
}

func backupMetadataInspectionFromStore(backupSet recovery.BackupSet) BackupMetadataInspection {
	return BackupMetadataInspection{
		SchemaID:                              BackupMetadataInspectionSchemaID,
		BackupSetID:                           backupSet.BackupSetID.String(),
		ConsistencyPointAt:                    backupSet.ConsistencyPointAt,
		PostgresRestoreAnchor:                 backupSet.PostgresRestoreAnchor,
		ObjectStoreRestoreAnchor:              backupSet.ObjectStoreRestoreAnchor,
		PostgresArtifactKey:                   backupSet.PostgresArtifactKey,
		PostgresArtifactSHA256:                backupSet.PostgresArtifactSHA256,
		PostgresArtifactSizeBytes:             backupSet.PostgresArtifactSizeBytes,
		ObjectStoreArtifactKey:                backupSet.ObjectStoreArtifactKey,
		ObjectStoreArtifactSHA256:             backupSet.ObjectStoreArtifactSHA256,
		ObjectStoreArtifactSizeBytes:          backupSet.ObjectStoreArtifactSizeBytes,
		IntegrityManifestKey:                  backupSet.IntegrityManifestKey,
		IntegrityManifestSHA256:               backupSet.IntegrityManifestSHA256,
		IntegrityManifestSizeBytes:            backupSet.IntegrityManifestSizeBytes,
		CreatedAt:                             backupSet.CreatedAt,
		RetainedUntil:                         backupSet.RetainedUntil,
		PostgresRestoreAnchorRetainedUntil:    backupSet.PostgresRestoreAnchorRetainedUntil,
		ObjectStoreRestoreAnchorRetainedUntil: backupSet.ObjectStoreRestoreAnchorRetainedUntil,
		VerificationState:                     string(backupSet.VerificationState),
		LastVerifiedRestoreAt:                 backupSet.LastVerifiedRestoreAt,
		LastVerificationBasisSHA256:           backupSet.LastVerificationBasisSHA256,
	}
}

func backupCaptureResultFromStore(backupSet recovery.BackupSet, summary recovery.ObjectStoreBackupSummary, storageProof recovery.BackupStorageEncryptionProof, quiescenceProof OperatorBackupCaptureQuiescenceProof) OperatorBackupCaptureResult {
	return OperatorBackupCaptureResult{
		SchemaID:                              OperatorBackupCaptureResultSchemaID,
		BackupSetID:                           backupSet.BackupSetID.String(),
		ConsistencyPointAt:                    backupSet.ConsistencyPointAt,
		CreatedAt:                             backupSet.CreatedAt,
		RetainedUntil:                         backupSet.RetainedUntil,
		PostgresRestoreAnchor:                 backupSet.PostgresRestoreAnchor,
		ObjectStoreRestoreAnchor:              backupSet.ObjectStoreRestoreAnchor,
		PostgresArtifactKey:                   backupSet.PostgresArtifactKey,
		PostgresArtifactSHA256:                backupSet.PostgresArtifactSHA256,
		PostgresArtifactSizeBytes:             backupSet.PostgresArtifactSizeBytes,
		ObjectStoreArtifactKey:                backupSet.ObjectStoreArtifactKey,
		ObjectStoreArtifactSHA256:             backupSet.ObjectStoreArtifactSHA256,
		ObjectStoreArtifactSizeBytes:          backupSet.ObjectStoreArtifactSizeBytes,
		IntegrityManifestKey:                  backupSet.IntegrityManifestKey,
		IntegrityManifestSHA256:               backupSet.IntegrityManifestSHA256,
		IntegrityManifestSizeBytes:            backupSet.IntegrityManifestSizeBytes,
		PostgresRestoreAnchorRetainedUntil:    backupSet.PostgresRestoreAnchorRetainedUntil,
		ObjectStoreRestoreAnchorRetainedUntil: backupSet.ObjectStoreRestoreAnchorRetainedUntil,
		VerificationState:                     string(backupSet.VerificationState),
		LastVerifiedRestoreAt:                 backupSet.LastVerifiedRestoreAt,
		ObjectStoreBackupSummary:              summary,
		StorageEncryption:                     storageProof,
		QuiescenceProof:                       quiescenceProof,
	}
}

type backupCaptureStorageEncryptionReporter interface {
	BackupStorageEncryptionProof() recovery.BackupStorageEncryptionProof
}

func backupCaptureStorageEncryptionProof(storage recovery.BackupStorage) (recovery.BackupStorageEncryptionProof, error) {
	reporter, ok := storage.(backupCaptureStorageEncryptionReporter)
	if !ok {
		return recovery.BackupStorageEncryptionProof{}, recovery.ErrEncryptedBackupStorage
	}
	proof := reporter.BackupStorageEncryptionProof()
	if proof.Mode != recovery.BackupStorageEncryptionModeAESGCM ||
		proof.EnvelopeSchemaID != recovery.BackupArtifactEnvelopeSchemaID ||
		strings.TrimSpace(proof.KeyFingerprintSHA256) == "" {
		return recovery.BackupStorageEncryptionProof{}, recovery.ErrEncryptedBackupStorage
	}
	return proof, nil
}

func loadBackupCaptureQuiescenceProof(path string) (OperatorBackupCaptureQuiescenceProof, error) {
	body, err := os.ReadFile(filepath.Clean(path)) // #nosec G304 -- operator supplied local proof path is intentionally read by deployment-local CLI.
	if err != nil {
		return OperatorBackupCaptureQuiescenceProof{}, fmt.Errorf("read backup capture quiescence proof: %w", err)
	}
	var proof OperatorBackupCaptureQuiescenceProof
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&proof); err != nil {
		return OperatorBackupCaptureQuiescenceProof{}, fmt.Errorf("decode backup capture quiescence proof: %w", err)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return OperatorBackupCaptureQuiescenceProof{}, errors.New("decode backup capture quiescence proof: trailing JSON content")
	}
	return proof, nil
}

func validateBackupCaptureQuiescenceProof(proof OperatorBackupCaptureQuiescenceProof) error {
	if proof.SchemaID != BackupCaptureQuiescenceProofSchemaID {
		return fmt.Errorf("%w: unsupported backup capture quiescence proof schema %q", recovery.ErrInvalidBackupArtifact, proof.SchemaID)
	}
	if proof.ProofKind != "process_stopped" {
		return fmt.Errorf("%w: backup capture proof_kind must be process_stopped", recovery.ErrInvalidBackupArtifact)
	}
	if proof.CheckedAt.IsZero() {
		return fmt.Errorf("%w: backup capture checked_at is required", recovery.ErrInvalidBackupArtifact)
	}
	switch proof.ProcessState {
	case "absent", "stopped_by_supervisor":
	default:
		return fmt.Errorf("%w: backup capture process_state must prove process absence or supervisor stop", recovery.ErrInvalidBackupArtifact)
	}
	if !proof.HTTPListenerClosed || !proof.WebSocketListenerClosed {
		return fmt.Errorf("%w: backup capture listeners must both be closed", recovery.ErrInvalidBackupArtifact)
	}
	return nil
}

func loadBackupCaptureObjectBlobIndex(ctx context.Context, db postgres.DB) (map[string]uuid.UUID, error) {
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

func backupCaptureObjectStoreBucket(cfg config.Config) (string, error) {
	settings, err := objectstore.ResolveSettings(cfg, nil)
	if err != nil {
		return "", fmt.Errorf("resolve object-store settings for backup capture: %w", err)
	}
	switch settings.BindingKind {
	case "managed_service":
		if strings.TrimSpace(settings.Bucket) == "" {
			return "", errors.New("backup capture requires configured object-store bucket")
		}
		return settings.Bucket, nil
	case "filesystem_root":
		if strings.TrimSpace(settings.RootPath) == "" {
			return "", errors.New("backup capture requires configured filesystem object-store root")
		}
		return "filesystem-root:" + filepath.Clean(settings.RootPath), nil
	default:
		return "", fmt.Errorf("backup capture unsupported object-store binding kind %q", settings.BindingKind)
	}
}

func (runner operatorRunner) openObjectStoreMigrationRuntime(ctx context.Context, parsed operatorCLIResult) (config.Config, config.Config, operatorPostgresPool, objectstore.Store, objectstore.Store, recovery.BackupStorage, error) {
	sourceCfg, err := runner.loadConfig(parsed.sourceConfigPath)
	if err != nil {
		return config.Config{}, config.Config{}, nil, nil, nil, nil, fmt.Errorf("load source config: %w", err)
	}
	targetCfg, err := runner.loadConfig(parsed.targetConfigPath)
	if err != nil {
		return config.Config{}, config.Config{}, nil, nil, nil, nil, fmt.Errorf("load target config: %w", err)
	}
	sourcePool, err := runner.setupPostgres(ctx, sourceCfg)
	if err != nil {
		return config.Config{}, config.Config{}, nil, nil, nil, nil, fmt.Errorf("open source postgres: %w", err)
	}
	sourceObjectStore, err := runner.setupObjectStore(ctx, sourceCfg)
	if err != nil {
		sourcePool.Close()
		return config.Config{}, config.Config{}, nil, nil, nil, nil, fmt.Errorf("open source object store: %w", err)
	}
	targetObjectStore, err := runner.setupObjectStore(ctx, targetCfg)
	if err != nil {
		sourcePool.Close()
		_ = sourceObjectStore.Close()
		return config.Config{}, config.Config{}, nil, nil, nil, nil, fmt.Errorf("open target object store: %w", err)
	}
	backupStorage, err := runner.newBackupStorage(sourceCfg)
	if err != nil {
		sourcePool.Close()
		_ = sourceObjectStore.Close()
		_ = targetObjectStore.Close()
		return config.Config{}, config.Config{}, nil, nil, nil, nil, fmt.Errorf("open source backup storage: %w", err)
	}
	return sourceCfg, targetCfg, sourcePool, sourceObjectStore, targetObjectStore, backupStorage, nil
}

func (runner operatorRunner) preflightObjectStoreMigrationConfig(parsed operatorCLIResult, sourceCfg config.Config, targetCfg config.Config) (objectstore.Settings, objectstore.Settings, error) {
	if sameConfigPath(parsed.sourceConfigPath, parsed.targetConfigPath) {
		return objectstore.Settings{}, objectstore.Settings{}, errors.New("object-store migration preflight failed: source-config and target-config must be different files")
	}
	sourceObject, err := objectstore.ResolveSettings(sourceCfg, nil)
	if err != nil {
		return objectstore.Settings{}, objectstore.Settings{}, fmt.Errorf("resolve source object-store settings: %w", err)
	}
	targetObject, err := objectstore.ResolveSettings(targetCfg, nil)
	if err != nil {
		return objectstore.Settings{}, objectstore.Settings{}, fmt.Errorf("resolve target object-store settings: %w", err)
	}
	if sourceObject.BindingKind != "managed_service" || targetObject.BindingKind != "managed_service" {
		return objectstore.Settings{}, objectstore.Settings{}, errors.New("object-store migration preflight failed: source and target object stores must be managed_service S3-compatible bindings")
	}
	if objectStoreBindingID(sourceObject) == objectStoreBindingID(targetObject) {
		return objectstore.Settings{}, objectstore.Settings{}, errors.New("object-store migration preflight failed: source and target object stores must differ")
	}
	if strings.TrimSpace(sourceObject.Bucket) != strings.TrimSpace(targetObject.Bucket) {
		return objectstore.Settings{}, objectstore.Settings{}, errors.New("object-store migration preflight failed: default migration requires identical source and target bucket names")
	}
	if strings.TrimSpace(parsed.sourceBackend) == "" || strings.TrimSpace(parsed.targetBackend) == "" {
		return objectstore.Settings{}, objectstore.Settings{}, errors.New("object-store migration preflight failed: source-backend and target-backend are required")
	}
	return sourceObject, targetObject, nil
}

func loadObjectStoreMigrationQuiescenceProof(path string) (recovery.ObjectStoreMigrationWriteQuiescenceProof, error) {
	body, err := os.ReadFile(filepath.Clean(path)) // #nosec G304 -- operator supplied local proof path is intentionally read by deployment-local CLI.
	if err != nil {
		return recovery.ObjectStoreMigrationWriteQuiescenceProof{}, fmt.Errorf("read migration quiescence proof: %w", err)
	}
	var proof recovery.ObjectStoreMigrationWriteQuiescenceProof
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&proof); err != nil {
		return recovery.ObjectStoreMigrationWriteQuiescenceProof{}, fmt.Errorf("decode migration quiescence proof: %w", err)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return recovery.ObjectStoreMigrationWriteQuiescenceProof{}, errors.New("decode migration quiescence proof: trailing JSON content")
	}
	return proof, nil
}

type operatorMigrationArtifact struct {
	path string
	ref  recovery.ObjectStoreMigrationArtifactRef
}

func writeOperatorMigrationArtifact(dir string, name string, body []byte) (operatorMigrationArtifact, error) {
	if len(body) == 0 {
		return operatorMigrationArtifact{}, errors.New("write migration artifact: body is empty")
	}
	if strings.TrimSpace(dir) == "" {
		return operatorMigrationArtifact{}, errors.New("write migration artifact: artifacts-dir is required")
	}
	if strings.TrimSpace(name) == "" || filepath.Base(name) != name {
		return operatorMigrationArtifact{}, errors.New("write migration artifact: artifact name must be a file name")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return operatorMigrationArtifact{}, fmt.Errorf("create migration artifact dir: %w", err)
	}
	path := filepath.Join(dir, name)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600) // #nosec G304 -- path is beneath operator-selected retained artifact directory.
	if err != nil {
		return operatorMigrationArtifact{}, fmt.Errorf("create migration artifact %s: %w", path, err)
	}
	_, writeErr := file.Write(body)
	closeErr := file.Close()
	if writeErr != nil {
		return operatorMigrationArtifact{}, fmt.Errorf("write migration artifact %s: %w", path, writeErr)
	}
	if closeErr != nil {
		return operatorMigrationArtifact{}, fmt.Errorf("close migration artifact %s: %w", path, closeErr)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return operatorMigrationArtifact{}, fmt.Errorf("resolve migration artifact path: %w", err)
	}
	return operatorMigrationArtifact{
		path: abs,
		ref:  recovery.ArtifactRefForBody(abs, body, "application/json"),
	}, nil
}

func (artifact operatorMigrationArtifact) payload() operatorMigrationArtifactPathAndProof {
	return operatorMigrationArtifactPathAndProof{
		Path:        artifact.path,
		Key:         artifact.ref.Key,
		SHA256:      artifact.ref.SHA256,
		SizeBytes:   artifact.ref.SizeBytes,
		ContentType: artifact.ref.ContentType,
	}
}

func (runner operatorRunner) openRestoreRuntime(ctx context.Context, parsed operatorCLIResult) (config.Config, config.Config, operatorPostgresPool, operatorPostgresPool, objectstore.Store, objectstore.Store, recovery.BackupStorage, error) {
	sourceCfg, err := runner.loadConfig(parsed.sourceConfigPath)
	if err != nil {
		return config.Config{}, config.Config{}, nil, nil, nil, nil, nil, fmt.Errorf("load source config: %w", err)
	}
	targetCfg, err := runner.loadConfig(parsed.targetConfigPath)
	if err != nil {
		return config.Config{}, config.Config{}, nil, nil, nil, nil, nil, fmt.Errorf("load target config: %w", err)
	}
	sourcePool, err := runner.setupPostgres(ctx, sourceCfg)
	if err != nil {
		return config.Config{}, config.Config{}, nil, nil, nil, nil, nil, fmt.Errorf("open source postgres: %w", err)
	}
	targetPool, err := runner.setupPostgres(ctx, targetCfg)
	if err != nil {
		sourcePool.Close()
		return config.Config{}, config.Config{}, nil, nil, nil, nil, nil, fmt.Errorf("open target postgres: %w", err)
	}
	sourceObjectStore, err := runner.setupObjectStore(ctx, sourceCfg)
	if err != nil {
		sourcePool.Close()
		targetPool.Close()
		return config.Config{}, config.Config{}, nil, nil, nil, nil, nil, fmt.Errorf("open source object store: %w", err)
	}
	targetObjectStore, err := runner.setupObjectStore(ctx, targetCfg)
	if err != nil {
		sourcePool.Close()
		targetPool.Close()
		_ = sourceObjectStore.Close()
		return config.Config{}, config.Config{}, nil, nil, nil, nil, nil, fmt.Errorf("open target object store: %w", err)
	}
	backupStorage, err := runner.newBackupStorage(sourceCfg)
	if err != nil {
		sourcePool.Close()
		targetPool.Close()
		_ = sourceObjectStore.Close()
		_ = targetObjectStore.Close()
		return config.Config{}, config.Config{}, nil, nil, nil, nil, nil, fmt.Errorf("open source backup storage: %w", err)
	}
	return sourceCfg, targetCfg, sourcePool, targetPool, sourceObjectStore, targetObjectStore, backupStorage, nil
}

func (runner operatorRunner) preflightRestoreTarget(ctx context.Context, sourceConfigPath string, targetConfigPath string, sourceCfg config.Config, targetCfg config.Config, targetPool postgres.DB, targetObjectStore objectstore.Store) error {
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
	if schemaVersion < phase10RestoreMinimumSchemaVersion {
		return fmt.Errorf("restore target preflight failed: target schema version %d is below required %d", schemaVersion, phase10RestoreMinimumSchemaVersion)
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

func (runner operatorRunner) preflightRestoreVerificationTarget(ctx context.Context, sourceConfigPath string, targetConfigPath string, sourceCfg config.Config, targetCfg config.Config, targetPool postgres.DB, targetObjectStore objectstore.Store) error {
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
	if schemaVersion < phase10RestoreMinimumSchemaVersion {
		return fmt.Errorf("restore verification target preflight failed: target schema version %d is below required %d", schemaVersion, phase10RestoreMinimumSchemaVersion)
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
	if marker.SchemaID != RestoreVerificationTargetMarkerSchemaID || marker.Purpose != "restore_verification_target" {
		return errors.New("restore verification target preflight failed: target marker is not a restore-verification target")
	}
	return nil
}

func tryRestoreVerificationAdvisoryLock(ctx context.Context, pool postgres.DB) (bool, error) {
	var locked bool
	if err := pool.QueryRow(ctx, `SELECT pg_try_advisory_lock($1)`, restoreVerificationAdvisoryLockKey).Scan(&locked); err != nil {
		return false, fmt.Errorf("acquire restore verification advisory lock: %w", err)
	}
	return locked, nil
}

func unlockRestoreVerificationAdvisoryLock(ctx context.Context, pool postgres.DB) error {
	var unlocked bool
	if err := pool.QueryRow(ctx, `SELECT pg_advisory_unlock($1)`, restoreVerificationAdvisoryLockKey).Scan(&unlocked); err != nil {
		return fmt.Errorf("release restore verification advisory lock: %w", err)
	}
	return nil
}

func authorizeDeploymentAdmin(ctx context.Context, pool postgres.DB, email string) error {
	user, err := authn.NewStore(pool).GetUserByNormalizedEmail(ctx, email)
	if err != nil || !user.IsActive || !user.IsDeploymentAdmin {
		return errors.New("deployment admin authorization failed")
	}
	return nil
}

func (runner operatorRunner) encodeJSON(payload any) error {
	encoder := json.NewEncoder(runner.stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(payload); err != nil {
		return fmt.Errorf("encode operator JSON: %w", err)
	}
	return nil
}

func operatorObjectStoreInitResultCode(result objectstore.EnsureBucketResult) string {
	if result.Created {
		return "created"
	}
	return "already_exists"
}

func sanitizeObjectStoreInitError(err error) error {
	if err == nil {
		return nil
	}
	reasonCode := "dependency_unavailable"
	var diagnosticsErr *config.DiagnosticsError
	if errors.As(err, &diagnosticsErr) && diagnosticsErr.Code != "" {
		reasonCode = diagnosticsErr.Code
	} else if adapterErr, ok := objectstore.AsAdapterError(err); ok && adapterErr.Reason != "" {
		reasonCode = string(adapterErr.Reason)
	} else {
		lower := strings.ToLower(err.Error())
		switch {
		case strings.Contains(lower, "service_ref"):
			reasonCode = "managed_service_ref_invalid"
		case strings.Contains(lower, "missing object-store"):
			reasonCode = "missing_service_binding"
		case strings.Contains(lower, "parse cartulary_s3_"):
			reasonCode = "invalid_service_binding"
		}
	}
	return fmt.Errorf("object-store init failed: reason_code=%s", reasonCode)
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

func restoreVerificationBasisForConfig(cfg config.Config) (string, error) {
	return recovery.RestoreVerificationBasisSHA256(map[string]string{
		"backup_mechanism":         "cartulary.phase10.filesystem_snapshot.v1",
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

func (runner operatorRunner) logger() *slog.Logger {
	return slog.New(slog.NewTextHandler(runner.stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func normalizeOperatorWriter(writer io.Writer) io.Writer {
	if writer == nil {
		return io.Discard
	}
	return writer
}
