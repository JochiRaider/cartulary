package entities_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
	"github.com/google/uuid"
)

func exerciseEntityPartyActionOrder(t *testing.T, h *appsupport.ServerHarness, login authcookietest.AuthCookies, incident uuid.UUID, party string) {
	add := map[string]any{"op": "add_party_ref", "party_id": party}
	remove := map[string]any{"op": "remove_party_ref", "item_ref": "party_ref:" + party}
	payload := func(actions ...map[string]any) map[string]any {
		return map[string]any{"kind": "collection_actions_v1", "actions": actions}
	}
	for _, family := range []string{"host", "identity"} {
		view := "cartulary.view.hosts.v1"
		if family == "identity" {
			view = "cartulary.view.identities.v1"
		}
		for _, role := range []string{"subject", "responsible", "observer"} {
			field := family + "." + role + "_party_ids"
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
				t.Run("ordered/"+field+"/"+tc.name, func(t *testing.T) {
					prefix := family + "-" + role + "-" + tc.name
					body := map[string]any{"client_txn_id": prefix + "-create", family + ".display_name": prefix}
					if tc.seeded {
						body[field] = payload(add)
					}
					created := createEntityRow(t, h.Server.HTTP.URL, incident.String(), view, login, body, http.StatusCreated)
					id := created["row"].(map[string]any)["record_id"].(string)
					initialLink := ""
					if tc.seeded {
						if err := h.DB.QueryRowContext(t.Context(), `SELECT record_link_id::text FROM record_links WHERE src_record_id=$1 AND field_key=$2 AND deleted_at IS NULL`, id, field).Scan(&initialLink); err != nil {
							t.Fatal(err)
						}
					}
					patch := map[string]any{"client_txn_id": prefix + "-patch", "view_schema_id": view, "base_row_version": 1, "changes": []map[string]any{{"field_key": field, "action_payload": payload(tc.actions...)}}}
					send := func() *http.Response {
						return httptestx.DoJSON(t, http.DefaultClient, http.MethodPatch, h.Server.HTTP.URL+"/api/v1/records/"+id, patch, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
					}
					before := snapshotEntityPartyOrderState(t, h, incident)
					if tc.invalid {
						failure := httptestx.RequireErrorEnvelope(t, send(), http.StatusBadRequest, "invalid_mutation_payload")
						details := failure["error"].(map[string]any)["details"].(map[string]any)
						if details["field"] != field || details["reason_code"] != "invalid_value" {
							t.Fatalf("invalid action detail=%#v", details)
						}
						if after := snapshotEntityPartyOrderState(t, h, incident); after != before {
							t.Fatal("invalid ordered actions changed source, history, receipt or publication")
						}
						return
					}
					changed := httptestx.RequireSuccessData(t, send(), http.StatusOK)
					row := changed["row"].(map[string]any)
					if row["record_id"] != id || row["row_version"] != float64(2) {
						t.Fatal("ordered operation did not advance the original parent once")
					}
					var want []string
					if tc.active {
						want = []string{party}
					}
					assertEntityPartyItems(t, map[string]string{party: "Alpha"}, row, field, want)
					var raw []byte
					if err := h.DB.QueryRowContext(t.Context(), `SELECT jsonb_agg(jsonb_build_object('sequence',sequence_no,'target',target_id,'op',operation_kind,'before',before_value,'after',after_value) ORDER BY sequence_no) FROM change_set_mutations WHERE change_set_id=$1 AND target_kind='record_link'`, changed["change_set_id"]).Scan(&raw); err != nil {
						t.Fatal(err)
					}
					var mutations []map[string]any
					if err := json.Unmarshal(raw, &mutations); err != nil {
						t.Fatal(err)
					}
					if len(mutations) != len(tc.operations) {
						t.Fatalf("ordered mutation count %d want %d", len(mutations), len(tc.operations))
					}
					lastID := initialLink
					createdIDs := map[string]bool{}
					for i, m := range mutations {
						after := m["after"].(map[string]any)
						target := m["target"].(string)
						if m["sequence"] != float64(i+2) || m["op"] != tc.operations[i] || after["record_link_id"] != target || after["src_record_id"] != id || after["dst_record_id"] != party || after["field_key"] != field {
							t.Fatalf("wrong ordered history at %d", i)
						}
						if tc.operations[i] == "create" {
							if m["before"] != nil || after["deleted_at"] != nil || target == lastID || createdIDs[target] {
								t.Fatal("addition reused a deleted Link identity")
							}
							createdIDs[target] = true
							lastID = target
						} else {
							prior := m["before"].(map[string]any)
							if target != lastID || prior["record_link_id"] != target || prior["deleted_at"] != nil || after["deleted_at"] == nil {
								t.Fatal("removal did not retain the immediately active Link identity")
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
					if active != len(want) || physical != expectedPhysical {
						t.Fatalf("membership/physical links %d/%d want %d/%d", active, physical, len(want), expectedPhysical)
					}
					before = snapshotEntityPartyOrderState(t, h, incident)
					replay := httptestx.RequireSuccessData(t, send(), http.StatusOK)
					if !reflect.DeepEqual(changed, replay) {
						t.Fatal("ordered operation replay changed its receipt")
					}
					if after := snapshotEntityPartyOrderState(t, h, incident); after != before {
						t.Fatal("ordered operation replay wrote effects")
					}
				})
			}
		}
	}
}

type entityPartyOrderState struct {
	source  [32]byte
	effects workbookscenariotest.MutationEffects
}

func snapshotEntityPartyOrderState(t testing.TB, h *appsupport.ServerHarness, incident uuid.UUID) entityPartyOrderState {
	t.Helper()
	var source string
	// Observe physical association identities and both typed sources/projections;
	// the shared effect observer covers envelope, history, receipts and publication.
	if err := h.DB.QueryRowContext(t.Context(), `SELECT jsonb_build_object(
 'hosts',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY record_id),'[]') FROM hosts x WHERE incident_id=$1),
 'identities',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY record_id),'[]') FROM identities x WHERE incident_id=$1),
 'host_projection',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY record_id),'[]') FROM host_grid_projection x WHERE incident_id=$1),
 'identity_projection',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY record_id),'[]') FROM identity_grid_projection x WHERE incident_id=$1),
 'links',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY record_link_id),'[]') FROM record_links x WHERE incident_id=$1))::text`, incident).Scan(&source); err != nil {
		t.Fatal(fmt.Errorf("snapshot entity assignment source: %w", err))
	}
	return entityPartyOrderState{sha256.Sum256([]byte(source)), workbookscenariotest.SnapshotIncidentMutationEffects(t, h.DB, incident.String())}
}
