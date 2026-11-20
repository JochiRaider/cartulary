package revisions_test

// Record history, delete, restore and rollback route admission order (Core 01
// REQ-01-707, Core 04 AC-617).
import (
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// TestRecordRevisionRouteAdmissionOrder_Integration drives the four revisions
// operations through central admission. Authentication and CSRF precede
// record_id; a malformed record_id receives incident_not_found, identical to
// a missing record; an encoded slash stays inside its segment; undeclared
// methods and HEAD are enveloped 405s. Visibility precedes the role gate, and
// both precede query and body validation. No rejected request commits an
// effect.
func TestRecordRevisionRouteAdmissionOrder_Integration(t *testing.T) {
	harness := appsupport.StartRuntime(t).StartServer(t, appsupport.ServerOptions{Prefix: "revisions-admission-order", TestRouteMode: httptestx.TestRouteModeDisabled})
	base := harness.Server.HTTP.URL
	admin, adminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, base)
	incidentID, recordID := seedRecord(t, harness.DB, harness.Server, admin, adminID, "IR-REVISIONS-ADMISSION-ORDER")
	viewerPassword := "RevisionsAdmissionViewer1!"
	viewer := flowtest.SeedLocalUserRecord(t, harness.DB, "revisions-admission-viewer@example.test", "Revisions Admission Viewer", viewerPassword, false, false, true)
	appsupport.SeedMembershipSQL(t, harness.DB, incidentID, viewer.ID, "viewer", adminID)
	viewerLogin := flowtest.LoginLocalUser(t, http.DefaultClient, base, viewer.Email, viewerPassword, nil)
	outsiderPassword := "RevisionsAdmissionOutsider1!"
	outsider := flowtest.SeedLocalUserRecord(t, harness.DB, "revisions-admission-outsider@example.test", "Revisions Admission Outsider", outsiderPassword, false, false, true)
	outsiderLogin := flowtest.LoginLocalUser(t, http.DefaultClient, base, outsider.Email, outsiderPassword, nil)
	// A redirect is never followed, so a 3xx would surface as itself.
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	send := func(t testing.TB, method string, path string, login authcookietest.AuthCookies, body any) *http.Response {
		t.Helper()
		return httptestx.DoJSON(t, client, method, base+path, body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
	}
	requireAlike := func(t testing.TB, existing, missing *http.Response) {
		t.Helper()
		bodies := make([]map[string]any, 0, 2)
		for _, response := range []*http.Response{existing, missing} {
			body := httptestx.RequireErrorEnvelope(t, response, http.StatusNotFound, "incident_not_found")
			delete(body["error"].(map[string]any), "request_id")
			bodies = append(bodies, body)
		}
		if !reflect.DeepEqual(bodies[0], bodies[1]) {
			t.Fatalf("existing-record response %v differs from missing-record response %v", bodies[0], bodies[1])
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
	recordPath := "/api/v1/records/" + recordID.String()
	missingPath := "/api/v1/records/" + uuid.NewString()
	type route struct {
		name, method, suffix string
		csrf                 bool
	}
	routes := []route{
		{name: "getRecordHistory", method: http.MethodGet, suffix: "/history"},
		{name: "deleteRecord", method: http.MethodDelete, csrf: true},
		{name: "restoreRecord", method: http.MethodPost, suffix: "/restore", csrf: true},
		{name: "rollbackRecord", method: http.MethodPost, suffix: "/rollback", csrf: true},
	}
	// invalid is a request body no route accepts, or none for a read.
	invalid := func(r route) any {
		if r.method == http.MethodGet {
			return nil
		}
		return []any{1}
	}
	httptestx.RequireSuccessEnvelope(t, send(t, http.MethodGet, "/api/v1/records/"+strings.ToUpper(recordID.String())+"/history", viewerLogin, nil), http.StatusOK)
	effects := func(t testing.TB) string {
		t.Helper()
		var state string
		if err := harness.DB.QueryRow(`SELECT jsonb_build_object(
 'record',(SELECT jsonb_build_object('row_version',row_version,'deleted_at',deleted_at,'updated_at',updated_at) FROM records WHERE record_id=$1),
 'change_sets',(SELECT count(*) FROM change_sets WHERE incident_id=$2),
 'revisions',(SELECT count(*) FROM record_revisions WHERE record_id=$1),
 'receipts',(SELECT count(*) FROM route_idempotency)
 )::text`, recordID, incidentID).Scan(&state); err != nil {
			t.Fatalf("read revisions state: %v", err)
		}
		return state + " collaboration_intents=" + strconv.Itoa(collaborationsupport.CountIntents(t, harness.DB, collaborationsupport.IntentSelector{}))
	}
	before := effects(t)

	t.Run("authentication and CSRF precede path parameters", func(t *testing.T) {
		for _, r := range routes {
			path := base + "/api/v1/records/not-a-record-id" + r.suffix
			httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, client, r.method, path, invalid(r)), http.StatusUnauthorized, "session_required")
			if r.csrf {
				httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, client, r.method, path, invalid(r), httptestx.WithCookies(admin.Session)), http.StatusForbidden, "csrf_verification_failed")
			}
		}
	})
	t.Run("a malformed record_id is a missing record", func(t *testing.T) {
		for _, malformed := range []string{"not-an-id", strings.ReplaceAll(recordID.String(), "-", ""), "%7B" + recordID.String() + "%7D"} {
			for _, r := range routes {
				requireAlike(t, send(t, r.method, "/api/v1/records/"+malformed+r.suffix, admin, invalid(r)), send(t, r.method, missingPath+r.suffix, admin, invalid(r)))
			}
		}
	})
	t.Run("an encoded slash stays inside record_id", func(t *testing.T) {
		for _, r := range routes {
			requireAlike(t, send(t, r.method, recordPath+"%2Fx"+r.suffix, admin, invalid(r)), send(t, r.method, missingPath+r.suffix, admin, invalid(r)))
		}
		// The record template declares only DELETE and PATCH, so the encoded
		// slash never selects history, restore or rollback.
		requireMethodNotAllowed(t, send(t, http.MethodGet, recordPath+"%2Fhistory", admin, nil), http.MethodDelete, http.MethodPatch)
		requireMethodNotAllowed(t, send(t, http.MethodPost, recordPath+"%2Frestore", admin, map[string]any{"base_row_version": 1, "client_txn_id": "txn-revisions-admission-slash"}), http.MethodDelete, http.MethodPatch)
		requireMethodNotAllowed(t, send(t, http.MethodPost, recordPath+"%2Frollback", admin, map[string]any{"client_txn_id": "txn-revisions-admission-slash"}), http.MethodDelete, http.MethodPatch)
	})
	t.Run("route resolution rejections are enveloped", func(t *testing.T) {
		head := httptestx.DoJSON(t, client, http.MethodHead, base+recordPath+"/history", nil, httptestx.WithCookies(admin.Session))
		headBody, err := io.ReadAll(head.Body)
		_ = head.Body.Close()
		if err != nil || head.StatusCode != http.StatusMethodNotAllowed || head.Header.Get("Allow") != http.MethodGet || len(headBody) != 0 {
			t.Fatalf("HEAD = %d Allow=%q body=%q err=%v, want 405 Allow GET without a body", head.StatusCode, head.Header.Get("Allow"), headBody, err)
		}
		requireMethodNotAllowed(t, send(t, http.MethodPut, recordPath, admin, map[string]any{}), http.MethodDelete, http.MethodPatch)
		requireMethodNotAllowed(t, send(t, http.MethodPost, recordPath+"/history", admin, map[string]any{}), http.MethodGet)
		requireMethodNotAllowed(t, send(t, http.MethodGet, recordPath+"/restore", admin, nil), http.MethodPost)
		requireMethodNotAllowed(t, send(t, http.MethodGet, recordPath+"/rollback", admin, nil), http.MethodPost)
		httptestx.RequireErrorEnvelope(t, send(t, http.MethodGet, recordPath+"/unknown", admin, nil), http.StatusNotFound, "route_not_found")
	})
	t.Run("visibility precedes the role gate and query and body validation", func(t *testing.T) {
		for _, r := range routes {
			requireAlike(t, send(t, r.method, recordPath+r.suffix+"?limit=0", outsiderLogin, invalid(r)), send(t, r.method, missingPath+r.suffix+"?limit=0", outsiderLogin, invalid(r)))
		}
		// A visible viewer is denied the mutation routes before the body is
		// validated, and meets the history query validation.
		for _, r := range routes[1:] {
			httptestx.RequireErrorEnvelope(t, send(t, r.method, recordPath+r.suffix, viewerLogin, invalid(r)), http.StatusForbidden, "authorization_denied")
		}
		httptestx.RequireErrorEnvelope(t, send(t, http.MethodGet, recordPath+"/history?limit=0", viewerLogin, nil), http.StatusBadRequest, "invalid_pagination_request")
		// A permitted caller then meets the body validation.
		for path, code := range map[string]string{recordPath: "invalid_mutation_payload", recordPath + "/restore": "invalid_mutation_payload", recordPath + "/rollback": "invalid_rollback_request"} {
			method := http.MethodPost
			if path == recordPath {
				method = http.MethodDelete
			}
			httptestx.RequireErrorEnvelope(t, send(t, method, path, admin, []any{1}), http.StatusBadRequest, code)
		}
	})
	if after := effects(t); after != before {
		t.Fatalf("rejected revisions admission requests committed an effect:\nbefore=%s\nafter=%s", before, after)
	}
}
