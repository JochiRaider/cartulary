package revisions_test

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

func TestRevisionsIncidentBundleExactSourceState(t *testing.T) {
	t.Run("exact_integer_storage", exerciseRevisionIntegerStorage)
	t.Run("actually_empty_incident_history", exerciseEmptyIncidentHistory)
	h := newRevisionsPortabilityHarness(t, "exact-source-state")
	ctx := t.Context()
	const version int64 = 1
	for _, table := range []string{"records", "hosts"} {
		if _, err := h.db.Exec(ctx, "UPDATE "+table+" SET row_version=$2 WHERE record_id=$1", h.recordID, version); err != nil {
			t.Fatal(err)
		}
		h.snapshot[map[string]string{"records": "record", "hosts": "source"}[table]].(map[string]any)["row_version"] = json.Number("1")
	}
	otherActor := authstoretest.SeedActor(t, h.db, authstoretest.ActorAttributes{Email: "revisions-state-" + uuid.NewString() + "@example.test", DisplayName: "Other actor"})
	foreign := appsupport.CreateIncidentInStore(t, h.db, h.actor, "txn-revisions-state-"+uuid.NewString(), "IR-RS-"+uuid.NewString()[:8], "Other incident")
	otherRecord := uuid.New()
	portableActor, alternateChangeSet := uuid.New(), uuid.New()
	actors, err := sourceport.NewActorCatalog([]sourceport.ActorDescriptor{{SourceActorID: portableActor.String()}})
	if err != nil {
		t.Fatal(err)
	}
	input := h.context
	input.Actors = actors
	base := func(changesOnly bool) portabilitysupport.MapBundle {
		bundle := h.validBundle(t)
		changes := decodePortableRows(t, bundle[revisionsTestChangeSetsPath])
		changes[0]["actor_user_id"] = portableActor.String()
		changes[0]["reason"], changes[0]["client_txn_id"], changes[0]["request_id"] = "Portable reason", "portable-client", "portable-request"
		if !changesOnly {
			alternate := clonePortableRow(t, changes[0])
			alternate["change_set_id"] = alternateChangeSet.String()
			alternate["history_order"] = 2
			changes = append(changes, alternate)
		}
		sort.Slice(changes, func(i, j int) bool {
			return changes[i]["change_set_id"].(string) < changes[j]["change_set_id"].(string)
		})
		bundle[revisionsTestChangeSetsPath] = encodePortableRows(t, changes)
		for _, path := range []string{revisionsTestMutationsPath, revisionsTestHistoryPath} {
			if changesOnly {
				bundle[path] = nil
				continue
			}
			rows := decodePortableRows(t, bundle[path])
			replacePortableActorIDs(rows[0], h.actor.ID.String(), portableActor.String())
			if path == revisionsTestMutationsPath {
				rows[0]["after_version_id"] = fmt.Sprintf("host:%s:%d", h.recordID, version)
			} else {
				rows[0]["row_version"] = json.Number("1")
			}
			bundle[path] = encodePortableRows(t, rows)
		}
		return bundle
	}
	withApplied := func(t *testing.T, bundle portabilitysupport.MapBundle, check func(pgx.Tx, sourceport.Prepared, sourceport.ImportContext)) {
		t.Helper()
		caseInput := input
		caseInput.Attributions = &revisionsAttributionRecorder{localActor: h.actor.ID}
		prepared, err := h.port.PrepareImport(t.Context(), bundle, caseInput)
		if err != nil {
			t.Fatalf("prepare exact source fixture: %v", err)
		}
		tx, err := h.db.BeginTx(t.Context(), pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(context.Background()) }()
		if err := h.port.ApplyImportTx(t.Context(), tx, prepared, caseInput); err != nil {
			t.Fatalf("apply exact source fixture: %v", err)
		}
		check(tx, prepared, caseInput)
	}
	for _, changesOnly := range []bool{false, true} {
		t.Run(fmt.Sprintf("exact_positive_changes_only=%t", changesOnly), func(t *testing.T) {
			withApplied(t, base(changesOnly), func(tx pgx.Tx, prepared sourceport.Prepared, caseInput sourceport.ImportContext) {
				validationErr := h.port.ValidateImportTx(t.Context(), tx, prepared, caseInput)
				if changesOnly {
					requireRevisionsInvariant(t, validationErr, "revisions.history_reconstruction")
				} else if validationErr != nil {
					t.Fatalf("exact complete source rejected: %v", validationErr)
				}
				var actor uuid.UUID
				var source, reason, client, request string
				var created time.Time
				if err := tx.QueryRow(t.Context(), `SELECT actor_user_id,source,reason,client_txn_id,request_id,created_at FROM change_sets WHERE change_set_id=$1`, h.changeSet).Scan(&actor, &source, &reason, &client, &request, &created); err != nil {
					t.Fatal(err)
				}
				if actor != h.actor.ID || source != "portable.test" || reason != "Portable reason" || client != "portable-client" || request != "portable-request" || !created.Equal(h.createdAt) {
					t.Fatal("stored change-set values differ from independent expectations")
				}
				if !changesOnly {
					var mutationJSON, revisionJSON []byte
					var storedVersion int64
					var historyIDs, entryIDs []uuid.UUID
					if err := tx.QueryRow(t.Context(), `SELECT after_value,history_record_ids,history_entry_record_ids FROM change_set_mutations WHERE change_set_id=$1 AND sequence_no=1`, h.changeSet).Scan(&mutationJSON, &historyIDs, &entryIDs); err != nil {
						t.Fatal(err)
					}
					if err := tx.QueryRow(t.Context(), `SELECT after_json,row_version FROM record_revisions WHERE change_set_id=$1`, h.changeSet).Scan(&revisionJSON, &storedVersion); err != nil {
						t.Fatal(err)
					}
					if storedVersion != version || !reflect.DeepEqual(historyIDs, []uuid.UUID{h.recordID}) || !reflect.DeepEqual(entryIDs, []uuid.UUID{h.recordID}) {
						t.Fatal("stored history version or association arrays differ from expected values")
					}
					for _, raw := range [][]byte{mutationJSON, revisionJSON} {
						snapshot := decodePortableRows(t, append(raw, '\n'))[0]
						for _, member := range []string{"record", "source"} {
							value := snapshot[member].(map[string]any)
							if value["row_version"] != json.Number("1") || value["created_by_user_id"] != portableActor.String() {
								t.Fatalf("stored %s lost exact history integer or portable actor", member)
							}
						}
					}
				}
			})
		})
	}
	portableSnapshot := clonePortableRow(t, h.snapshot)
	replacePortableActorIDs(portableSnapshot, h.actor.ID.String(), portableActor.String())
	wrongIntegerSnapshot := clonePortableRow(t, portableSnapshot)
	wrongIntegerSnapshot["source"].(map[string]any)["row_version"] = json.Number("9007199254740992")
	historyIDs := []uuid.UUID{h.recordID, otherRecord}
	sort.Slice(historyIDs, func(i, j int) bool { return historyIDs[i].String() < historyIDs[j].String() })
	for _, field := range []struct {
		table, column string
		value         any
		json          bool
	}{
		{"change_sets", "change_set_id", alternateChangeSet, false},
		{"change_sets", "incident_id", foreign.ID, false},
		{"change_sets", "actor_user_id", otherActor.ID, false},
		{"change_sets", "source", "Changed source", false},
		{"change_sets", "history_order", int64(17), false},
		{"change_sets", "actor_import_contexts", []any{}, true},
		{"change_sets", "actor_import_contexts", []any{map[string]any{"import_id": h.context.OperationID, "actor_user_id": h.actor.ID.String(), "unknown": true}}, true},
		{"change_sets", "actor_import_contexts", []any{map[string]any{"import_id": h.context.OperationID, "ACTOR_USER_ID": h.actor.ID.String()}}, true},
		{"change_sets", "reason", nil, false},
		{"change_sets", "client_txn_id", nil, false},
		{"change_sets", "request_id", nil, false},
		{"change_sets", "created_at", h.createdAt.Add(time.Microsecond), false},
		{"change_set_mutations", "change_set_id", alternateChangeSet, false},
		{"change_set_mutations", "sequence_no", 2, false},
		{"change_set_mutations", "target_kind", "identity", false},
		{"change_set_mutations", "target_id", otherRecord.String(), false},
		{"change_set_mutations", "operation_kind", "patch", false},
		{"change_set_mutations", "before_version_id", "new-before", false},
		{"change_set_mutations", "after_version_id", nil, false},
		{"change_set_mutations", "before_value", portableSnapshot, true},
		{"change_set_mutations", "after_value", nil, true},
		{"change_set_mutations", "after_value", wrongIntegerSnapshot, true},
		{"change_set_mutations", "history_record_ids", historyIDs, false},
		{"change_set_mutations", "history_entry_record_ids", []uuid.UUID{}, false},
		{"record_revisions", "revision_id", int64(2), false},
		{"record_revisions", "change_set_id", alternateChangeSet, false},
		{"record_revisions", "record_id", otherRecord, false},
		{"record_revisions", "row_version", version + 1, false},
		{"record_revisions", "before_json", portableSnapshot, true},
		{"record_revisions", "after_json", nil, true},
		{"record_revisions", "after_json", wrongIntegerSnapshot, true},
		{"record_revisions", "created_at", h.createdAt.Add(time.Microsecond), false},
	} {
		t.Run("drift_"+field.table+"_"+field.column, func(t *testing.T) {
			withApplied(t, base(field.table == "change_sets"), func(tx pgx.Tx, prepared sourceport.Prepared, caseInput sourceport.ImportContext) {
				if field.column == "record_id" || field.table == "change_set_mutations" && field.column == "target_id" {
					if _, err := tx.Exec(t.Context(), `
INSERT INTO records (record_id,incident_id,record_type,created_by_user_id,created_at,updated_by_user_id,updated_at,row_version)
VALUES ($1,$2,'host',$3,$4,$3,$4,1)
`, otherRecord, h.incidentID, h.actor.ID, h.createdAt); err != nil {
						t.Fatal(err)
					}
					if _, err := tx.Exec(t.Context(), `
INSERT INTO hosts (record_id,incident_id,display_name,host_state,row_version,created_at,updated_at,created_by_user_id,updated_by_user_id)
VALUES ($1,$2,'Other host','canonical',1,$3,$3,$4,$4)
`, otherRecord, h.incidentID, h.createdAt, h.actor.ID); err != nil {
						t.Fatal(err)
					}
				}
				value, cast := field.value, ""
				if field.json && value != nil {
					encoded, err := json.Marshal(value)
					if err != nil {
						t.Fatal(err)
					}
					value, cast = string(encoded), "::jsonb"
				}
				query := "UPDATE " + field.table + " SET " + field.column + "=$2" + cast + " WHERE change_set_id=$1"
				if tag, err := tx.Exec(t.Context(), query, h.changeSet, value); err != nil || tag.RowsAffected() != 1 {
					t.Fatalf("apply one stored-field drift: affected=%d err=%v", tag.RowsAffected(), err)
				}
				requireRevisionsInvariant(t, h.port.ValidateImportTx(t.Context(), tx, prepared, caseInput), "revisions.source_state_exact")
			})
		})
	}
	for _, table := range []string{"change_sets", "change_set_mutations", "record_revisions"} {
		t.Run("missing_"+table, func(t *testing.T) {
			withApplied(t, base(false), func(tx pgx.Tx, prepared sourceport.Prepared, caseInput sourceport.ImportContext) {
				if tag, err := tx.Exec(t.Context(), "DELETE FROM "+table+" WHERE change_set_id=$1", h.changeSet); err != nil || tag.RowsAffected() != 1 {
					t.Fatalf("remove stored row: affected=%d err=%v", tag.RowsAffected(), err)
				}
				requireRevisionsInvariant(t, h.port.ValidateImportTx(t.Context(), tx, prepared, caseInput), "revisions.source_state_exact")
			})
		})
	}
	for _, empty := range []bool{false, true} {
		for _, table := range []string{"change_sets", "change_set_mutations", "record_revisions"} {
			t.Run(fmt.Sprintf("extra_%s_empty=%t", table, empty), func(t *testing.T) {
				bundle := base(false)
				if empty {
					bundle = base(true)
					if table == "change_sets" {
						bundle[revisionsTestChangeSetsPath] = nil
					}
				}
				withApplied(t, bundle, func(tx pgx.Tx, prepared sourceport.Prepared, caseInput sourceport.ImportContext) {
					var err error
					switch table {
					case "change_sets":
						_, err = tx.Exec(t.Context(), `INSERT INTO change_sets(change_set_id,incident_id,actor_user_id,source,created_at) VALUES($1,$2,$3,'extra',$4)`, uuid.New(), h.incidentID, h.actor.ID, h.createdAt)
					case "change_set_mutations":
						_, err = tx.Exec(t.Context(), `INSERT INTO change_set_mutations(change_set_id,sequence_no,target_kind,target_id,operation_kind,history_record_ids,history_entry_record_ids) VALUES($1,2,'host',$2,'patch',$3,$3)`, h.changeSet, h.recordID.String(), []uuid.UUID{h.recordID})
					case "record_revisions":
						_, err = tx.Exec(t.Context(), `INSERT INTO record_revisions(revision_id,change_set_id,record_id,row_version,created_at) VALUES(2,$1,$2,$3,$4)`, h.changeSet, h.recordID, version+1, h.createdAt)
					}
					if err != nil {
						t.Fatal(err)
					}
					requireRevisionsInvariant(t, h.port.ValidateImportTx(t.Context(), tx, prepared, caseInput), "revisions.source_state_exact")
				})
			})
		}
	}
	t.Run("empty_history_rejects_existing_record", func(t *testing.T) {
		withApplied(t, portabilitysupport.MapBundle{revisionsTestChangeSetsPath: nil, revisionsTestMutationsPath: nil, revisionsTestHistoryPath: nil}, func(tx pgx.Tx, prepared sourceport.Prepared, caseInput sourceport.ImportContext) {
			requireRevisionsInvariant(t, h.port.ValidateImportTx(t.Context(), tx, prepared, caseInput), "revisions.history_reconstruction")
		})
	})
	t.Run("reference_precedes_state", func(t *testing.T) {
		withApplied(t, base(false), func(tx pgx.Tx, prepared sourceport.Prepared, caseInput sourceport.ImportContext) {
			if _, err := tx.Exec(t.Context(), `UPDATE change_sets SET source='drift' WHERE change_set_id=$1`, h.changeSet); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(t.Context(), `DELETE FROM hosts WHERE record_id=$1`, h.recordID); err != nil {
				t.Fatal(err)
			}
			requireRevisionsInvariant(t, h.port.ValidateImportTx(t.Context(), tx, prepared, caseInput), "revisions.references_complete")
		})
	})
}
