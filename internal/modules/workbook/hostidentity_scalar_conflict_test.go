package workbook_test

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

func exerciseHostIdentityScalarConflicts(t *testing.T) {
	h, login, actor, incident := ConflictFixture(t, "entity-scalar-conflicts", "IR-ENTITY-SCALAR")
	for _, tc := range []struct{ kind, view, identifier string }{
		{"host", "cartulary.view.hosts.v1", "host.hostname"},
		{"identity", "cartulary.view.identities.v1", "identity.upn"},
	} {
		for _, mode := range []string{"same-field", "disjoint"} {
			t.Run(tc.kind+"/"+mode, func(t *testing.T) {
				field := tc.kind + ".display_name"
				prefix := tc.kind + "-" + mode
				row := requireWorkbookCreate(t, h, login, incident, tc.view, map[string]any{"client_txn_id": prefix + "-create", field: "Original", tc.identifier: prefix + "@example.test"})["row"].(map[string]any)
				id := appsupport.MustUUID(t, row["record_id"].(string))
				patch := func(base int, txn, key string, value any) map[string]any {
					return map[string]any{"view_schema_id": tc.view, "base_row_version": base, "client_txn_id": prefix + txn, "changes": []map[string]any{{"field_key": key, "value": value}}}
				}
				saved := requireWorkbookPatch(t, h, login, id, patch(1, "-saved", field, "Saved"))
				if saved["row"].(map[string]any)["row_version"] != float64(2) {
					t.Fatal("saved version is not 2")
				}
				if mode == "disjoint" {
					body := patch(1, "-rebased", tc.identifier, "rebased-"+prefix+"@example.test")
					rebased := requireWorkbookPatch(t, h, login, id, body)
					changed := rebased["row"].(map[string]any)
					requireCellValue(t, changed, field, "Saved")
					requireCellValue(t, changed, tc.identifier, "rebased-"+prefix+"@example.test")
					if changed["row_version"] != float64(3) {
						t.Fatal("rebased version is not 3")
					}
					var coherent bool
					if err := h.DB.QueryRowContext(t.Context(), `SELECT m.before_version_id=$3 AND m.after_version_id=$4 AND (m.before_value #>> '{record,row_version}')::bigint=2 AND (m.after_value #>> '{record,row_version}')::bigint=3 FROM change_set_mutations m WHERE m.change_set_id=$1 AND m.target_id=$2`, rebased["change_set_id"], id.String(), fmt.Sprintf("%s:%s:2", tc.kind, id), fmt.Sprintf("%s:%s:3", tc.kind, id)).Scan(&coherent); err != nil {
						t.Fatal(err)
					}
					if !coherent {
						t.Fatal("rebased history did not use protected version 2")
					}
					before := snapshotWorkbookState(t, h, login, incident, tc.view)
					replay := requireWorkbookPatch(t, h, login, id, body)
					if !reflect.DeepEqual(replay, rebased) {
						t.Fatal("rebased receipt changed")
					}
					requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, h, login, incident, tc.view))
					return
				}
				before := snapshotWorkbookState(t, h, login, incident, tc.view)
				response := doWorkbookJSON(t, h, login, http.MethodPatch, uuid.Nil, "", id, patch(1, "-local", field, "Local"))
				failure := httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "same_field_conflict")
				requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, h, login, incident, tc.view))
				conflict := failure["error"].(map[string]any)["conflict"].(map[string]any)
				if conflict["record_id"] != id.String() || conflict["field_key"] != field || conflict["base_value"] != "Original" || conflict["server_value"] != "Saved" || conflict["client_value"] != "Local" || conflict["server_updated_by"] != actor.String() || conflict["base_row_version"] != float64(1) || conflict["current_row_version"] != float64(2) {
					t.Fatalf("scalar conflict=%#v", conflict)
				}
				var attributed bool
				if err := h.DB.QueryRowContext(t.Context(), `SELECT created_at=$2::timestamptz AND actor_user_id=$3 FROM change_sets WHERE change_set_id=$1`, saved["change_set_id"], conflict["server_updated_at"], actor).Scan(&attributed); err != nil {
					t.Fatal(err)
				}
				if !attributed {
					t.Fatal("conflict attribution differs from saved change set")
				}
				token, ok := conflict["conflict_token"].(string)
				if !ok || token == "" {
					t.Fatal("missing conflict token")
				}
				keep := map[string]any{"conflict_token": token, "resolution_kind": "keep_saved", "client_txn_id": prefix + "-keep"}
				kept := ResolveConflict(t, h, login, id, token, keep)
				RequireNoChangeSetID(t, kept)
				requireCellValue(t, kept["row"].(map[string]any), field, "Saved")
				requireWorkbookKeepSaved(t, before, snapshotWorkbookState(t, h, login, incident, tc.view))
				before = snapshotWorkbookState(t, h, login, incident, tc.view)
				replay := ResolveConflict(t, h, login, id, token, keep)
				if !reflect.DeepEqual(replay, kept) {
					t.Fatal("keep saved receipt changed")
				}
				requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, h, login, incident, tc.view))
				use := map[string]any{"conflict_token": token, "resolution_kind": "use_unsaved", "client_txn_id": prefix + "-use", "resolved_value": "Local"}
				used := ResolveConflict(t, h, login, id, token, use)
				requireCellValue(t, used["row"].(map[string]any), field, "Local")
				if used["row"].(map[string]any)["row_version"] != float64(3) {
					t.Fatal("resolved version is not 3")
				}
				before = snapshotWorkbookState(t, h, login, incident, tc.view)
				usedReplay := ResolveConflict(t, h, login, id, token, use)
				if !reflect.DeepEqual(usedReplay, used) {
					t.Fatal("use unsaved receipt changed")
				}
				requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, h, login, incident, tc.view))
				stale := map[string]any{"conflict_token": token, "resolution_kind": "use_unsaved", "client_txn_id": prefix + "-stale", "resolved_value": "Stale local"}
				staleFailure := httptestx.RequireErrorEnvelope(t, ResolveConflictRaw(t, h, login, id, token, stale), http.StatusConflict, "same_field_conflict")
				fresh := staleFailure["error"].(map[string]any)["conflict"].(map[string]any)
				if fresh["conflict_token"] == token || fresh["base_row_version"] != float64(2) || fresh["current_row_version"] != float64(3) || fresh["base_value"] != "Saved" || fresh["server_value"] != "Local" || fresh["client_value"] != "Stale local" {
					t.Fatalf("stale resolution=%#v", fresh)
				}
				requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, h, login, incident, tc.view))
			})
		}
	}
}
