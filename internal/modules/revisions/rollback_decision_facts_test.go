package revisions_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/tasksdecisionassembly"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	envelopetest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func exerciseDecisionInverseFacts(t *testing.T) {
	for _, tc := range []struct {
		name, relation, status string
		valid, hidden, fail    bool
	}{
		{"no relation", "", "proposed", true, false, false},
		{"incoming legal", "incoming", "superseded", true, false, false},
		{"incoming illegal", "incoming", "proposed", false, false, false},
		{"outgoing legal", "outgoing", "approved", true, false, false},
		{"outgoing illegal", "outgoing", "rejected", false, false, false},
		{"unrelated type", "reference", "proposed", true, false, false},
		{"hidden incoming", "incoming", "superseded", false, true, false},
		{"partial fact failure", "incoming", "superseded", false, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			h := appsupport.StartStore(t, "decision-inverse-facts")
			actor := authstoretest.SeedActor(t, h.DB, authstoretest.ActorAttributes{Email: "decision-inverse@example.test", DisplayName: "Decision inverse"})
			incident := appsupport.CreateIncidentInStore(t, h.DB, actor, "decision-inverse-incident", "IR-DECISION-INVERSE", "Decision inverse")
			selected, peer := uuid.New(), uuid.New()
			now := time.Date(2026, 11, 14, 2, 0, 0, 0, time.UTC)
			for _, id := range []uuid.UUID{selected, peer} {
				envelopetest.SeedRecordEnvelope(t, h.DB, incident.ID, actor.ID, id, "decision")
				status := "executed"
				if id == peer && tc.relation == "incoming" {
					status = "approved"
				}
				if _, err := h.DB.Exec(ctx, `INSERT INTO decisions(record_id,incident_id,summary,status,owner_user_id,decision_type,decided_at,rationale,created_at,updated_at) VALUES($1,$2,'Current summary',$3,$4,'containment',$5,'Current rationale',$5,$5)`, id, incident.ID, status, actor.ID, now); err != nil {
					t.Fatal(err)
				}
			}
			if tc.relation != "" {
				source, target, kind := selected, peer, "supersedes"
				if tc.relation == "incoming" {
					source, target = peer, selected
				}
				if tc.relation == "reference" {
					kind = "references_record"
				}
				if _, err := h.DB.Exec(ctx, `INSERT INTO record_links(incident_id,src_record_id,dst_record_id,link_type,provenance,owner_user_id,created_by_user_id,decided_at,created_at) VALUES($1,$2,$3,$4,'manual',$5,$5,$6,$6)`, incident.ID, source, target, kind, actor.ID, now); err != nil {
					t.Fatal(err)
				}
			}
			if tc.hidden {
				if _, err := h.DB.Exec(ctx, `UPDATE records SET deleted_at=$2,deleted_by_user_id=$3 WHERE record_id=$1`, peer, now, actor.ID); err != nil {
					t.Fatal(err)
				}
			}
			witness := &decisionInverseFactsWitness{delegate: tasksdecisionassembly.NewLinkFactsCapability()}
			if tc.fail {
				witness.err = errors.New("injected Links read failure")
			}
			contribution, err := tasksdecisions.NewRevisionContribution(witness, tasksdecisionassembly.NewFieldReferenceReader(), tasksdecisionassembly.NewSupersessionReader(), tasksdecisionassembly.NewReferenceTargetCapability())
			if err != nil {
				t.Fatal(err)
			}
			var provider rollbackcontract.RowSourceProvider
			for _, record := range contribution.Records {
				if record.RecordType == "decision" {
					provider = record.RowRollbackProvider
				}
			}
			if provider == nil {
				t.Fatal("missing Decision inverse provider")
			}
			tx, err := h.DB.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(context.Background())
			snapshot := func(table string) string {
				t.Helper()
				var state string
				if err := tx.QueryRow(ctx, "SELECT coalesce(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text),'[]'::jsonb)::text FROM "+table+" t").Scan(&state); err != nil {
					t.Fatal(err)
				}
				return state
			}
			records, decisions, relations := snapshot("records"), snapshot("decisions"), snapshot("record_links")
			err = provider.RestoreTx(ctx, tx, rollbackcontract.RestoreRequest{RecordID: selected, ActorUserID: actor.ID, Now: now.Add(time.Minute), NextRowVersion: 2, RetainedValue: map[string]any{"source": map[string]any{"status": tc.status, "summary": "Restored summary"}}})
			if witness.calls != 1 || witness.ctx != ctx || witness.tx != tx || witness.incident != incident.ID || witness.record != selected {
				t.Fatalf("Decision inverse did not use exact borrowed Links capability: %+v", witness)
			}
			if tc.fail {
				if !errors.Is(err, witness.err) {
					t.Fatalf("fact error was lost: %v", err)
				}
			} else if !tc.valid {
				if !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
					t.Fatalf("illegal inverse error=%v", err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				var exact bool
				if err := tx.QueryRow(ctx, `SELECT summary='Restored summary' AND status=$2 AND owner_user_id=$3 AND decided_at=$4 AND created_at=$4 AND updated_at=$5 AND rationale='Current rationale' AND decision_type='containment' FROM decisions WHERE record_id=$1`, selected, tc.status, actor.ID, now, now.Add(time.Minute)).Scan(&exact); err != nil || !exact {
					t.Fatalf("restored Decision source exact=%v error=%v", exact, err)
				}
			}
			if records != snapshot("records") || relations != snapshot("record_links") || (!tc.valid && decisions != snapshot("decisions")) {
				t.Fatal("Decision inverse changed unrelated or rejected source state")
			}
		})
	}
}

type decisionInverseFactsWitness struct {
	delegate         tasksdecisions.LinkFactsCapability
	err              error
	calls            int
	ctx              context.Context
	tx               pgx.Tx
	incident, record uuid.UUID
}

func (w *decisionInverseFactsWitness) LoadRecordLinkFactsTx(ctx context.Context, tx pgx.Tx, incident, record uuid.UUID) ([]tasksdecisions.LinkFact, error) {
	w.calls++
	w.ctx, w.tx, w.incident, w.record = ctx, tx, incident, record
	facts, err := w.delegate.LoadRecordLinkFactsTx(ctx, tx, incident, record)
	if err != nil {
		return nil, err
	}
	return facts, w.err
}
