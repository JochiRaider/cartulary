package incidents_test

import (
	"context"
	"database/sql"
	"net/http"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"example.com/todo/cartulary/internal/platform/authn"
	"example.com/todo/cartulary/internal/testutil/fixtures"
	"example.com/todo/cartulary/internal/testutil/httptestx"
	"example.com/todo/cartulary/internal/testutil/pgtest"
	"example.com/todo/cartulary/internal/testutil/s3test"
)

func TestPhase2_IncidentCreateBootstrap_I_2_01(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase2Server(t, postgresHarness, s3Harness, "phase2-i-2-01")
	defer db.Close()

	adminLogin, adminID := provisionBootstrapAdmin(t, server)
	createResp := doPhase2JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-i-2-01-create",
			"incident_key":  "  IR-I201  ",
			"title":         "  Integration Incident  ",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	body := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)
	data := body["data"].(map[string]any)
	incidentID := data["incident_id"].(string)
	if got := createResp.Header.Get("Location"); got != "/api/v1/incidents/"+incidentID {
		t.Fatalf("unexpected incident Location header: got %q", got)
	}
	if data["status"] != "active" || data["incident_version"] != float64(1) {
		t.Fatalf("unexpected incident create payload: %#v", data)
	}

	if got := queryCount(t, db, `SELECT COUNT(*) FROM incidents WHERE id::text = $1`, incidentID); got != 1 {
		t.Fatalf("expected one incident row, got %d", got)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM incident_memberships WHERE incident_id::text = $1 AND user_id::text = $2 AND role = 'admin'`, incidentID, adminID); got != 1 {
		t.Fatalf("expected one bootstrap admin membership, got %d", got)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM incident_workbook_preferences WHERE incident_id::text = $1`, incidentID); got != 1 {
		t.Fatalf("expected one incident workbook preferences row, got %d", got)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM user_workbook_preferences WHERE incident_id::text = $1 AND user_id::text = $2`, incidentID, adminID); got != 1 {
		t.Fatalf("expected one user workbook preferences row, got %d", got)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events WHERE incident_id::text = $1`, incidentID); got != 2 {
		t.Fatalf("expected two incident audit events, got %d", got)
	}

	sessionResp := doPhase2JSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, withCookies(adminLogin.sessionCookie))
	sessionBody := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	memberships := sessionBody["memberships"].([]any)
	if len(memberships) != 1 {
		t.Fatalf("expected one session membership summary, got %#v", memberships)
	}
	summary := memberships[0].(map[string]any)
	if summary["incident_id"] != incidentID || summary["role"] != "admin" {
		t.Fatalf("unexpected session membership summary: %#v", summary)
	}

	defaultPrefsResp := doPhase2JSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/default", nil, withCookies(adminLogin.sessionCookie))
	defaultPrefs := httptestx.RequireSuccessEnvelope(t, defaultPrefsResp, http.StatusOK)["data"].(map[string]any)
	if defaultPrefs["default_sheet_ref"] != nil {
		t.Fatalf("expected default_sheet_ref bootstrap null, got %#v", defaultPrefs)
	}

	mePrefsResp := doPhase2JSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/me", nil, withCookies(adminLogin.sessionCookie))
	mePrefs := httptestx.RequireSuccessEnvelope(t, mePrefsResp, http.StatusOK)["data"].(map[string]any)
	if mePrefs["home_sheet_ref"] != nil {
		t.Fatalf("expected home_sheet_ref bootstrap null, got %#v", mePrefs)
	}
}

