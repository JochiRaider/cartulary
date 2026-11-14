package revisions_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func exerciseNonRowHistoryReferenceScope(t *testing.T) {
	h := newRevisionsPortabilityHarness(t, "nonrow-references")
	reference, missingSourceRecord, foreignRecord := uuid.New(), uuid.New(), uuid.New()
	foreign := appsupport.CreateIncidentInStore(t, h.db, h.actor, "txn-nonrow-"+uuid.NewString(), "IR-NR-"+uuid.NewString()[:8], "Foreign non-row scope")
	seedRevisionsReferenceHost(t, h, reference, h.incidentID)
	seedRevisionsReferenceHost(t, h, missingSourceRecord, h.incidentID)
	seedRevisionsReferenceHost(t, h, foreignRecord, foreign.ID)
	for _, target := range []struct {
		kind             string
		records          []string
		explicitIncident bool
	}{
		{"entity_mention", []string{"source_record_id", "resolved_record_id"}, false},
		{"entity_alias", []string{"record_id"}, true},
		{"entity_preserved_identifier", []string{"record_id"}, true},
		{"indicator_observation", []string{"source_record_id", "resolved_indicator_record_id"}, true},
		{"indicator_state_interval", []string{"indicator_record_id", "support_refs"}, true},
	} {
		tests := []struct{ side, field, defect string }{{"before", target.records[0], "none"}}
		for _, side := range []string{"before", "after"} {
			for _, field := range target.records {
				for _, defect := range []string{"foreign_record", "missing_record", "missing_source"} {
					tests = append(tests, struct{ side, field, defect string }{side, field, defect})
				}
			}
			if target.explicitIncident {
				tests = append(tests, struct{ side, field, defect string }{side, "incident_id", "foreign_incident"})
			}
		}
		for _, test := range tests {
			t.Run(target.kind+"/"+test.side+"/"+test.field+"/"+test.defect, func(t *testing.T) {
				// Deliberately incomplete values and sequence 2 isolate reference
				// precedence. This fixture does not claim full mutation admission
				// or source-type legality from record identity alone.
				before, after := map[string]any{}, map[string]any{}
				for _, field := range target.records {
					before[field], after[field] = reference.String(), reference.String()
					if field == "support_refs" {
						before[field], after[field] = []string{reference.String()}, []string{reference.String()}
					}
				}
				if target.explicitIncident {
					before["incident_id"], after["incident_id"] = h.incidentID.String(), h.incidentID.String()
				}
				value := before
				if test.side == "after" {
					value = after
				}
				switch test.defect {
				case "foreign_record":
					value[test.field] = foreignRecord.String()
				case "missing_record":
					value[test.field] = uuid.NewString()
				case "foreign_incident":
					value[test.field] = foreign.ID.String()
				case "missing_source":
					value[test.field] = missingSourceRecord.String()
				}
				if test.field == "support_refs" && test.defect != "none" {
					value[test.field] = []any{value[test.field]}
				}
				bundle := h.validBundle(t)
				mutations := decodePortableRows(t, bundle[revisionsTestMutationsPath])
				mutations[0]["target_kind"], mutations[0]["target_id"] = target.kind, uuid.NewString()
				mutations[0]["sequence_no"], mutations[0]["before_value"], mutations[0]["after_value"] = 2, before, after
				bundle[revisionsTestMutationsPath] = encodePortableRows(t, mutations)
				input := h.freshImportContext()
				prepared, err := h.port.PrepareImport(t.Context(), bundle, input)
				if err != nil {
					t.Fatalf("prepare reference-priority fixture: %v", err)
				}
				tx, err := h.db.BeginTx(t.Context(), pgx.TxOptions{})
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = tx.Rollback(context.Background()) }()
				if test.defect == "missing_source" {
					if _, err := tx.Exec(t.Context(), `DELETE FROM hosts WHERE record_id=$1`, missingSourceRecord); err != nil {
						t.Fatal(err)
					}
				}
				applyWant, validateWant := "revisions.references_complete", "revisions.references_complete"
				if test.defect == "none" {
					applyWant, validateWant = "revisions.mutation_sequence_contiguous", "revisions.source_state_exact"
				}
				requireRevisionsInvariant(t, h.port.ApplyImportTx(t.Context(), tx, prepared, input), applyWant)
				requireRevisionsInvariant(t, h.port.ValidateImportTx(t.Context(), tx, prepared, input), validateWant)
				for _, table := range []string{"change_sets", "change_set_mutations", "record_revisions"} {
					var count int
					if err := tx.QueryRow(t.Context(), "SELECT count(*) FROM "+table+" WHERE change_set_id=$1", h.changeSet).Scan(&count); err != nil || count != 0 {
						t.Fatalf("pre-write %s count=%d error=%v", table, count, err)
					}
				}
				if err := tx.Rollback(t.Context()); err != nil {
					t.Fatal(err)
				}
				for _, table := range []string{"change_sets", "change_set_mutations", "record_revisions"} {
					var count int
					if err := h.db.QueryRow(t.Context(), "SELECT count(*) FROM "+table+" WHERE change_set_id=$1", h.changeSet).Scan(&count); err != nil || count != 0 {
						t.Fatalf("rollback %s count=%d error=%v", table, count, err)
					}
				}
			})
		}
	}
}
