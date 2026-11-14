package timeline_test

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func exerciseTimelineAttachmentActionOrder(t *testing.T) {
	h := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "timeline-attachment-order", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	incident := appsupport.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "attachment-order-incident", "incident_key": "IR-ATTACHMENT-ORDER", "title": "Ordered attachments"})["incident_id"].(string)
	send := func(method, url string, body any) *http.Response {
		return httptestx.DoJSON(t, http.DefaultClient, method, h.Server.HTTP.URL+url, body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
	}
	evidence := httptestx.RequireSuccessData(t, send(http.MethodPost, "/api/v1/incidents/"+incident+"/views/cartulary.view.evidence.v1/rows", map[string]any{"client_txn_id": "ordered-evidence", "evidence.title": "Reference"}), http.StatusCreated)["row"].(map[string]any)["record_id"].(string)
	const field = "timeline.attached_evidence_ids"
	add := map[string]any{"op": "add_record_ref", "linked_record_id": evidence}
	remove := map[string]any{"op": "remove_record_ref", "item_ref": "record_ref:" + evidence}
	payload := func(actions ...map[string]any) map[string]any {
		return map[string]any{"kind": "collection_actions_v1", "actions": actions}
	}
	for _, tc := range []struct {
		name                    string
		seeded, active, invalid bool
		actions                 []map[string]any
		operations              []string
	}{
		{"replace", true, true, false, []map[string]any{remove, add}, []string{"delete", "create"}},
		{"transient", false, false, false, []map[string]any{add, remove}, []string{"create", "delete"}},
		{"duplicate", false, true, false, []map[string]any{add, add}, []string{"create"}},
		{"recreate", false, true, false, []map[string]any{add, remove, add}, []string{"create", "delete", "create"}},
		{"missing-first", false, false, true, []map[string]any{remove, add}, nil},
		{"later-invalid", true, true, true, []map[string]any{remove, add, remove, remove}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := map[string]any{"client_txn_id": "attachment-order-create-" + tc.name}
			if tc.seeded {
				body[field] = payload(add)
			}
			created := httptestx.RequireSuccessData(t, send(http.MethodPost, "/api/v1/incidents/"+incident+"/views/cartulary.view.timeline.v2/rows", body), http.StatusCreated)
			id := created["row"].(map[string]any)["record_id"].(string)
			initialLink := ""
			if tc.seeded {
				if err := h.DB.QueryRowContext(t.Context(), `SELECT record_link_id::text FROM record_links WHERE src_record_id=$1 AND field_key=$2 AND deleted_at IS NULL`, id, field).Scan(&initialLink); err != nil {
					t.Fatal(err)
				}
			}
			patch := map[string]any{"client_txn_id": "attachment-order-patch-" + tc.name, "view_schema_id": "cartulary.view.timeline.v2", "base_row_version": 1, "changes": []map[string]any{{"field_key": field, "action_payload": payload(tc.actions...)}}}
			before := snapshotTimelineMutationState(t, h.DB, incident, id)
			response := send(http.MethodPatch, "/api/v1/records/"+id, patch)
			if tc.invalid {
				failure := httptestx.RequireErrorEnvelope(t, response, http.StatusBadRequest, "invalid_mutation_payload")
				details := failure["error"].(map[string]any)["details"].(map[string]any)
				if len(details) != 2 || details["field"] != field || details["reason_code"] != "invalid_value" {
					t.Fatalf("invalid action details=%#v", details)
				}
				if after := snapshotTimelineMutationState(t, h.DB, incident, id); after != before {
					t.Fatal("rejected ordered attachment actions left partial effects")
				}
				return
			}
			changed := httptestx.RequireSuccessData(t, response, http.StatusOK)
			row := changed["row"].(map[string]any)
			if row["record_id"] != id || row["row_version"] != float64(2) || row["cells"].(map[string]any)["timeline.capture_state"].(map[string]any)["value"] != "enriched" {
				t.Fatal("attachment actions changed identity, capture state or version count")
			}
			collection := row["cells"].(map[string]any)[field].(map[string]any)["value"].(map[string]any)
			items := collection["items"].([]any)
			expectedActive := 0
			if tc.active {
				expectedActive = 1
			}
			if collection["kind"] != "collection_value_v1" || collection["ordered"] != false || len(items) != expectedActive {
				t.Fatal("wrong final attachment membership")
			}
			if tc.active {
				item := items[0].(map[string]any)
				if item["item_ref"] != "record_ref:"+evidence || item["linked_record_id"] != evidence || item["item_kind"] != "record_ref" || item["display_text"] != "Reference" {
					t.Fatalf("wrong attachment item=%#v", item)
				}
			}
			var raw []byte
			if err := h.DB.QueryRowContext(t.Context(), `SELECT jsonb_agg(jsonb_build_object('sequence',sequence_no,'target',target_id,'op',operation_kind,'before',before_value,'after',after_value) ORDER BY sequence_no) FROM change_set_mutations WHERE change_set_id=$1 AND target_kind='record_link'`, changed["change_set_id"]).Scan(&raw); err != nil {
				t.Fatal(err)
			}
			var mutations []map[string]any
			if err := json.Unmarshal(raw, &mutations); err != nil {
				t.Fatal(err)
			}
			if len(mutations) != len(tc.operations) {
				t.Fatalf("mutation count=%d want=%d", len(mutations), len(tc.operations))
			}
			lastID := initialLink
			createdIDs := map[string]bool{}
			for i, m := range mutations {
				after := m["after"].(map[string]any)
				target := m["target"].(string)
				if m["sequence"] != float64(i+2) || m["op"] != tc.operations[i] || after["record_link_id"] != target || after["src_record_id"] != id || after["dst_record_id"] != evidence || after["field_key"] != field {
					t.Fatalf("wrong ordered Link history at %d", i)
				}
				if tc.operations[i] == "create" {
					if m["before"] != nil || after["deleted_at"] != nil || target == lastID || createdIDs[target] {
						t.Fatal("attachment addition reused a deleted Link identity")
					}
					createdIDs[target] = true
					lastID = target
				} else {
					prior := m["before"].(map[string]any)
					if target != lastID || prior["record_link_id"] != target || prior["deleted_at"] != nil || after["deleted_at"] == nil {
						t.Fatal("attachment removal lost its immediately active Link identity")
					}
				}
			}
			var active, physical int
			if err := h.DB.QueryRowContext(t.Context(), `SELECT count(*) FILTER(WHERE deleted_at IS NULL),count(*) FROM record_links WHERE src_record_id=$1 AND field_key=$2`, id, field).Scan(&active, &physical); err != nil {
				t.Fatal(err)
			}
			expectedPhysical := len(createdIDs)
			if tc.seeded {
				expectedPhysical++
			}
			if active != expectedActive || physical != expectedPhysical {
				t.Fatal("unexpected physical attachment Link set")
			}
			before = snapshotTimelineMutationState(t, h.DB, incident, id)
			replay := httptestx.RequireSuccessData(t, send(http.MethodPatch, "/api/v1/records/"+id, patch), http.StatusOK)
			if !reflect.DeepEqual(changed, replay) {
				t.Fatal("ordered attachment replay changed receipt")
			}
			if after := snapshotTimelineMutationState(t, h.DB, incident, id); after != before {
				t.Fatal("ordered attachment replay wrote effects")
			}
		})
	}
}