func TestPhase2_IncidentReplayAndConflict_I_2_02(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase2Server(t, postgresHarness, s3Harness, "phase2-i-2-02")
	defer db.Close()

	adminLogin, _ := provisionBootstrapAdmin(t, server)
	create := doPhase2JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-i-2-02-create",
			"incident_key":  " IR-I202 ",
			"title":         "Replay Incident",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	first := httptestx.RequireSuccessEnvelope(t, create, http.StatusCreated)["data"].(map[string]any)
	incidentID := first["incident_id"]

	replay := doPhase2JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-i-2-02-create",
			"incident_key":  "IR-I202",
			"title":         "Replay Incident",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	replayBody := httptestx.RequireSuccessEnvelope(t, replay, http.StatusOK)["data"].(map[string]any)
	if replayBody["incident_id"] != incidentID {
		t.Fatalf("expected idempotent replay to return original incident, got %#v", replayBody)
	}

	divergent := doPhase2JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-i-2-02-create",
			"incident_key":  "IR-I202",
			"title":         "Different title",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, divergent, http.StatusConflict, "client_txn_conflict")

	duplicateKey := doPhase2JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-i-2-02-duplicate",
			"incident_key":  "  IR-I202  ",
			"title":         "Duplicate key",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, duplicateKey, http.StatusConflict, "incident_key_conflict")

	if got := queryCount(t, db, `SELECT COUNT(*) FROM incidents WHERE incident_key_canonical = $1`, "IR-I202"); got != 1 {
		t.Fatalf("expected exactly one stored canonical incident key, got %d", got)
	}
}

