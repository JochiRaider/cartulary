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

func exerciseTimelineHiddenMentionPreview(t *testing.T) {
	h := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "timeline-hidden-preview", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	incident := appsupport.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "hidden-preview-incident", "incident_key": "IR-HIDDEN-PREVIEW", "title": "Hidden preview"})["incident_id"].(string)
	send := func(method, path string, body any) *http.Response {
		return httptestx.DoJSON(t, http.DefaultClient, method, h.Server.HTTP.URL+path, body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
	}
	payload := func(actions ...map[string]any) map[string]any {
		return map[string]any{"kind": "collection_actions_v1", "actions": actions}
	}
	for _, kind := range []string{"host", "identity"} {
		field := "timeline." + kind + "_refs"
		view := "cartulary.view.hosts.v1"
		if kind == "identity" {
			view = "cartulary.view.identities.v1"
		}
		target := httptestx.RequireSuccessData(t, send(http.MethodPost, "/api/v1/incidents/"+incident+"/views/"+view+"/rows", map[string]any{"client_txn_id": "hidden-target-" + kind, kind + ".display_name": "Target"}), http.StatusCreated)["row"].(map[string]any)["record_id"].(string)
		created := httptestx.RequireSuccessData(t, send(http.MethodPost, "/api/v1/incidents/"+incident+"/views/cartulary.view.timeline.v2/rows", map[string]any{"client_txn_id": "hidden-source-" + kind, field: payload(
			map[string]any{"op": "add_resolved_ref", "raw_text": "First", "resolved_record_id": target},
			map[string]any{"op": "add_token", "raw_text": "Middle"},
			map[string]any{"op": "add_resolved_ref", "raw_text": "Last", "resolved_record_id": target},
		)}), http.StatusCreated)["row"].(map[string]any)
		id := created["record_id"].(string)
		cell := func(row map[string]any) map[string]any {
			return row["cells"].(map[string]any)[field].(map[string]any)["value"].(map[string]any)
		}
		original := cell(created)["items"].([]any)
		if len(original) != 3 {
			t.Fatal("hidden Mention fixture count differs")
		}
		first, middle, last := original[0].(map[string]any), original[1].(map[string]any), original[2].(map[string]any)
		expectedItem := func(item map[string]any, text string, version float64) map[string]any {
			return map[string]any{
				"entity_mention_id": item["entity_mention_id"], "item_ref": "entity_mention:" + item["entity_mention_id"].(string),
				"item_kind": "unresolved_mention", "entity_type": kind, "raw_text": text, "display_text": text, "mention_row_version": version,
			}
		}
		patch := func(txn string, version int, actions ...map[string]any) map[string]any {
			return map[string]any{"client_txn_id": "hidden-" + kind + "-" + txn, "view_schema_id": "cartulary.view.timeline.v2", "base_row_version": version, "changes": []map[string]any{{"field_key": field, "action_payload": payload(actions...)}}}
		}
		action := func(op string, item map[string]any) map[string]any {
			a := map[string]any{"op": op, "item_ref": item["item_ref"]}
			if op == "resolve_item" {
				a["resolved_record_id"] = target
			}
			return a
		}
		hidden := httptestx.RequireSuccessData(t, send(http.MethodPatch, "/api/v1/records/"+id, patch("dismiss", 1, action("dismiss_item", first), action("dismiss_item", last))), http.StatusOK)["row"].(map[string]any)
		base := map[string]any{"kind": "collection_value_v1", "ordered": true, "items": []any{expectedItem(middle, "Middle", 1)}}
		if hidden["row_version"] != float64(2) || !reflect.DeepEqual(cell(hidden), base) {
			t.Fatal("dismissed base fixture differs")
		}
		saved := httptestx.RequireSuccessData(t, send(http.MethodPatch, "/api/v1/records/"+id, patch("server", 2, map[string]any{"op": "add_token", "raw_text": "Tail"})), http.StatusOK)["row"].(map[string]any)
		serverItems := cell(saved)["items"].([]any)
		if saved["row_version"] != float64(3) || len(serverItems) != 2 {
			t.Fatal("hidden preview server fixture differs")
		}
		tail := serverItems[1].(map[string]any)
		server := map[string]any{"kind": "collection_value_v1", "ordered": true, "items": []any{expectedItem(middle, "Middle", 1), expectedItem(tail, "Tail", 1)}}
		if !reflect.DeepEqual(cell(saved), server) {
			t.Fatal("hidden preview server values differ")
		}
		for _, mode := range []string{"restore", "restore both reversed", "restore and resolve", "add then restore", "dismiss hidden", "resolve hidden", "restore twice", "active dismiss", "active resolve", "active revert"} {
			t.Run(kind+"/"+mode, func(t *testing.T) {
				actions := []map[string]any{action("revert_to_unresolved", first)}
				expected := []any{expectedItem(first, "First", 2), expectedItem(middle, "Middle", 1)}
				from, to := "", ""
				invalid := false
				addition := false
				switch mode {
				case "restore both reversed":
					actions = []map[string]any{action("revert_to_unresolved", last), action("revert_to_unresolved", first)}
					expected = append(expected, expectedItem(last, "Last", 2))
				case "restore and resolve":
					actions = append(actions, action("resolve_item", first))
					expected[0].(map[string]any)["item_kind"] = "resolved_ref"
					expected[0].(map[string]any)["resolved_record_id"] = target
				case "add then restore":
					actions = append([]map[string]any{{"op": "add_token", "raw_text": "Client"}}, actions...)
					addition = true
				case "dismiss hidden":
					actions = []map[string]any{action("dismiss_item", first)}
					from, to = "dismissed", "dismissed"
				case "resolve hidden":
					actions = []map[string]any{action("resolve_item", first)}
					from, to = "dismissed", "resolved"
				case "restore twice":
					actions = append(actions, action("revert_to_unresolved", first))
					from, to = "unresolved", "unresolved"
				case "active dismiss":
					actions = []map[string]any{action("dismiss_item", tail)}
					invalid = true
				case "active resolve":
					actions = []map[string]any{action("resolve_item", tail)}
					invalid = true
				case "active revert":
					actions = []map[string]any{action("revert_to_unresolved", tail)}
					invalid = true
				}
				request := patch(mode, 2, actions...)
				before := snapshotTimelineMutationState(t, h.DB, incident, id)
				var firstPreview any
				for attempt := 0; attempt < 2; attempt++ {
					response := send(http.MethodPatch, "/api/v1/records/"+id, request)
					if invalid {
						failure := httptestx.RequireErrorEnvelope(t, response, http.StatusBadRequest, "invalid_mutation_payload")
						want := map[string]any{"field": field, "reason_code": "invalid_value"}
						if !reflect.DeepEqual(failure["error"].(map[string]any)["details"], want) {
							t.Fatal("current-only active item failure differs")
						}
					} else if from != "" {
						failure := httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "illegal_transition")
						want := map[string]any{"from_status": from, "to_status": to, "violated_guards": []any{"from_status_not_allowed"}}
						if !reflect.DeepEqual(failure["error"].(map[string]any)["details"], want) {
							t.Fatal("hidden transition failure differs")
						}
					} else {
						conflict := httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "same_field_conflict")["error"].(map[string]any)["conflict"].(map[string]any)
						if conflict["field_key"] != field || conflict["conflict_resolution_class"] != "collection_review" || conflict["base_row_version"] != float64(2) || conflict["current_row_version"] != float64(3) || !reflect.DeepEqual(conflict["base_value"], base) || !reflect.DeepEqual(conflict["server_value"], server) {
							t.Fatal("hidden preview changed base/server facts")
						}
						client := conflict["client_value"].(map[string]any)
						items := client["items"].([]any)
						count := len(expected)
						if addition {
							count++
						}
						if client["kind"] != "collection_value_v1" || client["ordered"] != true || len(items) != count || !reflect.DeepEqual(items[:len(expected)], expected) {
							t.Fatalf("hidden client preview=%#v expected existing=%#v", client, expected)
						}
						if addition {
							item := items[len(expected)].(map[string]any)
							if len(item) != 5 || item["item_kind"] != "unresolved_mention" || item["entity_type"] != kind || item["raw_text"] != "Client" || item["display_text"] != "Client" {
								t.Fatal("new Mention preview fields differ")
							}
							if ref, ok := item["item_ref"].(string); !ok || ref == "" {
								t.Fatal("new Mention preview identity missing")
							}
						}
						if attempt == 0 {
							firstPreview = client
						} else if !reflect.DeepEqual(client, firstPreview) {
							t.Fatal("hidden client preview is not repeatable")
						}
					}
					if after := snapshotTimelineMutationState(t, h.DB, incident, id); after != before {
						t.Fatal("hidden preview wrote persisted effects")
					}
				}
			})
		}
	}
}
