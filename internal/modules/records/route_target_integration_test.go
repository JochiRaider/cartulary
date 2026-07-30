package records

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestRouteTargetResolverContract_Integration(t *testing.T) {
	ctx := context.Background()
	db := pgtest.Start(t).BeginRollbackDBT(t, "records-route-target-contract")
	actorID, incidentID := seedEnvelopeOwnerContext(t, db, "route-target")
	recordID := uuid.New()
	now := time.Date(2026, time.July, 29, 16, 45, 0, 0, time.UTC)
	if _, err := db.Exec(ctx, `
INSERT INTO records (
    record_id, incident_id, record_type,
    created_by_user_id, created_at, updated_by_user_id, updated_at, row_version
) VALUES ($1, $2, 'identity', $3, $4, $3, $4, 4)
`, recordID, incidentID, actorID, now); err != nil {
		t.Fatalf("seed route-target envelope: %v", err)
	}

	resolver := NewRouteTargetResolver(db)
	got, err := resolver.Resolve(ctx, recordID)
	if err != nil {
		t.Fatalf("resolve active route target: %v", err)
	}
	want := RouteTarget{
		IncidentID: incidentID,
		RecordType: "identity",
		Deleted:    false,
		RowVersion: 4,
	}
	if got != want {
		t.Fatalf("active route target = %#v; want %#v", got, want)
	}
	gotIncidentID, err := resolver.ResolveIncident(ctx, recordID)
	if err != nil {
		t.Fatalf("resolve incident: %v", err)
	}
	if gotIncidentID != incidentID {
		t.Fatalf("resolved incident id = %s; want %s", gotIncidentID, incidentID)
	}
	if _, err := resolver.Resolve(ctx, uuid.New()); err == nil {
		t.Fatal("resolve missing route target succeeded")
	}

	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin route-target transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	gotTx, err := resolver.ResolveTx(ctx, tx, recordID)
	if err != nil {
		t.Fatalf("resolve transaction route target: %v", err)
	}
	if gotTx != want {
		t.Fatalf("transaction route target = %#v; want %#v", gotTx, want)
	}

	deletedAt := now.Add(time.Hour)
	if _, err := tx.Exec(ctx, `
UPDATE records
   SET deleted_at = $2,
       deleted_by_user_id = $3,
       updated_at = $2,
       updated_by_user_id = $3,
       row_version = 5
 WHERE record_id = $1
`, recordID, deletedAt, actorID); err != nil {
		t.Fatalf("soft-delete route-target envelope: %v", err)
	}
	gotDeleted, err := resolver.ResolveTx(ctx, tx, recordID)
	if err != nil {
		t.Fatalf("resolve deleted transaction route target: %v", err)
	}
	want.Deleted = true
	want.RowVersion = 5
	if gotDeleted != want {
		t.Fatalf("deleted route target = %#v; want %#v", gotDeleted, want)
	}

	if _, err := resolver.ResolveTx(ctx, tx, uuid.New()); err == nil {
		t.Fatal("resolve missing transaction route target succeeded")
	}
}
