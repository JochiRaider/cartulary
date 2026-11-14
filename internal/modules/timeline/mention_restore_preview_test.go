package timeline_test

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func exerciseTimelineMentionRestorePreview(t *testing.T) {
	h := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "timeline-mention-restore-preview", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	incident := appsupport.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "mention-preview-incident", "incident_key": "IR-MENTION-RESTORE-PREVIEW", "title": "Mention restore preview"})["incident_id"].(string)
	send := func(method, path string, body any) *http.Response {
		return httptestx.DoJSON(t, http.DefaultClient, method, h.Server.HTTP.URL+path, body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
	}
	payload := func(actions ...map[string]any) map[string]any {
		return map[string]any{"kind": "collection_actions_v1", "actions": actions}
	}
	for _, kind := range []string{"host", "identity"} {
		view := "cartulary.view.hosts.v1"
		if kind == "identity" {
			view = "cartulary.view.identities.v1"
		}
		target := httptestx.RequireSuccessData(t, send(http.MethodPost, "/api/v1/incidents/"+incident+"/views/"+view+"/rows", map[string]any{"client_txn_id": "mention-preview-target-" + kind, kind + ".display_name": "Target"}), http.StatusCreated)["row"].(map[string]any)["record_id"].(string)
		field := "timeline." + kind + "_refs"
		for _, initial := range []string{"unresolved", "resolved"} {
			prefix := "mention-preview-" + kind + "-" + initial
			first := map[string]any{"op": "add_token", "raw_text": "First"}
			if initial == "resolved" {
				first = map[string]any{"op": "add_resolved_ref", "raw_text": "First", "resolved_record_id": target}
			}
			created := httptestx.RequireSuccessData(t, send(http.MethodPost, "/api/v1/incidents/"+incident+"/views/cartulary.view.timeline.v2/rows", map[string]any{"client_txn_id": prefix + "-create", field: payload(first, map[string]any{"op": "add_token", "raw_text": "Second"})}), http.StatusCreated)
			row := created["row"].(map[string]any)
			id := row["record_id"].(string)
			base := row["cells"].(map[string]any)[field].(map[string]any)["value"].(map[string]any)
			baseItems := base["items"].([]any)
			if len(baseItems) != 2 {
				t.Fatal("Mention preview base fixture missing items")
			}
			firstItem := baseItems[0].(map[string]any)
			secondItem := baseItems[1].(map[string]any)
			expectedItem := func(item map[string]any, text string) map[string]any {
				return map[string]any{"entity_mention_id": item["entity_mention_id"], "item_ref": "entity_mention:" + item["entity_mention_id"].(string), "item_kind": "unresolved_mention", "entity_type": kind, "display_text": text, "raw_text": text, "mention_row_version": float64(1)}
			}
			if firstItem["raw_text"] != "First" || secondItem["raw_text"] != "Second" || firstItem["entity_type"] != kind || secondItem["item_kind"] != "unresolved_mention" {
				t.Fatal("Mention fixture has wrong text, type or order")
			}
			if initial == "unresolved" && firstItem["item_kind"] != "unresolved_mention" {
				t.Fatal("unresolved fixture unexpectedly resolved")
			}
			if initial == "resolved" && (firstItem["item_kind"] != "resolved_ref" || firstItem["resolved_record_id"] != target) {
				t.Fatal("resolved fixture target differs")
			}
			patch := func(txn string, actions ...map[string]any) map[string]any {
				return map[string]any{"view_schema_id": "cartulary.view.timeline.v2", "client_txn_id": prefix + txn, "base_row_version": 1, "changes": []map[string]any{{"field_key": field, "action_payload": payload(actions...)}}}
			}
			saved := httptestx.RequireSuccessData(t, send(http.MethodPatch, "/api/v1/records/"+id, patch("-server", map[string]any{"op": "add_token", "raw_text": "Third"})), http.StatusOK)
			server := saved["row"].(map[string]any)["cells"].(map[string]any)[field].(map[string]any)["value"]
			for _, mode := range []string{"dismiss", "restore"} {
				t.Run(kind+"/"+initial+"/"+mode, func(t *testing.T) {
					actions := []map[string]any{{"op": "dismiss_item", "item_ref": firstItem["item_ref"]}}
					expected := []any{expectedItem(secondItem, "Second")}
					if mode == "restore" {
						actions = append(actions, map[string]any{"op": "revert_to_unresolved", "item_ref": firstItem["item_ref"]})
						expected = []any{expectedItem(firstItem, "First"), expectedItem(secondItem, "Second")}
					}
					request := patch("-client-"+mode, actions...)
					before := snapshotTimelineMutationState(t, h.DB, incident, id)
					for attempt := 0; attempt < 2; attempt++ {
						failure := httptestx.RequireErrorEnvelope(t, send(http.MethodPatch, "/api/v1/records/"+id, request), http.StatusConflict, "same_field_conflict")
						conflict := failure["error"].(map[string]any)["conflict"].(map[string]any)
						if conflict["field_key"] != field || conflict["conflict_resolution_class"] != "collection_review" || conflict["base_row_version"] != float64(1) || conflict["current_row_version"] != float64(2) || !reflect.DeepEqual(conflict["base_value"], base) || !reflect.DeepEqual(conflict["server_value"], server) {
							t.Fatal("Mention preview changed base/server conflict facts")
						}
						want := map[string]any{"kind": "collection_value_v1", "ordered": true, "items": expected}
						if !reflect.DeepEqual(conflict["client_value"], want) {
							t.Fatalf("Mention preview=%#v want=%#v", conflict["client_value"], want)
						}
						if after := snapshotTimelineMutationState(t, h.DB, incident, id); after != before {
							t.Fatal("Mention preview wrote source, projection, history, receipt or publication effects")
						}
					}
				})
			}
			action := func(op string) map[string]any {
				value := map[string]any{"op": op, "item_ref": firstItem["item_ref"]}
				if op == "resolve_item" {
					value["resolved_record_id"] = target
				}
				return value
			}
			cases := []struct {
				name, from, to string
				ops            []string
			}{
				{"double dismissal", "dismissed", "dismissed", []string{"dismiss_item", "dismiss_item"}},
				{"resolve dismissed", "dismissed", "resolved", []string{"dismiss_item", "resolve_item"}},
				{"restore twice", "unresolved", "unresolved", []string{"dismiss_item", "revert_to_unresolved", "revert_to_unresolved"}},
			}
			if initial == "unresolved" {
				cases = append(cases, struct {
					name, from, to string
					ops            []string
				}{"restore unresolved", "unresolved", "unresolved", []string{"revert_to_unresolved"}})
			}
			for _, tc := range cases {
				t.Run(kind+"/"+initial+"/"+tc.name, func(t *testing.T) {
					actions := make([]map[string]any, len(tc.ops))
					for i, op := range tc.ops {
						actions[i] = action(op)
					}
					request := patch("-illegal-"+tc.name, actions...)
					before := snapshotTimelineMutationState(t, h.DB, incident, id)
					for attempt := 0; attempt < 2; attempt++ {
						failure := httptestx.RequireErrorEnvelope(t, send(http.MethodPatch, "/api/v1/records/"+id, request), http.StatusConflict, "illegal_transition")
						details := failure["error"].(map[string]any)["details"]
						expected := map[string]any{"from_status": tc.from, "to_status": tc.to, "violated_guards": []any{"from_status_not_allowed"}}
						if !reflect.DeepEqual(details, expected) {
							t.Fatalf("preview transition details=%#v want=%#v", details, expected)
						}
						if after := snapshotTimelineMutationState(t, h.DB, incident, id); after != before {
							t.Fatal("illegal preview transition wrote persisted effects")
						}
					}
				})
			}

		}
	}
}
