package revisions_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

func TestRevisionsIncidentBundleExactAttribution(t *testing.T) {
	h := newRevisionsPortabilityHarness(t, "exact-attribution")
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
		var changes []map[string]any
		if !empty {
			for i, id := range changeSets {
				change := clonePortableRow(t, row)
				change["change_set_id"], change["actor_user_id"] = id, portableActors[i].String()
				changes = append(changes, change)
			}
		}
		bundle[revisionsTestChangeSetsPath] = encodePortableRows(t, changes)
		bundle[revisionsTestMutationsPath], bundle[revisionsTestHistoryPath] = nil, nil
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
		check(tx, prepared, caseInput, recorder)
	}
	apply := func(t *testing.T, tx pgx.Tx, prepared sourceport.Prepared, input sourceport.ImportContext) {
		t.Helper()
		if err := h.port.ApplyImportTx(t.Context(), tx, prepared, input); err != nil {
			t.Fatalf("apply attribution fixture: %v", err)
		}
	}
	for _, empty := range []bool{false, true} {
		t.Run(fmt.Sprintf("exact/empty=%t", empty), func(t *testing.T) {
			withPrepared(t, base(empty), func(tx pgx.Tx, prepared sourceport.Prepared, input sourceport.ImportContext, recorder *revisionsAttributionRecorder) {
				apply(t, tx, prepared, input)
				if empty {
					if len(recorder.rows) != 0 {
						t.Fatalf("empty input recorded attribution: %#v", recorder.rows)
					}
				} else if !reflect.DeepEqual(recorder.rows, want) {
					t.Fatalf("attribution=%#v want=%#v", recorder.rows, want)
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
		t.Run(fmt.Sprintf("extra/empty=%t", empty), func(t *testing.T) {
			withPrepared(t, base(empty), func(tx pgx.Tx, prepared sourceport.Prepared, input sourceport.ImportContext, recorder *revisionsAttributionRecorder) {
				apply(t, tx, prepared, input)
				extra := want[0]
				extra.SourceRowID = uuid.NewString()
				recorder.rows = append(recorder.rows, extra)
				requireRevisionsInvariant(t, h.port.ValidateImportTx(t.Context(), tx, prepared, input), "revisions.actor_references_complete")
			})
		})
	}
	for _, defect := range []string{"missing", "duplicate", "table", "row", "column", "source_actor", "local_actor"} {
		for _, index := range []int{0, 1} {
			t.Run(fmt.Sprintf("%s/row=%d", defect, index), func(t *testing.T) {
				withPrepared(t, base(false), func(tx pgx.Tx, prepared sourceport.Prepared, input sourceport.ImportContext, recorder *revisionsAttributionRecorder) {
					apply(t, tx, prepared, input)
					switch defect {
					case "missing":
						recorder.rows = append(recorder.rows[:index], recorder.rows[index+1:]...)
					case "duplicate":
						recorder.rows = append(recorder.rows, recorder.rows[index])
					case "table":
						recorder.rows[index].SourceTable = "record_revisions"
					case "row":
						recorder.rows[index].SourceRowID = uuid.NewString()
					case "column":
						recorder.rows[index].SourceColumn = "owner_user_id"
					case "source_actor":
						recorder.rows[index].SourceActorID = portableActors[1-index].String()
					case "local_actor":
						recorder.rows[index].LocalUserID = portableActors[index]
					}
					requireRevisionsInvariant(t, h.port.ValidateImportTx(t.Context(), tx, prepared, input), "revisions.actor_references_complete")
				})
			})
		}
	}
	for _, phase := range []string{"record", "visit"} {
		for _, empty := range []bool{false, true} {
			if phase == "record" && empty {
				continue
			}
			for _, failure := range []error{errors.New("attribution capability unavailable"), &pgconn.PgError{Code: "40001", ConstraintName: "record_revisions_pkey"}} {
				t.Run(fmt.Sprintf("ordinary_error/%s/empty=%t/%T", phase, empty, failure), func(t *testing.T) {
					withPrepared(t, base(empty), func(tx pgx.Tx, prepared sourceport.Prepared, input sourceport.ImportContext, recorder *revisionsAttributionRecorder) {
						var got error
						if phase == "record" {
							recorder.recordErr = failure
							got = h.port.ApplyImportTx(t.Context(), tx, prepared, input)
							requireCounts(t, tx, [3]int{})
						} else {
							apply(t, tx, prepared, input)
							recorder.visitErr = failure
							got = h.port.ValidateImportTx(t.Context(), tx, prepared, input)
						}
						var semantic *sourceport.Failure
						if !errors.Is(got, failure) || errors.As(got, &semantic) {
							t.Fatalf("capability cause lost or reclassified: got=%v want=%v", got, failure)
						}
					})
				})
			}
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
		withPrepared(t, base(false), func(tx pgx.Tx, prepared sourceport.Prepared, input sourceport.ImportContext, recorder *revisionsAttributionRecorder) {
			apply(t, tx, prepared, input)
			if _, err := tx.Exec(t.Context(), `UPDATE change_sets SET source='drifted' WHERE change_set_id=$1`, changeSets[0]); err != nil {
				t.Fatal(err)
			}
			recorder.rows = nil
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
		withPrepared(t, h.validBundle(t), func(tx pgx.Tx, prepared sourceport.Prepared, input sourceport.ImportContext, recorder *revisionsAttributionRecorder) {
			apply(t, tx, prepared, input)
			if _, err := tx.Exec(t.Context(), `UPDATE hosts SET display_name='drifted' WHERE record_id=$1`, h.recordID); err != nil {
				t.Fatal(err)
			}
			requireRevisionsInvariant(t, h.port.ValidateImportTx(t.Context(), tx, prepared, input), "revisions.history_reconstruction")
			recorder.rows = nil
			requireRevisionsInvariant(t, h.port.ValidateImportTx(t.Context(), tx, prepared, input), "revisions.actor_references_complete")
		})
	})
}
