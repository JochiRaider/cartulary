package administrativeaudit_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestAdministrativeAuditProjectionPersistence_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	testDB := harness.PrepareIsolatedDatabaseT(t, "administrative-audit-projection")
	pool, err := pgxpool.New(context.Background(), testDB.DSN)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(pool.Close)

	now := time.Date(2026, time.July, 25, 19, 45, 0, 0, time.UTC)
	targetID := uuid.NewString()
	eventID, err := administrativeaudit.Append(context.Background(), pool, administrativeaudit.RawEvent{
		EventSource: "operator.recovery.backup_create",
		EventKind:   "operator_recovery_succeeded",
		After:       map[string]any{"backup_set_id": targetID},
		OccurredAt:  now,
	}, administrativeaudit.Event{
		ScopeKind:  administrativeaudit.ScopeDeployment,
		OccurredAt: now,
		ActorKind:  administrativeaudit.ActorOperator,
		Source:     administrativeaudit.SourceOperator,
		ActionCode: administrativeaudit.ActionBackupCreated,
		TargetKind: administrativeaudit.TargetBackupSet,
		TargetID:   &targetID,
		Changes: []administrativeaudit.Change{
			administrativeaudit.Visible("result", nil, "succeeded"),
			administrativeaudit.Visible("backup_set_id", nil, targetID),
		},
	})
	if err != nil {
		t.Fatalf("append projected event: %v", err)
	}

	var rawOccurredAt time.Time
	var projectedOccurredAt time.Time
	var changesJSON []byte
	if err := pool.QueryRow(context.Background(), `
SELECT raw.created_at, projected.occurred_at, projected.changes
  FROM deployment_admin_audit_events AS raw
  JOIN administrative_audit_projections AS projected ON projected.audit_event_id = raw.id
 WHERE raw.id = $1
`, eventID).Scan(&rawOccurredAt, &projectedOccurredAt, &changesJSON); err != nil {
		t.Fatalf("read raw/projection pair: %v", err)
	}
	if !rawOccurredAt.Equal(projectedOccurredAt) || !rawOccurredAt.Equal(now) {
		t.Fatalf("raw/projection occurrence mismatch: raw=%s projected=%s want=%s", rawOccurredAt, projectedOccurredAt, now)
	}
	var changes []administrativeaudit.Change
	if err := json.Unmarshal(changesJSON, &changes); err != nil {
		t.Fatalf("decode projected changes: %v", err)
	}
	if got := changes[0].FieldPath; got != "backup_set_id" {
		t.Fatalf("changes were not persisted in field order: got first %q", got)
	}

	for _, statement := range []string{
		`UPDATE deployment_admin_audit_events SET event_kind = 'tampered' WHERE id = $1`,
		`DELETE FROM deployment_admin_audit_events WHERE id = $1`,
		`UPDATE administrative_audit_projections SET source = 'system' WHERE audit_event_id = $1`,
		`DELETE FROM administrative_audit_projections WHERE audit_event_id = $1`,
	} {
		if _, err := pool.Exec(context.Background(), statement, eventID); err == nil || !strings.Contains(err.Error(), "immutable") {
			t.Fatalf("expected immutable-row rejection for %q, got %v", statement, err)
		}
	}

	unsafeTargetID := uuid.NewString()
	if _, err := administrativeaudit.Append(context.Background(), pool, administrativeaudit.RawEvent{
		EventSource: "test",
		EventKind:   "unsafe",
		OccurredAt:  now,
	}, administrativeaudit.Event{
		ScopeKind:  administrativeaudit.ScopeDeployment,
		OccurredAt: now,
		ActorKind:  administrativeaudit.ActorSystem,
		Source:     administrativeaudit.SourceSystem,
		ActionCode: administrativeaudit.ActionUserCreated,
		TargetKind: administrativeaudit.TargetUser,
		TargetID:   &unsafeTargetID,
		Changes:    []administrativeaudit.Change{administrativeaudit.Visible("password", nil, "unsafe")},
	}); err == nil {
		t.Fatal("expected unsafe projection to fail")
	}
	var unsafeRawCount int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM deployment_admin_audit_events WHERE event_kind = 'unsafe'`).Scan(&unsafeRawCount); err != nil {
		t.Fatalf("count unsafe raw events: %v", err)
	}
	if unsafeRawCount != 0 {
		t.Fatalf("invalid projection left %d raw rows", unsafeRawCount)
	}

	if !recovery.IsAuthoritativePostgresSnapshotTable("administrative_audit_projections") {
		t.Fatal("administrative audit projections must be included in deployment backup snapshots")
	}
}

func TestAdministrativeAuditLegacyProjectionCleanupPreservesRawJournal_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.MigrationDatabaseT(t, "administrative-audit-cleanup", "up-to", "38")
	insertLegacyRaw(t, db, "auth", "user_created")
	insertLegacyRaw(t, db, "network_flow", "network_flow_import_started")

	if _, err := postgres.Migrate(context.Background(), db, dbmigrations.Source(), "up-to", "39"); err != nil {
		t.Fatalf("apply administrative audit migration: %v", err)
	}
	var projectedCount int
	if err := db.QueryRowContext(context.Background(), `
