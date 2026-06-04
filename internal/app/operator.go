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
	BackupMetadataInspectionSchemaID              = "cartulary.operator.backup_metadata.v1"
	OperatorRestoreResultSchemaID                 = "cartulary.operator.restore_result.v1"
	OperatorRestoreVerificationSchemaID           = "cartulary.operator.restore_verification_result.v1"
	OperatorRestoreVerificationDueSchemaID        = "cartulary.operator.restore_verification_due_result.v1"
	RestoreVerificationTargetMarkerSchemaID       = "cartulary.restore_verification_target.v1"
	phase10RestoreMinimumSchemaVersion      int64 = 22
	restoreVerificationAdvisoryLockKey      int64 = 401010
)

type operatorPostgresPool interface {
	postgres.DB
	Close()
}

type operatorRunner struct {
	stdout           io.Writer
	stderr           io.Writer
	loadConfig       func(string) (config.Config, error)
	setupPostgres    func(context.Context, config.Config) (operatorPostgresPool, error)
	setupObjectStore func(context.Context, config.Config) (objectstore.Store, error)
	newBackupStorage func(config.Config) (recovery.BackupStorage, error)
	now              func() time.Time
}

type operatorCLIResult struct {
	stop               bool
	exitCode           int
	command            string
	email              string
	asOf               time.Time
	sourceConfigPath   string
	targetConfigPath   string
	confirmBackupSetID uuid.UUID
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
	case "backup-metadata latest":
		return runner.runBackupMetadataLatest(ctx, parsed)
	case "restore latest":
		return runner.runRestoreLatest(ctx, parsed)
	case "restore-verify latest":
		return runner.runRestoreVerifyLatest(ctx, parsed)
	case "restore-verify due":
		return runner.runRestoreVerifyDue(ctx, parsed)
	default:
		return fmt.Errorf("unsupported operator command %q", parsed.command)
	}
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

func parseOperatorCLIArgs(args []string, stderr io.Writer) operatorCLIResult {
	if len(args) < 2 {
		_, _ = fmt.Fprintln(normalizeOperatorWriter(stderr), operatorUsage())
		return operatorCLIResult{stop: true, exitCode: 2}
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
		"  operator backup-metadata latest -deployment-admin-email <email> [-source-config <path>] [-as-of <RFC3339>]",
		"  operator restore latest -source-config <path> -target-config <path> -deployment-admin-email <email> -confirm-backup-set-id <uuid> [-as-of <RFC3339>]",
		"  operator restore-verify latest -source-config <path> -target-config <path> -deployment-admin-email <email> [-as-of <RFC3339>]",
		"  operator restore-verify due -source-config <path> -target-config <path> -deployment-admin-email <email> [-as-of <RFC3339>]",
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
