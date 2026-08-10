package operator

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/modules/database_migrations/migrationevidence"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
)

func TestMigrationEvidenceTransport_Unit(t *testing.T) {
	t.Run("parse and validate CLI flags", runMigrationEvidenceCaptureArgsParseAndValidate)
	t.Run("reject invalid CLI inputs", runMigrationEvidenceCaptureArgsRejectsInvalidInputs)
	t.Run("emit one redacted JSON object and close the pool", runMigrationEvidenceCaptureTransport)
	t.Run("preserve exact v2 output bytes", runMigrationEvidenceCaptureV2GoldenDigest)
	t.Run("remain invariant across manifest relocation", runMigrationEvidenceRelocationInvariance)
	t.Run("redact manifest locator failures", runMigrationEvidenceManifestFailureRedaction)
	t.Run("source failure has no external side effects", runMigrationEvidenceSourceFailureNoExternalEffects)
}

func TestMigrationEvidenceSemantics_Unit(t *testing.T) {
	t.Run("audit embedded sources without authorizing rewrite", runMigrationEvidenceCaptureProjectionSemantics)
	t.Run("emit evidence when goose metadata is missing", runMigrationEvidenceCaptureCommandMissingGooseMetadataStillEmitsEvidencePayload)
}

func runMigrationEvidenceCaptureArgsParseAndValidate(t *testing.T) {
	var stderr bytes.Buffer
	result, stop, exitCode := parseMigrationEvidenceCaptureArgs([]string{
		"-source-config",
		"/etc/cartulary/config.toml",
		"-manifest",
		"/srv/cartulary/migration_history_manifest.json",
		"-as-of",
		"2026-04-17T12:00:00Z",
	}, &stderr)
	if stop {
		t.Fatalf("parse stopped: exit=%d stderr=%s", exitCode, stderr.String())
	}
	if result.sourceConfigPath != "/etc/cartulary/config.toml" {
		t.Fatalf("unexpected source config: %q", result.sourceConfigPath)
	}
	if result.manifestPath != "/srv/cartulary/migration_history_manifest.json" {
		t.Fatalf("unexpected manifest path: %q", result.manifestPath)
	}
	if got := result.asOf.Format(time.RFC3339); got != "2026-04-17T12:00:00Z" {
		t.Fatalf("unexpected as-of: %s", got)
	}

	defaulted, stop, exitCode := parseMigrationEvidenceCaptureArgs(nil, &stderr)
	if stop {
		t.Fatalf("defaulted parse stopped: exit=%d stderr=%s", exitCode, stderr.String())
	}
	if defaulted.sourceConfigPath != "" {
		t.Fatalf("unexpected default source config path: %q", defaulted.sourceConfigPath)
	}
	if defaulted.manifestPath != migrationevidence.DefaultManifestPath {
		t.Fatalf("unexpected default manifest path: %q", defaulted.manifestPath)
	}
	if _, stop, exitCode := parseMigrationEvidenceCaptureArgs([]string{"-help"}, &stderr); !stop || exitCode != 0 {
		t.Fatalf("help got stop=%t exit=%d stderr=%s", stop, exitCode, stderr.String())
	}
}

func runMigrationEvidenceCaptureArgsRejectsInvalidInputs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantMessage string
	}{
		{
			name:        "deprecated admin flag",
			args:        []string{"migration-evidence", "capture", "-deployment-admin-email", "not-an-email"},
			wantMessage: "deployment-admin-email",
		},
		{
			name:        "invalid timestamp",
			args:        []string{"migration-evidence", "capture", "-as-of", "yesterday"},
			wantMessage: "as-of must be RFC3339",
		},
		{
			name:        "empty manifest",
			args:        []string{"migration-evidence", "capture", "-manifest", ""},
			wantMessage: "manifest is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stderr bytes.Buffer
			_, stop, exitCode := parseMigrationEvidenceCaptureArgs(test.args[2:], &stderr)
			if !stop || exitCode != 2 {
				t.Fatalf("expected parse exit 2, got stop=%v exit=%d stderr=%s", stop, exitCode, stderr.String())
			}
			if !strings.Contains(stderr.String(), test.wantMessage) {
				t.Fatalf("stderr %q does not contain %q", stderr.String(), test.wantMessage)
			}
		})
	}
}

