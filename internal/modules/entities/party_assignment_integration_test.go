package entities_test

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	incidentstoretest "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestEntityPartyAssignmentWrites_Integration(t *testing.T) {
	h := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "entity-party-assignment-writes", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, actor := appsupport.ProvisionBootstrapAdmin(t, h.Server)
	incident := appsupport.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "assignment-incident", "incident_key": "IR-ASSIGN-WRITE", "title": "Party assignment writes"})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	foreign := appsupport.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "assignment-foreign", "incident_key": "IR-ASSIGN-FOREIGN", "title": "Foreign assignment"})
	foreignID := foreign["incident_id"].(string)
	viewer := authflowtest.SeedLocalUserRecord(t, h.DB, "assignment-viewer@example.test", "Assignment viewer", "AssignmentViewer1!", false, false, true)
	incidentstoretest.SeedMembership(t, h.DB, incidentID, viewer.ID, viewer.DisplayName, "viewer", actor)
	viewerLogin := loginLocalUser(t, h.Server, viewer.Email, "AssignmentViewer1!")
	party := func(incident, label string) string {
		result := createEntityRow(t, h.Server.HTTP.URL, incident, "cartulary.view.parties.v1", login, map[string]any{"client_txn_id": "assignment-" + label, "party.display_name": label, "party.party_kind": "organization"}, http.StatusCreated)
		return result["row"].(map[string]any)["record_id"].(string)
	}
	a, b, foreignParty := party(incidentID.String(), "Alpha"), party(incidentID.String(), "Beta"), party(foreignID, "Foreign")
	add := func(ids ...string) map[string]any {
		actions := []map[string]any{}
		for _, id := range ids {
			actions = append(actions, map[string]any{"op": "add_party_ref", "party_id": id})
		}
		return map[string]any{"kind": "collection_actions_v1", "actions": actions}
	}
	for _, family := range []string{"host", "identity"} {
		t.Run(family, func(t *testing.T) {
			view := "cartulary.view.hosts.v1"
			seed := family + ".hostname"
			seedValue := "assignment-host"
			if family == "identity" {
				view = "cartulary.view.identities.v1"
				seed = family + ".email"
				seedValue = "assignment@example.test"
			}
			field := func(kind string) string { return family + "." + kind + "_party_ids" }
			create := map[string]any{"client_txn_id": "assignment-create-" + family, family + ".display_name": "Assigned " + family, seed: seedValue, field("subject"): add(a, a)}
			created := createEntityRow(t, h.Server.HTTP.URL, incidentID.String(), view, login, create, http.StatusCreated)
			row := created["row"].(map[string]any)
			id := row["record_id"].(string)
			assertEntityPartyItems(t, row, field("subject"), []string{a})
			if got := row["row_version"].(float64); got != 1 {
				t.Fatalf("create version %v", got)
			}
			count := func(sql string, args ...any) int { return appsupport.QueryCount(t, h.DB, sql, args...) }
			if got := count(`SELECT count(*) FROM change_set_mutations WHERE change_set_id=$1 AND target_kind='record_link'`, created["change_set_id"]); got != 1 {
				t.Fatalf("create link history count %d", got)
			}
			replay := createEntityRow(t, h.Server.HTTP.URL, incidentID.String(), view, login, create, http.StatusOK)
			if !reflect.DeepEqual(created, replay) {
				t.Fatal("create replay changed receipt")
			}
			create["client_txn_id"] = "assignment-reuse-" + family
			create[field("subject")] = add(a, b)
			reused := createEntityRow(t, h.Server.HTTP.URL, incidentID.String(), view, login, create, http.StatusOK)
			row = reused["row"].(map[string]any)
			if row["record_id"] != id || row["row_version"] != float64(2) {
				t.Fatalf("reuse identity/version: %#v", row)
			}
			assertEntityPartyItems(t, row, field("subject"), []string{a, b})
			patch := func(txn string, version int, changes []map[string]any) map[string]any {
				return map[string]any{"client_txn_id": txn, "view_schema_id": view, "base_row_version": version, "changes": changes}
			}
			change := func(key string, payload any) map[string]any {
				return map[string]any{"field_key": key, "action_payload": payload}
			}
			route := h.Server.HTTP.URL + "/api/v1/records/" + id
			send := func(body any, status int) map[string]any {
				response := httptestx.DoJSON(t, http.DefaultClient, http.MethodPatch, route, body, httptestx.WithCookies(login.SessionCookie, login.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
				if status == http.StatusOK {
					return httptestx.RequireSuccessData(t, response, status)
				}
				code := "invalid_mutation_payload"
				if status == http.StatusConflict {
					code = "row_version_conflict"
				}
				httptestx.RequireErrorEnvelope(t, response, status, code)
				return nil
			}
			body := patch("assignment-patch-"+family, 2, []map[string]any{change(field("responsible"), add(b)), change(field("observer"), add(a))})
			changed := send(body, http.StatusOK)
			row = changed["row"].(map[string]any)
			if row["row_version"] != float64(3) {
				t.Fatalf("multi-field patch version %#v", row["row_version"])
			}
			assertEntityPartyItems(t, row, field("responsible"), []string{b})
			assertEntityPartyItems(t, row, field("observer"), []string{a})
			if count(`SELECT count(*) FROM change_set_mutations WHERE change_set_id=$1 AND target_kind='record_link'`, changed["change_set_id"]) != 2 {
				t.Fatal("missing patch link history")
			}
			if replay := send(body, http.StatusOK); !reflect.DeepEqual(changed, replay) {
				t.Fatal("patch replay changed receipt")
			}
			// Each assignment role independently protects its Party from deletion.
			for _, kind := range []string{"subject", "responsible", "observer"} {
				guardedParty := party(incidentID.String(), family+"-"+kind)
				guarded := createEntityRow(t, h.Server.HTTP.URL, incidentID.String(), view, login, map[string]any{
					"client_txn_id":          "assignment-guard-" + family + "-" + kind,
					family + ".display_name": "Guard " + kind,
					field(kind):              add(guardedParty),
				}, http.StatusCreated)
				deleteParty := func(txn string) *http.Response {
					return httptestx.DoJSON(t, http.DefaultClient, http.MethodDelete, h.Server.HTTP.URL+"/api/v1/records/"+guardedParty, map[string]any{"client_txn_id": txn, "base_row_version": 1}, httptestx.WithCookies(login.SessionCookie, login.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
				}
				blocked := httptestx.RequireErrorEnvelope(t, deleteParty("assignment-guarded-delete-"+family+"-"+kind), http.StatusConflict, "record_delete_blocked")
				if blocked["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "active_incoming_party_reference" {
					t.Fatal("assignment deletion guard lost its reason")
				}
				if count(`SELECT count(*) FROM records WHERE record_id=$1 AND row_version=1 AND deleted_at IS NULL`, guardedParty) != 1 {
					t.Fatal("blocked Party deletion changed its envelope")
				}
				guardID := guarded["row"].(map[string]any)["record_id"].(string)
				detach := patch("assignment-detach-"+family+"-"+kind, 1, []map[string]any{change(field(kind), map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{{"op": "remove_party_ref", "item_ref": "party_ref:" + guardedParty}}})})
				httptestx.RequireSuccessData(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPatch, h.Server.HTTP.URL+"/api/v1/records/"+guardID, detach, httptestx.WithCookies(login.SessionCookie, login.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value)), http.StatusOK)
				httptestx.RequireSuccessData(t, deleteParty("assignment-detached-delete-"+family+"-"+kind), http.StatusOK)
				send(patch("assignment-deleted-"+family+"-"+kind, 3, []map[string]any{change(field(kind), add(guardedParty))}), http.StatusBadRequest)
			}
			beforeLinks := count(`SELECT count(*) FROM record_links WHERE src_record_id=$1`, id)
			beforeChanges := count(`SELECT count(*) FROM change_sets WHERE incident_id=$1`, incidentID)
			// A valid earlier field must roll back when a later membership is invalid.
			for index, bad := range []string{foreignParty, id, uuid.NewString()} {
				send(patch(fmt.Sprintf("assignment-invalid-%s-%d", family, index), 3, []map[string]any{change(field("observer"), add(b)), change(field("subject"), add(bad))}), http.StatusBadRequest)
			}
			send(patch("assignment-stale-"+family, 2, []map[string]any{change(field("observer"), add(b))}), http.StatusConflict)
			denied := httptestx.DoJSON(t, http.DefaultClient, http.MethodPatch, route, patch("assignment-viewer-"+family, 3, []map[string]any{change(field("observer"), add(b))}), httptestx.WithCookies(viewerLogin.SessionCookie, viewerLogin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, viewerLogin.CSRFCookie.Value))
			httptestx.RequireErrorEnvelope(t, denied, http.StatusForbidden, "authorization_denied")
			var version int
			if err := h.DB.QueryRowContext(context.Background(), `SELECT row_version FROM records WHERE record_id=$1`, id).Scan(&version); err != nil {
				t.Fatal(err)
			}
			if version != 3 || count(`SELECT count(*) FROM record_links WHERE src_record_id=$1`, id) != beforeLinks || count(`SELECT count(*) FROM change_sets WHERE incident_id=$1`, incidentID) != beforeChanges {
				t.Fatal("rejected assignment changed source/version/history")
			}
			tx, err := h.Pool.BeginTx(context.Background(), pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(context.Background()) }()
			facts, err := (links.FactReader{}).LoadPartyAssignmentsForRecords(context.Background(), tx, incidentID, []uuid.UUID{appsupport.MustUUID(t, id)})
			if err != nil {
				t.Fatal(err)
			}
			if len(facts) != 4 {
				t.Fatalf("snapshot assignment count %d", len(facts))
			}
			remove := map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{{"op": "remove_party_ref", "item_ref": "party_ref:" + a}}}
			removed := send(patch("assignment-remove-"+family, 3, []map[string]any{change(field("observer"), remove)}), http.StatusOK)
			assertEntityPartyItems(t, removed["row"].(map[string]any), field("observer"), nil)
			retained, err := (links.FactReader{}).LoadPartyAssignmentsForRecords(context.Background(), tx, incidentID, []uuid.UUID{appsupport.MustUUID(t, id)})
			if err != nil || !reflect.DeepEqual(facts, retained) {
				t.Fatal("source snapshot changed with later removal", err)
			}
			// Query hydration must expose the surviving collections, not only write receipts.
			query := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, h.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+view+"/query", map[string]any{}, httptestx.WithCookies(login.SessionCookie, login.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
			data := httptestx.RequireSuccessData(t, query, http.StatusOK)
			rows := data["rows"].([]any)
			var selected map[string]any
			for _, candidate := range rows {
				if candidate.(map[string]any)["record_id"] == id {
					selected = candidate.(map[string]any)
				}
			}
			if selected == nil {
				t.Fatal("assigned entity missing from query")
			}
			assertEntityPartyItems(t, selected, field("subject"), []string{a, b})
			assertEntityPartyItems(t, selected, field("observer"), nil)
		})
	}
}

func assertEntityPartyItems(t *testing.T, row map[string]any, field string, want []string) {
	t.Helper()
	value := row["cells"].(map[string]any)[field].(map[string]any)["value"].(map[string]any)
	if value["kind"] != "collection_value_v1" || value["ordered"] != false {
		t.Fatalf("invalid assignment collection %#v", value)
	}
	items := value["items"].([]any)
	if len(items) != len(want) {
		t.Fatalf("%s count %d want %d", field, len(items), len(want))
	}
	for i, item := range items {
		ref := item.(map[string]any)
		if ref["party_id"] != want[i] || ref["item_ref"] != "party_ref:"+want[i] || ref["item_kind"] != "party_ref" || (ref["display_text"] != "Alpha" && ref["display_text"] != "Beta") {
			t.Fatalf("invalid party item %#v", ref)
		}
	}
}
