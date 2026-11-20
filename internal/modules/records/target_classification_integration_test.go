package records

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/records/enveloperead"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

// TestRecordTargetClassification_Integration pins the shared REQ-01-705
// classifier: missing, deleted and live targets under every lock mode, the lock
// each mode really holds, and that a database failure is never a
// classification.
func TestRecordTargetClassification_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	t.Run("presence matrix", func(t *testing.T) {
		deletedAt := time.Date(2026, time.November, 18, 9, 0, 0, 0, time.UTC)
		for _, test := range []struct {
			name      string
			found     bool
			deletedAt *time.Time
			want      error
		}{
			{name: "missing", want: ErrTargetNotFound},
			{name: "missing ignores deletion state", deletedAt: &deletedAt, want: ErrTargetNotFound},
			{name: "deleted", found: true, deletedAt: &deletedAt, want: ErrTargetDeleted},
			{name: "live", found: true},
		} {
			if got := ClassifyTargetPresence(test.found, test.deletedAt); got != test.want {
				t.Fatalf("%s: classification = %v, want %v", test.name, got, test.want)
			}
		}
		if !errors.Is(ErrTargetNotFound, enveloperead.ErrEnvelopeNotFound) {
			t.Fatal("ErrTargetNotFound must match enveloperead.ErrEnvelopeNotFound")
		}
		if errors.Is(ErrTargetDeleted, enveloperead.ErrEnvelopeNotFound) || errors.Is(ErrTargetNotFound, ErrTargetDeleted) {
			t.Fatal("the missing and deleted classifications must stay distinct")
		}
	})

	testDB := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "records-target-classification")
	holder, err := pgx.Connect(ctx, testDB.DSN)
	if err != nil {
		t.Fatalf("connect classifier: %v", err)
	}
	t.Cleanup(func() { _ = holder.Close(context.Background()) })
	contender, err := pgx.Connect(ctx, testDB.DSN)
	if err != nil {
		t.Fatalf("connect contender: %v", err)
	}
	t.Cleanup(func() { _ = contender.Close(context.Background()) })

	actorID, incidentID := seedEnvelopeOwnerContext(t, holder, "target-classification")
	now := time.Date(2026, time.November, 18, 8, 0, 0, 0, time.UTC)
	deletedAt := now.Add(time.Hour)
	liveID, deletedID := uuid.New(), uuid.New()
	if _, err := holder.Exec(ctx, `
INSERT INTO records (
    record_id, incident_id, record_type,
    created_by_user_id, created_at, updated_by_user_id, updated_at, row_version
) VALUES
    ($1, $3, 'timeline_event', $4, $5, $4, $5, 3),
    ($2, $3, 'party', $4, $5, $4, $5, 4)
`, liveID, deletedID, incidentID, actorID, now); err != nil {
		t.Fatalf("seed classification envelopes: %v", err)
	}
	if _, err := holder.Exec(ctx, `
UPDATE records
   SET deleted_at = $2, deleted_by_user_id = $3, updated_at = $2, updated_by_user_id = $3, row_version = 5
 WHERE record_id = $1
`, deletedID, deletedAt, actorID); err != nil {
		t.Fatalf("soft-delete classification envelope: %v", err)
	}
	live := enveloperead.Envelope{
		RecordID: liveID, IncidentID: incidentID, RecordType: "timeline_event", RowVersion: 3,
		CreatedByUserID: actorID, CreatedAt: now, UpdatedByUserID: actorID, UpdatedAt: now,
	}
	deletedBy := actorID
	deleted := enveloperead.Envelope{
		RecordID: deletedID, IncidentID: incidentID, RecordType: "party", RowVersion: 5,
		CreatedByUserID: actorID, CreatedAt: now, UpdatedByUserID: actorID, UpdatedAt: deletedAt,
		DeletedAt: &deletedAt, DeletedByUserID: &deletedBy,
	}

	begin := func(t *testing.T, conn *pgx.Conn) pgx.Tx {
		t.Helper()
		tx, err := conn.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
		return tx
	}
	// probe reports whether the contender can take the given row lock now.
	probe := func(t *testing.T, recordID uuid.UUID, clause string) bool {
		t.Helper()
		tx := begin(t, contender)
		defer func() { _ = tx.Rollback(ctx) }()
		_, err := tx.Exec(ctx, "SELECT record_id FROM records WHERE record_id = $1 "+clause+" NOWAIT", recordID)
		var pgErr *pgconn.PgError
		switch {
		case err == nil:
			return true
		case errors.As(err, &pgErr) && pgErr.Code == "55P03":
			return false
		default:
			t.Fatalf("probe %s: %v", clause, err)
			return false
		}
	}

	for _, mode := range []struct {
		name               string
		lock               TargetLock
		allowsShare        bool
		allowsKeyShare     bool
		allowsNoKeyUpdates bool
	}{
		{name: "none", lock: TargetLockNone, allowsShare: true, allowsKeyShare: true, allowsNoKeyUpdates: true},
		{name: "share", lock: TargetLockShare, allowsShare: true, allowsKeyShare: true},
		{name: "no key update", lock: TargetLockNoKeyUpdate, allowsKeyShare: true},
		{name: "update", lock: TargetLockUpdate},
	} {
		t.Run(mode.name, func(t *testing.T) {
			for _, state := range []struct {
				name     string
				recordID uuid.UUID
				want     enveloperead.Envelope
				wantErr  error
			}{
				{name: "live", recordID: liveID, want: live},
				{name: "deleted", recordID: deletedID, want: deleted, wantErr: ErrTargetDeleted},
				{name: "missing", recordID: uuid.New(), wantErr: ErrTargetNotFound},
			} {
				t.Run(state.name, func(t *testing.T) {
					tx := begin(t, holder)
					got, err := ClassifyTargetTx(ctx, tx, state.recordID, mode.lock)
					if err != state.wantErr {
						t.Fatalf("classification error = %v, want %v", err, state.wantErr)
					}
					if !reflect.DeepEqual(got, state.want) {
						t.Fatalf("classified envelope = %#v, want %#v", got, state.want)
					}
					storeGot, storeErr := NewStore().ClassifyTargetTx(ctx, tx, state.recordID, mode.lock)
					if storeErr != state.wantErr || !reflect.DeepEqual(storeGot, state.want) {
						t.Fatalf("Store classification = %#v, %v; want %#v, %v", storeGot, storeErr, state.want, state.wantErr)
					}
					if state.want.RecordID == uuid.Nil {
						return
					}
					// The classifying read holds its lock for live and deleted targets.
					if got := probe(t, state.recordID, "FOR SHARE"); got != mode.allowsShare {
						t.Fatalf("contender FOR SHARE allowed = %v, want %v", got, mode.allowsShare)
					}
					if got := probe(t, state.recordID, "FOR KEY SHARE"); got != mode.allowsKeyShare {
						t.Fatalf("contender FOR KEY SHARE allowed = %v, want %v", got, mode.allowsKeyShare)
					}
					if got := probe(t, state.recordID, "FOR NO KEY UPDATE"); got != mode.allowsNoKeyUpdates {
						t.Fatalf("contender FOR NO KEY UPDATE allowed = %v, want %v", got, mode.allowsNoKeyUpdates)
					}
					if err := tx.Rollback(ctx); err != nil {
						t.Fatal(err)
					}
					if !probe(t, state.recordID, "FOR UPDATE") {
						t.Fatal("classification lock outlived its transaction")
					}
				})
			}
		})
	}

	t.Run("unknown lock is rejected before reading", func(t *testing.T) {
		tx := begin(t, holder)
		_, err := ClassifyTargetTx(ctx, tx, liveID, TargetLock(99))
		if err == nil || errors.Is(err, ErrTargetNotFound) || errors.Is(err, ErrTargetDeleted) {
			t.Fatalf("unknown lock error = %v", err)
		}
	})

	t.Run("database failure is not a classification", func(t *testing.T) {
		tx := begin(t, holder)
		if _, err := tx.Exec(ctx, "SELECT 1/0"); err == nil {
			t.Fatal("expected the setup statement to abort the transaction")
		}
		_, err := ClassifyTargetTx(ctx, tx, liveID, TargetLockUpdate)
		if err == nil || errors.Is(err, ErrTargetNotFound) || errors.Is(err, ErrTargetDeleted) || errors.Is(err, enveloperead.ErrEnvelopeNotFound) {
			t.Fatalf("aborted transaction classification error = %v, want an unclassified failure", err)
		}
	})
}
