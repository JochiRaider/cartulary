package records

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/records/enveloperead"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func exerciseSharedEnvelopeLock(t *testing.T, holder, contender *pgx.Conn, incident, id, actor uuid.UUID, now time.Time) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	begin := func(conn *pgx.Conn) pgx.Tx {
		t.Helper()
		tx, err := conn.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
		return tx
	}
	rollback := func(tx pgx.Tx) {
		t.Helper()
		if err := tx.Rollback(ctx); err != nil {
			t.Fatal(err)
		}
	}
	requireContention := func(err error) {
		t.Helper()
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != "55P03" {
			t.Fatalf("lock error=%v want PostgreSQL 55P03", err)
		}
	}
	store := NewStore()
	want := enveloperead.Envelope{RecordID: id, IncidentID: incident, RecordType: "timeline_event", RowVersion: 1, CreatedByUserID: actor, CreatedAt: now, UpdatedByUserID: actor, UpdatedAt: now}
	first := begin(holder)
	got, err := store.LoadEnvelopeForShareNowaitTx(ctx, first, id)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("shared envelope=%#v error=%v", got, err)
	}
	second := begin(contender)
	got, err = store.LoadEnvelopeForShareNowaitTx(ctx, second, id)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("concurrent reader=%#v error=%v", got, err)
	}
	if _, err := second.Exec(ctx, "SELECT 1"); err != nil {
		t.Fatalf("borrowed transaction unusable: %v", err)
	}
	rollback(second)
	writer := begin(contender)
	_, err = writer.Exec(ctx, "SELECT record_id FROM records WHERE record_id=$1 FOR NO KEY UPDATE NOWAIT", id)
	requireContention(err)
	rollback(writer)
	rollback(first)
	exclusive := begin(holder)
	if _, err := exclusive.Exec(ctx, "SELECT record_id FROM records WHERE record_id=$1 FOR UPDATE", id); err != nil {
		t.Fatal(err)
	}
	blocked := begin(contender)
	_, err = store.LoadEnvelopeForShareNowaitTx(ctx, blocked, id)
	requireContention(err)
	rollback(blocked)
	rollback(exclusive)
	released := begin(contender)
	got, err = store.LoadEnvelopeForShareNowaitTx(ctx, released, id)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("released envelope=%#v error=%v", got, err)
	}
	if _, err := store.LoadEnvelopeForShareNowaitTx(ctx, released, uuid.New()); !errors.Is(err, enveloperead.ErrEnvelopeNotFound) {
		t.Fatalf("missing envelope=%v", err)
	}
	rollback(released)
	for _, nowait := range []bool{false, true} {
		writer := begin(holder)
		if nowait {
			got, err = store.LoadEnvelopeForNoKeyUpdateNowaitTx(ctx, writer, id)
		} else {
			got, err = store.LoadEnvelopeForNoKeyUpdateTx(ctx, writer, id)
		}
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("writer envelope nowait=%t value=%#v error=%v", nowait, got, err)
		}
		// Foreign-key readers remain compatible: this distinguishes the intended
		// NO KEY UPDATE posture from the stronger FOR UPDATE lock.
		keyReader := begin(contender)
		if _, err := keyReader.Exec(ctx, "SELECT record_id FROM records WHERE record_id=$1 FOR KEY SHARE NOWAIT", id); err != nil {
			t.Fatalf("writer blocked key reader: %v", err)
		}
		rollback(keyReader)
		blocked := begin(contender)
		_, err = store.LoadEnvelopeForNoKeyUpdateNowaitTx(ctx, blocked, id)
		requireContention(err)
		rollback(blocked)
		if _, err := writer.Exec(ctx, "SELECT 1"); err != nil {
			t.Fatalf("writer capability ended borrowed transaction: %v", err)
		}
		rollback(writer)
		released := begin(contender)
		got, err = store.LoadEnvelopeForNoKeyUpdateNowaitTx(ctx, released, id)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("released writer envelope=%#v error=%v", got, err)
		}
		if _, err := store.LoadEnvelopeForNoKeyUpdateNowaitTx(ctx, released, uuid.New()); !errors.Is(err, enveloperead.ErrEnvelopeNotFound) {
			t.Fatalf("missing writer envelope=%v", err)
		}
		rollback(released)
	}
}
