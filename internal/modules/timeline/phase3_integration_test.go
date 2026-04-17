package timeline_test

import (
	"context"
	"database/sql"
	"net/http"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"example.com/todo/cartulary/internal/modules/timeline"
	"example.com/todo/cartulary/internal/platform/authn"
	"example.com/todo/cartulary/internal/testutil/fixtures"
	"example.com/todo/cartulary/internal/testutil/httptestx"
	"example.com/todo/cartulary/internal/testutil/pgtest"
	"example.com/todo/cartulary/internal/testutil/s3test"
)

func TestPhase3_TimelineCreatePatchHistoryAndReplay_I_3_01(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase3Server(t, postgresHarness, s3Harness, "phase3-i-3-01")
	defer db.Close()

	adminLogin, _ := provisionBootstrapAdmin(t, server)
	incident := createIncident(t, server, adminLogin, map[string]any{
		"client_txn_id": "txn-i-3-01-incident",
		"incident_key":  "IR-I301",
		"title":         "Timeline substrate",
	})
	incidentID := incident["incident_id"].(string)

	createResp := doPhase3JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
		map[string]any{
			"client_txn_id":    "txn-i-3-01-row-create",
			"timeline.summary": "Initial capture",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	row := createData["row"].(map[string]any)
	recordID := row["record_id"].(string)
	if row["row_version"] != float64(1) {
		t.Fatalf("unexpected create row_version: %#v", row)
	}
	createCells := row["cells"].(map[string]any)
	if got := createCells["timeline.capture_state"].(map[string]any)["value"]; got != "rough" {
		t.Fatalf("unexpected initial capture state: %#v", createCells["timeline.capture_state"])
	}
	if got := createCells["timeline.summary"].(map[string]any)["value"]; got != "Initial capture" {
		t.Fatalf("unexpected initial summary: %#v", createCells["timeline.summary"])
	}

	if got := queryCount(t, db, `SELECT COUNT(*) FROM timeline_events WHERE incident_id::text = $1`, incidentID); got != 1 {
		t.Fatalf("expected one timeline source row, got %d", got)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM timeline_grid_projection WHERE incident_id::text = $1`, incidentID); got != 1 {
		t.Fatalf("expected one timeline projection row, got %d", got)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM change_sets WHERE incident_id::text = $1`, incidentID); got != 1 {
		t.Fatalf("expected one change set, got %d", got)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM change_set_mutations`); got != 1 {
		t.Fatalf("expected one mutation row after create, got %d", got)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM record_revisions WHERE record_id::text = $1`, recordID); got != 1 {
		t.Fatalf("expected one record revision after create, got %d", got)
	}
	if got := len(server.Runtime.WSHub.SnapshotRecordChanges()); got != 1 {
		t.Fatalf("expected one record change emission, got %d", got)
	}

	replayResp := doPhase3JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
		map[string]any{
			"client_txn_id":    "txn-i-3-01-row-create",
			"timeline.summary": "Initial capture",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	replayData := httptestx.RequireSuccessEnvelope(t, replayResp, http.StatusOK)["data"].(map[string]any)
	if replayData["change_set_id"] != createData["change_set_id"] {
		t.Fatalf("expected create replay to return original payload, got %#v", replayData)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM change_sets WHERE incident_id::text = $1`, incidentID); got != 1 {
		t.Fatalf("create replay must not duplicate change_sets, got %d", got)
	}
	if got := len(server.Runtime.WSHub.SnapshotRecordChanges()); got != 1 {
		t.Fatalf("create replay must not emit another record change, got %d", got)
	}

	patchResp := doPhase3JSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/records/"+recordID,
		map[string]any{
			"view_schema_id":   timeline.TimelineViewSchemaID,
			"base_row_version": 1,
			"client_txn_id":    "txn-i-3-01-row-patch",
			"changes": []map[string]any{
				{"field_key": "timeline.summary", "value": "Enriched capture"},
				{"field_key": "timeline.details", "value": "Details from patch"},
			},
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	patchData := httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)["data"].(map[string]any)
	patchedRow := patchData["row"].(map[string]any)
	if patchedRow["row_version"] != float64(2) {
		t.Fatalf("unexpected patch row_version: %#v", patchedRow)
	}
	patchedCells := patchedRow["cells"].(map[string]any)
	if got := patchedCells["timeline.capture_state"].(map[string]any)["value"]; got != "enriched" {
		t.Fatalf("expected rough -> enriched transition, got %#v", patchedCells["timeline.capture_state"])
	}
	if got := patchedCells["timeline.details"].(map[string]any)["value"]; got != "Details from patch" {
		t.Fatalf("unexpected patched details: %#v", patchedCells["timeline.details"])
	}

	if got := queryCount(t, db, `SELECT COUNT(*) FROM change_sets WHERE incident_id::text = $1`, incidentID); got != 2 {
		t.Fatalf("expected second change set after patch, got %d", got)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM change_set_mutations`); got != 2 {
		t.Fatalf("expected two mutation rows after patch, got %d", got)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM record_revisions WHERE record_id::text = $1`, recordID); got != 2 {
		t.Fatalf("expected second record revision after patch, got %d", got)
	}
	if got := len(server.Runtime.WSHub.SnapshotRecordChanges()); got != 2 {
		t.Fatalf("expected one more record change emission after patch, got %d", got)
	}

	patchReplay := doPhase3JSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/records/"+recordID,
		map[string]any{
			"view_schema_id":   timeline.TimelineViewSchemaID,
			"base_row_version": 1,
			"client_txn_id":    "txn-i-3-01-row-patch",
			"changes": []map[string]any{
				{"field_key": "timeline.details", "value": "Details from patch"},
				{"field_key": "timeline.summary", "value": "Enriched capture"},
			},
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	patchReplayData := httptestx.RequireSuccessEnvelope(t, patchReplay, http.StatusOK)["data"].(map[string]any)
	if patchReplayData["change_set_id"] != patchData["change_set_id"] {
		t.Fatalf("expected patch replay to return original payload, got %#v", patchReplayData)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM change_sets WHERE incident_id::text = $1`, incidentID); got != 2 {
		t.Fatalf("patch replay must not duplicate change_sets, got %d", got)
	}
	if got := len(server.Runtime.WSHub.SnapshotRecordChanges()); got != 2 {
		t.Fatalf("patch replay must not emit another record change, got %d", got)
	}

	stalePatch := doPhase3JSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/records/"+recordID,
		map[string]any{
			"view_schema_id":   timeline.TimelineViewSchemaID,
			"base_row_version": 1,
			"client_txn_id":    "txn-i-3-01-stale",
			"changes": []map[string]any{
				{"field_key": "timeline.summary", "value": "stale write"},
			},
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, stalePatch, http.StatusConflict, "row_version_conflict")
}

func TestPhase3_TimelineProjectionBackedQuery_I_3_02(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase3Server(t, postgresHarness, s3Harness, "phase3-i-3-02")
	defer db.Close()

	adminLogin, _ := provisionBootstrapAdmin(t, server)
	incident := createIncident(t, server, adminLogin, map[string]any{
		"client_txn_id": "txn-i-3-02-incident",
		"incident_key":  "IR-I302",
		"title":         "Projection reads",
	})
	incidentID := incident["incident_id"].(string)

	createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
		"client_txn_id":        "txn-i-3-02-row-a",
		"timeline.summary":     "Earlier",
		"timeline.occurred_at": "2026-04-10T10:00:00Z",
	})
	second := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
		"client_txn_id":        "txn-i-3-02-row-b",
		"timeline.summary":     "Later",
		"timeline.occurred_at": "2026-04-10T11:00:00Z",
	})
	secondID := second["row"].(map[string]any)["record_id"].(string)

	patch := doPhase3JSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/records/"+secondID,
		map[string]any{
			"view_schema_id":   timeline.TimelineViewSchemaID,
			"base_row_version": 1,
			"client_txn_id":    "txn-i-3-02-row-b-patch",
			"changes": []map[string]any{
				{"field_key": "timeline.details", "value": "Projected details"},
			},
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, patch, http.StatusOK)

	queryResp := doPhase3JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/query",
		map[string]any{},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
	)
	queryData := httptestx.RequireSuccessEnvelope(t, queryResp, http.StatusOK)["data"].(map[string]any)
	rows := queryData["rows"].([]any)
	if len(rows) != 2 {
		t.Fatalf("expected two projected rows, got %#v", rows)
	}
	firstRow := rows[0].(map[string]any)
	secondRow := rows[1].(map[string]any)
	if firstRow["record_id"] == secondRow["record_id"] {
		t.Fatalf("query rows must preserve stable record identity, got %#v", rows)
	}
	firstSummary := firstRow["cells"].(map[string]any)["timeline.summary"].(map[string]any)["value"]
	secondDetails := secondRow["cells"].(map[string]any)["timeline.details"].(map[string]any)["value"]
	if firstSummary != "Earlier" {
		t.Fatalf("unexpected projection-backed ordering: %#v", rows)
	}
	if secondDetails != "Projected details" {
		t.Fatalf("expected query to read patched projection row, got %#v", secondRow)
	}

	if got := queryCount(t, db, `SELECT COUNT(*) FROM timeline_grid_projection WHERE incident_id::text = $1`, incidentID); got != 2 {
		t.Fatalf("expected two projection rows, got %d", got)
	}
}

