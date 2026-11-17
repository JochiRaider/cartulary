package administrativeaudit_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
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

	// Hold a reader before Append so its committed pair must be visible through
	// a different physical connection, not a savepoint or the writer's session.
	reader, err := pool.Acquire(context.Background())
	if err != nil {
		t.Fatalf("acquire independent reader: %v", err)
	}
	defer reader.Release()
	eventID, now := appendAuditFixture(t, pool)

	var rawOccurredAt time.Time
	var projectedOccurredAt time.Time
	var changesJSON []byte
	if err := reader.QueryRow(context.Background(), `
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
	gotPaths := make([]string, 0, len(changes))
	for _, change := range changes {
		gotPaths = append(gotPaths, change.FieldPath)
	}
	if want := []string{"backup_set_id", "result"}; !reflect.DeepEqual(gotPaths, want) {
		t.Fatalf("changes were not persisted whole in field order: got %v, want %v", gotPaths, want)
	}

}

func TestAdministrativeAuditProjectionStore_Integration(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "administrative-audit-store")
	eventID, now := appendAuditFixture(t, db)

	for _, statement := range []string{
		`UPDATE deployment_admin_audit_events SET event_kind = 'tampered' WHERE id = $1`,
		`DELETE FROM deployment_admin_audit_events WHERE id = $1`,
		`UPDATE administrative_audit_projections SET source = 'system' WHERE audit_event_id = $1`,
		`DELETE FROM administrative_audit_projections WHERE audit_event_id = $1`,
	} {
		// A rejected SQL statement aborts its transaction. Isolate that expected
		// rejection in a savepoint and settle it before borrowing the connection.
		tx, err := db.BeginTx(context.Background(), pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin rejection probe: %v", err)
		}
		_, rejection := tx.Exec(context.Background(), statement, eventID)
		if err := tx.Rollback(context.Background()); err != nil {
			t.Fatalf("settle rejection probe: %v", err)
		}
		if rejection == nil || !strings.Contains(rejection.Error(), "immutable") {
			t.Fatalf("expected immutable-row rejection for %q, got %v", statement, rejection)
		}
	}

	unsafeTargetID := uuid.NewString()
	if _, err := administrativeaudit.Append(context.Background(), db, administrativeaudit.RawEvent{
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
	}); !errors.Is(err, administrativeaudit.ErrUnsafeChanges) {
		t.Fatalf("expected unsafe projection to fail validation, got %v", err)
	}
	var unsafeRawCount int
	if err := db.QueryRow(context.Background(), `SELECT count(*) FROM deployment_admin_audit_events WHERE event_kind = 'unsafe'`).Scan(&unsafeRawCount); err != nil {
		t.Fatalf("count unsafe raw events: %v", err)
	}
	if unsafeRawCount != 0 {
		t.Fatalf("invalid projection left %d raw rows", unsafeRawCount)
	}

	// REQ-02-257: a projection-insert failure after the raw insert rolls the
	// raw event back with it.
	countRaw := func(query func(context.Context, string, ...any) pgx.Row) int {
		t.Helper()
		var count int
		if err := query(context.Background(), `SELECT count(*) FROM deployment_admin_audit_events WHERE event_kind = 'operator_recovery_succeeded'`).Scan(&count); err != nil {
			t.Fatalf("count raw recovery events: %v", err)
		}
		return count
	}
	baseline := countRaw(db.QueryRow)
	faultDB := &projectionFaultDB{DB: db, countRaw: countRaw, rawAtFault: -1}
	faultTargetID := uuid.NewString()
	_, faultErr := administrativeaudit.Append(context.Background(), faultDB, administrativeaudit.RawEvent{
		EventSource: "operator.recovery.backup_create",
		EventKind:   "operator_recovery_succeeded",
		After:       map[string]any{"backup_set_id": faultTargetID},
		OccurredAt:  now,
	}, administrativeaudit.Event{
		ScopeKind:  administrativeaudit.ScopeDeployment,
		OccurredAt: now,
		ActorKind:  administrativeaudit.ActorOperator,
		Source:     administrativeaudit.SourceOperator,
		ActionCode: administrativeaudit.ActionBackupCreated,
		TargetKind: administrativeaudit.TargetBackupSet,
		TargetID:   &faultTargetID,
		Changes: []administrativeaudit.Change{
			administrativeaudit.Visible("result", nil, "succeeded"),
			administrativeaudit.Visible("backup_set_id", nil, faultTargetID),
		},
	})
	if faultErr == nil || !strings.Contains(faultErr.Error(), "insert administrative audit projection") || !errors.Is(faultErr, errInjectedProjectionFault) {
		t.Fatalf("projection fault was not returned: %v", faultErr)
	}
	if faultDB.rawAtFault != baseline+1 {
		t.Fatalf("raw rows inside the transaction at the projection fault = %d, want %d", faultDB.rawAtFault, baseline+1)
	}
	if after := countRaw(db.QueryRow); after != baseline {
		t.Fatalf("projection fault left %d raw rows, want %d", after, baseline)
	}

}

var errInjectedProjectionFault = errors.New("injected projection fault")

// projectionFaultDB fails only the projection insert, after the raw insert.
type projectionFaultDB struct {
	dbport.DB
	countRaw   func(func(context.Context, string, ...any) pgx.Row) int
	rawAtFault int
}

func (db *projectionFaultDB) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	tx, err := db.DB.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	return projectionFaultTx{Tx: tx, db: db}, nil
}

type projectionFaultTx struct {
	pgx.Tx
	db *projectionFaultDB
}

func (tx projectionFaultTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if strings.Contains(sql, "INSERT INTO administrative_audit_projections") {
		tx.db.rawAtFault = tx.db.countRaw(tx.Tx.QueryRow)
		return pgconn.CommandTag{}, errInjectedProjectionFault
	}
	return tx.Tx.Exec(ctx, sql, args...)
}

func appendAuditFixture(t *testing.T, db dbport.DB) (uuid.UUID, time.Time) {
	t.Helper()
	now := time.Date(2026, time.July, 25, 19, 45, 0, 0, time.UTC)
	targetID := uuid.NewString()
	eventID, err := administrativeaudit.Append(context.Background(), db, administrativeaudit.RawEvent{
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

	return eventID, now
}
