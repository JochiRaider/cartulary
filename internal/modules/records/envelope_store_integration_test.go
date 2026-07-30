package records

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestEnvelopeStoreContract_Integration(t *testing.T) {
	ctx := context.Background()
	db := pgtest.Start(t).BeginRollbackDBT(t, "records-envelope-store-contract")
	actorID, incidentID := seedEnvelopeOwnerContext(t, db, "envelope-store")
	store := NewStore()
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin envelope store transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	location := time.FixedZone("fixture-offset", -7*60*60)
	createdAt := time.Date(2026, time.July, 29, 8, 15, 0, 123456000, location)
	updatedAt := createdAt.Add(5 * time.Minute)
	suppliedRecordID := uuid.New()
	gotSuppliedID, err := store.InsertTx(ctx, tx, InsertParams{
		RecordID:        &suppliedRecordID,
		IncidentID:      incidentID,
		RecordType:      "timeline_event",
		CreatedByUserID: actorID,
		CreatedAt:       createdAt,
		UpdatedByUserID: actorID,
		UpdatedAt:       updatedAt,
		RowVersion:      7,
	})
	if err != nil {
		t.Fatalf("insert supplied envelope: %v", err)
	}
	if gotSuppliedID != suppliedRecordID {
		t.Fatalf("supplied record id = %s; want %s", gotSuppliedID, suppliedRecordID)
	}

	generatedRecordID, err := store.InsertTx(ctx, tx, InsertParams{
		IncidentID:      incidentID,
		RecordType:      "host",
		CreatedByUserID: actorID,
		CreatedAt:       createdAt,
		UpdatedByUserID: actorID,
		UpdatedAt:       updatedAt,
	})
	if err != nil {
		t.Fatalf("insert generated envelope: %v", err)
	}
	if generatedRecordID == uuid.Nil {
		t.Fatal("generated record id is nil")
	}

	supplied, err := store.LoadEnvelopeTx(ctx, tx, suppliedRecordID, false)
	if err != nil {
		t.Fatalf("load supplied envelope: %v", err)
	}
	if supplied.RecordID != suppliedRecordID ||
		supplied.IncidentID != incidentID ||
		supplied.RecordType != "timeline_event" ||
		supplied.RowVersion != 7 ||
		supplied.CreatedByUserID != actorID ||
		supplied.UpdatedByUserID != actorID ||
		supplied.DeletedAt != nil {
		t.Fatalf("supplied envelope = %#v", supplied)
	}
	if supplied.CreatedAt.Location() != time.UTC || supplied.UpdatedAt.Location() != time.UTC {
		t.Fatalf("envelope timestamps are not normalized to UTC: created=%v updated=%v", supplied.CreatedAt.Location(), supplied.UpdatedAt.Location())
	}
	if !supplied.CreatedAt.Equal(createdAt) || !supplied.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("envelope timestamps = (%s, %s); want (%s, %s)", supplied.CreatedAt, supplied.UpdatedAt, createdAt, updatedAt)
	}

	generated, err := store.LoadEnvelopeTx(ctx, tx, generatedRecordID, true)
	if err != nil {
		t.Fatalf("load generated envelope with lock: %v", err)
	}
	if generated.RowVersion != 1 {
		t.Fatalf("default row version = %d; want 1", generated.RowVersion)
	}

	loaded, err := store.LoadEnvelopesTx(ctx, tx, []uuid.UUID{
		generatedRecordID,
		suppliedRecordID,
		generatedRecordID,
		uuid.New(),
	}, false)
	if err != nil {
		t.Fatalf("batch-load envelopes: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("batch-loaded envelope count = %d; want 2", len(loaded))
	}
	empty, err := store.LoadEnvelopesTx(ctx, tx, nil, false)
	if err != nil {
		t.Fatalf("empty batch-load: %v", err)
	}
	if empty == nil || len(empty) != 0 {
		t.Fatalf("empty batch-load = %#v; want non-nil empty map", empty)
	}

	advancedAt := updatedAt.Add(10 * time.Minute)
	advancedVersion, err := store.AdvanceVersionTx(ctx, tx, suppliedRecordID, actorID, advancedAt)
	if err != nil {
		t.Fatalf("advance envelope version: %v", err)
	}
	if advancedVersion != 8 {
		t.Fatalf("advanced row version = %d; want 8", advancedVersion)
	}
	loadedVersion, err := store.LoadRowVersionTx(ctx, tx, suppliedRecordID)
	if err != nil {
		t.Fatalf("load row version: %v", err)
	}
	if loadedVersion != 8 {
		t.Fatalf("loaded row version = %d; want 8", loadedVersion)
	}
	deletedAt := advancedAt.Add(time.Minute)
	deletedVersion, err := store.SetDeleteStateTx(ctx, tx, suppliedRecordID, actorID, deletedAt, true)
	if err != nil {
		t.Fatalf("set envelope deleted state: %v", err)
	}
	if deletedVersion != 9 {
		t.Fatalf("deleted row version = %d; want 9", deletedVersion)
	}
	deleted, err := store.LoadEnvelopeTx(ctx, tx, suppliedRecordID, false)
	if err != nil {
		t.Fatalf("load deleted envelope: %v", err)
	}
	if deleted.DeletedAt == nil || !deleted.DeletedAt.Equal(deletedAt) ||
		deleted.DeletedByUserID == nil || *deleted.DeletedByUserID != actorID {
		t.Fatalf("deleted envelope tuple = (%v, %v)", deleted.DeletedAt, deleted.DeletedByUserID)
	}
	restoredVersion, err := store.SetDeleteStateTx(ctx, tx, suppliedRecordID, actorID, deletedAt.Add(time.Minute), false)
	if err != nil {
		t.Fatalf("clear envelope deleted state: %v", err)
	}
	if restoredVersion != 10 {
		t.Fatalf("restored row version = %d; want 10", restoredVersion)
	}
	restored, err := store.LoadEnvelopeTx(ctx, tx, suppliedRecordID, false)
	if err != nil {
		t.Fatalf("load restored envelope: %v", err)
	}
	if restored.DeletedAt != nil || restored.DeletedByUserID != nil {
		t.Fatalf("restored envelope retained delete tuple: %#v", restored)
	}

	if _, err := store.LoadEnvelopeTx(ctx, tx, uuid.New(), false); !errors.Is(err, ErrEnvelopeNotFound) {
		t.Fatalf("missing envelope error = %v; want %v", err, ErrEnvelopeNotFound)
	}
	if _, err := store.LoadRowVersionTx(ctx, tx, uuid.New()); !errors.Is(err, ErrEnvelopeNotFound) {
		t.Fatalf("missing row-version error = %v; want %v", err, ErrEnvelopeNotFound)
	}
	if _, err := store.AdvanceVersionTx(ctx, tx, uuid.New(), actorID, advancedAt); !errors.Is(err, ErrEnvelopeNotFound) {
		t.Fatalf("missing advance-version error = %v; want %v", err, ErrEnvelopeNotFound)
	}
	if _, err := store.SetDeleteStateTx(ctx, tx, uuid.New(), actorID, deletedAt, true); !errors.Is(err, ErrEnvelopeNotFound) {
		t.Fatalf("missing delete-state error = %v; want %v", err, ErrEnvelopeNotFound)
	}
}

func seedEnvelopeOwnerContext(t testing.TB, db postgres.DB, label string) (uuid.UUID, uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	actorID := uuid.New()
	incidentID := uuid.New()
	if _, err := db.Exec(ctx, `
INSERT INTO users (id, email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, 'Records owner evidence actor', 'test-only', false, true, true)
`, actorID, "records-"+label+"-"+actorID.String()+"@example.test"); err != nil {
		t.Fatalf("seed records actor: %v", err)
	}
	if _, err := db.Exec(ctx, `
INSERT INTO incidents (
    id, incident_key, incident_key_canonical, title, status,
    created_by_user_id, updated_by_user_id
) VALUES ($1, $2, $2, 'Records owner evidence incident', 'active', $3, $3)
`, incidentID, "RECORDS-"+label+"-"+incidentID.String(), actorID); err != nil {
		t.Fatalf("seed records incident: %v", err)
	}
	return actorID, incidentID
}
