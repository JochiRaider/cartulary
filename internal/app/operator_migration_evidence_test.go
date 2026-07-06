package app

import (
	"bytes"
	"context"
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

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/migrationevidence"
)

func TestPhase0_MigrationEvidenceCommand_U_0_10(t *testing.T) {
	t.Run("parse and validate CLI flags", runOperatorMigrationEvidenceCaptureArgsParseAndValidate)
	t.Run("reject invalid CLI inputs", runOperatorMigrationEvidenceCaptureArgsRejectsInvalidInputs)
	t.Run("emit redacted evidence-only JSON", runOperatorMigrationEvidenceCaptureCommandOutputsRedactedEvidenceOnlyJSON)
	t.Run("emit evidence when goose metadata is missing", runOperatorMigrationEvidenceCaptureCommandMissingGooseMetadataStillEmitsEvidencePayload)
}

func runOperatorMigrationEvidenceCaptureArgsParseAndValidate(t *testing.T) {
	var stderr bytes.Buffer
	result := parseOperatorCLIArgs([]string{
		"migration-evidence",
		"capture",
		"-deployment-admin-email",
		"admin@example.test",
		"-source-config",
		"/etc/cartulary/config.toml",
		"-manifest",
		"/srv/cartulary/migration_history_manifest.json",
		"-as-of",
		"2026-04-17T12:00:00Z",
	}, &stderr)
	if result.stop {
		t.Fatalf("parse stopped: exit=%d stderr=%s", result.exitCode, stderr.String())
	}
	if result.command != "migration-evidence capture" {
		t.Fatalf("unexpected command: %q", result.command)
	}
	if result.email != "admin@example.test" {
		t.Fatalf("unexpected email: %q", result.email)
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

	defaulted := parseOperatorCLIArgs([]string{
		"migration-evidence",
		"capture",
	}, &stderr)
	if defaulted.stop {
		t.Fatalf("defaulted parse stopped: exit=%d stderr=%s", defaulted.exitCode, stderr.String())
	}
	if defaulted.email != "" {
		t.Fatalf("unexpected default email authority: %q", defaulted.email)
	}
	if defaulted.manifestPath != defaultMigrationEvidenceManifestPath {
		t.Fatalf("unexpected default manifest path: %q", defaulted.manifestPath)
	}
}

func runOperatorMigrationEvidenceCaptureArgsRejectsInvalidInputs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantMessage string
	}{
		{
			name:        "invalid email",
			args:        []string{"migration-evidence", "capture", "-deployment-admin-email", "not-an-email"},
			wantMessage: "deployment-admin-email must be an email address",
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
			result := parseOperatorCLIArgs(test.args, &stderr)
			if !result.stop || result.exitCode != 2 {
				t.Fatalf("expected parse exit 2, got stop=%v exit=%d stderr=%s", result.stop, result.exitCode, stderr.String())
			}
			if !strings.Contains(stderr.String(), test.wantMessage) {
				t.Fatalf("stderr %q does not contain %q", stderr.String(), test.wantMessage)
			}
		})
	}
}

