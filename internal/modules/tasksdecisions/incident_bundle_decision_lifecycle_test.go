package tasksdecisions_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func requirePortableDecisionEndpointLifecycle(t *testing.T) {
	type scenario struct {
		name, status, defect, want                string
		targetDeleted, replacementDeleted, noLink bool
	}
	var cases []scenario
	for _, status := range []string{"superseded", "executed"} {
		for _, targetDeleted := range []bool{false, true} {
			for _, replacementDeleted := range []bool{false, true} {
				want := ""
				if status == "superseded" && !targetDeleted && replacementDeleted {
					want = "tasks_decisions.lifecycle_legal"
				}
				cases = append(cases, scenario{name: fmt.Sprintf("%s/target_deleted=%t/replacement_deleted=%t", status, targetDeleted, replacementDeleted), status: status, targetDeleted: targetDeleted, replacementDeleted: replacementDeleted, want: want})
			}
		}
	}
	cases = append(cases,
		scenario{name: "deleted superseded without retained relation", status: "superseded", targetDeleted: true, noLink: true},
		scenario{name: "deleted unknown status", status: "unknown", targetDeleted: true, want: "tasks_decisions.lifecycle_legal", defect: "prepare"},
		scenario{name: "deleted null owner", status: "superseded", targetDeleted: true, want: "tasks_decisions.dependent_fields_legal", defect: "owner_user_id"},
		scenario{name: "deleted null decision time", status: "superseded", targetDeleted: true, want: "tasks_decisions.dependent_fields_legal", defect: "decided_at"},
		scenario{name: "deleted wrong envelope type", status: "superseded", targetDeleted: true, want: "tasks_decisions.envelope_type_scope", defect: "envelope"},
		scenario{name: "deleted source drift", status: "superseded", targetDeleted: true, want: "tasks_decisions.source_state_exact", defect: "drift"},
	)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			h := newTasksDecisionsPortabilityHarness(t, "decision-endpoint")
			now := taskDecisionPortableTime()
			target := uuid.MustParse("00000000-0000-4000-8000-000000000002")
			replacement := uuid.MustParse("00000000-0000-4000-8000-000000000003")
			targetType := "decision"
			if tc.defect == "envelope" {
				targetType = "host"
			}
			seedTaskDecisionEnvelope(t, h, target, targetType, now)
			seedTaskDecisionEnvelope(t, h, replacement, "decision", now)
			for _, endpoint := range []struct {
				id      uuid.UUID
				deleted bool
			}{{target, tc.targetDeleted}, {replacement, tc.replacementDeleted}} {
				if endpoint.deleted {
					if _, err := h.db.Exec(ctx, `UPDATE records SET deleted_at=$2,deleted_by_user_id=$3 WHERE record_id=$1`, endpoint.id, now, h.actor.ID); err != nil {
						t.Fatal(err)
					}
				}
			}
			linkCount := 0
			if !tc.noLink {
				linkCount = 1
				if _, err := h.db.Exec(ctx, `INSERT INTO record_links
(incident_id,src_record_id,dst_record_id,link_type,field_key,provenance,confidence,owner_user_id,created_by_user_id,decided_at,created_at)
VALUES($1,$2,$3,'supersedes',NULL,'manual',NULL,$4,$4,$5,$5)`, h.incidentID, replacement, target, h.actor.ID, now); err != nil {
					t.Fatal(err)
				}
			}
			row := validPortableDecision(h, target, now)
			row["status"] = tc.status
			if tc.defect == "owner_user_id" || tc.defect == "decided_at" {
				row[tc.defect] = nil
			}
			port := newTaskDecisionIncidentBundleSourcePort(t)
			input := taskDecisionImportContext(h, "decision-endpoint")
			before := portableTaskAssociationState(t, h)
			prepared, err := port.PrepareImport(ctx, taskDecisionBundle(nil, []map[string]any{row, validPortableDecision(h, replacement, now)}), input)
			if tc.defect == "prepare" || tc.defect == "owner_user_id" || tc.defect == "decided_at" {
				requireTasksDecisionsPortabilityFailure(t, err, tc.want)
				if !reflect.DeepEqual(before, portableTaskAssociationState(t, h)) {
					t.Fatal("deleted source preparation changed retained state")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			tx, err := h.db.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			if tc.defect == "drift" {
				if _, err := tx.Exec(ctx, `UPDATE decisions SET summary='Drifted retained source' WHERE record_id=$1`, target); err != nil {
					t.Fatal(err)
				}
			}
			validationErr := port.ValidateImportTx(ctx, tx, prepared, input)
			var count int
			var exact bool
			if err := tx.QueryRow(ctx, `SELECT count(*),bool_and(
(d.record_id=$2 AND d.status=$3 AND (r.deleted_at IS NOT NULL)=$4)
OR (d.record_id=$5 AND d.status='approved' AND (r.deleted_at IS NOT NULL)=$6))
FROM decisions d JOIN records r ON r.record_id=d.record_id AND r.incident_id=d.incident_id
WHERE d.incident_id=$1`, h.incidentID, target, tc.status, tc.targetDeleted, replacement, tc.replacementDeleted).Scan(&count, &exact); err != nil || count != 2 || !exact {
				t.Fatalf("retained Decision states count=%d exact=%t error=%v", count, exact, err)
			}
			if err := tx.QueryRow(ctx, `SELECT count(*),coalesce(bool_and(src_record_id=$2 AND dst_record_id=$3 AND link_type='supersedes' AND field_key IS NULL AND deleted_at IS NULL AND owner_user_id=$4 AND created_by_user_id=$4 AND decided_at=$5 AND created_at=$5),true)
FROM record_links WHERE incident_id=$1`, h.incidentID, replacement, target, h.actor.ID, now).Scan(&count, &exact); err != nil || count != linkCount || !exact {
				t.Fatalf("retained supersession count=%d want=%d exact=%t error=%v", count, linkCount, exact, err)
			}
			if err := tx.Rollback(ctx); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before, portableTaskAssociationState(t, h)) {
				t.Fatal("Decision import rollback changed retained source or relations")
			}
			if tc.want == "" {
				if validationErr != nil {
					t.Fatalf("legal retained Decision lifecycle rejected: %v", validationErr)
				}
			} else {
				requireTasksDecisionsPortabilityFailure(t, validationErr, tc.want)
			}
		})
	}
}
