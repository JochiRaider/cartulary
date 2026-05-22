package app

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const BackupMetadataInspectionSchemaID = "cartulary.operator.backup_metadata.v1"

type operatorPostgresPool interface {
	postgres.DB
	Close()
}

type operatorRunner struct {
	stdout        io.Writer
	stderr        io.Writer
	loadConfig    func() (config.Config, error)
	setupPostgres func(context.Context, config.Config) (operatorPostgresPool, error)
	now           func() time.Time
}

type operatorCLIResult struct {
	stop     bool
	exitCode int
	command  string
	email    string
	asOf     time.Time
}

type BackupMetadataInspection struct {
	SchemaID                              string     `json:"schema_id"`
	BackupSetID                           string     `json:"backup_set_id"`
	ConsistencyPointAt                    time.Time  `json:"consistency_point_at"`
	PostgresRestoreAnchor                 string     `json:"postgres_restore_anchor"`
	ObjectStoreRestoreAnchor              string     `json:"object_store_restore_anchor"`
	PostgresArtifactKey                   string     `json:"postgres_artifact_key"`
	PostgresArtifactSHA256                string     `json:"postgres_artifact_sha256"`
	PostgresArtifactSizeBytes             int64      `json:"postgres_artifact_size_bytes"`
	ObjectStoreArtifactKey                string     `json:"object_store_artifact_key"`
	ObjectStoreArtifactSHA256             string     `json:"object_store_artifact_sha256"`
	ObjectStoreArtifactSizeBytes          int64      `json:"object_store_artifact_size_bytes"`
	IntegrityManifestKey                  string     `json:"integrity_manifest_key"`
	IntegrityManifestSHA256               string     `json:"integrity_manifest_sha256"`
	IntegrityManifestSizeBytes            int64      `json:"integrity_manifest_size_bytes"`
	CreatedAt                             time.Time  `json:"created_at"`
	RetainedUntil                         time.Time  `json:"retained_until"`
	PostgresRestoreAnchorRetainedUntil    time.Time  `json:"postgres_restore_anchor_retained_until"`
	ObjectStoreRestoreAnchorRetainedUntil time.Time  `json:"object_store_restore_anchor_retained_until"`
	VerificationState                     string     `json:"verification_state"`
	LastVerifiedRestoreAt                 *time.Time `json:"last_verified_restore_at"`
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
		stdout:     normalizeOperatorWriter(stdout),
		stderr:     normalizeOperatorWriter(stderr),
		loadConfig: config.Load,
		setupPostgres: func(ctx context.Context, cfg config.Config) (operatorPostgresPool, error) {
			return postgres.Setup(ctx, cfg)
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
	if parsed.command != "backup-metadata latest" {
		return fmt.Errorf("unsupported operator command %q", parsed.command)
	}
	cfg, err := runner.loadConfig()
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
	backupSet, err := recovery.NewStore(pool).LatestSuccessfulRetainedBackup(ctx, asOf)
	if err != nil {
		return err
	}
	payload := backupMetadataInspectionFromStore(backupSet)
	encoder := json.NewEncoder(runner.stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(payload); err != nil {
		return fmt.Errorf("encode backup metadata inspection: %w", err)
	}
	return nil
}

func parseOperatorCLIArgs(args []string, stderr io.Writer) operatorCLIResult {
	if len(args) < 2 || args[0] != "backup-metadata" || args[1] != "latest" {
		_, _ = fmt.Fprintln(normalizeOperatorWriter(stderr), "usage: operator backup-metadata latest -deployment-admin-email <email> [-as-of <RFC3339>]")
		return operatorCLIResult{stop: true, exitCode: 2}
	}

	flags := flag.NewFlagSet("operator backup-metadata latest", flag.ContinueOnError)
	flags.SetOutput(normalizeOperatorWriter(stderr))
	email := flags.String("deployment-admin-email", "", "active deployment-admin email authorized to inspect backup metadata")
	asOfRaw := flags.String("as-of", "", "RFC3339 timestamp for latest-success freshness evaluation")
	if err := flags.Parse(args[2:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return operatorCLIResult{stop: true, exitCode: 0}
		}
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	normalizedEmail := strings.TrimSpace(*email)
	if normalizedEmail == "" {
		_, _ = fmt.Fprintln(normalizeOperatorWriter(stderr), "deployment-admin-email is required")
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	var asOf time.Time
	if strings.TrimSpace(*asOfRaw) != "" {
		parsed, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(*asOfRaw))
		if err != nil {
			_, _ = fmt.Fprintf(normalizeOperatorWriter(stderr), "as-of must be RFC3339: %v\n", err)
			return operatorCLIResult{stop: true, exitCode: 2}
		}
		asOf = parsed.UTC()
	}
	return operatorCLIResult{
		command: "backup-metadata latest",
		email:   normalizedEmail,
		asOf:    asOf,
	}
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
