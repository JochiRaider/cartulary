package revisions_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

func TestRevisionsIncidentBundleExactAttribution(t *testing.T) {
	h := newRevisionsPortabilityHarness(t, "exact-attribution")
	// Each attribution change set carries a real host-create history for its
	// own record: every record must retain reconstructible revision history
	// and every change set needs a mutation (revisions.history_reconstruction).
	secondRecord, secondSnapshot := seedAttributionHostRecord(t, h, "portable-host-attribution")
	portableActors := []uuid.UUID{uuid.New(), uuid.New()}
	changeSets := []string{"11111111-1111-4111-8111-111111111111", "22222222-2222-4222-8222-222222222222"}
	actors, err := sourceport.NewActorCatalog([]sourceport.ActorDescriptor{
		{SourceActorID: portableActors[0].String()}, {SourceActorID: portableActors[1].String()}, {SourceActorID: h.actor.ID.String()},
	})
	if err != nil {
		t.Fatal(err)
	}
	input := h.context
	input.Actors = actors
	base := func(empty bool) portabilitysupport.MapBundle {
		bundle := h.validBundle(t)
		row := decodePortableRows(t, bundle[revisionsTestChangeSetsPath])[0]
		mutation := decodePortableRows(t, bundle[revisionsTestMutationsPath])[0]
		revision := decodePortableRows(t, bundle[revisionsTestHistoryPath])[0]
		var changes, mutations, revisionRows []map[string]any
		if !empty {
			for i, id := range changeSets {
				recordID, snapshot := h.recordID.String(), any(h.snapshot)
				if i == 1 {
					recordID, snapshot = secondRecord.String(), secondSnapshot
				}
				change := clonePortableRow(t, row)
				change["change_set_id"], change["actor_user_id"], change["history_order"] = id, portableActors[i].String(), i+1
				changes = append(changes, change)
				create := clonePortableRow(t, mutation)
				create["change_set_id"], create["target_id"] = id, recordID
				create["after_version_id"], create["after_value"] = "host:"+recordID+":1", snapshot
				mutations = append(mutations, create)
				created := clonePortableRow(t, revision)
				created["revision_id"], created["change_set_id"], created["record_id"], created["after_json"] = i+1, id, recordID, snapshot
				revisionRows = append(revisionRows, created)
			}
		}
		bundle[revisionsTestChangeSetsPath] = encodePortableRows(t, changes)
		bundle[revisionsTestMutationsPath] = encodePortableRows(t, mutations)
		bundle[revisionsTestHistoryPath] = encodePortableRows(t, revisionRows)
		return bundle
	}
	want := []incidentportability.ImportedAttribution{
		{SourceTable: "change_sets", SourceRowID: changeSets[0], SourceColumn: "actor_user_id", SourceActorID: portableActors[0].String(), LocalUserID: h.actor.ID},
		{SourceTable: "change_sets", SourceRowID: changeSets[1], SourceColumn: "actor_user_id", SourceActorID: portableActors[1].String(), LocalUserID: h.actor.ID},
	}
	requireCounts := func(t *testing.T, query incidentportability.Queryer, counts [3]int) {
		t.Helper()
		for i, sql := range []string{
			`SELECT count(*) FROM change_sets WHERE incident_id=$1`,
			`SELECT count(*) FROM change_set_mutations m JOIN change_sets c USING (change_set_id) WHERE c.incident_id=$1`,
			`SELECT count(*) FROM record_revisions r JOIN change_sets c USING (change_set_id) WHERE c.incident_id=$1`,
		} {
			var got int
			if err := query.QueryRow(t.Context(), sql, h.incidentID).Scan(&got); err != nil || got != counts[i] {
				t.Fatalf("history table %d count=%d want=%d err=%v", i, got, counts[i], err)
			}
		}
	}
	withPrepared := func(t *testing.T, bundle portabilitysupport.MapBundle, check func(pgx.Tx, sourceport.Prepared, sourceport.ImportContext, *revisionsAttributionRecorder)) {
		t.Helper()
		caseInput := input
		recorder := &revisionsAttributionRecorder{localActor: h.actor.ID}
		caseInput.Attributions = recorder
		prepared, err := h.port.PrepareImport(t.Context(), bundle, caseInput)
		if err != nil {
			t.Fatalf("prepare attribution fixture: %v", err)
		}
		tx, err := h.db.BeginTx(t.Context(), pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := tx.Rollback(context.Background()); err != nil {
				t.Errorf("rollback attribution transaction: %v", err)
			}
			requireCounts(t, h.db, [3]int{})
		}()
		if len(decodePortableRows(t, bundle[revisionsTestChangeSetsPath])) == 0 {
			// An empty history bundle is legal only for an incident without
			// records; remove them inside this rolled-back transaction.
			for _, recordID := range []uuid.UUID{h.recordID, secondRecord} {
				for _, statement := range []string{`DELETE FROM hosts WHERE record_id=$1`, `DELETE FROM records WHERE record_id=$1`} {
					if _, err := tx.Exec(t.Context(), statement, recordID); err != nil {
						t.Fatalf("remove record for empty history bundle: %v", err)
					}
				}
			}
		}
		check(tx, prepared, caseInput, recorder)
	}
	apply := func(t *testing.T, tx pgx.Tx, prepared sourceport.Prepared, input sourceport.ImportContext) {
		t.Helper()
		if err := h.port.ApplyImportTx(t.Context(), tx, prepared, input); err != nil {
			t.Fatalf("apply attribution fixture: %v", err)
		}
	}
	for _, empty := range []bool{false, true} {
		// Core 01 REQ-01-641: attribution is staged while preparing and flushed
		// by the core in step 9; Revisions records nothing row by row.
		t.Run(fmt.Sprintf("exact/empty=%t", empty), func(t *testing.T) {
			withPrepared(t, base(empty), func(tx pgx.Tx, prepared sourceport.Prepared, input sourceport.ImportContext, recorder *revisionsAttributionRecorder) {
				apply(t, tx, prepared, input)
				if len(recorder.rows) != 0 {
					t.Fatalf("apply recorded attribution row by row: %#v", recorder.rows)
				}
				got := revisionsStagedAttributions(t, h.port, prepared, input, h.actor.ID)
				if empty {
					if len(got) != 0 {
						t.Fatalf("empty input staged attribution: %#v", got)
					}
				} else if !reflect.DeepEqual(got, want) {
					t.Fatalf("attribution=%#v want=%#v", got, want)
				}
				if err := h.port.ValidateImportTx(t.Context(), tx, prepared, input); err != nil {
					t.Fatalf("exact attribution rejected: %v", err)
				}
			})
		})
		for _, phase := range []string{"apply", "validate"} {
			t.Run(fmt.Sprintf("missing_capability/%s/empty=%t", phase, empty), func(t *testing.T) {
				withPrepared(t, base(empty), func(tx pgx.Tx, prepared sourceport.Prepared, input sourceport.ImportContext, _ *revisionsAttributionRecorder) {
					if phase == "validate" {
						apply(t, tx, prepared, input)
					}
					input.Attributions = nil
					if phase == "apply" {
						requireRevisionsInvariant(t, h.port.ApplyImportTx(t.Context(), tx, prepared, input), "revisions.actor_references_complete")
						requireCounts(t, tx, [3]int{})
					} else {
						requireRevisionsInvariant(t, h.port.ValidateImportTx(t.Context(), tx, prepared, input), "revisions.actor_references_complete")
					}
				})
			})
		}
	}
	for _, phase := range []string{"apply", "validate"} {
		t.Run("references_before_attribution/"+phase, func(t *testing.T) {
			withPrepared(t, h.validBundle(t), func(tx pgx.Tx, prepared sourceport.Prepared, input sourceport.ImportContext, _ *revisionsAttributionRecorder) {
				if phase == "validate" {
					apply(t, tx, prepared, input)
				}
				if _, err := tx.Exec(t.Context(), `DELETE FROM hosts WHERE record_id=$1`, h.recordID); err != nil {
					t.Fatal(err)
				}
				input.Attributions = nil
				var err error
				if phase == "apply" {
					err = h.port.ApplyImportTx(t.Context(), tx, prepared, input)
					requireCounts(t, tx, [3]int{})
				} else {
					err = h.port.ValidateImportTx(t.Context(), tx, prepared, input)
				}
				requireRevisionsInvariant(t, err, "revisions.references_complete")
			})
		})
	}
	t.Run("source_state_before_attribution", func(t *testing.T) {
		withPrepared(t, base(false), func(tx pgx.Tx, prepared sourceport.Prepared, input sourceport.ImportContext, _ *revisionsAttributionRecorder) {
			apply(t, tx, prepared, input)
			if _, err := tx.Exec(t.Context(), `UPDATE change_sets SET source='drifted' WHERE change_set_id=$1`, changeSets[0]); err != nil {
				t.Fatal(err)
			}
			input.Attributions = nil
			requireRevisionsInvariant(t, h.port.ValidateImportTx(t.Context(), tx, prepared, input), "revisions.source_state_exact")
		})
	})
	t.Run("missing_capability_before_sequence", func(t *testing.T) {
		bundle := h.validBundle(t)
		mutations := decodePortableRows(t, bundle[revisionsTestMutationsPath])
		mutations[0]["sequence_no"] = 2
		bundle[revisionsTestMutationsPath] = encodePortableRows(t, mutations)
		withPrepared(t, bundle, func(tx pgx.Tx, prepared sourceport.Prepared, input sourceport.ImportContext, _ *revisionsAttributionRecorder) {
			requireRevisionsInvariant(t, h.port.ApplyImportTx(t.Context(), tx, prepared, input), "revisions.mutation_sequence_contiguous")
			input.Attributions = nil
			requireRevisionsInvariant(t, h.port.ApplyImportTx(t.Context(), tx, prepared, input), "revisions.actor_references_complete")
			requireCounts(t, tx, [3]int{})
		})
	})
	t.Run("attribution_before_history", func(t *testing.T) {
		withPrepared(t, h.validBundle(t), func(tx pgx.Tx, prepared sourceport.Prepared, input sourceport.ImportContext, _ *revisionsAttributionRecorder) {
			apply(t, tx, prepared, input)
			if _, err := tx.Exec(t.Context(), `UPDATE hosts SET display_name='drifted' WHERE record_id=$1`, h.recordID); err != nil {
				t.Fatal(err)
			}
			requireRevisionsInvariant(t, h.port.ValidateImportTx(t.Context(), tx, prepared, input), "revisions.history_reconstruction")
			input.Attributions = nil
			requireRevisionsInvariant(t, h.port.ValidateImportTx(t.Context(), tx, prepared, input), "revisions.actor_references_complete")
		})
	})
}

