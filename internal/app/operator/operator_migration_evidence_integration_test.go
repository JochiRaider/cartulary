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
	head := migrationEvidenceManifestMaxVersionForTest(t, migrationEvidenceManifestPathForTest(t))
	wantStates := make([]migrationevidence.GooseState, head+1)
	for version := int64(0); version <= head; version++ {
		wantStates[version] = migrationevidence.GooseState{Version: version, IsApplied: true}
	}
	assertMigrationLedgerProjection(t, payload, head, head+1, wantStates)
	assertMigrationLedgerFindings(t, payload.Findings, nil)
	if payload.Manifest.ExpectedMaxVersion != head || payload.Manifest.ExpectedVersionCount != int(head) || len(payload.SourceAudit) != int(head) {
		t.Fatal("manifest/source projection disagrees with authored fixture allocation")
	}
	assertEditableMigrationBaseline(t, payload)

	dbOnlyVersion := head + 1
	later := time.Date(2026, 4, 18, 12, 0, 0, 0, time.UTC)
	earlier := later.Add(-time.Hour)
	insertState := func(version int64, applied bool, stamp time.Time) {
		t.Helper()
		if _, err := sqlDB.ExecContext(ctx, `INSERT INTO public.goose_db_version (version_id, is_applied, tstamp) VALUES ($1, $2, $3)`, version, applied, stamp); err != nil {
			t.Fatalf("insert synthetic goose ledger state: %v", err)
		}
	}
	insertState(dbOnlyVersion, true, later)
	wantStates = append(wantStates, migrationevidence.GooseState{Version: dbOnlyVersion, IsApplied: true, TStamp: later})
	payload = runMigrationEvidenceCaptureForDatabase(t, testDB.DSN).payload
	assertMigrationLedgerProjection(t, payload, dbOnlyVersion, head+2, wantStates)
	assertMigrationLedgerFindings(t, payload.Findings, map[string]int64{"db_version_not_in_manifest": dbOnlyVersion})

	insertState(1, false, later)
	insertState(dbOnlyVersion, false, later)
	wantStates[1] = migrationevidence.GooseState{Version: 1, IsApplied: false, TStamp: later}
	wantStates[dbOnlyVersion].IsApplied = false
	payload = runMigrationEvidenceCaptureForDatabase(t, testDB.DSN).payload
	assertMigrationLedgerProjection(t, payload, head, head+4, wantStates)
	assertMigrationLedgerFindings(t, payload.Findings, map[string]int64{"db_version_not_in_manifest": dbOnlyVersion, "db_applied_version_gap": 1})

	// A later ledger ID wins even with an earlier timestamp. These are evidence
	// fixtures on a disposable database, not an authorized production repair.
	insertState(1, true, earlier)
	insertState(dbOnlyVersion, true, earlier)
	wantStates[1] = migrationevidence.GooseState{Version: 1, IsApplied: true, TStamp: earlier}
	wantStates[dbOnlyVersion] = migrationevidence.GooseState{Version: dbOnlyVersion, IsApplied: true, TStamp: earlier}
	payload = runMigrationEvidenceCaptureForDatabase(t, testDB.DSN).payload
	assertMigrationLedgerProjection(t, payload, dbOnlyVersion, head+6, wantStates)
	assertMigrationLedgerFindings(t, payload.Findings, map[string]int64{"db_version_not_in_manifest": dbOnlyVersion})
}

func assertMigrationLedgerProjection(t *testing.T, payload migrationevidence.Result, head, rows int64, states []migrationevidence.GooseState) {
	t.Helper()
	ledger := payload.GooseLedger
	if !ledger.MetadataPresent || ledger.CurrentEffectiveAppliedVersion != head || ledger.RowCount != rows || len(ledger.LatestEffectiveStates) != len(states) {
		t.Fatalf("unexpected ledger projection: %#v; want head=%d rows=%d states=%d", ledger, head, rows, len(states))
	}
	for index, want := range states {
		got := ledger.LatestEffectiveStates[index]
		if got.Version != want.Version || got.IsApplied != want.IsApplied || got.TStamp.IsZero() || (!want.TStamp.IsZero() && !got.TStamp.Equal(want.TStamp)) {
			t.Fatalf("ledger state %d: got %#v want %#v", index, got, want)
		}
	}
}

func assertMigrationLedgerFindings(t *testing.T, findings []migrationevidence.Finding, expected map[string]int64) {
	t.Helper()
	seen := make(map[string]bool)
	for _, finding := range findings {
		version, ok := expected[finding.ReasonCode]
		if !ok || seen[finding.ReasonCode] || finding.Version == nil || *finding.Version != version || finding.Severity != "blocking" {
			t.Fatalf("unexpected ledger finding: %#v", finding)
		}
		seen[finding.ReasonCode] = true
	}
	if len(seen) != len(expected) {
		t.Fatalf("ledger findings=%v, want %v", seen, expected)
	}
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
	t.Setenv("CARTULARY_POSTGRES_POSTGRES_PRIMARY_RUNTIME_DSN", dsn)
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
