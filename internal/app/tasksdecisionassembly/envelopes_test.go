package tasksdecisionassembly

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/records/enveloperead"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func exerciseEnvelopeAdapter(t *testing.T) {
	ctx := t.Context()
	tx := &struct{ pgx.Tx }{}
	record, incident := uuid.New(), uuid.New()
	reader := &fakeEnvelopeReader{value: enveloperead.Envelope{RecordID: record, IncidentID: incident, RecordType: "decision"}}
	adapter := envelopeAdapter{reader: reader}
	fact, found, err := adapter.LoadEnvelopeFactTx(ctx, tx, record)
	if err != nil || !found || fact != (tasksdecisions.EnvelopeFact{IncidentID: incident, RecordType: "decision"}) {
		t.Fatalf("envelope projection=%#v found=%t error=%v", fact, found, err)
	}
	if reader.ctx != ctx || reader.tx != tx || reader.record != record || reader.calls != 1 {
		t.Fatal("envelope read changed borrowed scope or repeated lookup")
	}
	deletedAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	reader.value.DeletedAt = &deletedAt
	deleted, found, err := adapter.LoadEnvelopeFactTx(ctx, tx, record)
	if err != nil || !found || !deleted.Deleted || deleted.IncidentID != incident || deleted.RecordType != "decision" {
		t.Fatalf("deleted envelope fact=%#v found=%t error=%v", deleted, found, err)
	}
	reader.value.DeletedAt = nil
	if !deleted.Deleted {
		t.Fatal("deletion fact changed with source storage")
	}
	reader.value.IncidentID = uuid.New()
	reader.value.RecordType = "party"
	if fact.IncidentID != incident || fact.RecordType != "decision" {
		t.Fatal("fact changed with reader storage")
	}
	for _, missing := range []bool{false, true} {
		sentinel := errors.New("envelope read failed")
		if missing {
			sentinel = enveloperead.ErrEnvelopeNotFound
		}
		reader.err = fmt.Errorf("reader: %w", sentinel)
		partial, found, err := adapter.LoadEnvelopeFactTx(ctx, tx, record)
		if found || partial != (tasksdecisions.EnvelopeFact{}) {
			t.Fatalf("failed read exposed partial fact=%#v found=%t", partial, found)
		}
		if missing {
			if err != nil {
				t.Fatalf("absent envelope error=%v", err)
			}
		} else if !errors.Is(err, sentinel) {
			t.Fatalf("read error lost cause: %v", err)
		}
	}
	if port, err := tasksdecisions.NewIncidentBundleSourcePort(NewLinkFactsCapability(), NewFieldReferenceReader(), nil); err == nil || port != nil {
		t.Fatalf("missing envelope reader produced port=%#v error=%v", port, err)
	}
}

type fakeEnvelopeReader struct {
	value  enveloperead.Envelope
	err    error
	ctx    context.Context
	tx     pgx.Tx
	record uuid.UUID
	calls  int
}

func (r *fakeEnvelopeReader) LoadEnvelopeTx(ctx context.Context, tx pgx.Tx, record uuid.UUID) (enveloperead.Envelope, error) {
	r.ctx, r.tx, r.record = ctx, tx, record
	r.calls++
	return r.value, r.err
}