// seedAttributionHostRecord adds a host record to the harness incident and
// returns its canonical snapshot from the host snapshot provider.
func seedAttributionHostRecord(t *testing.T, h revisionsPortabilityHarness, hostname string) (uuid.UUID, map[string]any) {
	t.Helper()
	recordID := uuid.New()
	if _, err := h.db.Exec(context.Background(), `
INSERT INTO records (
    record_id, incident_id, record_type, created_by_user_id, created_at,
    updated_by_user_id, updated_at, row_version
) VALUES ($1, $2, 'host', $3, $4, $3, $4, 1)
`, recordID, h.incidentID, h.actor.ID, h.createdAt); err != nil {
		t.Fatalf("seed attribution record envelope: %v", err)
	}
	if _, err := h.db.Exec(context.Background(), `
INSERT INTO hosts (
    record_id, incident_id, display_name, hostname, host_state,
    row_version, created_at, updated_at, created_by_user_id, updated_by_user_id
) VALUES ($1, $2, 'Attribution host', $3, 'canonical', 1, $4, $4, $5, $5)
`, recordID, h.incidentID, hostname, h.createdAt, h.actor.ID); err != nil {
		t.Fatalf("seed attribution host source: %v", err)
	}
	var snapshot map[string]any
	for _, contribution := range mustRevisionProviderContributions(t) {
		for _, record := range contribution.Records {
			if record.RecordType != "host" {
				continue
			}
			tx, err := h.db.BeginTx(context.Background(), pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			snapshot, err = record.DeleteRestoreSource.SnapshotTx(context.Background(), tx, recordID)
			_ = tx.Rollback(context.Background())
			if err != nil {
				t.Fatalf("load attribution host snapshot: %v", err)
			}
			snapshot["snapshot_schema_id"] = record.SnapshotSchemaID
		}
	}
	if snapshot == nil {
		t.Fatal("host snapshot provider is missing")
	}
	return recordID, snapshot
}
