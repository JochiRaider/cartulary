package tasksdecisions_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func requirePortableTaskReferenceLifecycle(t *testing.T) {
	for _, kind := range []string{"party", "decision"} {
		for _, deleted := range []bool{false, true} {
			for _, targetState := range []string{"active", "deleted", "missing", "missing source", "wrong type", "foreign", "null"} {
				t.Run(fmt.Sprintf("%s/task_deleted=%t/%s", kind, deleted, targetState), func(t *testing.T) {
					ctx := context.Background()
					h := newTasksDecisionsPortabilityHarness(t, "reference-lifecycle")
					now := taskDecisionPortableTime()
					task, target := uuid.New(), uuid.New()
					seedTaskDecisionEnvelope(t, h, task, "task_request", now)
					if deleted {
						if _, err := h.db.Exec(ctx, `UPDATE records SET deleted_at=$2,deleted_by_user_id=$3 WHERE record_id=$1`, task, now, h.actor.ID); err != nil {
							t.Fatal(err)
						}
					}
					targetIncident := h.incidentID
					if targetState == "foreign" {
						foreign := appsupport.CreateIncidentInStore(t, h.db, h.actor, "txn-portable-target-"+uuid.NewString(), "IR-TARGET-"+uuid.NewString(), "Foreign target")
						targetIncident = foreign.ID
					}
					var decisions []map[string]any
					if targetState != "null" && targetState != "missing" {
						recordType := kind
						if targetState == "wrong type" {
							recordType = "host"
						}
						seedEnvelopeForIncident(t, h, targetIncident, target, recordType, now)
						if targetState != "missing source" && targetState != "wrong type" {
							if kind == "party" {
								if _, err := h.db.Exec(ctx, `INSERT INTO parties(record_id,incident_id,display_name,party_kind) VALUES($1,$2,'Retained requester','organization')`, target, targetIncident); err != nil {
									t.Fatal(err)
								}
							} else if targetState == "foreign" {
								if _, err := h.db.Exec(ctx, `INSERT INTO decisions(record_id,incident_id,summary,status,owner_user_id,decision_type,decided_at,rationale) VALUES($1,$2,'Foreign decision','approved',$3,'containment',$4,'Rationale')`, target, targetIncident, h.actor.ID, now); err != nil {
									t.Fatal(err)
								}
							} else {
								decisions = []map[string]any{validPortableDecision(h, target, now)}
							}
						}
						if targetState == "deleted" {
							if _, err := h.db.Exec(ctx, `UPDATE records SET deleted_at=$2,deleted_by_user_id=$3 WHERE record_id=$1`, target, now, h.actor.ID); err != nil {
								t.Fatal(err)
							}
						}
					}
					row := validPortableTask(h, task, now)
					row["requester_party_text"] = "Independent retained requester text"
					var wantParty, wantDecision *uuid.UUID
					if targetState != "null" {
						if kind == "party" {
							row["requester_party_id"] = target.String()
							wantParty = &target
						} else {
							row["decision_record_id"] = target.String()
							wantDecision = &target
						}
					}
					port := newTaskDecisionIncidentBundleSourcePort(t)
					input := taskDecisionImportContext(h, "reference-lifecycle")
					prepared, err := port.PrepareImport(ctx, taskDecisionBundle([]map[string]any{row}, decisions), input)
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
					validationErr := port.ValidateImportTx(ctx, tx, prepared, input)
					var actualParty, actualDecision *uuid.UUID
					var actualText string
					if err := tx.QueryRow(ctx, `SELECT requester_party_id,decision_record_id,requester_party_text FROM task_requests WHERE record_id=$1`, task).Scan(&actualParty, &actualDecision, &actualText); err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(actualParty, wantParty) || !reflect.DeepEqual(actualDecision, wantDecision) || actualText != "Independent retained requester text" {
						t.Fatal("portable reference admission normalized or cleared source values")
					}
					if err := tx.Rollback(ctx); err != nil {
						t.Fatal(err)
					}
					if after := portableTaskAssociationState(t, h); !reflect.DeepEqual(before, after) {
						t.Fatal("portable reference transaction rollback changed retained state")
					}
					valid := targetState == "null" || targetState == "active" || (deleted && targetState == "deleted")
					if valid {
						if validationErr != nil {
							t.Fatalf("legal retained Task reference rejected: %v", validationErr)
						}
					} else {
						requireTasksDecisionsPortabilityFailure(t, validationErr, "tasks_decisions.references_same_incident")
					}
				})
			}
		}
	}
}