func runOperatorMigrationEvidenceCaptureCommandOutputsRedactedEvidenceOnlyJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	collectedAt := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	pool := newMigrationEvidenceFakePool(true, migrationEvidenceAppliedStates(40, collectedAt))
	manifestPath := migrationEvidenceManifestPathForTest(t)
	runner := operatorRunner{
		stdout: &stdout,
		stderr: &stderr,
		loadConfig: func(path string) (config.Config, error) {
			if path != "/etc/cartulary/config.toml" {
				t.Fatalf("unexpected source config path: %q", path)
			}
			return config.Config{
				ConfigSchemaID: "cartulary.deployment_config.v1",
				Roots: config.RootBindings{
					DatabaseStorage: config.RootBinding{
						BindingKind: "managed_service",
						ServiceRef:  "postgres-primary",
						Path:        "/srv/cartulary/secrets/postgres",
					},
				},
			}, nil
		},
		setupPostgres: func(context.Context, config.Config) (operatorPostgresPool, error) {
			return pool, nil
		},
		now: func() time.Time {
			return collectedAt
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
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr on success, got %s", stderr.String())
	}
	if !strings.Contains(stdout.String(), "\n  \"schema_id\": \""+migrationevidence.SchemaID+"\"") {
		t.Fatalf("operator JSON is not indented as expected: %s", stdout.String())
	}
	for _, forbidden := range []string{
		"postgres://cartulary:secret@db.example.test/cartulary",
		"secret",
		"db.example.test",
		"/srv/cartulary/secrets/postgres",
	} {
		if strings.Contains(stdout.String(), forbidden) {
			t.Fatalf("migration evidence leaked forbidden value %q in stdout %s", forbidden, stdout.String())
		}
	}

	var payload OperatorMigrationEvidenceResult
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode migration evidence payload: %v\nstdout=%s", err, stdout.String())
	}
	if payload.SchemaID != migrationevidence.SchemaID {
		t.Fatalf("unexpected schema_id: %q", payload.SchemaID)
	}
	if !payload.EvidenceOnly || payload.RewriteAuthorized {
		t.Fatalf("payload must remain evidence-only and non-authorizing: %#v", payload)
	}
	if payload.DatabaseBinding.BindingKind != "managed_service" || payload.DatabaseBinding.ServiceRef != "postgres-primary" {
		t.Fatalf("unexpected database binding summary: %#v", payload.DatabaseBinding)
	}
	if payload.GooseLedger.CurrentEffectiveAppliedVersion != 40 {
		t.Fatalf("unexpected current effective applied version: %d", payload.GooseLedger.CurrentEffectiveAppliedVersion)
	}
	if len(payload.SourceAudit) < 40 {
		t.Fatalf("expected embedded migration source audit through current head, got %d rows", len(payload.SourceAudit))
	}
	assertMigrationEvidenceFinding(t, payload.Findings, "protected_boundary_applied")
}

func runOperatorMigrationEvidenceCaptureCommandMissingGooseMetadataStillEmitsEvidencePayload(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	pool := newMigrationEvidenceFakePool(false, nil)
	manifestPath := migrationEvidenceManifestPathForTest(t)
	runner := operatorRunner{
		stdout: &stdout,
		stderr: &stderr,
		loadConfig: func(string) (config.Config, error) {
			return config.Config{
				ConfigSchemaID: "cartulary.deployment_config.v1",
				Roots: config.RootBindings{
					DatabaseStorage: config.RootBinding{BindingKind: "managed_service", ServiceRef: "postgres-primary"},
				},
			}, nil
		},
		setupPostgres: func(context.Context, config.Config) (operatorPostgresPool, error) {
			return pool, nil
		},
		now: func() time.Time {
			return time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
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

	var payload OperatorMigrationEvidenceResult
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

func assertMigrationEvidenceFinding(t *testing.T, findings []OperatorMigrationEvidenceFinding, reasonCode string) {
	t.Helper()
	for _, finding := range findings {
		if finding.ReasonCode == reasonCode {
			return
		}
	}
	t.Fatalf("finding %q not present in %#v", reasonCode, findings)
}

func migrationEvidenceManifestPathForTest(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for {
		candidate := filepath.Join(dir, defaultMigrationEvidenceManifestPath)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find %s above %s", defaultMigrationEvidenceManifestPath, dir)
		}
		dir = parent
	}
}

type migrationEvidenceFakePool struct {
	metadataPresent bool
	states          []OperatorMigrationEvidenceGooseState
	rowCount        int64
	closed          bool
}

func newMigrationEvidenceFakePool(metadataPresent bool, states []OperatorMigrationEvidenceGooseState) *migrationEvidenceFakePool {
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

func migrationEvidenceAppliedStates(maxVersion int64, tstamp time.Time) []OperatorMigrationEvidenceGooseState {
	states := make([]OperatorMigrationEvidenceGooseState, 0, maxVersion)
	for version := int64(1); version <= maxVersion; version++ {
		states = append(states, OperatorMigrationEvidenceGooseState{
			Version:   version,
			IsApplied: true,
			TStamp:    tstamp,
		})
	}
	return states
}
