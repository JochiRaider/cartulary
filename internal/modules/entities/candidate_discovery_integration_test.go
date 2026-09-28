package entities_test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/entities/candidates"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

func TestEntityCandidateDiscovery_Integration(t *testing.T) {
	h := appsupport.StartServer(t, "entity-candidate-discovery")
	login, actor := appsupport.ProvisionBootstrapAdmin(t, h.Server)
	incident := appsupport.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "candidate-incident", "incident_key": "IR-CANDIDATES", "title": "Candidate discovery"})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	for i := 0; i < 105; i++ {
		entitytest.SeedHostRecord(t, h.DB, incidentID, actor, uuid.New(), fmt.Sprintf("A host %03d", i), fmt.Sprintf("host-%03d", i), "", "")
	}
	target := uuid.New()
	entitytest.SeedHostRecord(t, h.DB, incidentID, actor, target, "Zulu Café", "zulu", "zulu.example.test", "device-unique")
	entitytest.SeedEntityAlias(t, h.DB, incidentID, actor, target, "host", "Distinct Alias")
	deleted, merged := uuid.New(), uuid.New()
	entitytest.SeedHostRecord(t, h.DB, incidentID, actor, deleted, "Deleted target", "deleted", "", "")
	entitytest.SeedHostRecord(t, h.DB, incidentID, actor, merged, "Merged target", "merged", "", "")
	transitionEntityClaimLifecycle(t, h.DB, deleted, actor, true)
	setEntityClaimMergeState(t, h.DB, merged, target, actor, true)
	identity := uuid.New()
	entitytest.SeedIdentityRecord(t, h.DB, incidentID, actor, identity, "Identity", "unique@example.test", "mail@example.test", "sam-unique")
	for _, entry := range []struct{ value, classification string }{{"preserved-match", "suggestion_only"}, {"exact-marker", "exact_match_reuse"}, {"excluded-marker", "provenance_only"}} {
		_, err := h.DB.Exec(`INSERT INTO entity_preserved_identifiers(incident_id,record_id,entity_type,identifier_type,raw_value,normalized_value,classification,created_by_user_id) VALUES($1,$2,'host','hostname',$3,$3,$4,$5)`, incidentID, target, entry.value, entry.classification, actor)
		if err != nil {
			t.Fatal(err)
		}
	}
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := h.DB.Exec(query, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`UPDATE identities SET sid='sid-unique', aad_object_id='object-unique' WHERE record_id=$1`, identity)
	entitytest.SeedEntityAlias(t, h.DB, incidentID, actor, target, "host", "retired-alias")
	exec(`UPDATE entity_aliases SET deleted_at=now() WHERE record_id=$1 AND raw_text='retired-alias'`, target)
	exec(`INSERT INTO entity_preserved_identifiers(incident_id,record_id,entity_type,identifier_type,raw_value,normalized_value,classification,created_by_user_id,deleted_at) VALUES($1,$2,'host','hostname','retired-identifier','retired-identifier','suggestion_only',$3,now())`, incidentID, target, actor)
	base := h.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/entity-candidates"
	before := appsupport.QueryCount(t, h.DB, `SELECT count(*) FROM change_sets WHERE incident_id=$1`, incidentID)
	get := func(query string) map[string]any {
		t.Helper()
		resp := appsupport.DoJSON(t, http.MethodGet, base+query, nil, appsupport.WithCookies(login.SessionCookie))
		return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
	}
	for _, search := range []string{"distinct ali", "ZULU café", "zulu example", "preserved mat", "exact mark", "device uni", " café café "} {
		envelope := get("?entity_type=host&search=" + url.QueryEscape(search))
		rows := envelope["data"].(map[string]any)["candidates"].([]any)
		if len(rows) != 1 || rows[0].(map[string]any)["record_id"] != target.String() {
			t.Fatalf("search %q: %#v", search, rows)
		}
	}
	for _, search := range []string{"excluded marker", "cafe", "retired alias", "retired identifier", "deleted target", "merged target"} {
		rows := get("?entity_type=host&search=" + url.QueryEscape(search))["data"].(map[string]any)["candidates"].([]any)
		if len(rows) != 0 {
			t.Fatalf("excluded %q: %#v", search, rows)
		}
	}
	for _, search := range []string{"sam uni", "unique example", "mail example", "sid uni", "object uni"} {
		rows := get("?entity_type=identity&search=" + url.QueryEscape(search))["data"].(map[string]any)["candidates"].([]any)
		if len(rows) != 1 || rows[0].(map[string]any)["record_id"] != identity.String() {
			t.Fatal(rows)
		}
	}
	first := get("?entity_type=host&limit=1")
	token := first["meta"].(map[string]any)["paging"].(map[string]any)["next_cursor"].(string)
	second := get("?entity_type=host&limit=1&cursor_token=" + url.QueryEscape(token))
	if first["data"].(map[string]any)["candidates"].([]any)[0].(map[string]any)["record_id"] == second["data"].(map[string]any)["candidates"].([]any)[0].(map[string]any)["record_id"] {
		t.Fatal("continuation repeated its anchor")
	}
	for _, query := range []string{"?entity_type=identity&limit=1", "?entity_type=host&limit=2", "?entity_type=host&limit=1&search=zulu"} {
		resp := appsupport.DoJSON(t, http.MethodGet, base+query+"&cursor_token="+url.QueryEscape(token), nil, appsupport.WithCookies(login.SessionCookie))
		appsupport.RequireErrorBody(t, resp, http.StatusBadRequest, "invalid_pagination_request")
	}
	for _, query := range []string{"", "?entity_type=host&entity_type=host", "?entity_type=host&search=%01", "?entity_type=host&unknown=1"} {
		resp := appsupport.DoJSON(t, http.MethodGet, base+query, nil, appsupport.WithCookies(login.SessionCookie))
		appsupport.RequireErrorBody(t, resp, http.StatusBadRequest, "invalid_list_query")
	}
	resp := appsupport.DoJSON(t, http.MethodGet, base+"?entity_type=host", nil)
	appsupport.RequireErrorBody(t, resp, http.StatusUnauthorized, "session_required")
	resp = appsupport.DoJSON(t, http.MethodGet, h.Server.HTTP.URL+"/api/v1/incidents/"+uuid.NewString()+"/entity-candidates?entity_type=host", nil, appsupport.WithCookies(login.SessionCookie))
	appsupport.RequireErrorBody(t, resp, http.StatusNotFound, "incident_not_found")
	if after := appsupport.QueryCount(t, h.DB, `SELECT count(*) FROM change_sets WHERE incident_id=$1`, incidentID); after != before {
		t.Fatal("discovery mutated revisions")
	}
	// Cursor identity and authorization are checked live, independently of mutation permission.
	other := authflowtest.SeedLocalUserRecord(t, h.DB, "candidate-reader@example.test", "Candidate reader", "CandidatePass1!", false, false, true)
	otherLogin := loginLocalUser(t, h.Server, other.Email, "CandidatePass1!")
	otherIncident := appsupport.CreateIncidentInStore(t, h.Pool, other, "candidate-other", "IR-CANDIDATE-OTHER", "Other incident")
	entitytest.SeedHostRecord(t, h.DB, otherIncident.ID, other.ID, uuid.New(), "Elsewhere target", "elsewhere", "", "")
	if rows := get("?entity_type=host&search=elsewhere")["data"].(map[string]any)["candidates"].([]any); len(rows) != 0 {
		t.Fatalf("search leaked another incident: %#v", rows)
	}
	resp = appsupport.DoJSON(t, http.MethodGet, h.Server.HTTP.URL+"/api/v1/incidents/"+otherIncident.ID.String()+"/entity-candidates?entity_type=host", nil, appsupport.WithCookies(login.SessionCookie))
	appsupport.RequireErrorBody(t, resp, http.StatusNotFound, "incident_not_found")
	exec(`INSERT INTO incident_memberships(incident_id,user_id,role,added_by_user_id,updated_by_user_id) VALUES($1,$2,'viewer',$3,$3)`, incidentID, other.ID, actor)
	resp = appsupport.DoJSON(t, http.MethodGet, base+"?entity_type=host&limit=1&cursor_token="+url.QueryEscape(token), nil, appsupport.WithCookies(otherLogin.SessionCookie))
	appsupport.RequireErrorBody(t, resp, http.StatusBadRequest, "invalid_pagination_request")
	viewer := appsupport.DoJSON(t, http.MethodGet, base+"?entity_type=host&limit=1", nil, appsupport.WithCookies(otherLogin.SessionCookie))
	viewerToken := httptestx.RequireSuccessEnvelope(t, viewer, http.StatusOK)["meta"].(map[string]any)["paging"].(map[string]any)["next_cursor"].(string)
	resp = appsupport.DoJSON(t, http.MethodGet, h.Server.HTTP.URL+"/api/v1/incidents/"+otherIncident.ID.String()+"/entity-candidates?entity_type=host&limit=1&cursor_token="+url.QueryEscape(viewerToken), nil, appsupport.WithCookies(otherLogin.SessionCookie))
	appsupport.RequireErrorBody(t, resp, http.StatusBadRequest, "invalid_pagination_request")
	exec(`DELETE FROM incident_memberships WHERE incident_id=$1 AND user_id=$2`, incidentID, other.ID)
	resp = appsupport.DoJSON(t, http.MethodGet, base+"?entity_type=host&limit=1&cursor_token="+url.QueryEscape(viewerToken), nil, appsupport.WithCookies(otherLogin.SessionCookie))
	appsupport.RequireErrorBody(t, resp, http.StatusNotFound, "incident_not_found")
	exec(`UPDATE user_sessions SET revoked_at=now(),revoke_reason_code='logout' WHERE user_id=$1`, actor)
	resp = appsupport.DoJSON(t, http.MethodGet, base+"?entity_type=host", nil, appsupport.WithCookies(login.SessionCookie))
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("revoked session status: %d", resp.StatusCode)
	}
	resp.Body.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := candidates.NewReader(nil).Page(ctx, candidates.Request{IncidentID: incidentID, EntityType: "host", Limit: 100}); err != context.Canceled {
		t.Fatalf("cancellation: %v", err)
	}
}
