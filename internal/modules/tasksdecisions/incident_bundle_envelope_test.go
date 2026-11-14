package tasksdecisions_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/tasksdecisionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func requirePortableEnvelopeCapability(t *testing.T) {
	for _, kind := range []string{"task_request", "decision"} {
		for _, scenario := range []string{"valid", "missing", "foreign", "wrong type", "error with partial fact"} {
			t.Run(kind+"/"+scenario, func(t *testing.T) {
				ctx := t.Context()
				h := newTasksDecisionsPortabilityHarness(t, "envelope-capability")
				id := uuid.New()
				now := taskDecisionPortableTime()
				seedTaskDecisionEnvelope(t, h, id, kind, now)
				before := portableTaskAssociationState(t, h)
				fact := tasksdecisions.EnvelopeFact{IncidentID: h.incidentID, RecordType: kind}
				found := scenario != "missing"
				if scenario == "foreign" {
					fact.IncidentID = uuid.New()
				}
				if scenario == "wrong type" {
					fact.RecordType = "party"
				}
				sentinel := errors.New("envelope capability failed")
				var failure error
				if scenario == "error with partial fact" {
					failure = sentinel
				}
				var wantTx pgx.Tx
				calls := 0
				reader := portableEnvelopeReader(func(gotCtx context.Context, gotTx pgx.Tx, gotID uuid.UUID) (tasksdecisions.EnvelopeFact, bool, error) {
					calls++
					if gotCtx != ctx || gotTx != wantTx || gotID != id {
						t.Fatal("portable lookup changed borrowed context, transaction or identity")
					}
					return fact, found, failure
				})
				port, err := tasksdecisions.NewIncidentBundleSourcePort(tasksdecisionassembly.NewLinkFactsCapability(), tasksdecisionassembly.NewFieldReferenceReader(), reader)
				if err != nil {
					t.Fatal(err)
				}
				input := taskDecisionImportContext(h, "envelope-capability")
				var tasks, decisions []map[string]any
				if kind == "task_request" {
					tasks = []map[string]any{validPortableTask(h, id, now)}
				} else {
					decisions = []map[string]any{validPortableDecision(h, id, now)}
				}
				prepared, err := port.PrepareImport(ctx, taskDecisionBundle(tasks, decisions), input)
				if err != nil {
					t.Fatal(err)
				}
				tx, err := h.db.BeginTx(ctx, pgx.TxOptions{})
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
				wantTx = tx
				if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
					t.Fatal(err)
				}
				err = port.ValidateImportTx(ctx, tx, prepared, input)
				switch scenario {
				case "valid":
					if err != nil {
						t.Fatal(err)
					}
				case "error with partial fact":
					if !errors.Is(err, sentinel) {
						t.Fatalf("capability error lost: %v", err)
					}
				default:
					requireTasksDecisionsPortabilityFailure(t, err, "tasks_decisions.envelope_type_scope")
				}
				if calls != 1 {
					t.Fatalf("envelope lookups=%d want 1", calls)
				}
				if err := tx.Rollback(ctx); err != nil {
					t.Fatal(err)
				}
				if after := portableTaskAssociationState(t, h); !reflect.DeepEqual(before, after) {
					t.Fatal("portable envelope validation changed retained state after rollback")
				}
			})
		}
	}
}

type portableEnvelopeReader func(context.Context, pgx.Tx, uuid.UUID) (tasksdecisions.EnvelopeFact, bool, error)

func (r portableEnvelopeReader) LoadEnvelopeFactTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (tasksdecisions.EnvelopeFact, bool, error) {
	return r(ctx, tx, id)
}
