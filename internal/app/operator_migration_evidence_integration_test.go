package app

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestPhase0_MigrationEvidenceCommand_I_0_07(t *testing.T) {
	ctx := context.Background()
	postgresHarness := pgtest.Start(t)
	testDB, err := postgresHarness.NewMigrationDatabase(ctx, "operator-migration-evidence")
	if err != nil {
		t.Fatalf("create migration evidence database: %v", err)
	}
	t.Cleanup(func() {
		if err := postgresHarness.DropMigrationDatabase(context.Background(), testDB.Name); err != nil {
			t.Fatalf("drop migration evidence database: %v", err)
		}
	})

	sqlDB, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open migration evidence database: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	if _, err := postgres.Migrate(ctx, sqlDB, dbmigrations.Source(), "up"); err != nil {
		t.Fatalf("migrate database: %v", err)
	}
	payload := runMigrationEvidenceCaptureForDatabase(t, testDB.DSN)
	if !payload.GooseLedger.MetadataPresent {
		t.Fatalf("expected goose metadata table to be present: %#v", payload.GooseLedger)
	}
	if payload.GooseLedger.CurrentEffectiveAppliedVersion != payload.Manifest.ExpectedMaxVersion {
		t.Fatalf("expected current version %d, got %d", payload.Manifest.ExpectedMaxVersion, payload.GooseLedger.CurrentEffectiveAppliedVersion)
	}
	if len(payload.SourceAudit) != payload.Manifest.ExpectedVersionCount {
		t.Fatalf("expected %d source audit rows, got %d", payload.Manifest.ExpectedVersionCount, len(payload.SourceAudit))
	}
	assertMigrationEvidenceFinding(t, payload.Findings, "protected_boundary_applied")

	dbOnlyVersion := payload.Manifest.ExpectedMaxVersion + 1
	if _, err := sqlDB.ExecContext(ctx, `INSERT INTO goose_db_version (version_id, is_applied) VALUES ($1, true)`, dbOnlyVersion); err != nil {
		t.Fatalf("insert synthetic goose ledger version: %v", err)
	}
	payload = runMigrationEvidenceCaptureForDatabase(t, testDB.DSN)
	assertMigrationEvidenceFinding(t, payload.Findings, "db_version_not_in_manifest")
}

func runMigrationEvidenceCaptureForDatabase(t *testing.T, dsn string) OperatorMigrationEvidenceResult {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := operatorRunner{
		stdout: &stdout,
		stderr: &stderr,
		loadConfig: func(string) (config.Config, error) {
			return config.Config{
				ConfigSchemaID: "cartulary.deployment_config.v1",
				Roots: config.RootBindings{
					DatabaseStorage: config.RootBinding{
						BindingKind: "managed_service",
						ServiceRef:  "postgres-primary",
					},
				},
			}, nil
		},
		setupPostgres: func(ctx context.Context, config config.Config) (operatorPostgresPool, error) {
			return pgxpool.New(ctx, dsn)
		},
		now: func() time.Time {
			return time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
		},
	}
	exitCode := runner.runCLI(context.Background(), []string{
		"migration-evidence",
		"capture",
		"-manifest",
		migrationEvidenceManifestPathForTest(t),
	})
	if exitCode != 0 {
		t.Fatalf("migration-evidence capture failed: exit=%d stdout=%s stderr=%s", exitCode, stdout.String(), stderr.String())
	}
	var payload OperatorMigrationEvidenceResult
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode migration evidence payload: %v\nstdout=%s", err, stdout.String())
	}
	return payload
}
