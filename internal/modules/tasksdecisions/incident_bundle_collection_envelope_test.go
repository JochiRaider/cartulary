package tasksdecisions_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/tasksdecisionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func requirePortableCollectionEnvelopes(t *testing.T) {
	for _, kind := range []string{"task_request", "decision"} {
		for _, scenario := range []string{"valid", "source missing", "source foreign", "source wrong type", "source deleted", "source error", "target missing", "target foreign", "target deleted", "target error"} {
			t.Run(kind+"/"+scenario, func(t *testing.T) {
				ctx := t.Context()
				h := newTasksDecisionsPortabilityHarness(t, "collection-envelope")
				source, target := uuid.New(), uuid.New()
				now := taskDecisionPortableTime()
				seedTaskDecisionEnvelope(t, h, source, kind, now)
				seedTaskDecisionEnvelope(t, h, target, "party", now)
				if _, err := h.db.Exec(ctx, `INSERT INTO parties(record_id,incident_id,display_name,party_kind) VALUES($1,$2,'Target','organization')`, target, h.incidentID); err != nil {
					t.Fatal(err)
				}
				before := portableTaskAssociationState(t, h)
				sentinel := errors.New("collection endpoint read failed")
				var wantTx pgx.Tx
				var reads []uuid.UUID
				actual := tasksdecisionassembly.NewEnvelopeFactReader()
				reader := portableEnvelopeReader(func(gotCtx context.Context, gotTx pgx.Tx, id uuid.UUID) (tasksdecisions.EnvelopeFact, bool, error) {
					if gotCtx != ctx || gotTx != wantTx {
						t.Fatal("endpoint lookup changed borrowed transaction/context")
					}
					reads = append(reads, id)
					fact, found, err := actual.LoadEnvelopeFactTx(gotCtx, gotTx, id)
					if err != nil || len(reads) == 1 {
						return fact, found, err
					}
					selected := id == source && strings.HasPrefix(scenario, "source ") || id == target && strings.HasPrefix(scenario, "target ")
					if selected {
						switch {
						case strings.HasSuffix(scenario, "missing"):
							found = false
						case strings.HasSuffix(scenario, "foreign"):
							fact.IncidentID = uuid.New()
						case strings.HasSuffix(scenario, "wrong type"):
							fact.RecordType = "party"
						case strings.HasSuffix(scenario, "deleted"):
							fact.Deleted = true
						case strings.HasSuffix(scenario, "error"):
							err = sentinel
						}
					}
					return fact, found, err
				})
				field := "task.linked_record_ids"
				if kind == "decision" {
					field = "decision.affected_record_ids"
				}
				links := portableCollectionLinks(func(gotCtx context.Context, gotTx pgx.Tx, incident, id uuid.UUID) ([]tasksdecisions.LinkFact, error) {
					if gotCtx != ctx || gotTx != wantTx || incident != h.incidentID || id != source {
						t.Fatal("collection facts changed borrowed scope")
					}
					return []tasksdecisions.LinkFact{{SourceRecordID: source, DestinationRecordID: target, LinkType: "references_record", HasFieldKey: true, FieldKey: field}}, nil
				})
				port, err := tasksdecisions.NewIncidentBundleSourcePort(links, tasksdecisionassembly.NewFieldReferenceReader(), reader)
				if err != nil {
					t.Fatal(err)
				}
				input := taskDecisionImportContext(h, "collection-envelope")
				var tasks, decisions []map[string]any
				if kind == "task_request" {
					tasks = []map[string]any{validPortableTask(h, source, now)}
				} else {
					decisions = []map[string]any{validPortableDecision(h, source, now)}
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
				if scenario == "valid" {
					if err != nil {
						t.Fatal(err)
					}
				} else if strings.HasSuffix(scenario, "error") {
					if !errors.Is(err, sentinel) {
						t.Fatalf("endpoint capability error lost: %v", err)
					}
				} else {
					requireTasksDecisionsPortabilityFailure(t, err, "tasks_decisions.references_same_incident")
				}
				wantReads := []uuid.UUID{source, source, target}
				if strings.HasPrefix(scenario, "source ") {
					wantReads = wantReads[:2]
				}
				if !reflect.DeepEqual(reads, wantReads) {
					t.Fatalf("endpoint reads=%v want %v", reads, wantReads)
				}
				if err := tx.Rollback(ctx); err != nil {
					t.Fatal(err)
				}
				if after := portableTaskAssociationState(t, h); !reflect.DeepEqual(before, after) {
					t.Fatal("collection validation changed retained state after rollback")
				}
			})
		}
	}
}

type portableCollectionLinks func(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) ([]tasksdecisions.LinkFact, error)

func (r portableCollectionLinks) LoadRecordLinkFactsTx(ctx context.Context, tx pgx.Tx, incident, id uuid.UUID) ([]tasksdecisions.LinkFact, error) {
	return r(ctx, tx, incident, id)
}
