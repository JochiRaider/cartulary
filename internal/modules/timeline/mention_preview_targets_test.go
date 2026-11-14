package timeline_test

import (
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

func exerciseTimelineMentionPreviewTargets(t *testing.T) {
	h := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "timeline-mention-preview-targets", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	createIncident := func(key string) string {
		return appsupport.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": key, "incident_key": key, "title": key})["incident_id"].(string)
	}
	incident, foreign := createIncident("IR-MENTION-PREVIEW-TARGETS"), createIncident("IR-MENTION-PREVIEW-FOREIGN")
	send := func(method, path string, body any) *http.Response {
		return httptestx.DoJSON(t, http.DefaultClient, method, h.Server.HTTP.URL+path, body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
	}
	create := func(scope, view, txn string, values map[string]any) map[string]any {
		values["client_txn_id"] = txn
		return httptestx.RequireSuccessData(t, send(http.MethodPost, "/api/v1/incidents/"+scope+"/views/"+view+"/rows", values), http.StatusCreated)["row"].(map[string]any)
	}
	wrong := create(incident, "cartulary.view.evidence.v1", "mention-preview-wrong-type", map[string]any{"evidence.title": "Wrong type"})["record_id"].(string)
	payload := func(actions ...map[string]any) map[string]any {
		return map[string]any{"kind": "collection_actions_v1", "actions": actions}
	}
	for _, kind := range []string{"host", "identity"} {
		view := "cartulary.view.hosts.v1"
		if kind == "identity" {
			view = "cartulary.view.identities.v1"
		}
		entity := func(scope, label string) string {
			return create(scope, view, "preview-"+kind+"-"+label, map[string]any{kind + ".display_name": label})["record_id"].(string)
		}
		valid := entity(incident, "Valid")
		otherIncident := entity(foreign, "Foreign")
		deleted := entity(incident, "Deleted")
		httptestx.RequireSuccessData(t, send(http.MethodDelete, "/api/v1/records/"+deleted, map[string]any{"client_txn_id": "preview-delete-" + kind, "base_row_version": 1}), http.StatusOK)
		merged, survivor := entity(incident, "Merged"), entity(incident, "Survivor")
		httptestx.RequireSuccessData(t, send(http.MethodPost, "/api/v1/records/"+survivor+"/merge", map[string]any{"loser_record_id": merged, "survivor_base_row_version": 1, "loser_base_row_version": 1, "client_txn_id": "preview-merge-" + kind, "reason": "Preview target fixture"}), http.StatusOK)
		var mergedLive bool
		query := `SELECT r.deleted_at IS NULL AND h.host_state='merged' FROM records r JOIN hosts h ON h.record_id=r.record_id WHERE r.record_id=$1`
		if kind == "identity" {
			query = `SELECT r.deleted_at IS NULL AND i.identity_state='merged' FROM records r JOIN identities i ON i.record_id=r.record_id WHERE r.record_id=$1`
		}
		if err := h.DB.QueryRowContext(t.Context(), query, merged).Scan(&mergedLive); err != nil || !mergedLive {
			t.Fatalf("merged fixture is not a live merged source: %v", err)
		}
		field := "timeline." + kind + "_refs"
		otherField := "timeline.identity_refs"
		if kind == "identity" {
			otherField = "timeline.host_refs"
		}
		row := create(incident, "cartulary.view.timeline.v2", "preview-target-source-"+kind, map[string]any{
			field:      payload(map[string]any{"op": "add_token", "raw_text": "Original"}),
			otherField: payload(map[string]any{"op": "add_token", "raw_text": "Other field"}),
		})
		id := row["record_id"].(string)
		original := row["cells"].(map[string]any)[field].(map[string]any)["value"].(map[string]any)["items"].([]any)[0].(map[string]any)
		otherRef := row["cells"].(map[string]any)[otherField].(map[string]any)["value"].(map[string]any)["items"].([]any)[0].(map[string]any)["item_ref"].(string)
		foreignRow := create(incident, "cartulary.view.timeline.v2", "preview-foreign-row-"+kind, map[string]any{field: payload(map[string]any{"op": "add_token", "raw_text": "Foreign row"})})
		foreignRef := foreignRow["cells"].(map[string]any)[field].(map[string]any)["value"].(map[string]any)["items"].([]any)[0].(map[string]any)["item_ref"].(string)

		patch := func(txn string, action map[string]any) map[string]any {
			return map[string]any{"client_txn_id": "target-preview-" + kind + "-" + txn, "view_schema_id": "cartulary.view.timeline.v2", "base_row_version": 1, "changes": []map[string]any{{"field_key": field, "action_payload": payload(action)}}}
		}
		saved := httptestx.RequireSuccessData(t, send(http.MethodPatch, "/api/v1/records/"+id, patch("saved", map[string]any{"op": "add_token", "raw_text": "Saved"})), http.StatusOK)
		if saved["row"].(map[string]any)["row_version"] != float64(2) {
			t.Fatal("stale Mention fixture version differs")
		}
		for _, op := range []string{"add_resolved_ref", "resolve_item"} {
			for _, tc := range []struct {
				name, id string
				valid    bool
			}{
				{"missing", uuid.NewString(), false}, {"foreign", otherIncident, false}, {"wrong type", wrong, false}, {"deleted", deleted, false}, {"merged", merged, false}, {"valid", valid, true},
			} {
				t.Run(kind+"/"+op+"/"+tc.name, func(t *testing.T) {
					action := map[string]any{"op": op, "resolved_record_id": tc.id}
					if op == "add_resolved_ref" {
						action["raw_text"] = "Client"
					} else {
						action["item_ref"] = original["item_ref"]
					}
					request := patch(op+"-"+tc.name, action)
					before := snapshotTimelineMutationState(t, h.DB, incident, id)
					foreignBefore := snapshotTimelineMutationState(t, h.DB, foreign, id)
					for attempt := 0; attempt < 2; attempt++ {
						response := send(http.MethodPatch, "/api/v1/records/"+id, request)
						if !tc.valid {
							failure := httptestx.RequireErrorEnvelope(t, response, http.StatusBadRequest, "invalid_mutation_payload")
							details := failure["error"].(map[string]any)["details"].(map[string]any)
							if len(details) != 2 || details["field"] != field || details["reason_code"] != "invalid_value" {
								t.Fatalf("target preview failure=%#v", details)
							}
						} else {
							conflict := httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "same_field_conflict")["error"].(map[string]any)["conflict"].(map[string]any)
							if conflict["field_key"] != field || conflict["base_row_version"] != float64(1) || conflict["current_row_version"] != float64(2) {
								t.Fatal("eligible target changed conflict scope/version")
							}
							items := conflict["client_value"].(map[string]any)["items"].([]any)
							wantCount, index, text := 1, 0, "Original"
							if op == "add_resolved_ref" {
								wantCount, index, text = 2, 1, "Client"
							}
							if len(items) != wantCount {
								t.Fatal("eligible preview changed Mention count")
							}
							item := items[index].(map[string]any)
							if item["item_kind"] != "resolved_ref" || item["resolved_record_id"] != valid || item["entity_type"] != kind || item["raw_text"] != text {
								t.Fatal("eligible preview lost intended resolved target")
							}
							if op == "resolve_item" && item["item_ref"] != original["item_ref"] {
								t.Fatal("resolve preview changed Mention identity")
							}
						}
						if after := snapshotTimelineMutationState(t, h.DB, incident, id); after != before {
							t.Fatal("target preview wrote source incident effects")
						}
						if after := snapshotTimelineMutationState(t, h.DB, foreign, id); after != foreignBefore {
							t.Fatal("target preview wrote foreign incident effects")
						}
					}
				})
			}
		}
		for _, op := range []string{"resolve_item", "dismiss_item", "revert_to_unresolved"} {
			for _, tc := range []struct{ name, ref string }{
				{"missing", "entity_mention:" + uuid.NewString()}, {"foreign row", foreignRef}, {"wrong field", otherRef},
			} {
				t.Run(kind+"/item/"+op+"/"+tc.name, func(t *testing.T) {
					action := map[string]any{"op": op, "item_ref": tc.ref}
					if op == "resolve_item" {
						action["resolved_record_id"] = valid
					}
					request := patch("item-"+op+"-"+tc.name, action)
					before := snapshotTimelineMutationState(t, h.DB, incident, id)
					foreignBefore := snapshotTimelineMutationState(t, h.DB, foreign, id)
					for attempt := 0; attempt < 2; attempt++ {
						failure := httptestx.RequireErrorEnvelope(t, send(http.MethodPatch, "/api/v1/records/"+id, request), http.StatusBadRequest, "invalid_mutation_payload")
						details := failure["error"].(map[string]any)["details"].(map[string]any)
						if len(details) != 2 || details["field"] != field || details["reason_code"] != "invalid_value" {
							t.Fatalf("Mention item preview rejection=%#v", details)
						}
						if after := snapshotTimelineMutationState(t, h.DB, incident, id); after != before {
							t.Fatal("invalid Mention item wrote source incident effects")
						}
						if after := snapshotTimelineMutationState(t, h.DB, foreign, id); after != foreignBefore {
							t.Fatal("invalid Mention item wrote foreign incident effects")
						}
					}
				})
			}
		}

	}
}
