package operator

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/modules/database_migrations/migrationevidence"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestMigrationEvidenceTransport_Integration(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "operator-migration-evidence-transport")
	capture := runMigrationEvidenceCaptureForDatabase(t, testDB.DSN)
	if capture.stderr != "" || strings.Count(capture.stdout, "\n") != 1 {
		t.Fatalf("expected one JSON object plus LF and empty stderr: stdout=%q stderr=%q", capture.stdout, capture.stderr)
	}
	if !capture.poolClosed {
		t.Fatal("migration-evidence transport did not close its acquired Postgres pool")
	}
	if capture.payload.SchemaID != migrationevidence.SchemaID {
		t.Fatalf("unexpected schema_id: %q", capture.payload.SchemaID)
	}
}

func TestMigrationEvidenceSemantics_Integration(t *testing.T) {
	ctx := context.Background()
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "operator-migration-evidence-semantics")

	sqlDB, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open migration evidence database: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	payload := runMigrationEvidenceCaptureForDatabase(t, testDB.DSN).payload
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
	payload = runMigrationEvidenceCaptureForDatabase(t, testDB.DSN).payload
	assertMigrationEvidenceFinding(t, payload.Findings, "db_version_not_in_manifest")
}

type migrationEvidenceIntegrationCapture struct {
	stdout     string
	stderr     string
	payload    migrationevidence.Result
	poolClosed bool
}

type migrationEvidenceTrackingPool struct {
	*pgxpool.Pool
	closed bool
}

func (pool *migrationEvidenceTrackingPool) Close() {
	pool.closed = true
	pool.Pool.Close()
}

func runMigrationEvidenceCaptureForDatabase(t *testing.T, dsn string) migrationEvidenceIntegrationCapture {
	t.Helper()
	t.Setenv("CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN", dsn)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var acquiredPool *migrationEvidenceTrackingPool
	runner := operatorRunner{
		migrationEvidence: migrationEvidenceExecutor{
			transport: operatorTransport{stdout: &stdout, stderr: &stderr},
			source:    dbmigrations.Source,
			loadConfig: func(string) (configassembly.Loaded, error) {
				return migrationEvidenceTestDeployment(t), nil
			},
			setupPostgres: func(ctx context.Context, settings postgres.Settings) (operatorPostgresPool, error) {
				pool, err := pgxpool.New(ctx, dsn)
				if err != nil {
					return nil, err
				}
				acquiredPool = &migrationEvidenceTrackingPool{Pool: pool}
				return acquiredPool, nil
			},
			now: func() time.Time {
				return time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
			},
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
	var payload migrationevidence.Result
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode migration evidence payload: %v\nstdout=%s", err, stdout.String())
	}
	return migrationEvidenceIntegrationCapture{
		stdout:     stdout.String(),
		stderr:     stderr.String(),
		payload:    payload,
		poolClosed: acquiredPool != nil && acquiredPool.closed,
	}
}
