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

func exerciseTimelineCollectionValidationTransport(t *testing.T) {
	h := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "timeline-collection-validation", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	incident := appsupport.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "validation-incident", "incident_key": "IR-TIMELINE-COLLECTION-VALIDATION", "title": "Collection validation"})["incident_id"].(string)
	foreign := appsupport.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "validation-foreign", "incident_key": "IR-TIMELINE-COLLECTION-FOREIGN", "title": "Foreign collection"})["incident_id"].(string)
	actions := func(entries ...map[string]any) map[string]any {
		return map[string]any{"kind": "collection_actions_v1", "actions": entries}
	}
	create := func(incident, view, txn string, fields map[string]any) map[string]any {
		fields["client_txn_id"] = txn
		response := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, h.Server.HTTP.URL+"/api/v1/incidents/"+incident+"/views/"+view+"/rows", fields, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
		return httptestx.RequireSuccessData(t, response, http.StatusCreated)["row"].(map[string]any)
	}
	current := create(incident, "cartulary.view.timeline.v2", "validation-source", map[string]any{"timeline.tags": actions(map[string]any{"op": "add_tag", "tag_name": "original"})})
	id := current["record_id"].(string)
	other := create(incident, "cartulary.view.timeline.v2", "validation-other", map[string]any{"timeline.tags": actions(map[string]any{"op": "add_tag", "tag_name": "other"})})
	otherItems := other["cells"].(map[string]any)["timeline.tags"].(map[string]any)["value"].(map[string]any)["items"].([]any)
	if len(otherItems) != 1 {
		t.Fatal("foreign tag fixture missing")
	}
	foreignTag := otherItems[0].(map[string]any)["item_ref"].(string)
	evidence := create(incident, "cartulary.view.evidence.v1", "validation-evidence", map[string]any{"evidence.title": "Reference"})["record_id"].(string)
	foreignEvidence := create(foreign, "cartulary.view.evidence.v1", "validation-foreign-evidence", map[string]any{"evidence.title": "Foreign"})["record_id"].(string)
	for _, tc := range []struct {
		name, field string
		payload     map[string]any
	}{
		{"missing tag", "timeline.tags", actions(map[string]any{"op": "remove_tag", "item_ref": "record_tag:" + id + ":" + uuid.NewString()})},
		{"foreign tag", "timeline.tags", actions(map[string]any{"op": "remove_tag", "item_ref": foreignTag})},
		{"late invalid tag", "timeline.tags", actions(map[string]any{"op": "add_tag", "tag_name": "rolled-back"}, map[string]any{"op": "remove_tag", "item_ref": foreignTag})},
		{"missing Evidence", "timeline.attached_evidence_ids", actions(map[string]any{"op": "add_record_ref", "linked_record_id": uuid.NewString()})},
		{"foreign Evidence", "timeline.attached_evidence_ids", actions(map[string]any{"op": "add_record_ref", "linked_record_id": foreignEvidence})},
		{"wrong Evidence type", "timeline.attached_evidence_ids", actions(map[string]any{"op": "add_record_ref", "linked_record_id": other["record_id"]})},
		{"late invalid membership", "timeline.attached_evidence_ids", actions(map[string]any{"op": "add_record_ref", "linked_record_id": evidence}, map[string]any{"op": "remove_record_ref", "item_ref": "record_ref:" + uuid.NewString()})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := snapshotTimelineMutationState(t, h.DB, incident, id)
			body := map[string]any{"view_schema_id": "cartulary.view.timeline.v2", "base_row_version": 1, "client_txn_id": "validation-" + tc.name, "changes": []map[string]any{{"field_key": tc.field, "action_payload": tc.payload}}}
			response := httptestx.DoJSON(t, http.DefaultClient, http.MethodPatch, h.Server.HTTP.URL+"/api/v1/records/"+id, body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
			failure := httptestx.RequireErrorEnvelope(t, response, http.StatusBadRequest, "invalid_mutation_payload")
			details := failure["error"].(map[string]any)["details"].(map[string]any)
			if len(details) != 2 || details["field"] != tc.field || details["reason_code"] != "invalid_value" {
				t.Fatalf("collection error details=%#v", details)
			}
			if after := snapshotTimelineMutationState(t, h.DB, incident, id); after != before {
				t.Fatal("invalid collection action changed source, tags, links, projection, history, receipt or publication")
			}
		})
	}
	send := func(method, path string, body any) *http.Response {
		return httptestx.DoJSON(t, http.DefaultClient, method, h.Server.HTTP.URL+path, body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
	}
	deleted := create(incident, "cartulary.view.evidence.v1", "validation-deleted-evidence", map[string]any{"evidence.title": "Deleted"})["record_id"].(string)
	httptestx.RequireSuccessData(t, send(http.MethodDelete, "/api/v1/records/"+deleted, map[string]any{"client_txn_id": "validation-delete", "base_row_version": 1}), http.StatusOK)
	const field = "timeline.attached_evidence_ids"
	patch := func(txn, target string) map[string]any {
		return map[string]any{"view_schema_id": "cartulary.view.timeline.v2", "base_row_version": 1, "client_txn_id": txn, "changes": []map[string]any{{"field_key": field, "action_payload": actions(map[string]any{"op": "add_record_ref", "linked_record_id": target})}}}
	}
	attached := httptestx.RequireSuccessData(t, send(http.MethodPatch, "/api/v1/records/"+id, patch("validation-attach", evidence)), http.StatusOK)
	if attached["row"].(map[string]any)["row_version"] != float64(2) {
		t.Fatal("stale attachment fixture version differs")
	}
	for _, tc := range []struct{ name, target string }{
		{"missing", uuid.NewString()}, {"foreign", foreignEvidence}, {"wrong type", other["record_id"].(string)}, {"deleted", deleted},
	} {
		t.Run("stale Evidence/"+tc.name, func(t *testing.T) {
			before := snapshotTimelineMutationState(t, h.DB, incident, id)
			request := patch("validation-stale-"+tc.name, tc.target)
			for attempt := 0; attempt < 2; attempt++ {
				failure := httptestx.RequireErrorEnvelope(t, send(http.MethodPatch, "/api/v1/records/"+id, request), http.StatusBadRequest, "invalid_mutation_payload")
				details := failure["error"].(map[string]any)["details"].(map[string]any)
				if len(details) != 2 || details["field"] != field || details["reason_code"] != "invalid_value" {
					t.Fatalf("stale Evidence validation=%#v", details)
				}
				if after := snapshotTimelineMutationState(t, h.DB, incident, id); after != before {
					t.Fatal("ineligible stale Evidence target changed persisted effects")
				}
			}
		})
	}

}
