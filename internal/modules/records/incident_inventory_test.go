package records

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
)

func exerciseIncidentRecordEnumeration(t *testing.T, db dbport.DB) {
	actorID, incidentID := seedEnvelopeOwnerContext(t, db, "record-enumeration")
	_, otherIncident := seedEnvelopeOwnerContext(t, db, "other-record-enumeration")
	tx, err := db.BeginTx(t.Context(), pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	ids := make([]uuid.UUID, 257)
	for i := range ids {
		ids[i] = uuid.New()
	}
	slices.SortFunc(ids, func(a, b uuid.UUID) int { return strings.Compare(a.String(), b.String()) })
	if _, err := tx.Exec(t.Context(), `INSERT INTO records(record_id,incident_id,record_type,created_by_user_id,updated_by_user_id,created_at,updated_at) SELECT id,$2,'host',$3,$3,$4,$4 FROM unnest($1::uuid[]) id`, ids, incidentID, actorID, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(t.Context(), `UPDATE records SET deleted_at=updated_at,deleted_by_user_id=$2 WHERE record_id=$1`, ids[128], actorID); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(t.Context(), `INSERT INTO records(record_id,incident_id,record_type,created_by_user_id,updated_by_user_id) VALUES($1,$2,'host',$3,$3)`, uuid.New(), otherIncident, actorID); err != nil {
		t.Fatal(err)
	}
	var got []uuid.UUID
	store := NewStore()
	if err := store.VisitIncidentRecordIDsTx(t.Context(), tx, incidentID, func(id uuid.UUID) error {
		var found uuid.UUID
		if err := tx.QueryRow(t.Context(), `SELECT record_id FROM records WHERE record_id=$1`, id).Scan(&found); err != nil {
			return err
		}
		if found != id {
			t.Fatal("callback identity changed")
		}
		got = append(got, id)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, ids) {
		t.Fatal("inventory lost, duplicated, reordered, concealed tombstones or included foreign records")
	}
	if err := store.VisitIncidentRecordIDsTx(t.Context(), tx, uuid.New(), func(uuid.UUID) error { t.Fatal("empty incident invoked visitor"); return nil }); err != nil {
		t.Fatal(err)
	}
	stop := errors.New("stop inventory")
	calls := 0
	if err := store.VisitIncidentRecordIDsTx(t.Context(), tx, incidentID, func(uuid.UUID) error { calls++; return stop }); !errors.Is(err, stop) || calls != 1 {
		t.Fatalf("callback failure=%v calls=%d", err, calls)
	}
	if _, err := tx.Exec(t.Context(), `SELECT 1`); err != nil {
		t.Fatalf("borrowed transaction closed: %v", err)
	}
}
