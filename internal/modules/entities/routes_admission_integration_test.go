package entities_test

// Entity candidate, merge and mention route admission order (Core 01
// REQ-01-707, Core 04 AC-617).
import (
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"

	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// TestEntityRouteAdmissionOrder_Integration drives the three entity
// operations through central admission. Authentication and CSRF precede path
// parameters; a malformed identifier receives the parameter's declared
// not-found response, identical to a missing resource; an encoded slash stays
// inside its segment; undeclared methods and HEAD are enveloped 405s.
// Visibility precedes the role gate, and both precede query and body
// validation. No rejected request commits an effect.
func TestEntityRouteAdmissionOrder_Integration(t *testing.T) {
	harness := appsupport.StartRuntime(t).StartServer(t, appsupport.ServerOptions{Prefix: "entities-admission-order", TestRouteMode: httptestx.TestRouteModeDisabled})
	base := harness.Server.HTTP.URL
	admin, adminID := authflowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, base)
	incidentID := appsupport.MustUUID(t, appsupport.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "txn-entities-admission-order-incident",
		"incident_key":  "IR-ENTITIES-ADMISSION-ORDER",
		"title":         "Entity route admission order",
	})["incident_id"].(string))
	entitytest.SeedHostRecord(t, harness.DB, incidentID, adminID, entitytest.CanonicalHostRecordID, "Survivor", "SURVIVOR", "", "")
	entitytest.SeedHostRecord(t, harness.DB, incidentID, adminID, entitytest.DuplicateHostRecordID, "Loser", "LOSER", "", "")
	timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, adminID, timelinetest.RecordID)
	entitytest.SeedMention(t, harness.DB, adminID, entitytest.HostMentionID, timelinetest.RecordID, timelinetest.FieldHostRefs, "host", "SURVIVOR", "unresolved", nil, nil)
	viewerPassword := "EntitiesAdmissionViewer1!"
	viewer := authflowtest.SeedLocalUserRecord(t, harness.DB, "entities-admission-viewer@example.test", "Entities Admission Viewer", viewerPassword, false, false, true)
	appsupport.SeedMembershipSQL(t, harness.DB, incidentID, viewer.ID, "viewer", adminID)
	viewerLogin := authflowtest.LoginLocalUser(t, http.DefaultClient, base, viewer.Email, viewerPassword, nil)
	outsiderPassword := "EntitiesAdmissionOutsider1!"
	outsider := authflowtest.SeedLocalUserRecord(t, harness.DB, "entities-admission-outsider@example.test", "Entities Admission Outsider", outsiderPassword, false, false, true)
	outsiderLogin := authflowtest.LoginLocalUser(t, http.DefaultClient, base, outsider.Email, outsiderPassword, nil)
	// A redirect is never followed, so a 3xx would surface as itself.
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	send := func(t testing.TB, method string, path string, login authcookietest.AuthCookies, body any) *http.Response {
		t.Helper()
		return httptestx.DoJSON(t, client, method, base+path, body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
	}
	requireAlike := func(t testing.TB, existing, missing *http.Response, code string) {
		t.Helper()
		bodies := make([]map[string]any, 0, 2)
		for _, response := range []*http.Response{existing, missing} {
			body := httptestx.RequireErrorEnvelope(t, response, http.StatusNotFound, code)
			delete(body["error"].(map[string]any), "request_id")
			bodies = append(bodies, body)
		}
		if !reflect.DeepEqual(bodies[0], bodies[1]) {
			t.Fatalf("existing-resource response %v differs from missing-resource response %v", bodies[0], bodies[1])
		}
	}
	requireMethodNotAllowed := func(t testing.TB, response *http.Response, allow ...string) {
		t.Helper()
		if got := response.Header.Get("Allow"); got != strings.Join(allow, ", ") {
			t.Fatalf("Allow = %q, want %q", got, strings.Join(allow, ", "))
		}
		body := httptestx.RequireErrorEnvelope(t, response, http.StatusMethodNotAllowed, "method_not_allowed")
		want := make([]any, 0, len(allow))
		for _, method := range allow {
			want = append(want, method)
		}
		if details := httptestx.RequireErrorDetails(t, body); !reflect.DeepEqual(details, map[string]any{"allowed_methods": want}) {
			t.Fatalf("details = %#v, want allowed_methods %v", details, allow)
		}
	}
	type route struct {
		name, method string
		at           func(value string) string
		value, code  string
		csrf         bool
		query        string
	}
	routes := []route{
		{name: "listEntityCandidates", method: http.MethodGet, at: func(value string) string { return "/api/v1/incidents/" + value + "/entity-candidates" },
			value: incidentID.String(), code: "incident_not_found", query: "?entity_type=host"},
		{name: "mergeEntityRecord", method: http.MethodPost, at: func(value string) string { return "/api/v1/records/" + value + "/merge" },
			value: entitytest.CanonicalHostRecordID.String(), code: "incident_not_found", csrf: true},
		{name: "resolveEntityMention", method: http.MethodPost, at: func(value string) string { return "/api/v1/entity-mentions/" + value + "/resolve" },
			value: entitytest.HostMentionID.String(), code: "entity_mention_not_found", csrf: true},
	}
	// invalid is a request body no route accepts, or none for a read.
	invalid := func(r route) any {
		if r.method == http.MethodGet {
			return nil
		}
		return []any{1}
	}
	httptestx.RequireSuccessEnvelope(t, send(t, http.MethodGet, routes[0].at(strings.ToUpper(incidentID.String()))+"?entity_type=host", viewerLogin, nil), http.StatusOK)
	effects := func(t testing.TB) string {
		t.Helper()
		var state string
		if err := harness.DB.QueryRow(`SELECT jsonb_build_object(
 'records',(SELECT coalesce(jsonb_agg(jsonb_build_object('record_id',record_id,'row_version',row_version,'deleted_at',deleted_at) ORDER BY record_id),'[]') FROM records WHERE incident_id=$1),
 'mentions',(SELECT coalesce(jsonb_agg(jsonb_build_object('id',entity_mention_id,'status',resolution_status,'row_version',row_version) ORDER BY entity_mention_id),'[]') FROM entity_mentions),
 'change_sets',(SELECT count(*) FROM change_sets WHERE incident_id=$1),
 'receipts',(SELECT count(*) FROM route_idempotency)
 )::text`, incidentID).Scan(&state); err != nil {
			t.Fatalf("read entity state: %v", err)
		}
		return state + " collaboration_intents=" + strconv.Itoa(collaborationsupport.CountIntents(t, harness.DB, collaborationsupport.IntentSelector{}))
	}
	before := effects(t)

	t.Run("authentication and CSRF precede path parameters", func(t *testing.T) {
		for _, r := range routes {
			path := base + r.at("not-a-valid-id") + r.query
			httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, client, r.method, path, invalid(r)), http.StatusUnauthorized, "session_required")
			if r.csrf {
				httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, client, r.method, path, invalid(r), httptestx.WithCookies(admin.Session)), http.StatusForbidden, "csrf_verification_failed")
			}
		}
	})
	t.Run("a malformed identifier is a missing resource", func(t *testing.T) {
		for _, malformed := range []string{"not-an-id", strings.ReplaceAll(incidentID.String(), "-", ""), "%7B" + incidentID.String() + "%7D"} {
			for _, r := range routes {
				requireAlike(t, send(t, r.method, r.at(malformed)+r.query, admin, invalid(r)), send(t, r.method, r.at(uuid.NewString())+r.query, admin, invalid(r)), r.code)
			}
		}
	})
	t.Run("an encoded slash stays inside the identifier", func(t *testing.T) {
		for _, r := range routes {
			requireAlike(t, send(t, r.method, r.at(r.value+"%2Fx")+r.query, admin, invalid(r)), send(t, r.method, r.at(uuid.NewString())+r.query, admin, invalid(r)), r.code)
		}
		// The record template declares only DELETE and PATCH, so the encoded
		// slash never selects merge.
		requireMethodNotAllowed(t, send(t, http.MethodPost, "/api/v1/records/"+entitytest.CanonicalHostRecordID.String()+"%2Fmerge", admin, map[string]any{}), http.MethodDelete, http.MethodPatch)
		httptestx.RequireErrorEnvelope(t, send(t, http.MethodPost, "/api/v1/entity-mentions/"+entitytest.HostMentionID.String()+"%2Fresolve", admin, map[string]any{}), http.StatusNotFound, "route_not_found")
	})
	t.Run("route resolution rejections are enveloped", func(t *testing.T) {
		head := httptestx.DoJSON(t, client, http.MethodHead, base+routes[0].at(incidentID.String())+"?entity_type=host", nil, httptestx.WithCookies(admin.Session))
		headBody, err := io.ReadAll(head.Body)
		_ = head.Body.Close()
		if err != nil || head.StatusCode != http.StatusMethodNotAllowed || head.Header.Get("Allow") != http.MethodGet || len(headBody) != 0 {
			t.Fatalf("HEAD = %d Allow=%q body=%q err=%v, want 405 Allow GET without a body", head.StatusCode, head.Header.Get("Allow"), headBody, err)
		}
		requireMethodNotAllowed(t, send(t, http.MethodPost, routes[0].at(incidentID.String()), admin, map[string]any{}), http.MethodGet)
		requireMethodNotAllowed(t, send(t, http.MethodGet, routes[1].at(routes[1].value), admin, nil), http.MethodPost)
		requireMethodNotAllowed(t, send(t, http.MethodGet, routes[2].at(routes[2].value), admin, nil), http.MethodPost)
		httptestx.RequireErrorEnvelope(t, send(t, http.MethodGet, "/api/v1/entity-mentions/"+routes[2].value, admin, nil), http.StatusNotFound, "route_not_found")
	})
	t.Run("visibility precedes the role gate and query and body validation", func(t *testing.T) {
		for _, r := range routes {
			requireAlike(t, send(t, r.method, r.at(r.value)+"?entity_type=bogus", outsiderLogin, invalid(r)), send(t, r.method, r.at(uuid.NewString())+"?entity_type=bogus", outsiderLogin, invalid(r)), r.code)
		}
		// A visible viewer is denied merge and mention resolution before the
		// body is validated, and meets the candidate query validation.
		for _, r := range routes[1:] {
			httptestx.RequireErrorEnvelope(t, send(t, r.method, r.at(r.value), viewerLogin, invalid(r)), http.StatusForbidden, "authorization_denied")
		}
		httptestx.RequireErrorEnvelope(t, send(t, http.MethodGet, routes[0].at(incidentID.String())+"?entity_type=bogus", viewerLogin, nil), http.StatusBadRequest, "invalid_list_query")
		// A permitted caller then meets the body validation.
		for _, r := range routes[1:] {
			httptestx.RequireErrorEnvelope(t, send(t, r.method, r.at(r.value), admin, invalid(r)), http.StatusBadRequest, "invalid_mutation_payload")
		}
	})
	if after := effects(t); after != before {
		t.Fatalf("rejected entity admission requests committed an effect:\nbefore=%s\nafter=%s", before, after)
	}
}
