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
	"github.com/google/uuid"
)

func exerciseTimelineTagActionOrder(t *testing.T) {
	h := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "timeline-tag-order", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	incident := appsupport.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "tag-order-incident", "incident_key": "IR-TAG-ORDER", "title": "Ordered tags"})["incident_id"].(string)
	send := func(method, url string, body any) *http.Response {
		return httptestx.DoJSON(t, http.DefaultClient, method, h.Server.HTTP.URL+url, body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
	}
	add := map[string]any{"op": "add_tag", "tag_name": "original"}
	duplicate := map[string]any{"op": "add_tag", "tag_name": " ORIGINAL "}
	for _, tc := range []struct {
		name   string
		steps  []string
		ops    []string
		active int
		reason string
	}{
		{"remove-add", []string{"remove", "add"}, []string{"delete", "create"}, 1, ""},
		{"add-remove", []string{"add", "remove"}, []string{"delete"}, 0, ""},
		{"remove-add-duplicate", []string{"remove", "add", "duplicate"}, []string{"delete", "create"}, 1, ""},
		{"late-invalid", []string{"remove", "add", "remove"}, nil, 1, "invalid_value"},
		{"missing-first", []string{"missing", "add"}, nil, 1, "invalid_value"},
		{"duplicate-only", []string{"add", "duplicate"}, nil, 1, "no_effective_change"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			created := httptestx.RequireSuccessData(t, send(http.MethodPost, "/api/v1/incidents/"+incident+"/views/cartulary.view.timeline.v2/rows", map[string]any{"client_txn_id": "tag-order-create-" + tc.name, "timeline.tags": map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{add}}}), http.StatusCreated)
			row := created["row"].(map[string]any)
			id := row["record_id"].(string)
			collection := row["cells"].(map[string]any)["timeline.tags"].(map[string]any)["value"].(map[string]any)
			items := collection["items"].([]any)
			if len(items) != 1 {
				t.Fatal("initial tag missing")
			}
			original := items[0].(map[string]any)
			originalID := original["tag_id"].(string)
			originalRef := original["item_ref"].(string)
			if originalRef != "record_tag:"+id+":"+originalID || original["item_kind"] != "tag" || original["display_text"] != "original" {
				t.Fatal("invalid initial tag fixture")
			}
			actions := []map[string]any{}
			for _, step := range tc.steps {
				switch step {
				case "add":
					actions = append(actions, add)
				case "duplicate":
					actions = append(actions, duplicate)
				case "remove":
					actions = append(actions, map[string]any{"op": "remove_tag", "item_ref": originalRef})
				case "missing":
					actions = append(actions, map[string]any{"op": "remove_tag", "item_ref": "record_tag:" + id + ":" + uuid.NewString()})
				}
			}
			patch := map[string]any{"client_txn_id": "tag-order-patch-" + tc.name, "view_schema_id": "cartulary.view.timeline.v2", "base_row_version": 1, "changes": []map[string]any{{"field_key": "timeline.tags", "action_payload": map[string]any{"kind": "collection_actions_v1", "actions": actions}}}}
			before := snapshotTimelineMutationState(t, h.DB, incident, id)
			response := send(http.MethodPatch, "/api/v1/records/"+id, patch)
			if tc.reason != "" {
				failure := httptestx.RequireErrorEnvelope(t, response, http.StatusBadRequest, "invalid_mutation_payload")
				if failure["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != tc.reason {
					t.Fatal("ordered tag rejection reason changed")
				}
				if after := snapshotTimelineMutationState(t, h.DB, incident, id); after != before {
					t.Fatal("rejected ordered tags left partial effects")
				}
				return
			}
			changed := httptestx.RequireSuccessData(t, response, http.StatusOK)
			row = changed["row"].(map[string]any)
			if row["record_id"] != id || row["row_version"] != float64(2) || row["cells"].(map[string]any)["timeline.capture_state"].(map[string]any)["value"] != "rough" {
				t.Fatal("tag operation changed identity, capture state or version count")
			}
			collection = row["cells"].(map[string]any)["timeline.tags"].(map[string]any)["value"].(map[string]any)
			items = collection["items"].([]any)
			if collection["kind"] != "collection_value_v1" || collection["ordered"] != false || len(items) != tc.active {
				t.Fatal("wrong final tag membership")
			}
			var raw []byte
			if err := h.DB.QueryRowContext(t.Context(), `SELECT jsonb_agg(jsonb_build_object('sequence',sequence_no,'target',target_id,'op',operation_kind,'before',before_value,'after',after_value) ORDER BY sequence_no) FROM change_set_mutations WHERE change_set_id=$1 AND target_kind='record_tag'`, changed["change_set_id"]).Scan(&raw); err != nil {
				t.Fatal(err)
			}
			var mutations []map[string]any
			if err := json.Unmarshal(raw, &mutations); err != nil {
				t.Fatal(err)
			}
			if len(mutations) != len(tc.ops) {
				t.Fatal("tag mutation history count differs")
			}
			for i, m := range mutations {
				after := m["after"].(map[string]any)
				if m["sequence"] != float64(i+2) || m["op"] != tc.ops[i] || after["record_id"] != id || after["tag_name"] != "original" {
					t.Fatal("tag history lost request order or source values")
				}
				if tc.ops[i] == "delete" {
					prior := m["before"].(map[string]any)
					if m["target"] != originalRef || prior["record_tag_id"] != originalID || prior["deleted_at"] != nil || after["record_tag_id"] != originalID || after["deleted_at"] == nil {
						t.Fatal("removal lost original physical tag identity")
					}
				} else {
					item := items[0].(map[string]any)
					replacement := after["record_tag_id"].(string)
					if replacement == originalID || m["before"] != nil || after["deleted_at"] != nil || m["target"] != "record_tag:"+id+":"+replacement || item["tag_id"] != replacement || item["item_ref"] != m["target"] || item["item_kind"] != "tag" || item["display_text"] != "original" {
						t.Fatal("replacement did not create and expose its own active tag identity")
					}
				}
			}
			var physical, active int
			if err := h.DB.QueryRowContext(t.Context(), `SELECT count(*),count(*) FILTER(WHERE deleted_at IS NULL) FROM record_tags WHERE record_id=$1`, id).Scan(&physical, &active); err != nil {
				t.Fatal(err)
			}
			if physical != 1+tc.active || active != tc.active {
				t.Fatal("unexpected physical tag set")
			}
			before = snapshotTimelineMutationState(t, h.DB, incident, id)
			if replay := httptestx.RequireSuccessData(t, send(http.MethodPatch, "/api/v1/records/"+id, patch), http.StatusOK); !reflect.DeepEqual(replay, changed) {
				t.Fatal("ordered tag replay changed receipt")
			}
			if after := snapshotTimelineMutationState(t, h.DB, incident, id); after != before {
				t.Fatal("ordered tag replay wrote effects")
			}
		})
	}
}
