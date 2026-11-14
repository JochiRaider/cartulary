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

func exerciseTimelineAttachmentNoEffect(t *testing.T) {
	h := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "timeline-attachment-no-effect", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	incident := appsupport.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "attachment-no-effect-incident", "incident_key": "IR-ATTACHMENT-NO-EFFECT", "title": "Attachment no effect"})["incident_id"].(string)
	send := func(method, url string, body any) *http.Response {
		return httptestx.DoJSON(t, http.DefaultClient, method, h.Server.HTTP.URL+url, body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
	}
	evidence := httptestx.RequireSuccessData(t, send(http.MethodPost, "/api/v1/incidents/"+incident+"/views/cartulary.view.evidence.v1/rows", map[string]any{"client_txn_id": "no-effect-evidence", "evidence.title": "Reference"}), http.StatusCreated)["row"].(map[string]any)["record_id"].(string)
	const field = "timeline.attached_evidence_ids"
	attachment := map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{{"op": "add_record_ref", "linked_record_id": evidence}, {"op": "add_record_ref", "linked_record_id": evidence}}}
	for _, state := range []string{"rough", "reviewed"} {
		for _, companion := range []string{"none", "tag", "material"} {
			t.Run(state+"/"+companion, func(t *testing.T) {
				prefix := "attachment-no-effect-" + state + "-" + companion
				created := httptestx.RequireSuccessData(t, send(http.MethodPost, "/api/v1/incidents/"+incident+"/views/cartulary.view.timeline.v2/rows", map[string]any{"client_txn_id": prefix + "-create", field: attachment}), http.StatusCreated)
				id := created["row"].(map[string]any)["record_id"].(string)
				version := 1
				if state == "reviewed" {
					reviewed := httptestx.RequireSuccessData(t, send(http.MethodPost, "/api/v1/records/"+id+"/mark-reviewed", map[string]any{"client_txn_id": prefix + "-review", "base_row_version": 1, "reason": "Reviewed attachment"}), http.StatusOK)
					if reviewed["row_version"] != float64(2) || reviewed["capture_state"] != "reviewed" {
						t.Fatal("review fixture did not enter reviewed state")
					}
					version = 2
				}
				changes := []map[string]any{{"field_key": field, "action_payload": attachment}}
				if companion == "tag" {
					changes = append(changes, map[string]any{"field_key": "timeline.tags", "action_payload": map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{{"op": "add_tag", "tag_name": "nonmaterial"}}}})
				}
				if companion == "material" {
					changes = append(changes, map[string]any{"field_key": "timeline.analyst_text", "value": "Changed analyst"})
				}
				patch := map[string]any{"client_txn_id": prefix + "-patch", "view_schema_id": "cartulary.view.timeline.v2", "base_row_version": version, "changes": changes}
				before := snapshotTimelineMutationState(t, h.DB, incident, id)
				response := send(http.MethodPatch, "/api/v1/records/"+id, patch)
				if companion == "none" {
					for attempt := 0; attempt < 2; attempt++ {
						if attempt > 0 {
							response = send(http.MethodPatch, "/api/v1/records/"+id, patch)
						}
						failure := httptestx.RequireErrorEnvelope(t, response, http.StatusBadRequest, "invalid_mutation_payload")
						details := failure["error"].(map[string]any)["details"].(map[string]any)
						if len(details) != 2 || details["field"] != "changes" || details["reason_code"] != "no_effective_change" {
							t.Fatalf("wrong no-effect rejection=%#v", details)
						}
						if after := snapshotTimelineMutationState(t, h.DB, incident, id); after != before {
							t.Fatal("duplicate-only attachment changed persisted effects")
						}
					}
					return
				}
				changed := httptestx.RequireSuccessData(t, response, http.StatusOK)
				row := changed["row"].(map[string]any)
				cells := row["cells"].(map[string]any)
				expectedState := state
				if companion == "material" {
					expectedState = "enriched"
				}
				if row["record_id"] != id || row["row_version"] != float64(version+1) || cells["timeline.capture_state"].(map[string]any)["value"] != expectedState {
					t.Fatal("duplicate attachment changed companion version or material classification")
				}
				items := cells[field].(map[string]any)["value"].(map[string]any)["items"].([]any)
				if len(items) != 1 || items[0].(map[string]any)["item_ref"] != "record_ref:"+evidence || items[0].(map[string]any)["linked_record_id"] != evidence {
					t.Fatal("duplicate attachment changed membership")
				}
				if companion == "material" && cells["timeline.analyst_text"].(map[string]any)["value"] != "Changed analyst" {
					t.Fatal("material companion was lost")
				}
				if companion == "tag" {
					tags := cells["timeline.tags"].(map[string]any)["value"].(map[string]any)["items"].([]any)
					if len(tags) != 1 || tags[0].(map[string]any)["display_text"] != "nonmaterial" {
						t.Fatal("tag companion was lost")
					}
				}
				var links, physical, tags, total int
				if err := h.DB.QueryRowContext(t.Context(), `SELECT count(*) FILTER(WHERE target_kind='record_link'),count(*) FILTER(WHERE target_kind='record_tag'),count(*) FROM change_set_mutations WHERE change_set_id=$1`, changed["change_set_id"]).Scan(&links, &tags, &total); err != nil {
					t.Fatal(err)
				}
				expectedTags := 0
				if companion == "tag" {
					expectedTags = 1
				}
				if links != 0 || tags != expectedTags || total != 1+expectedTags {
					t.Fatal("duplicate attachment manufactured history")
				}
				if err := h.DB.QueryRowContext(t.Context(), `SELECT count(*) FROM record_links WHERE src_record_id=$1 AND field_key=$2`, id, field).Scan(&physical); err != nil {
					t.Fatal(err)
				}
				if physical != 1 {
					t.Fatal("duplicate attachment manufactured a physical Link")
				}
				before = snapshotTimelineMutationState(t, h.DB, incident, id)
				replay := httptestx.RequireSuccessData(t, send(http.MethodPatch, "/api/v1/records/"+id, patch), http.StatusOK)
				if !reflect.DeepEqual(changed, replay) {
					t.Fatal("companion replay changed receipt")
				}
				if after := snapshotTimelineMutationState(t, h.DB, incident, id); after != before {
					t.Fatal("companion replay wrote effects")
				}
			})
		}
	}
}