func TestPhase2_IncidentPatchVersioning_I_2_03(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase2Server(t, postgresHarness, s3Harness, "phase2-i-2-03")
	defer db.Close()

	adminLogin, _ := provisionBootstrapAdmin(t, server)
	incident := createIncident(t, server, adminLogin, map[string]any{
		"client_txn_id": "txn-i-2-03-create",
		"incident_key":  "IR-I203",
		"title":         "Patchable Incident",
	})
	incidentID := incident["incident_id"].(string)
	initialUpdatedAt := incident["updated_at"]

	noOpPatch := doPhase2JSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID,
		map[string]any{
			"base_incident_version": 1,
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	noOpBody := httptestx.RequireSuccessEnvelope(t, noOpPatch, http.StatusOK)["data"].(map[string]any)
	if noOpBody["incident_version"] != float64(1) || noOpBody["updated_at"] != initialUpdatedAt {
		t.Fatalf("unexpected no-op patch payload: %#v", noOpBody)
	}

	time.Sleep(20 * time.Millisecond)
	patch := doPhase2JSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID,
		map[string]any{
			"base_incident_version":     1,
			"tlp":                       "amber",
			"current_phase":             "containment",
			"primary_external_case_ref": "CASE-I203",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	patchBody := httptestx.RequireSuccessEnvelope(t, patch, http.StatusOK)["data"].(map[string]any)
	if patchBody["incident_version"] != float64(2) || patchBody["tlp"] != "amber" || patchBody["current_phase"] != "containment" {
		t.Fatalf("unexpected incident patch payload: %#v", patchBody)
	}
	if patchBody["updated_at"] == initialUpdatedAt {
		t.Fatalf("expected material patch to advance updated_at: before=%v after=%v", initialUpdatedAt, patchBody["updated_at"])
	}

	stale := doPhase2JSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID,
		map[string]any{
			"base_incident_version": 1,
			"tlp":                   "green",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, stale, http.StatusConflict, "incident_version_conflict")
}

func TestPhase2_MembershipAuthorizationLifecycle_I_2_04(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase2Server(t, postgresHarness, s3Harness, "phase2-i-2-04")
	defer db.Close()

	adminLogin, adminID := provisionBootstrapAdmin(t, server)
	targetID := seedLocalUserFlags(t, db, "viewer-target@example.test", "Viewer Target", "ViewerTargetPass1!", false, false, true)
	deploymentOnlyID := seedLocalUserFlags(t, db, "deployment-only@example.test", "Deployment Only", "DeploymentOnly1!", false, true, true)
	incident := createIncident(t, server, adminLogin, map[string]any{
		"client_txn_id": "txn-i-2-04-create",
		"incident_key":  "IR-I204",
		"title":         "Membership Lifecycle",
	})
	incidentID := incident["incident_id"].(string)

	membershipCreate := doPhase2JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-i-2-04-membership",
			"email":         " viewer-target@example.test ",
			"role":          "viewer",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	createdMembership := httptestx.RequireSuccessEnvelope(t, membershipCreate, http.StatusCreated)["data"].(map[string]any)
	if createdMembership["user_id"] != targetID || createdMembership["role"] != "viewer" {
		t.Fatalf("unexpected created membership payload: %#v", createdMembership)
	}

	targetSession, targetCSRF := loginLocalUser(t, server, "viewer-target@example.test", "ViewerTargetPass1!")
	targetGet := doPhase2JSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/incidents/"+incidentID, nil, withCookies(targetSession))
	httptestx.RequireSuccessEnvelope(t, targetGet, http.StatusOK)

	targetPatchDenied := doPhase2JSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID,
		map[string]any{
			"base_incident_version": 1,
			"tlp":                   "amber",
		},
		withCookies(targetSession, targetCSRF),
		withHeader(authn.CSRFHeaderName, targetCSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, targetPatchDenied, http.StatusForbidden, "authorization_denied")

	memberPatch := doPhase2JSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetID,
		map[string]any{
			"base_membership_version": 1,
			"role":                    "reviewer",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, memberPatch, http.StatusOK)

	targetPatchAllowed := doPhase2JSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID,
		map[string]any{
			"base_incident_version": 1,
			"current_phase":         "eradication",
		},
		withCookies(targetSession, targetCSRF),
		withHeader(authn.CSRFHeaderName, targetCSRF.Value),
	)
	httptestx.RequireSuccessEnvelope(t, targetPatchAllowed, http.StatusOK)

	deploymentOnlySession, _ := loginLocalUser(t, server, "deployment-only@example.test", "DeploymentOnly1!")
	deploymentOnlyDenied := doPhase2JSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/incidents/"+incidentID, nil, withCookies(deploymentOnlySession))
	httptestx.RequireErrorEnvelope(t, deploymentOnlyDenied, http.StatusNotFound, "incident_not_found")

	deleteMembership := doPhase2JSON(
		t,
		http.MethodDelete,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetID,
		nil,
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireStatus(t, deleteMembership, http.StatusNoContent)

	postDelete := doPhase2JSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/incidents/"+incidentID, nil, withCookies(targetSession))
	httptestx.RequireErrorEnvelope(t, postDelete, http.StatusNotFound, "incident_not_found")

	if got := queryCount(t, db, `SELECT COUNT(*) FROM incident_memberships WHERE incident_id::text = $1 AND user_id::text = $2`, incidentID, adminID); got != 1 {
		t.Fatalf("expected creator membership to remain, got %d", got)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM incident_memberships WHERE incident_id::text = $1 AND user_id::text = $2`, incidentID, deploymentOnlyID); got != 0 {
		t.Fatalf("deployment-only user must not gain implicit incident membership, got %d", got)
	}
}

func TestPhase2_MembershipGuards_I_2_05(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase2Server(t, postgresHarness, s3Harness, "phase2-i-2-05")
	defer db.Close()

	adminLogin, adminID := provisionBootstrapAdmin(t, server)
	inactiveID := seedLocalUserFlags(t, db, "inactive-member@example.test", "Inactive Member", "InactiveMember1!", false, false, false)
	incident := createIncident(t, server, adminLogin, map[string]any{
		"client_txn_id": "txn-i-2-05-create",
		"incident_key":  "IR-I205",
		"title":         "Admin Guards",
	})
	incidentID := incident["incident_id"].(string)

	inactiveCreate := doPhase2JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-i-2-05-inactive",
			"user_id":       inactiveID,
			"role":          "viewer",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, inactiveCreate, http.StatusConflict, "user_inactive")

	lastAdminPatch := doPhase2JSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+adminID,
		map[string]any{
			"base_membership_version": 1,
			"role":                    "reviewer",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, lastAdminPatch, http.StatusConflict, "last_incident_admin")

	lastAdminDelete := doPhase2JSON(
		t,
		http.MethodDelete,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+adminID,
		nil,
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, lastAdminDelete, http.StatusConflict, "last_incident_admin")
}

func TestPhase2_ExtensionDiscoveryAndReservedDispatch_I_2_06(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase2Server(t, postgresHarness, s3Harness, "phase2-i-2-06")
	defer db.Close()

	_, adminID := provisionBootstrapAdmin(t, server)
	userID := seedLocalUserFlags(t, db, "extension-user@example.test", "Extension User", "ExtensionUser1!", false, false, true)
	userSession, _ := loginLocalUser(t, server, "extension-user@example.test", "ExtensionUser1!")

	extensionsResp := doPhase2JSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/extensions", nil, withCookies(userSession))
	extensionsBody := httptestx.RequireSuccessEnvelope(t, extensionsResp, http.StatusOK)["data"].(map[string]any)
	extensions := extensionsBody["extensions"].([]any)
	if len(extensions) != 5 {
		t.Fatalf("unexpected extensions payload: %#v", extensionsBody)
	}

	rootReserved := doPhase2JSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/import-sessions", nil, withCookies(userSession))
	rootReservedBody := httptestx.RequireErrorEnvelope(t, rootReserved, http.StatusNotFound, "extension_profile_not_claimed")
	rootDetails := rootReservedBody["error"].(map[string]any)["details"].(map[string]any)
	if rootDetails["profile_id"] != "import" || rootDetails["route_family"] != "/api/v1/import-sessions" {
		t.Fatalf("unexpected reserved root dispatch details: %#v", rootDetails)
	}

	nestedReserved := doPhase2JSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/users/"+adminID+"/auth-bindings", nil, withCookies(userSession))
	nestedBody := httptestx.RequireErrorEnvelope(t, nestedReserved, http.StatusNotFound, "extension_profile_not_claimed")
	nestedDetails := nestedBody["error"].(map[string]any)["details"].(map[string]any)
	if nestedDetails["profile_id"] != "enterprise_authentication" || nestedDetails["route_family"] != "/api/v1/users/{user_id}/auth-bindings" {
		t.Fatalf("unexpected reserved nested dispatch details: %#v", nestedDetails)
	}

	_ = userID
}

type loginResult struct {
	sessionCookie *http.Cookie
	csrfCookie    *http.Cookie
}

func startPhase2Server(t testing.TB, postgresHarness *pgtest.Harness, s3Harness *s3test.Harness, prefix string) (*httptestx.Server, *sql.DB) {
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

	sessionResp := doPhase2JSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, withCookies(login.sessionCookie))
	sessionData := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	return login, sessionData["user_id"].(string)
}

func createIncident(t testing.TB, server *httptestx.Server, admin loginResult, body map[string]any) map[string]any {
	t.Helper()

	resp := doPhase2JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents",
		body,
		withCookies(admin.sessionCookie, admin.csrfCookie),
		withHeader(authn.CSRFHeaderName, admin.csrfCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func seedLocalUserFlags(t testing.TB, db *sql.DB, email string, displayName string, password string, mfaRequired bool, isDeploymentAdmin bool, isActive bool) string {
	t.Helper()

	hash, err := authn.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	var userID string
	if err := db.QueryRowContext(context.Background(), `
INSERT INTO users (email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id::text
`, email, displayName, hash, mfaRequired, isActive, isDeploymentAdmin).Scan(&userID); err != nil {
		t.Fatalf("seed local user with flags: %v", err)
	}
	return userID
}

func loginLocalUser(t testing.TB, server *httptestx.Server, username string, password string) (*http.Cookie, *http.Cookie) {
	t.Helper()

	resp := doPhase2JSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login failed: status=%d body=%#v", resp.StatusCode, httptestx.ReadJSONBody(t, resp))
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
	return sessionCookie, csrfCookie
}

func loginLocalUserWithSecondFactor(t testing.TB, server *httptestx.Server, username string, password string, code string) loginResult {
	t.Helper()

	resp := doPhase2JSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
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

func requireBootstrapLogin(t testing.TB, server *httptestx.Server, username string, password string) string {
	t.Helper()

	resp := doPhase2JSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusUnauthorized, "mfa_setup_required")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	return details["bootstrap_token"].(string)
}

func beginTOTPEnrollment(t testing.TB, server *httptestx.Server, bootstrapToken string, body map[string]any) map[string]any {
	t.Helper()

	resp := doPhase2JSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/begin", body, withHeader("Authorization", "Bearer "+bootstrapToken))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func completeInitialEnrollment(t testing.TB, server *httptestx.Server, bootstrapToken string, enrollmentID string, secretBase32 string, clientTxnID string) {
	t.Helper()

	resp := doPhase2JSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/complete", map[string]any{
		"client_txn_id": clientTxnID,
		"enrollment_id": enrollmentID,
		"code":          generateTOTPCode(t, secretBase32),
	}, withHeader("Authorization", "Bearer "+bootstrapToken))
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
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

func doPhase2JSON(t testing.TB, method string, url string, body any, options ...func(*http.Request)) *http.Response {
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
