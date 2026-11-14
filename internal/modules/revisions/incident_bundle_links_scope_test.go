package revisions_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

func seedRevisionsReferenceHost(t testing.TB, h revisionsPortabilityHarness, recordID, incidentID uuid.UUID) {
	t.Helper()
	if _, err := h.db.Exec(t.Context(), `INSERT INTO records
(record_id,incident_id,record_type,created_by_user_id,created_at,updated_by_user_id,updated_at,row_version)
VALUES($1,$2,'host',$3,$4,$3,$4,1)`, recordID, incidentID, h.actor.ID, h.createdAt); err != nil {
		t.Fatal(err)
	}
	if _, err := h.db.Exec(t.Context(), `INSERT INTO hosts
(record_id,incident_id,display_name,host_state,row_version,created_at,updated_at,created_by_user_id,updated_by_user_id)
VALUES($1,$2,'Reference host','canonical',1,$3,$3,$4,$4)`, recordID, incidentID, h.createdAt, h.actor.ID); err != nil {
		t.Fatal(err)
	}
}

func exerciseHistoricalLinksReferenceScope(t *testing.T) {
	t.Helper()
	h := newRevisionsPortabilityHarness(t, "historical-links-scope")
	destination, foreignRecord := uuid.New(), uuid.New()
	foreign := appsupport.CreateIncidentInStore(t, h.db, h.actor, "txn-history-scope-"+uuid.NewString(), "IR-HS-"+uuid.NewString()[:8], "Foreign history scope")
	seedRevisionsReferenceHost(t, h, destination, h.incidentID)
	seedRevisionsReferenceHost(t, h, foreignRecord, foreign.ID)
	base := func(kind string) portabilitysupport.MapBundle {
		bundle := h.validBundle(t)
		id := uuid.New()
		var before map[string]any
		target := id.String()
		if kind == "record_link" {
			before = canonicalPortableLinkValue(h, id, destination)
		} else {
			before = canonicalPortableTagValue(h, id)
			target = "record_tag:" + h.recordID.String() + ":" + id.String()
		}
		after := clonePortableRow(t, before)
		if kind == "record_link" {
			after["decided_at"] = h.createdAt.Add(time.Microsecond).Format(time.RFC3339Nano)
		} else {
			after["record_id"] = destination.String()
			after["updated_at"] = h.createdAt.Add(time.Microsecond).Format(time.RFC3339Nano)
		}
		mutations := decodePortableRows(t, bundle[revisionsTestMutationsPath])
		mutations = append(mutations, map[string]any{
			"change_set_id": h.changeSet.String(), "sequence_no": 2, "target_kind": kind, "target_id": target,
			"operation_kind": "patch", "before_version_id": nil, "after_version_id": nil, "before_value": before, "after_value": after,
		})
		bundle[revisionsTestMutationsPath] = encodePortableRows(t, mutations)
		source := clonePortableRow(t, h.snapshot["source"].(map[string]any))
		source["record_id"], source["display_name"], source["hostname"] = destination.String(), "Reference host", nil
		appendPortableCreationFixture(t, bundle, h, destination, "host", "cartulary.revisions.snapshot.host.v1", source)
		return bundle
	}
	requireCounts := func(t *testing.T, tx pgx.Tx, want [3]int) {
		t.Helper()
		for i, table := range []string{"change_sets", "change_set_mutations", "record_revisions"} {
			var count int
			if err := tx.QueryRow(t.Context(), "SELECT count(*) FROM "+table+" WHERE change_set_id=$1", h.changeSet).Scan(&count); err != nil || count != want[i] {
				t.Fatalf("%s count=%d want=%d err=%v", table, count, want[i], err)
			}
		}
	}
	withPrepared := func(t *testing.T, bundle portabilitysupport.MapBundle, check func(pgx.Tx, sourceport.Prepared, sourceport.ImportContext)) {
		t.Helper()
		input := h.freshImportContext()
		prepared, err := h.port.PrepareImport(t.Context(), bundle, input)
		if err != nil {
			t.Fatalf("prepare reference fixture: %v", err)
		}
		tx, err := h.db.BeginTx(t.Context(), pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := tx.Rollback(context.Background()); err != nil {
				t.Errorf("rollback references: %v", err)
			}
			for _, table := range []string{"change_sets", "change_set_mutations", "record_revisions"} {
				var count int
				if err := h.db.QueryRow(t.Context(), "SELECT count(*) FROM "+table+" WHERE change_set_id=$1", h.changeSet).Scan(&count); err != nil || count != 0 {
					t.Errorf("rollback retained %s rows=%d err=%v", table, count, err)
				}
			}
		}()
		check(tx, prepared, input)
	}
	for _, kind := range []string{"record_link", "record_tag"} {
		t.Run("historical_scope/"+kind+"/positive", func(t *testing.T) {
			withPrepared(t, base(kind), func(tx pgx.Tx, prepared sourceport.Prepared, input sourceport.ImportContext) {
				if err := h.port.ApplyImportTx(t.Context(), tx, prepared, input); err != nil {
					t.Fatalf("apply valid references: %v", err)
				}
				if err := h.port.ValidateImportTx(t.Context(), tx, prepared, input); err != nil {
					t.Fatalf("validate valid references: %v", err)
				}
				requireCounts(t, tx, [3]int{1, 3, 2})
			})
		})
		fields := []string{"incident_id", "record_id"}
		if kind == "record_link" {
			fields = []string{"incident_id", "src_record_id", "dst_record_id"}
		}
		for _, side := range []string{"before_value", "after_value"} {
			for _, field := range fields {
				for _, defect := range []string{"foreign", "missing", "zero", "malformed", "null"} {
					t.Run(fmt.Sprintf("historical_scope/%s/%s/%s/%s", kind, side, field, defect), func(t *testing.T) {
						bundle := base(kind)
						mutations := decodePortableRows(t, bundle[revisionsTestMutationsPath])
						value := mutations[1][side].(map[string]any)
						switch defect {
						case "foreign":
							value[field] = foreignRecord.String()
							if field == "incident_id" {
								value[field] = foreign.ID.String()
							}
						case "missing":
							value[field] = uuid.NewString()
						case "zero":
							value[field] = uuid.Nil.String()
						case "malformed":
							value[field] = "not-a-uuid"
						case "null":
							value[field] = nil
						}
						bundle[revisionsTestMutationsPath] = encodePortableRows(t, mutations)
						withPrepared(t, bundle, func(tx pgx.Tx, prepared sourceport.Prepared, input sourceport.ImportContext) {
							requireRevisionsInvariant(t, h.port.ApplyImportTx(t.Context(), tx, prepared, input), "revisions.references_complete")
							requireCounts(t, tx, [3]int{})
						})
					})
				}
			}
		}
		for _, phase := range []string{"apply", "validate"} {
			t.Run("historical_scope/"+kind+"/missing_source/"+phase, func(t *testing.T) {
				withPrepared(t, base(kind), func(tx pgx.Tx, prepared sourceport.Prepared, input sourceport.ImportContext) {
					if phase == "validate" {
						if err := h.port.ApplyImportTx(t.Context(), tx, prepared, input); err != nil {
							t.Fatal(err)
						}
					}
					if _, err := tx.Exec(t.Context(), `DELETE FROM hosts WHERE record_id=$1`, destination); err != nil {
						t.Fatal(err)
					}
					if phase == "apply" {
						requireRevisionsInvariant(t, h.port.ApplyImportTx(t.Context(), tx, prepared, input), "revisions.references_complete")
						requireCounts(t, tx, [3]int{})
					} else {
						if _, err := tx.Exec(t.Context(), `UPDATE change_sets SET source='drifted' WHERE change_set_id=$1`, h.changeSet); err != nil {
							t.Fatal(err)
						}
						requireRevisionsInvariant(t, h.port.ValidateImportTx(t.Context(), tx, prepared, input), "revisions.references_complete")
					}
				})
			})
		}
		for _, lower := range []string{"sequence", "history"} {
			t.Run("historical_scope/"+kind+"/reference_before_"+lower, func(t *testing.T) {
				bundle := base(kind)
				mutations := decodePortableRows(t, bundle[revisionsTestMutationsPath])
				if lower == "sequence" {
					mutations[0]["sequence_no"] = 3
				} else {
					mutations[0]["after_value"].(map[string]any)["snapshot_schema_id"] = "invalid"
				}
				value := mutations[1]["after_value"].(map[string]any)
				value["incident_id"] = foreign.ID.String()
				bundle[revisionsTestMutationsPath] = encodePortableRows(t, mutations)
				withPrepared(t, bundle, func(tx pgx.Tx, prepared sourceport.Prepared, input sourceport.ImportContext) {
					requireRevisionsInvariant(t, h.port.ApplyImportTx(t.Context(), tx, prepared, input), "revisions.references_complete")
					requireCounts(t, tx, [3]int{})
				})
			})
		}
	}
}
