package tasksdecisions_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/tasksdecisionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func requirePortableTaskDecisionBindings(t *testing.T) {
	t.Run("binding capability errors", requirePortableTaskBindingErrors)
	cases := []struct {
		name       string
		scalar     bool
		bindings   int
		kind       string
		hidden     string
		tombstoned bool
		valid      bool
	}{
		{name: "null with binding", bindings: 1, kind: "references_record"},
		{name: "wrong kind", scalar: true, bindings: 1, kind: "supported_by"},
		{name: "hidden destination mismatch", scalar: true, bindings: 1, kind: "references_record", hidden: "destination"},
		{name: "hidden source mismatch", scalar: true, bindings: 1, kind: "references_record", hidden: "source"},
		{name: "multiple", scalar: true, bindings: 2, kind: "references_record"},
		{name: "scalar without binding", scalar: true, valid: true},
		{name: "null without binding", valid: true},
		{name: "null with tombstoned binding", bindings: 1, kind: "references_record", tombstoned: true, valid: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			h := newTasksDecisionsPortabilityHarness(t, "binding")
			now := taskDecisionPortableTime()
			taskID := uuid.MustParse("00000000-0000-4000-8000-000000000001")
			decisionID := uuid.MustParse("00000000-0000-4000-8000-000000000002")
			otherID := uuid.MustParse("00000000-0000-4000-8000-000000000003")
			seedTaskDecisionEnvelope(t, h, taskID, "task_request", now)
			seedTaskDecisionEnvelope(t, h, decisionID, "decision", now)
			seedTaskDecisionEnvelope(t, h, otherID, "decision", now)
			row := validPortableTask(h, taskID, now)
			if tc.scalar {
				row["decision_record_id"] = decisionID.String()
			}
			for i := 0; i < tc.bindings; i++ {
				target := decisionID
				if tc.hidden != "" || i == 1 {
					target = otherID
				}
				if _, err := h.db.Exec(ctx, `INSERT INTO record_links
(incident_id,src_record_id,dst_record_id,link_type,field_key,provenance,owner_user_id,created_by_user_id,decided_at,created_at,deleted_at)
VALUES ($1,$2,$3,$4,'task.decision_record_id','manual',$5,$5,$6,$6,CASE WHEN $7 THEN $6::timestamptz ELSE NULL END)`,
					h.incidentID, taskID, target, tc.kind, h.actor.ID, now, tc.tombstoned); err != nil {
					t.Fatal(err)
				}
			}
			if tc.hidden != "" {
				id := otherID
				if tc.hidden == "source" {
					id = taskID
				}
				if _, err := h.db.Exec(ctx, `UPDATE records SET deleted_at=$2,deleted_by_user_id=$3 WHERE record_id=$1`, id, now, h.actor.ID); err != nil {
					t.Fatal(err)
				}
			}
			before := portableTaskAssociationState(t, h)
			err := applyAndValidateTaskDecisionBundle(t, h, taskDecisionBundle([]map[string]any{row}, []map[string]any{
				validPortableDecision(h, decisionID, now), validPortableDecision(h, otherID, now),
			}), "binding")
			if tc.valid {
				if err != nil {
					t.Fatalf("valid retained association rejected: %v", err)
				}
			} else {
				requireTasksDecisionsPortabilityFailure(t, err, "tasks_decisions.dependent_fields_legal")
			}
			if after := portableTaskAssociationState(t, h); !reflect.DeepEqual(before, after) {
				t.Fatal("source-port transaction rollback changed retained state")
			}
		})
	}
}

func portableTaskAssociationState(t testing.TB, h tasksDecisionsPortabilityHarness) map[string]string {
	t.Helper()
	state := make(map[string]string)
	for _, table := range []string{"records", "task_requests", "decisions", "record_links"} {
		var value string
		// Identifiers come exclusively from this closed table list.
		if err := h.db.QueryRow(context.Background(), "SELECT COALESCE(jsonb_agg(to_jsonb(s) ORDER BY to_jsonb(s)::text),'[]'::jsonb)::text FROM "+table+" s WHERE incident_id=$1", h.incidentID).Scan(&value); err != nil {
			t.Fatal(err)
		}
		state[table] = value
	}
	return state
}

func requirePortableTaskBindingErrors(t *testing.T) {
	for _, nonnull := range []bool{false, true} {
		name := "null"
		if nonnull {
			name = "nonnull"
		}
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			h := newTasksDecisionsPortabilityHarness(t, "binding-error")
			now := taskDecisionPortableTime()
			taskID, decisionID := uuid.New(), uuid.New()
			seedTaskDecisionEnvelope(t, h, taskID, "task_request", now)
			seedTaskDecisionEnvelope(t, h, decisionID, "decision", now)
			row := validPortableTask(h, taskID, now)
			if nonnull {
				row["decision_record_id"] = decisionID.String()
			}
			sentinel := errors.New("binding capability unavailable")
			reader := failingPortableFieldReader{err: sentinel, target: decisionID}
			port, err := tasksdecisions.NewIncidentBundleSourcePort(tasksdecisionassembly.NewLinkFactsCapability(), reader, tasksdecisionassembly.NewEnvelopeFactReader())
			if err != nil {
				t.Fatal(err)
			}
			input := taskDecisionImportContext(h, "binding-error")
			prepared, err := port.PrepareImport(ctx, taskDecisionBundle([]map[string]any{row}, []map[string]any{validPortableDecision(h, decisionID, now)}), input)
			if err != nil {
				t.Fatal(err)
			}
			before := portableTaskAssociationState(t, h)
			tx, err := h.db.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			err = port.ValidateImportTx(ctx, tx, prepared, input)
			var failure *sourceport.Failure
			if !errors.Is(err, sentinel) || errors.As(err, &failure) {
				t.Fatalf("capability failure converted to semantic outcome: %v", err)
			}
			if err := tx.Rollback(ctx); err != nil {
				t.Fatal(err)
			}
			if after := portableTaskAssociationState(t, h); !reflect.DeepEqual(before, after) {
				t.Fatal("capability failure rollback changed retained state")
			}
		})
	}
}

type failingPortableFieldReader struct {
	err    error
	target uuid.UUID
}

func (r failingPortableFieldReader) LoadFieldReferencesTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, string) ([]tasksdecisions.FieldReferenceFact, error) {
	return []tasksdecisions.FieldReferenceFact{{DestinationRecordID: r.target, LinkType: "references_record"}}, r.err
}