func TestPhase3_TimelineLifecycleTransitions_I_3_03(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase3Server(t, postgresHarness, s3Harness, "phase3-i-3-03")
	defer db.Close()

	adminLogin, _ := provisionBootstrapAdmin(t, server)
	incident := createIncident(t, server, adminLogin, map[string]any{
		"client_txn_id": "txn-i-3-03-incident",
		"incident_key":  "IR-I303",
		"title":         "Lifecycle",
	})
	incidentID := incident["incident_id"].(string)

	replacement := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
		"client_txn_id":    "txn-i-3-03-replacement",
		"timeline.summary": "Replacement row",
	})
	replacementID := replacement["row"].(map[string]any)["record_id"].(string)

	created := createTimelineRow(t, server, incidentID, adminLogin, map[string]any{
		"client_txn_id":    "txn-i-3-03-primary",
		"timeline.summary": "Primary row",
	})
	recordID := created["row"].(map[string]any)["record_id"].(string)

	markReviewed := doPhase3JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/records/"+recordID+"/mark-reviewed",
		map[string]any{
			"base_row_version": 1,
			"client_txn_id":    "txn-i-3-03-reviewed-1",
			"reason":           "Initial review",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	reviewedData := httptestx.RequireSuccessEnvelope(t, markReviewed, http.StatusOK)["data"].(map[string]any)
	if reviewedData["capture_state"] != "reviewed" || reviewedData["row_version"] != float64(2) {
		t.Fatalf("unexpected reviewed payload: %#v", reviewedData)
	}

	materialEdit := doPhase3JSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/records/"+recordID,
		map[string]any{
			"view_schema_id":   timeline.TimelineViewSchemaID,
			"base_row_version": 2,
			"client_txn_id":    "txn-i-3-03-demote",
			"changes": []map[string]any{
				{"field_key": "timeline.details", "value": "Material edit after review"},
			},
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	demoted := httptestx.RequireSuccessEnvelope(t, materialEdit, http.StatusOK)["data"].(map[string]any)
	demotedRow := demoted["row"].(map[string]any)
	if got := demotedRow["cells"].(map[string]any)["timeline.capture_state"].(map[string]any)["value"]; got != "enriched" {
		t.Fatalf("expected reviewed row to demote back to enriched, got %#v", demotedRow)
	}

	reviewAgain := doPhase3JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/records/"+recordID+"/mark-reviewed",
		map[string]any{
			"base_row_version": 3,
			"client_txn_id":    "txn-i-3-03-reviewed-2",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	reviewAgainData := httptestx.RequireSuccessEnvelope(t, reviewAgain, http.StatusOK)["data"].(map[string]any)
	if reviewAgainData["capture_state"] != "reviewed" || reviewAgainData["row_version"] != float64(4) {
		t.Fatalf("unexpected second reviewed payload: %#v", reviewAgainData)
	}

	supersede := doPhase3JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
		map[string]any{
			"base_row_version":      4,
			"client_txn_id":         "txn-i-3-03-supersede",
			"reason":                "Superseded by a better row",
			"replacement_record_id": replacementID,
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	superseded := httptestx.RequireSuccessEnvelope(t, supersede, http.StatusOK)["data"].(map[string]any)
	if superseded["capture_state"] != "superseded" || superseded["row_version"] != float64(5) {
		t.Fatalf("unexpected supersede payload: %#v", superseded)
	}
	if superseded["replacement_record_id"] != replacementID {
		t.Fatalf("expected replacement record id to echo, got %#v", superseded)
	}

	if got := queryCount(t, db, `SELECT COUNT(*) FROM record_links WHERE dst_record_id::text = $1 AND src_record_id::text = $2 AND link_type = 'supersedes' AND deleted_at IS NULL`, recordID, replacementID); got != 1 {
		t.Fatalf("expected one active supersedes link, got %d", got)
	}

	illegalReview := doPhase3JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/records/"+recordID+"/mark-reviewed",
		map[string]any{
			"base_row_version": 5,
			"client_txn_id":    "txn-i-3-03-illegal-review",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, illegalReview, http.StatusConflict, "illegal_transition")

	illegalPatch := doPhase3JSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/records/"+recordID,
		map[string]any{
			"view_schema_id":   timeline.TimelineViewSchemaID,
			"base_row_version": 5,
			"client_txn_id":    "txn-i-3-03-illegal-patch",
			"changes": []map[string]any{
				{"field_key": "timeline.summary", "value": "must fail"},
			},
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, illegalPatch, http.StatusConflict, "illegal_transition")
}

type loginResult struct {
	sessionCookie *http.Cookie
	csrfCookie    *http.Cookie
}

func startPhase3Server(t testing.TB, postgresHarness *pgtest.Harness, s3Harness *s3test.Harness, prefix string) (*httptestx.Server, *sql.DB) {
	t.Helper()

	testDB, _, err := postgresHarness.PrepareDatabase(context.Background(), prefix)
	if err != nil {
		t.Fatalf("prepare postgres database: %v", err)
	}
	t.Cleanup(func() {
		if err := postgresHarness.DropDatabase(context.Background(), testDB.Name); err != nil {
			t.Fatalf("drop postgres database: %v", err)
		}
	})

	bucket, err := s3Harness.BootstrapBucket(context.Background(), prefix)
	if err != nil {
		t.Fatalf("bootstrap bucket: %v", err)
	}
	t.Cleanup(func() {
		if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
			t.Logf("cleanup bucket: %v", err)
		}
	})

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")

	server := httptestx.StartServer(t, httptestx.ServerOptions{Env: env})
	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	return server, db
}

func provisionBootstrapAdmin(t testing.TB, server *httptestx.Server) (loginResult, string) {
	t.Helper()

	bootstrapToken := requireBootstrapLogin(t, server, "bootstrap-admin@example.test", "BootstrapPass1!")
	begin := beginTOTPEnrollment(t, server, bootstrapToken, map[string]any{
		"client_txn_id": "txn-bootstrap-admin-begin",
	})
	secretBase32 := begin["totp_setup"].(map[string]any)["secret_base32"].(string)
	completeInitialEnrollment(t, server, bootstrapToken, begin["enrollment_id"].(string), secretBase32, "txn-bootstrap-admin-complete")
	login := loginLocalUserWithSecondFactor(t, server, "bootstrap-admin@example.test", "BootstrapPass1!", generateTOTPCode(t, secretBase32))

	sessionResp := doPhase3JSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, withCookies(login.sessionCookie))
	sessionData := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	return login, sessionData["user_id"].(string)
}

func createIncident(t testing.TB, server *httptestx.Server, admin loginResult, body map[string]any) map[string]any {
	t.Helper()

	resp := doPhase3JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents",
		body,
		withCookies(admin.sessionCookie, admin.csrfCookie),
		withHeader(authn.CSRFHeaderName, admin.csrfCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func createTimelineRow(t testing.TB, server *httptestx.Server, incidentID string, admin loginResult, body map[string]any) map[string]any {
	t.Helper()

	resp := doPhase3JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
		body,
		withCookies(admin.sessionCookie, admin.csrfCookie),
		withHeader(authn.CSRFHeaderName, admin.csrfCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func requireBootstrapLogin(t testing.TB, server *httptestx.Server, username string, password string) string {
	t.Helper()

	resp := doPhase3JSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusUnauthorized, "mfa_setup_required")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	return details["bootstrap_token"].(string)
}

func beginTOTPEnrollment(t testing.TB, server *httptestx.Server, bootstrapToken string, body map[string]any) map[string]any {
	t.Helper()

	resp := doPhase3JSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/begin", body, withHeader("Authorization", "Bearer "+bootstrapToken))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func completeInitialEnrollment(t testing.TB, server *httptestx.Server, bootstrapToken string, enrollmentID string, secretBase32 string, clientTxnID string) {
	t.Helper()

	resp := doPhase3JSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/complete", map[string]any{
		"client_txn_id": clientTxnID,
		"enrollment_id": enrollmentID,
		"code":          generateTOTPCode(t, secretBase32),
	}, withHeader("Authorization", "Bearer "+bootstrapToken))
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
}

func loginLocalUserWithSecondFactor(t testing.TB, server *httptestx.Server, username string, password string, code string) loginResult {
	t.Helper()

	resp := doPhase3JSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
		"second_factor": map[string]any{
			"kind": "totp",
			"assertion": map[string]any{
				"code": code,
			},
		},
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login with second factor failed: status=%d body=%#v", resp.StatusCode, httptestx.ReadJSONBody(t, resp))
	}
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)

	var sessionCookie *http.Cookie
	var csrfCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		switch cookie.Name {
		case authn.SessionCookieName:
			sessionCookie = cookie
		case authn.CSRFCookieName:
			csrfCookie = cookie
		}
	}
	if sessionCookie == nil || csrfCookie == nil {
		t.Fatalf("expected login to set both session and csrf cookies, got %#v", resp.Cookies())
	}
	return loginResult{sessionCookie: sessionCookie, csrfCookie: csrfCookie}
}

func generateTOTPCode(t testing.TB, secretBase32 string) string {
	t.Helper()

	code, err := totp.GenerateCodeCustom(secretBase32, time.Now().UTC(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		t.Fatalf("generate totp code: %v", err)
	}
	return code
}

func doPhase3JSON(t testing.TB, method string, url string, body any, options ...func(*http.Request)) *http.Response {
	t.Helper()

	req := httptestx.NewJSONRequest(t, method, url, body)
	for _, option := range options {
		option(req)
	}
	return httptestx.Do(t, http.DefaultClient, req)
}

func withCookies(cookies ...*http.Cookie) func(*http.Request) {
	return func(req *http.Request) {
		for _, cookie := range cookies {
			if cookie != nil {
				req.AddCookie(cookie)
			}
		}
	}
}

func withHeader(key string, value string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}

func queryCount(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()

	var count int
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return count
}
