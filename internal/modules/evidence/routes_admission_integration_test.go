package evidence_test

// Evidence route admission order (Core 01 REQ-01-707, Core 04 AC-617).
import (
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"

	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// TestEvidenceRouteAdmissionOrder_Integration drives the eight Evidence
// operations through central admission. Authentication and CSRF precede path
// parameters; a malformed identifier or token receives the parameter's
// declared not-found response, identical to a missing resource, including the
// 409 blob_not_visible of Blob verification; an encoded slash stays inside its
// segment; undeclared methods and HEAD are enveloped 405s. Visibility precedes
// the role gate, and both precede body validation. No rejected request commits
// an effect.
func TestEvidenceRouteAdmissionOrder_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "evidence-admission-order", TestRouteMode: httptestx.TestRouteModeDisabled})
	base := harness.Server.HTTP.URL
	admin, adminID := authflowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, base)
	incidentID := appsupport.MustUUID(t, appsupport.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "txn-evidence-admission-order-incident",
		"incident_key":  "IR-EVIDENCE-ADMISSION-ORDER",
		"title":         "Evidence route admission order",
	})["incident_id"].(string))
	viewerPassword := "EvidenceAdmissionViewer1!"
	viewer := authflowtest.SeedLocalUserRecord(t, harness.DB, "evidence-admission-viewer@example.test", "Evidence Admission Viewer", viewerPassword, false, false, true)
	appsupport.SeedMembershipSQL(t, harness.DB, incidentID, viewer.ID, "viewer", adminID)
	viewerLogin := authflowtest.LoginLocalUser(t, http.DefaultClient, base, viewer.Email, viewerPassword, nil)
	outsiderPassword := "EvidenceAdmissionOutsider1!"
	outsider := authflowtest.SeedLocalUserRecord(t, harness.DB, "evidence-admission-outsider@example.test", "Evidence Admission Outsider", outsiderPassword, false, false, true)
	outsiderLogin := authflowtest.LoginLocalUser(t, http.DefaultClient, base, outsider.Email, outsiderPassword, nil)

	// An attached, available record with an issued download handle, and a
	// separate pending Blob slot with its upload capability.
	recordID := uuid.New()
	seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
	attachUploadedBlob(t, harness, admin, incidentID, recordID, []byte("admission body"), "txn-evidence-admission-order-blob", "txn-evidence-admission-order-attach")
	handle := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost,
		base+"/api/v1/evidence-records/"+strings.ToUpper(recordID.String())+"/download-handle", map[string]any{}, authOptions(admin)...), http.StatusOK)["data"].(map[string]any)
	handlePath := handle["href"].(string)
	slot := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, base+"/api/v1/object-blobs", map[string]any{
		"incident_id":       incidentID.String(),
		"client_txn_id":     "txn-evidence-admission-order-slot",
		"byte_size":         5,
		"content_type_hint": "text/plain",
	}, authOptions(admin)...), http.StatusCreated)["data"].(map[string]any)
	blobID := slot["object_blob_id"].(string)
	uploadPath := slot["upload_target"].(map[string]any)["href"].(string)

	// A redirect is never followed, so a 3xx would surface as itself.
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	send := func(t testing.TB, method string, path string, login authcookietest.AuthCookies, body any) *http.Response {
		t.Helper()
		return httptestx.DoJSON(t, client, method, base+path, body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
	}
	requireAlike := func(t testing.TB, existing, missing *http.Response, status int, code string) {
		t.Helper()
		bodies := make([]map[string]any, 0, 2)
		for _, response := range []*http.Response{existing, missing} {
			body := httptestx.RequireErrorEnvelope(t, response, status, code)
			delete(body["error"].(map[string]any), "request_id")
			bodies = append(bodies, body)
		}
		if !reflect.DeepEqual(bodies[0], bodies[1]) {
			t.Fatalf("existing-resource response %v differs from missing-resource response %v", bodies[0], bodies[1])
		}
	}
	requireMethodNotAllowed := func(t testing.TB, response *http.Response, allow string) {
		t.Helper()
		if got := response.Header.Get("Allow"); got != allow {
			t.Fatalf("Allow = %q, want %q", got, allow)
		}
		body := httptestx.RequireErrorEnvelope(t, response, http.StatusMethodNotAllowed, "method_not_allowed")
		if details := httptestx.RequireErrorDetails(t, body); !reflect.DeepEqual(details, map[string]any{"allowed_methods": []any{allow}}) {
			t.Fatalf("details = %#v, want allowed_methods [%s]", details, allow)
		}
	}
	recordPath := "/api/v1/evidence-records/" + recordID.String()
	metadataPath := "/api/v1/incidents/" + incidentID.String() + "/evidence-metadata"
	verifyPath := "/api/v1/object-blobs/" + blobID + "/verify"
	const createPath = "/api/v1/object-blobs"
	invalid := []any{1}
	// Each parameter-addressed operation with its existing and missing
	// parameter values and the declared not-found response.
	type route struct {
		name, method   string
		at             func(value string) string
		value, missing string
		csrf           bool
		status         int
		code           string
	}
	recordRoute := func(name, suffix string) route {
		return route{name: name, method: http.MethodPost, at: func(value string) string { return "/api/v1/evidence-records/" + value + suffix },
			value: recordID.String(), missing: uuid.NewString(), csrf: true, status: http.StatusNotFound, code: "evidence_record_not_found"}
	}
	uuidRoutes := []route{
		recordRoute("attachBlobToEvidenceRecord", "/attach-blob"),
		recordRoute("issueEvidencePreviewHandle", "/preview-handle"),
		recordRoute("issueEvidenceDownloadHandle", "/download-handle"),
		{name: "queryEvidenceMetadata", method: http.MethodPost, at: func(value string) string { return "/api/v1/incidents/" + value + "/evidence-metadata" },
			value: incidentID.String(), missing: uuid.NewString(), csrf: true, status: http.StatusNotFound, code: "incident_not_found"},
		{name: "verifyObjectBlob", method: http.MethodPost, at: func(value string) string { return "/api/v1/object-blobs/" + value + "/verify" },
			value: blobID, missing: uuid.NewString(), csrf: true, status: http.StatusConflict, code: "evidence_attach_rejected"},
	}
	tokenRoutes := []route{
		{name: "redeemEvidenceHandle", method: http.MethodGet, at: func(value string) string { return "/api/v1/evidence-handles/" + value },
			value: strings.TrimPrefix(handlePath, "/api/v1/evidence-handles/"), missing: "hdl_missing", status: http.StatusNotFound, code: "handle_not_found_or_revoked"},
		{name: "uploadObjectBlobContent", method: http.MethodPut, at: func(value string) string { return "/api/v1/object-uploads/" + value },
			value: strings.TrimPrefix(uploadPath, "/api/v1/object-uploads/"), missing: "upl_missing", csrf: true, status: http.StatusNotFound, code: "object_upload_not_found_or_revoked"},
	}
	allRoutes := append(append([]route{}, uuidRoutes...), tokenRoutes...)
	// body is a request body no route accepts, or none for a read.
	body := func(r route) any {
		if r.method == http.MethodGet {
			return nil
		}
		return invalid
	}
	effectTables := []string{"records", "evidence", "object_blobs", "evidence_object_upload_leases", "evidence_access_handles",
		"evidence_custody_events", "route_idempotency", "change_sets", "record_revisions"}
	effects := func(t testing.TB) string {
		t.Helper()
		var jobs int
		if err := harness.DB.QueryRow(`SELECT count(*) FROM jobs WHERE incident_id = $1`, incidentID).Scan(&jobs); err != nil {
			t.Fatalf("count incident jobs: %v", err)
		}
		return structuredTableText(t, harness, effectTables...) + "jobs=" + strconv.Itoa(jobs) +
			" collaboration_intents=" + strconv.Itoa(collaborationsupport.CountIntents(t, harness.DB, collaborationsupport.IntentSelector{}))
	}
	before := effects(t)

	t.Run("authentication and CSRF precede path parameters", func(t *testing.T) {
		for _, r := range allRoutes {
			path := r.at("not-a-valid!id")
			httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, client, r.method, base+path, body(r)), http.StatusUnauthorized, "session_required")
			if r.csrf {
				withoutCSRF := httptestx.DoJSON(t, client, r.method, base+path, body(r), httptestx.WithCookies(admin.Session))
				httptestx.RequireErrorEnvelope(t, withoutCSRF, http.StatusForbidden, "csrf_verification_failed")
			}
		}
		httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, client, http.MethodPost, base+createPath, invalid), http.StatusUnauthorized, "session_required")
		httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, client, http.MethodPost, base+createPath, invalid, httptestx.WithCookies(admin.Session)), http.StatusForbidden, "csrf_verification_failed")
	})
	t.Run("a malformed identifier is a missing resource", func(t *testing.T) {
		for _, malformed := range []string{"not-an-id", strings.ReplaceAll(recordID.String(), "-", ""), "%7B" + recordID.String() + "%7D"} {
			for _, r := range uuidRoutes {
				requireAlike(t, send(t, r.method, r.at(malformed), admin, body(r)), send(t, r.method, r.at(r.missing), admin, body(r)), r.status, r.code)
			}
		}
		for _, malformed := range []string{"bad!token", "bad%20token", strings.Repeat("t", 4097)} {
			for _, r := range tokenRoutes {
				requireAlike(t, send(t, r.method, r.at(malformed), admin, body(r)), send(t, r.method, r.at(r.missing), admin, body(r)), r.status, r.code)
			}
		}
		// Blob verification's not-found is its declared 409 with fixed details.
		response := send(t, http.MethodPost, "/api/v1/object-blobs/not-a-blob-id/verify", admin, invalid)
		if details := httptestx.RequireErrorDetails(t, httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "evidence_attach_rejected")); !reflect.DeepEqual(details, map[string]any{"reason_code": "blob_not_visible"}) {
			t.Fatalf("malformed object_blob_id details = %#v, want reason_code blob_not_visible", details)
		}
	})
	t.Run("an encoded slash stays inside the identifier", func(t *testing.T) {
		for _, r := range allRoutes {
			requireAlike(t, send(t, r.method, r.at(r.value+"%2Fx"), admin, body(r)), send(t, r.method, r.at(r.missing), admin, body(r)), r.status, r.code)
		}
		// No operation is the parent of these encoded paths, so the encoded
		// slash never selects verification or handle issuance.
		for _, path := range []string{"/api/v1/object-blobs/" + blobID + "%2Fverify", recordPath + "%2Fpreview-handle", recordPath + "%2Fdownload-handle"} {
			httptestx.RequireErrorEnvelope(t, send(t, http.MethodPost, path, admin, map[string]any{"client_txn_id": "txn-evidence-admission-order-slash"}), http.StatusNotFound, "route_not_found")
		}
	})
	t.Run("route resolution rejections are enveloped", func(t *testing.T) {
		// HEAD never redeems the single-use download handle.
		head := httptestx.DoJSON(t, client, http.MethodHead, base+handlePath, nil, httptestx.WithCookies(admin.Session))
		headBody, err := io.ReadAll(head.Body)
		_ = head.Body.Close()
		if err != nil || head.StatusCode != http.StatusMethodNotAllowed || head.Header.Get("Allow") != http.MethodGet || len(headBody) != 0 {
			t.Fatalf("HEAD = %d Allow=%q body=%q err=%v, want 405 Allow GET without a body", head.StatusCode, head.Header.Get("Allow"), headBody, err)
		}
		requireMethodNotAllowed(t, send(t, http.MethodPost, handlePath, admin, nil), http.MethodGet)
		requireMethodNotAllowed(t, send(t, http.MethodDelete, uploadPath, admin, nil), http.MethodPut)
		requireMethodNotAllowed(t, send(t, http.MethodGet, createPath, admin, nil), http.MethodPost)
		for _, r := range uuidRoutes {
			requireMethodNotAllowed(t, send(t, http.MethodGet, r.at(r.value), admin, nil), http.MethodPost)
		}
		httptestx.RequireErrorEnvelope(t, send(t, http.MethodGet, recordPath+"/unknown", admin, nil), http.StatusNotFound, "route_not_found")
	})
	t.Run("visibility precedes the role gate and body validation", func(t *testing.T) {
		for _, r := range allRoutes {
			requireAlike(t, send(t, r.method, r.at(r.value), outsiderLogin, body(r)), send(t, r.method, r.at(r.missing), outsiderLogin, body(r)), r.status, r.code)
		}
		// Blob-slot creation decodes only its body-carried incident_id before
		// visibility (Core 04 REQ-04-156, Table 2-B gate 4).
		hidden := map[string]any{"incident_id": incidentID.String(), "byte_size": "invalid"}
		requireAlike(t, send(t, http.MethodPost, createPath, outsiderLogin, hidden),
			send(t, http.MethodPost, createPath, outsiderLogin, map[string]any{"incident_id": uuid.NewString(), "byte_size": "invalid"}), http.StatusNotFound, "incident_not_found")
		// A visible viewer is denied the editor routes before the body is
		// validated; the viewer routes and a permitted editor then meet the
		// body validation.
		for _, check := range []struct {
			method, path, code string
			viewerDenied       bool
		}{
			{http.MethodPost, recordPath + "/attach-blob", "invalid_mutation_payload", true},
			{http.MethodPost, verifyPath, "invalid_mutation_payload", true},
			{http.MethodPost, createPath, "invalid_blob_create_request", true},
			{http.MethodPost, recordPath + "/preview-handle", "invalid_evidence_handle_request", false},
			{http.MethodPost, recordPath + "/download-handle", "invalid_evidence_handle_request", false},
			{http.MethodPost, metadataPath, "invalid_evidence_metadata_request", false},
		} {
			request := any(invalid)
			if check.path == createPath {
				request = hidden
			}
			if check.viewerDenied {
				httptestx.RequireErrorEnvelope(t, send(t, check.method, check.path, viewerLogin, request), http.StatusForbidden, "authorization_denied")
			} else {
				httptestx.RequireErrorEnvelope(t, send(t, check.method, check.path, viewerLogin, request), http.StatusBadRequest, check.code)
			}
			httptestx.RequireErrorEnvelope(t, send(t, check.method, check.path, admin, request), http.StatusBadRequest, check.code)
		}
		httptestx.RequireErrorEnvelope(t, send(t, http.MethodPut, uploadPath, viewerLogin, invalid), http.StatusForbidden, "authorization_denied")
	})
	if after := effects(t); after != before {
		t.Fatalf("rejected Evidence admission requests committed an effect:\nbefore=%s\nafter=%s", before, after)
	}
}