type migrationEvidenceUnitCapture struct {
	stdout  string
	stderr  string
	payload migrationevidence.Result
	pool    *migrationEvidenceFakePool
}

func captureMigrationEvidenceUnit(t *testing.T) migrationEvidenceUnitCapture {
	t.Helper()
	return captureMigrationEvidenceUnitAtManifest(t, migrationEvidenceManifestPathForTest(t))
}

func captureMigrationEvidenceUnitAtManifest(t *testing.T, manifestPath string) migrationEvidenceUnitCapture {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	t.Setenv("CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN", "postgres://unit-test")
	collectedAt := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	pool := newMigrationEvidenceFakePool(true, migrationEvidenceAppliedStates(migrationEvidenceManifestMaxVersionForTest(t, manifestPath), collectedAt))
	runner := operatorRunner{
		migrationEvidence: migrationEvidenceExecutor{
			transport: operatorTransport{stdout: &stdout, stderr: &stderr},
			source:    dbmigrations.Source,
			loadConfig: func(path string) (configassembly.Loaded, error) {
				if path != "/etc/cartulary/config.toml" {
					t.Fatalf("unexpected source config path: %q", path)
				}
				return migrationEvidenceTestDeployment(t), nil
			},
			setupPostgres: func(context.Context, postgres.Settings) (operatorPostgresPool, error) {
				return pool, nil
			},
			now: func() time.Time {
				return collectedAt
			},
		},
	}

	exitCode := runner.runCLI(context.Background(), []string{
		"migration-evidence",
		"capture",
		"-source-config",
		"/etc/cartulary/config.toml",
		"-manifest",
		manifestPath,
		"-as-of",
		collectedAt.Format(time.RFC3339),
	})
	if exitCode != 0 {
		t.Fatalf("migration-evidence capture failed: exit=%d stdout=%s stderr=%s", exitCode, stdout.String(), stderr.String())
	}

	var payload migrationevidence.Result
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode migration evidence payload: %v\nstdout=%s", err, stdout.String())
	}
	return migrationEvidenceUnitCapture{
		stdout:  stdout.String(),
		stderr:  stderr.String(),
		payload: payload,
		pool:    pool,
	}
}

func runMigrationEvidenceCaptureTransport(t *testing.T) {
	capture := captureMigrationEvidenceUnit(t)
	if !capture.pool.closed {
		t.Fatal("migration-evidence capture did not close its acquired Postgres pool")
	}
	if capture.stderr != "" {
		t.Fatalf("expected no stderr on success, got %s", capture.stderr)
	}
	if got := strings.Count(capture.stdout, "\n"); got != 1 {
		t.Fatalf("operator JSON must be one object followed by LF, got %d newlines: %s", got, capture.stdout)
	}
	for _, forbidden := range []string{
		"postgres://cartulary:secret@db.example.test/cartulary",
		"secret",
		"db.example.test",
		"/srv/cartulary/secrets/postgres",
		migrationEvidenceManifestPathForTest(t),
		`"path"`,
	} {
		if strings.Contains(capture.stdout, forbidden) {
			t.Fatalf("migration evidence leaked forbidden value %q in stdout %s", forbidden, capture.stdout)
		}
	}
	if capture.payload.SchemaID != migrationevidence.SchemaID {
		t.Fatalf("unexpected schema_id: %q", capture.payload.SchemaID)
	}
}

func runMigrationEvidenceCaptureV2GoldenDigest(t *testing.T) {
	capture := captureMigrationEvidenceUnit(t)
	digest := sha256.Sum256([]byte(capture.stdout))
	const wantDigest = "29bd46ab2d3d6a2e10c7d01972773baa070614dcf4ec75492b7c3241fdc45df4"
	if got := fmt.Sprintf("%x", digest); got != wantDigest {
		t.Fatalf("v2 migration evidence digest = %s, want %s", got, wantDigest)
	}
}

