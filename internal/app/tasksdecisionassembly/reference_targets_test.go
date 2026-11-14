package tasksdecisionassembly

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/records/enveloperead"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func exerciseReferenceTargetAdapter(t *testing.T) {
	ctx := t.Context()
	tx := &struct{ pgx.Tx }{}
	incident, id := uuid.New(), uuid.New()
	deletedAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	sentinel := errors.New("reference capability failure")
	for _, nowait := range []bool{false, true} {
		for _, kind := range []string{"party", "decision"} {
			want := tasksdecisions.EnvelopeFact{IncidentID: incident, RecordType: kind, Deleted: true}
			for _, mode := range []string{"valid", "missing", "error"} {
				r := &fakeReferenceEnvelopeLocker{value: enveloperead.Envelope{RecordID: id, IncidentID: incident, RecordType: kind, DeletedAt: &deletedAt}}
				if mode == "missing" {
					r.err = fmt.Errorf("read: %w", enveloperead.ErrEnvelopeNotFound)
				}
				if mode == "error" {
					r.err = fmt.Errorf("read: %w", sentinel)
				}
				a := referenceTargetAdapter{envelopes: r}
				fact, found, err := a.LockEnvelopeFactTx(ctx, tx, id, kind, nowait)
				if r.calls != 1 || r.ctx != ctx || r.tx != tx || r.record != id || r.nowait != nowait || r.writer != (kind == "decision") {
					t.Fatalf("envelope lock forwarding=%#v", r)
				}
				if mode == "valid" {
					if err != nil || !found || fact != want {
						t.Fatalf("lock fact=%#v found=%t error=%v", fact, found, err)
					}
				} else {
					if found || fact != (tasksdecisions.EnvelopeFact{}) {
						t.Fatal("failed lock exposed partial fact")
					}
					if mode == "missing" && err != nil || mode == "error" && !errors.Is(err, sentinel) {
						t.Fatalf("lock error=%v", err)
					}
				}
			}
		}
		for _, kind := range []string{"", "host"} {
			r := &fakeReferenceEnvelopeLocker{}
			a := referenceTargetAdapter{envelopes: r}
			fact, found, err := a.LockEnvelopeFactTx(ctx, tx, id, kind, nowait)
			if err == nil || found || fact != (tasksdecisions.EnvelopeFact{}) || r.calls != 0 {
				t.Fatalf("unsupported reference kind=%q fact=%#v found=%t calls=%d error=%v", kind, fact, found, r.calls, err)
			}
		}
		for _, mode := range []string{"valid", "missing", "error"} {
			r := &fakePartySourceLocker{found: mode != "missing"}
			if mode == "error" {
				r.err = sentinel
			}
			a := referenceTargetAdapter{parties: r}
			found, err := a.LockPartySourceTx(ctx, tx, incident, id, nowait)
			if r.calls != 1 || r.ctx != ctx || r.tx != tx || r.incident != incident || r.record != id || r.nowait != nowait {
				t.Fatalf("Party lock forwarding=%#v", r)
			}
			if found != (mode == "valid") || mode == "error" && !errors.Is(err, sentinel) || mode != "error" && err != nil {
				t.Fatalf("Party found=%t error=%v", found, err)
			}
		}
	}
	if contribution, err := tasksdecisions.NewRevisionContribution(NewLinkFactsCapability(), NewFieldReferenceReader(), NewSupersessionReader(), nil); err == nil || len(contribution.Records) != 0 {
		t.Fatalf("missing reference capability contribution=%#v error=%v", contribution, err)
	}
	if owner, err := tasksdecisions.NewImportContribution(tasksdecisions.TaskRequestsViewSchemaID, "tasksdecisions.task_request.import_create", tasksdecisions.ImportDependencies{RecordEnvelopes: records.NewStore()}); err == nil || owner != nil {
		t.Fatalf("missing reference capability import=%#v error=%v", owner, err)
	}
}

type fakeReferenceEnvelopeLocker struct {
	value  enveloperead.Envelope
	err    error
	ctx    context.Context
	tx     pgx.Tx
	record uuid.UUID
	nowait bool
	writer bool
	calls  int
}

func (r *fakeReferenceEnvelopeLocker) read(ctx context.Context, tx pgx.Tx, id uuid.UUID, nowait, writer bool) (enveloperead.Envelope, error) {
	r.ctx, r.tx, r.record, r.nowait, r.writer = ctx, tx, id, nowait, writer
	r.calls++
	return r.value, r.err
}
func (r *fakeReferenceEnvelopeLocker) LoadEnvelopeForShareTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (enveloperead.Envelope, error) {
	return r.read(ctx, tx, id, false, false)
}
func (r *fakeReferenceEnvelopeLocker) LoadEnvelopeForShareNowaitTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (enveloperead.Envelope, error) {
	return r.read(ctx, tx, id, true, false)
}

func (r *fakeReferenceEnvelopeLocker) LoadEnvelopeForNoKeyUpdateTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (enveloperead.Envelope, error) {
	return r.read(ctx, tx, id, false, true)
}
func (r *fakeReferenceEnvelopeLocker) LoadEnvelopeForNoKeyUpdateNowaitTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (enveloperead.Envelope, error) {
	return r.read(ctx, tx, id, true, true)
}

type fakePartySourceLocker struct {
	found            bool
	err              error
	ctx              context.Context
	tx               pgx.Tx
	incident, record uuid.UUID
	nowait           bool
	calls            int
}

func (r *fakePartySourceLocker) LockSourceTx(ctx context.Context, tx pgx.Tx, incident, id uuid.UUID, nowait bool) (bool, error) {
	r.ctx, r.tx, r.incident, r.record, r.nowait = ctx, tx, incident, id, nowait
	r.calls++
	return r.found, r.err
}
