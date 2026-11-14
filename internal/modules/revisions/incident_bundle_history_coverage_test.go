package revisions_test

import (
	"context"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

func exerciseEmptyIncidentHistory(t *testing.T) {
	h := newRevisionsPortabilityHarness(t, "empty-incident-history")
	input := h.freshImportContext()
	bundle := portabilitysupport.MapBundle{revisionsTestChangeSetsPath: nil, revisionsTestMutationsPath: nil, revisionsTestHistoryPath: nil}
	prepared, err := h.port.PrepareImport(t.Context(), bundle, input)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := h.db.BeginTx(t.Context(), pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	for _, table := range []string{"hosts", "records"} {
		if tag, err := tx.Exec(t.Context(), "DELETE FROM "+table+" WHERE record_id=$1", h.recordID); err != nil || tag.RowsAffected() != 1 {
			t.Fatalf("remove %s: %v affected=%d", table, err, tag.RowsAffected())
		}
	}
	if err := h.port.ApplyImportTx(t.Context(), tx, prepared, input); err != nil {
		t.Fatal(err)
	}
	if err := h.port.ValidateImportTx(t.Context(), tx, prepared, input); err != nil {
		t.Fatalf("actually empty incident rejected: %v", err)
	}
	for _, table := range []string{"records", "change_sets"} {
		var count int
		if err := tx.QueryRow(t.Context(), "SELECT count(*) FROM "+table+" WHERE incident_id=$1", h.incidentID).Scan(&count); err != nil || count != 0 {
			t.Fatalf("empty %s count=%d error=%v", table, count, err)
		}
	}
	if err := tx.Rollback(t.Context()); err != nil {
		t.Fatal(err)
	}
	var retained int
	if err := h.db.QueryRow(t.Context(), `SELECT count(*) FROM records WHERE record_id=$1`, h.recordID).Scan(&retained); err != nil || retained != 1 {
		t.Fatalf("borrowed rollback lost fixture record: count=%d error=%v", retained, err)
	}
}

func exerciseIncidentHistoryCoverage(t *testing.T) {
	for _, test := range []string{"complete_creation", "empty_members", "missing_revisions", "second_record_missing", "second_record_retained", "deleted_record_missing", "source_drift_precedes", "foreign_record_excluded"} {
		t.Run(test, func(t *testing.T) {
			h := newRevisionsPortabilityHarness(t, "history-coverage")
			bundle := h.validBundle(t)
			want := "revisions.history_reconstruction"
			changes, mutations, history := 1, 1, 1
			switch test {
			case "complete_creation":
				want = ""
			case "empty_members", "deleted_record_missing":
				bundle = portabilitysupport.MapBundle{revisionsTestChangeSetsPath: nil, revisionsTestMutationsPath: nil, revisionsTestHistoryPath: nil}
				changes, mutations, history = 0, 0, 0
			case "missing_revisions", "source_drift_precedes":
				bundle[revisionsTestHistoryPath] = nil
				history = 0
				if test == "source_drift_precedes" {
					want = "revisions.source_state_exact"
				}
			case "second_record_missing", "second_record_retained":
				other := uuid.New()
				seedRevisionsReferenceHost(t, h, other, h.incidentID)
				if test == "second_record_retained" {
					source := clonePortableRow(t, h.snapshot["source"].(map[string]any))
					source["record_id"], source["display_name"], source["hostname"] = other.String(), "Reference host", nil
					appendPortableCreationFixture(t, bundle, h, other, "host", "cartulary.revisions.snapshot.host.v1", source)
					mutations, history, want = 2, 2, ""
				}
			case "foreign_record_excluded":
				foreign := appsupport.CreateIncidentInStore(t, h.db, h.actor, "txn-coverage-"+uuid.NewString(), "IR-C-"+uuid.NewString()[:8], "Other coverage incident")
				seedRevisionsReferenceHost(t, h, uuid.New(), foreign.ID)
				want = ""
			}
			input := h.freshImportContext()
			prepared, err := h.port.PrepareImport(t.Context(), bundle, input)
			if err != nil {
				t.Fatal(err)
			}
			tx, err := h.db.BeginTx(t.Context(), pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(context.Background()) }()
			if test == "deleted_record_missing" {
				if _, err := tx.Exec(t.Context(), `UPDATE records SET deleted_at=updated_at,deleted_by_user_id=$2 WHERE record_id=$1`, h.recordID, h.actor.ID); err != nil {
					t.Fatal(err)
				}
			}
			if err := h.port.ApplyImportTx(t.Context(), tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			for table, expected := range map[string]int{"change_sets": changes, "change_set_mutations": mutations, "record_revisions": history} {
				var got int
				if err := tx.QueryRow(t.Context(), "SELECT count(*) FROM "+table+" WHERE change_set_id=$1", h.changeSet).Scan(&got); err != nil || got != expected {
					t.Fatalf("%s count=%d want=%d error=%v", table, got, expected, err)
				}
			}
			if test == "source_drift_precedes" {
				if _, err := tx.Exec(t.Context(), `UPDATE change_sets SET reason='drift' WHERE change_set_id=$1`, h.changeSet); err != nil {
					t.Fatal(err)
				}
			}
			err = h.port.ValidateImportTx(t.Context(), tx, prepared, input)
			if want == "" {
				if err != nil {
					t.Fatalf("complete incident history rejected: %v", err)
				}
			} else {
				requireRevisionsInvariant(t, err, want)
			}
			if err := tx.Rollback(t.Context()); err != nil {
				t.Fatal(err)
			}
			for _, table := range []string{"change_sets", "change_set_mutations", "record_revisions"} {
				var got int
				if err := h.db.QueryRow(t.Context(), "SELECT count(*) FROM "+table+" WHERE change_set_id=$1", h.changeSet).Scan(&got); err != nil || got != 0 {
					t.Fatalf("%s rollback count=%d error=%v", table, got, err)
				}
			}
		})
	}
}