func runMigrationEvidenceRelocationInvariance(t *testing.T) {
	original := migrationEvidenceManifestPathForTest(t)
	body, err := os.ReadFile(original)
	if err != nil {
		t.Fatalf("read canonical manifest: %v", err)
	}
	paths := make([]string, 2)
	for index := range paths {
		paths[index] = filepath.Join(t.TempDir(), fmt.Sprintf("relocated-%d.json", index+1))
		if err := os.WriteFile(paths[index], body, 0o600); err != nil {
			t.Fatalf("write relocated manifest: %v", err)
		}
	}
	left := captureMigrationEvidenceUnitAtManifest(t, paths[0])
	right := captureMigrationEvidenceUnitAtManifest(t, paths[1])
	if left.stdout != right.stdout {
		t.Fatalf("relocated logical evidence differs\nleft: %s\nright: %s", left.stdout, right.stdout)
	}
	for _, locator := range paths {
		if strings.Contains(left.stdout, locator) || strings.Contains(right.stdout, locator) {
			t.Fatalf("relocated evidence disclosed locator %q", locator)
		}
	}
}

func runMigrationEvidenceManifestFailureRedaction(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	t.Setenv("CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN", "postgres://unit-test")
	secretPath := filepath.Join(t.TempDir(), "operator-private-manifest.json")
	pool := newMigrationEvidenceFakePool(true, nil)
	runner := operatorRunner{
		migrationEvidence: migrationEvidenceExecutor{
			transport: operatorTransport{stdout: &stdout, stderr: &stderr},
			source:    dbmigrations.Source,
			loadConfig: func(string) (configassembly.Loaded, error) {
				return migrationEvidenceTestDeployment(t), nil
			},
			setupPostgres: func(context.Context, postgres.Settings) (operatorPostgresPool, error) {
				return pool, nil
			},
			now: time.Now,
		},
	}
	exitCode := runner.runCLI(context.Background(), []string{
		"migration-evidence", "capture", "-manifest", secretPath,
	})
	if exitCode != 1 || stdout.Len() != 0 || !pool.closed {
		t.Fatalf("manifest failure lifecycle mismatch: exit=%d stdout=%q closed=%t", exitCode, stdout.String(), pool.closed)
	}
	if strings.Contains(stderr.String(), secretPath) || strings.Contains(stderr.String(), "operator-private-manifest") {
		t.Fatalf("manifest failure disclosed locator: %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "migration evidence manifest unavailable") {
		t.Fatalf("manifest failure omitted safe reason: %s", stderr.String())
	}
}

func runMigrationEvidenceSourceFailureNoExternalEffects(t *testing.T) {
	wantErr := errors.New("source unavailable")
	configCalls := 0
	postgresCalls := 0
	executor := migrationEvidenceExecutor{
		source: func() (*database_migrations.Source, error) {
			return nil, wantErr
		},
		loadConfig: func(string) (configassembly.Loaded, error) {
			configCalls++
			return configassembly.Loaded{}, nil
		},
		setupPostgres: func(context.Context, postgres.Settings) (operatorPostgresPool, error) {
			postgresCalls++
			return nil, errors.New("postgres must not open")
		},
	}
	err := executor.capture(context.Background(), migrationEvidenceCaptureArgs{manifestPath: "unused.json"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("source failure = %v, want wrapped %v", err, wantErr)
	}
	if configCalls != 0 || postgresCalls != 0 {
		t.Fatalf("source failure reached external dependencies: config=%d postgres=%d", configCalls, postgresCalls)
	}
}

func runMigrationEvidenceCaptureProjectionSemantics(t *testing.T) {
	payload := captureMigrationEvidenceUnit(t).payload
	if payload.SchemaID != migrationevidence.SchemaID {
		t.Fatalf("unexpected schema_id: %q", payload.SchemaID)
	}
	if !payload.EvidenceOnly || payload.RewriteAuthorized {
		t.Fatalf("payload must remain evidence-only and non-authorizing: %#v", payload)
	}
	wantCollectedAt := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	if !payload.CollectedAt.Equal(wantCollectedAt) {
		t.Fatalf("collected_at was not preserved: got %s want %s", payload.CollectedAt, wantCollectedAt)
	}
	if payload.DatabaseBinding.BindingKind != "managed_service" || payload.DatabaseBinding.ServiceRef != "postgres-primary" {
		t.Fatalf("unexpected database binding summary: %#v", payload.DatabaseBinding)
	}
	if payload.GooseLedger.CurrentEffectiveAppliedVersion != payload.Manifest.ExpectedMaxVersion {
		t.Fatalf("unexpected current effective applied version: got %d want %d", payload.GooseLedger.CurrentEffectiveAppliedVersion, payload.Manifest.ExpectedMaxVersion)
	}
	if len(payload.SourceAudit) != payload.Manifest.ExpectedVersionCount {
		t.Fatalf("expected %d embedded migration source audit rows, got %d", payload.Manifest.ExpectedVersionCount, len(payload.SourceAudit))
	}
	assertMigrationEvidenceFinding(t, payload.Findings, "protected_boundary_applied")
}

func runMigrationEvidenceCaptureCommandMissingGooseMetadataStillEmitsEvidencePayload(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	t.Setenv("CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN", "postgres://unit-test")
	pool := newMigrationEvidenceFakePool(false, nil)
	manifestPath := migrationEvidenceManifestPathForTest(t)
	runner := operatorRunner{
		migrationEvidence: migrationEvidenceExecutor{
			transport: operatorTransport{stdout: &stdout, stderr: &stderr},
			source:    dbmigrations.Source,
			loadConfig: func(string) (configassembly.Loaded, error) {
				return migrationEvidenceTestDeployment(t), nil
			},
			setupPostgres: func(context.Context, postgres.Settings) (operatorPostgresPool, error) {
				return pool, nil
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
		manifestPath,
	})
	if exitCode != 0 {
		t.Fatalf("migration-evidence capture failed: exit=%d stdout=%s stderr=%s", exitCode, stdout.String(), stderr.String())
	}
	if !pool.closed {
		t.Fatal("migration-evidence capture did not close its acquired Postgres pool")
	}

	var payload migrationevidence.Result
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode migration evidence payload: %v\nstdout=%s", err, stdout.String())
	}
	if payload.GooseLedger.MetadataPresent {
		t.Fatalf("expected metadata_present=false, got %#v", payload.GooseLedger)
	}
	if payload.RewriteAuthorized {
		t.Fatalf("missing metadata must not authorize rewrite: %#v", payload)
	}
	assertMigrationEvidenceFinding(t, payload.Findings, "migration_metadata_missing")
}

func assertMigrationEvidenceFinding(t *testing.T, findings []migrationevidence.Finding, reasonCode string) {
	t.Helper()
	for _, finding := range findings {
		if finding.ReasonCode == reasonCode {
			return
		}
	}
	t.Fatalf("finding %q not present in %#v", reasonCode, findings)
}

func migrationEvidenceTestDeployment(t testing.TB) configassembly.Loaded {
	t.Helper()
	return configtest.LoadFixture(t, []string{"config", "valid.toml"}, map[string]string{
		"CARTULARY__DEPLOYMENT_PROFILE":                    "on_prem",
		"CARTULARY__ROOTS__DATABASE_STORAGE__BINDING_KIND": "managed_service",
		"CARTULARY__ROOTS__DATABASE_STORAGE__PATH":         "",
		"CARTULARY__ROOTS__DATABASE_STORAGE__SERVICE_REF":  "postgres-primary",
	})
}

func migrationEvidenceManifestPathForTest(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for {
		candidate := filepath.Join(dir, migrationevidence.DefaultManifestPath)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find %s above %s", migrationevidence.DefaultManifestPath, dir)
		}
		dir = parent
	}
}

func migrationEvidenceManifestMaxVersionForTest(t *testing.T, manifestPath string) int64 {
	t.Helper()
	body, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read migration evidence manifest: %v", err)
	}
	var payload struct {
		Entries []struct {
			Version int64 `json:"version"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode migration evidence manifest: %v", err)
	}
	var maxVersion int64
	for _, entry := range payload.Entries {
		if entry.Version > maxVersion {
			maxVersion = entry.Version
		}
	}
	if maxVersion == 0 {
		t.Fatal("migration evidence manifest has no migration versions")
	}
	return maxVersion
}

type migrationEvidenceFakePool struct {
	metadataPresent bool
	states          []migrationevidence.GooseState
	rowCount        int64
	closed          bool
}

func newMigrationEvidenceFakePool(metadataPresent bool, states []migrationevidence.GooseState) *migrationEvidenceFakePool {
	return &migrationEvidenceFakePool{
		metadataPresent: metadataPresent,
		states:          states,
		rowCount:        int64(len(states)),
	}
}

func (pool *migrationEvidenceFakePool) Close() {
	pool.closed = true
}

func (pool *migrationEvidenceFakePool) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected Exec call")
}

func (pool *migrationEvidenceFakePool) Query(_ context.Context, sql string, _ ...any) (pgx.Rows, error) {
	if strings.Contains(sql, "FROM goose_db_version") && strings.Contains(sql, "row_number()") {
		rows := make([][]any, 0, len(pool.states))
		for _, state := range pool.states {
			rows = append(rows, []any{state.Version, state.IsApplied, state.TStamp})
		}
		return &migrationEvidenceFakeRows{rows: rows}, nil
	}
	return nil, fmt.Errorf("unexpected Query SQL: %s", sql)
}

func (pool *migrationEvidenceFakePool) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	switch {
	case strings.Contains(sql, "to_regclass('public.goose_db_version')"):
		return migrationEvidenceFakeRow{values: []any{pool.metadataPresent}}
	case strings.Contains(sql, "COUNT(*)::bigint FROM goose_db_version"):
		return migrationEvidenceFakeRow{values: []any{pool.rowCount}}
	default:
		return migrationEvidenceFakeRow{err: fmt.Errorf("unexpected QueryRow SQL: %s", sql)}
	}
}

func (pool *migrationEvidenceFakePool) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	return nil, errors.New("unexpected BeginTx call")
}

type migrationEvidenceFakeRow struct {
	values []any
	err    error
}

func (row migrationEvidenceFakeRow) Scan(dest ...any) error {
	if row.err != nil {
		return row.err
	}
	return scanMigrationEvidenceFakeValues(dest, row.values)
}

type migrationEvidenceFakeRows struct {
	rows   [][]any
	index  int
	closed bool
}

func (rows *migrationEvidenceFakeRows) Close() {
	rows.closed = true
}

func (rows *migrationEvidenceFakeRows) Err() error {
	return nil
}

func (rows *migrationEvidenceFakeRows) CommandTag() pgconn.CommandTag {
	return pgconn.CommandTag{}
}

func (rows *migrationEvidenceFakeRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}

func (rows *migrationEvidenceFakeRows) Next() bool {
	return rows.index < len(rows.rows)
}

func (rows *migrationEvidenceFakeRows) Scan(dest ...any) error {
	if rows.index >= len(rows.rows) {
		return errors.New("scan without current row")
	}
	values := rows.rows[rows.index]
	rows.index++
	return scanMigrationEvidenceFakeValues(dest, values)
}

func (rows *migrationEvidenceFakeRows) Values() ([]any, error) {
	if rows.index == 0 || rows.index > len(rows.rows) {
		return nil, errors.New("no current row")
	}
	return rows.rows[rows.index-1], nil
}

func (rows *migrationEvidenceFakeRows) RawValues() [][]byte {
	return nil
}

func (rows *migrationEvidenceFakeRows) Conn() *pgx.Conn {
	return nil
}

func scanMigrationEvidenceFakeValues(dest []any, values []any) error {
	if len(dest) != len(values) {
		return fmt.Errorf("scan destination count %d does not match value count %d", len(dest), len(values))
	}
	for index := range dest {
		if err := assignMigrationEvidenceFakeValue(dest[index], values[index]); err != nil {
			return fmt.Errorf("scan column %d: %w", index, err)
		}
	}
	return nil
}

func assignMigrationEvidenceFakeValue(dest any, value any) error {
	switch target := dest.(type) {
	case *string:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
		*target = v
	case *time.Time:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("expected time.Time, got %T", value)
		}
		*target = v
	case *bool:
		v, ok := value.(bool)
		if !ok {
			return fmt.Errorf("expected bool, got %T", value)
		}
		*target = v
	case *int64:
		v, ok := value.(int64)
		if !ok {
			return fmt.Errorf("expected int64, got %T", value)
		}
		*target = v
	case **time.Time:
		if value == nil {
			*target = nil
			return nil
		}
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("expected nullable time.Time, got %T", value)
		}
		copyValue := v
		*target = &copyValue
	case *[]byte:
		if value == nil {
			*target = nil
			return nil
		}
		v, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("expected []byte, got %T", value)
		}
		*target = append([]byte(nil), v...)
	default:
		return fmt.Errorf("unsupported destination %T", dest)
	}
	return nil
}

func migrationEvidenceAppliedStates(maxVersion int64, tstamp time.Time) []migrationevidence.GooseState {
	states := make([]migrationevidence.GooseState, 0, maxVersion)
	for version := int64(1); version <= maxVersion; version++ {
		states = append(states, migrationevidence.GooseState{
			Version:   version,
			IsApplied: true,
			TStamp:    tstamp,
		})
	}
	return states
}