SELECT count(*)
  FROM administrative_audit_projections
 WHERE action_code = 'legacy_administrative_event'
`).Scan(&projectedCount); err != nil {
		t.Fatalf("count legacy projections before cleanup: %v", err)
	}
	if projectedCount != 1 {
		t.Fatalf("expected migration 39 fixture to contain one legacy projection, got %d", projectedCount)
	}

	rawCountBefore, rawDigestBefore := rawJournalIdentity(t, db)
	if _, err := postgres.Migrate(context.Background(), db, dbmigrations.Source(), "up-to", "40"); err != nil {
		t.Fatalf("apply administrative audit cleanup migration: %v", err)
	}
	rawCountAfter, rawDigestAfter := rawJournalIdentity(t, db)
	if rawCountAfter != rawCountBefore || rawDigestAfter != rawDigestBefore {
		t.Fatalf(
			"raw administrative audit journal changed during cleanup: before=(%d,%s) after=(%d,%s)",
			rawCountBefore,
			rawDigestBefore,
			rawCountAfter,
			rawDigestAfter,
		)
	}
	if err := db.QueryRowContext(context.Background(), `
SELECT count(*)
  FROM administrative_audit_projections
 WHERE action_code = 'legacy_administrative_event'
    OR target_kind = 'legacy_administrative_event'
`).Scan(&projectedCount); err != nil {
		t.Fatalf("count legacy projections after cleanup: %v", err)
	}
	if projectedCount != 0 {
		t.Fatalf("legacy projections remain after cleanup: %d", projectedCount)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO administrative_audit_projections (
    audit_event_id, scope_kind, occurred_at, actor_kind, source,
    action_code, target_kind, changes
)
SELECT id, 'deployment', created_at, 'system', 'system',
       'legacy_administrative_event', 'legacy_administrative_event', '[]'::jsonb
  FROM deployment_admin_audit_events
 LIMIT 1
`); err == nil {
		t.Fatal("cleaned projection constraints accepted legacy vocabulary")
	}
}

func rawJournalIdentity(t testing.TB, db *sql.DB) (int, string) {
	t.Helper()
	var count int
	var digest string
	if err := db.QueryRowContext(context.Background(), `
SELECT
    count(*),
    md5(COALESCE(string_agg(row_to_json(raw)::text, E'\n' ORDER BY raw.id), ''))
  FROM deployment_admin_audit_events AS raw
`).Scan(&count, &digest); err != nil {
		t.Fatalf("read raw administrative audit identity: %v", err)
	}
	return count, digest
}

func insertLegacyRaw(t testing.TB, db *sql.DB, source string, kind string) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO deployment_admin_audit_events (event_source, event_kind)
VALUES ($1, $2)
`, source, kind); err != nil {
		t.Fatalf("insert legacy raw event %s: %v", kind, err)
	}
}
