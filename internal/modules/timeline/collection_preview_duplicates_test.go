package timeline_test

import (
	"net/http"
	"reflect"
	"sort"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

func exerciseTimelineCollectionPreviews(t *testing.T) {
	h := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "timeline-preview-duplicates", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	incident := appsupport.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "preview-duplicate-incident", "incident_key": "IR-PREVIEW-DUPLICATES", "title": "Preview duplicates"})["incident_id"].(string)
	send := func(method, url string, body any) *http.Response {
		return httptestx.DoJSON(t, http.DefaultClient, method, h.Server.HTTP.URL+url, body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
	}
	evidence := map[string]string{}
	for _, label := range []string{"Original", "Saved", "Café"} {
		evidence[label] = httptestx.RequireSuccessData(t, send(http.MethodPost, "/api/v1/incidents/"+incident+"/views/cartulary.view.evidence.v1/rows", map[string]any{"client_txn_id": "preview-evidence-" + label, "evidence.title": label}), http.StatusCreated)["row"].(map[string]any)["record_id"].(string)
	}
	payload := func(actions ...map[string]any) map[string]any {
		return map[string]any{"kind": "collection_actions_v1", "actions": actions}
	}
	for _, field := range []string{"timeline.tags", "timeline.attached_evidence_ids"} {
		for _, mode := range []string{"existing", "new", "replace", "missing", "current-only", "twice-removed", "base-removed-on-server"} {
			t.Run(field+"/"+mode, func(t *testing.T) {
				prefix := "preview-" + field + "-" + mode
				tag := field == "timeline.tags"
				add := func(label string) map[string]any {
					if tag {
						return map[string]any{"op": "add_tag", "tag_name": label}
					}
					return map[string]any{"op": "add_record_ref", "linked_record_id": evidence[label]}
				}
				created := httptestx.RequireSuccessData(t, send(http.MethodPost, "/api/v1/incidents/"+incident+"/views/cartulary.view.timeline.v2/rows", map[string]any{"client_txn_id": prefix + "-create", field: payload(add("Original"))}), http.StatusCreated)
				row := created["row"].(map[string]any)
				id := row["record_id"].(string)
				base := row["cells"].(map[string]any)[field].(map[string]any)["value"].(map[string]any)
				original := base["items"].([]any)[0].(map[string]any)
				patch := func(txn string, actions ...map[string]any) map[string]any {
					return map[string]any{"client_txn_id": prefix + txn, "view_schema_id": "cartulary.view.timeline.v2", "base_row_version": 1, "changes": []map[string]any{{"field_key": field, "action_payload": payload(actions...)}}}
				}
				remove := func(ref any) map[string]any {
					op := "remove_record_ref"
					if tag {
						op = "remove_tag"
					}
					return map[string]any{"op": op, "item_ref": ref}
				}
				serverActions := []map[string]any{add("Saved")}
				if mode == "base-removed-on-server" {
					serverActions = append(serverActions, remove(original["item_ref"]))
				}
				saved := httptestx.RequireSuccessData(t, send(http.MethodPatch, "/api/v1/records/"+id, patch("-save", serverActions...)), http.StatusOK)
				server := saved["row"].(map[string]any)["cells"].(map[string]any)[field].(map[string]any)["value"]
				duplicate := add("Original")
				if tag {
					duplicate = add(" ORIGINAL ")
				}
				actions := []map[string]any{add("Original"), duplicate}
				if mode == "new" {
					duplicate = add("Café")
					if tag {
						duplicate = add(" CAFE\u0301 ")
					}
					actions = []map[string]any{add("Café"), duplicate, add("Café")}
				}
				if mode == "replace" {
					op := "remove_record_ref"
					if tag {
						op = "remove_tag"
					}
					actions = append([]map[string]any{{"op": op, "item_ref": original["item_ref"]}}, actions...)
				}
				switch mode {
				case "missing":
					missing := "record_ref:" + uuid.NewString()
					if tag {
						missing = "record_tag:" + id + ":" + uuid.NewString()
					}
					actions = []map[string]any{remove(missing)}
				case "current-only":
					var savedRef string
					for _, entry := range server.(map[string]any)["items"].([]any) {
						item := entry.(map[string]any)
						if item["item_ref"] != original["item_ref"] {
							savedRef = item["item_ref"].(string)
						}
					}
					if savedRef == "" {
						t.Fatal("current-only fixture missing")
					}
					actions = []map[string]any{remove(savedRef)}
				case "twice-removed":
					actions = []map[string]any{remove(original["item_ref"]), remove(original["item_ref"])}
				case "base-removed-on-server":
					actions = []map[string]any{remove(original["item_ref"])}
				}
				request := patch("-client", actions...)
				before := snapshotTimelineMutationState(t, h.DB, incident, id)
				if mode == "missing" || mode == "current-only" || mode == "twice-removed" {
					for attempt := 0; attempt < 2; attempt++ {
						failure := httptestx.RequireErrorEnvelope(t, send(http.MethodPatch, "/api/v1/records/"+id, request), http.StatusBadRequest, "invalid_mutation_payload")
						details := failure["error"].(map[string]any)["details"].(map[string]any)
						if len(details) != 2 || details["field"] != field || details["reason_code"] != "invalid_value" {
							t.Fatalf("invalid removal details=%#v", details)
						}
						if after := snapshotTimelineMutationState(t, h.DB, incident, id); after != before {
							t.Fatal("invalid preview removal wrote effects")
						}
					}
					return
				}
				failure := httptestx.RequireErrorEnvelope(t, send(http.MethodPatch, "/api/v1/records/"+id, request), http.StatusConflict, "same_field_conflict")
				conflict := failure["error"].(map[string]any)["conflict"].(map[string]any)
				if conflict["field_key"] != field || conflict["conflict_resolution_class"] != "collection_review" || conflict["base_row_version"] != float64(1) || conflict["current_row_version"] != float64(2) || !reflect.DeepEqual(conflict["base_value"], base) || !reflect.DeepEqual(conflict["server_value"], server) {
					t.Fatal("duplicate preview changed conflict base/server metadata")
				}
				client := conflict["client_value"].(map[string]any)
				items := client["items"].([]any)
				wantCount := 1
				if mode == "base-removed-on-server" {
					wantCount = 0
				}
				if mode == "new" {
					wantCount = 2
				}
				if client["kind"] != "collection_value_v1" || client["ordered"] != false || len(items) != wantCount {
					t.Fatalf("duplicate preview membership=%#v", client)
				}
				refs := []string{}
				originalSeen, newSeen := false, false
				for _, entry := range items {
					item := entry.(map[string]any)
					ref := item["item_ref"].(string)
					refs = append(refs, ref)
					if tag {
						if item["item_kind"] != "tag" {
							t.Fatal("tag preview changed item kind")
						}
						tagID, err := uuid.Parse(item["tag_id"].(string))
						if err != nil || tagID == uuid.Nil || ref != "record_tag:"+id+":"+tagID.String() {
							t.Fatal("tag preview lost canonical item identity")
						}
						switch item["display_text"] {
						case "Original":
							if originalSeen {
								t.Fatal("duplicate Original tag")
							}
							originalSeen = true
							if mode == "replace" {
								if ref == original["item_ref"] {
									t.Fatal("re-add reused removed tag identity")
								}
							} else if !reflect.DeepEqual(item, original) {
								t.Fatal("coalesced add replaced base tag")
							}
						case "Café":
							if mode != "new" || newSeen {
								t.Fatal("unexpected or duplicate new tag")
							}
							newSeen = true
						default:
							t.Fatalf("unexpected preview tag=%#v", item)
						}
					} else {
						if item["item_kind"] != "record_ref" || ref != "record_ref:"+item["linked_record_id"].(string) {
							t.Fatal("Evidence preview changed reference shape")
						}
						switch item["linked_record_id"] {
						case evidence["Original"]:
							if originalSeen {
								t.Fatal("duplicate Original Evidence")
							}
							originalSeen = true
						case evidence["Café"]:
							if mode != "new" || newSeen {
								t.Fatal("unexpected or duplicate new Evidence")
							}
							newSeen = true
						default:
							t.Fatalf("unexpected preview Evidence=%#v", item)
						}
					}
				}
				if originalSeen != (mode != "base-removed-on-server") || newSeen != (mode == "new") || !sort.StringsAreSorted(refs) {
					t.Fatal("preview omitted membership or canonical order")
				}
				repeated := httptestx.RequireErrorEnvelope(t, send(http.MethodPatch, "/api/v1/records/"+id, request), http.StatusConflict, "same_field_conflict")["error"].(map[string]any)["conflict"].(map[string]any)
				if len(conflict) != len(repeated) {
					t.Fatal("repeat preview changed conflict members")
				}
				for key, value := range conflict {
					if key == "conflict_token" {
						first, firstOK := value.(string)
						second, secondOK := repeated[key].(string)
						if !firstOK || !secondOK || first == "" || second == "" {
							t.Fatal("conflict omitted its opaque server token")
						}
					} else if !reflect.DeepEqual(value, repeated[key]) {
						t.Fatalf("repeat preview changed %s", key)
					}
				}
				if after := snapshotTimelineMutationState(t, h.DB, incident, id); after != before {
					t.Fatal("conflict preview persisted source, projection, history, receipt or publication effects")
				}
			})
		}
	}
}
