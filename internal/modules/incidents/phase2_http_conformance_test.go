package incidents_test

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase2test"
)

func TestSupportPhase2_IncidentCreateBootstrapsCreatorAndWorkbookPreferencesHTTPConformance(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase2-u-2-02")

	adminLogin, adminID := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	createResp := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-u-2-02-create",
			"incident_key":  "IR-U202",
			"title":         "Bootstrap Incident",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	createBody := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)
	incidentID := createBody["data"].(map[string]any)["incident_id"].(string)

	sessionResp := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/auth/session", nil, phase2test.WithCookies(adminLogin.SessionCookie))
	sessionBody := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)
	memberships := sessionBody["data"].(map[string]any)["memberships"].([]any)
	if len(memberships) != 1 {
		t.Fatalf("expected one bootstrap membership summary, got %#v", memberships)
	}
	sessionMembership := memberships[0].(map[string]any)
	if sessionMembership["incident_id"] != incidentID || sessionMembership["role"] != "admin" {
		t.Fatalf("unexpected bootstrap session membership: %#v", sessionMembership)
	}

	defaultPrefsResp := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/default", nil, phase2test.WithCookies(adminLogin.SessionCookie))
	defaultPrefs := httptestx.RequireSuccessEnvelope(t, defaultPrefsResp, http.StatusOK)["data"].(map[string]any)
	if defaultPrefs["incident_id"] != incidentID || defaultPrefs["default_sheet_ref"] != nil || defaultPrefs["updated_by_user_id"] != adminID {
		t.Fatalf("unexpected default workbook preferences payload: %#v", defaultPrefs)
	}

	userPrefsResp := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/me", nil, phase2test.WithCookies(adminLogin.SessionCookie))
	userPrefs := httptestx.RequireSuccessEnvelope(t, userPrefsResp, http.StatusOK)["data"].(map[string]any)
	if userPrefs["incident_id"] != incidentID || userPrefs["user_id"] != adminID || userPrefs["home_sheet_ref"] != nil {
		t.Fatalf("unexpected user workbook preferences payload: %#v", userPrefs)
	}
}

func TestSupportPhase2_PublicRouteInventoryEnvelopes(t *testing.T) {
	for _, route := range phase2test.PublicRouteInventory() {
		route := route
		t.Run(route.Name, func(t *testing.T) {
			fixtureCtx := newPhase2RouteFixture(t, "phase2-public-"+route.Name)
			resp := executeRouteRequest(
				t,
				fixtureCtx.harness.Server.HTTP.URL,
				route,
				fixtureCtx.routeFixture(route.Name),
				fixtureCtx.adminLogin.SessionCookie,
				fixtureCtx.adminLogin.CSRFCookie,
			)
			requireRouteSuccess(t, resp, route)
		})
	}
}

func TestSupportPhase2_IncidentCreateReturnsStableLocationHeaderHTTPConformance(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase2-u-2-03")

	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	createResp := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-u-2-03-create",
			"incident_key":  "IR-U203",
			"title":         "Location Incident",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	body := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)
	incidentID := body["data"].(map[string]any)["incident_id"].(string)
	if got := createResp.Header.Get("Location"); got != "/api/v1/incidents/"+incidentID {
		t.Fatalf("unexpected incident Location header: got %q", got)
	}
}

func TestSupportPhase2_IncidentCreateIdempotencyUsesActorAndNormalizedReplayHTTPConformance(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase2-u-2-04")

	firstActor, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	firstCreate := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-u-2-04",
			"incident_key":  "  IR-U204  ",
			"title":         "  Replay Incident  ",
		},
		phase2test.WithCookies(firstActor.SessionCookie, firstActor.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, firstActor.CSRFCookie.Value),
	)
	firstData := httptestx.RequireSuccessEnvelope(t, firstCreate, http.StatusCreated)["data"].(map[string]any)

	replay := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-u-2-04",
			"incident_key":  "IR-U204",
			"title":         "Replay Incident",
		},
		phase2test.WithCookies(firstActor.SessionCookie, firstActor.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, firstActor.CSRFCookie.Value),
	)
	replayData := httptestx.RequireSuccessEnvelope(t, replay, http.StatusOK)["data"].(map[string]any)
	if !reflect.DeepEqual(firstData, replayData) {
		t.Fatalf("expected replayed incident payload to match original result: first=%#v replay=%#v", firstData, replayData)
	}

	phase2test.SeedLocalUserFlags(t, harness.DB, "phase2-u204@example.test", "Phase2 U204", "Phase2U204Pass!", false, false, true)
	secondSession, secondCSRF := phase2test.LoginLocalUser(t, harness.Server, "phase2-u204@example.test", "Phase2U204Pass!")
	secondCreate := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-u-2-04",
			"incident_key":  "IR-U204-SECOND",
			"title":         "Second Actor Incident",
		},
		phase2test.WithCookies(secondSession, secondCSRF),
		phase2test.WithHeader(authn.CSRFHeaderName, secondCSRF.Value),
	)
	secondData := httptestx.RequireSuccessEnvelope(t, secondCreate, http.StatusCreated)["data"].(map[string]any)
	if secondData["incident_id"] == firstData["incident_id"] {
		t.Fatalf("expected actor-scoped idempotency to allow a distinct create, got %#v", secondData)
	}

	phase2test.RequireErrorContract(t, "client_txn_conflict", http.StatusConflict)
	divergent := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-u-2-04",
			"incident_key":  "IR-U204",
			"title":         "Different title",
		},
		phase2test.WithCookies(firstActor.SessionCookie, firstActor.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, firstActor.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, divergent, http.StatusConflict, "client_txn_conflict")
}

func TestSupportPhase2_MembershipCreateRequiresOneSelectorClosedRolesAndNoInvitationFieldsHTTPConformance(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase2-u-2-06")

	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-u-2-06-incident",
		"incident_key":  "IR-U206",
		"title":         "Membership Targets",
	})
	incidentID := incident["incident_id"].(string)

	missingSelector := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-u-2-06-missing-target",
			"role":          "viewer",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	missingSelectorBody := httptestx.RequireErrorEnvelope(t, missingSelector, http.StatusBadRequest, "invalid_mutation_payload")
	requireErrorDetails(t, missingSelectorBody, "user_id", "exactly_one_target_selector")

	dualSelector := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-u-2-06-dual-target",
			"user_id":       "00000000-0000-0000-0000-000000000602",
			"email":         "dual@example.test",
			"role":          "viewer",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	dualSelectorBody := httptestx.RequireErrorEnvelope(t, dualSelector, http.StatusBadRequest, "invalid_mutation_payload")
	requireErrorDetails(t, dualSelectorBody, "user_id", "exactly_one_target_selector")

	invalidRole := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-u-2-06-invalid-role",
			"email":         "solo@example.test",
			"role":          "owner",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	invalidRoleBody := httptestx.RequireErrorEnvelope(t, invalidRole, http.StatusBadRequest, "invalid_mutation_payload")
	requireErrorDetails(t, invalidRoleBody, "role", "invalid_role")

	unknownInvitation := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id":    "txn-u-2-06-no-invite",
			"email":            "solo@example.test",
			"role":             "viewer",
			"invitation_email": "new@example.test",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	unknownInvitationBody := httptestx.RequireErrorEnvelope(t, unknownInvitation, http.StatusBadRequest, "invalid_mutation_payload")
	requireErrorDetails(t, unknownInvitationBody, "invitation_email", "unknown_field")

	phase2test.RequireErrorContract(t, "user_not_found", http.StatusNotFound)
	missingUser := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-u-2-06-missing-user",
			"email":         "missing@example.test",
			"role":          "viewer",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, missingUser, http.StatusNotFound, "user_not_found")
	if got := phase2test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM users WHERE email = $1`, "missing@example.test"); got != 0 {
		t.Fatalf("membership create must not auto-create a missing user, got %d rows", got)
	}

	phase2test.SeedLocalUserFlags(t, harness.DB, "inactive@example.test", "Inactive User", "InactiveUser1!", false, false, false)
	phase2test.RequireErrorContract(t, "user_inactive", http.StatusConflict)
	inactiveUser := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-u-2-06-inactive-user",
			"email":         "inactive@example.test",
			"role":          "viewer",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, inactiveUser, http.StatusConflict, "user_inactive")
}

func TestSupportPhase2_MembershipPatchAndDeleteEnforceBaseVersionAndLastAdminGuardHTTPConformance(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase2-u-2-07")

	adminLogin, adminID := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-u-2-07-incident",
		"incident_key":  "IR-U207",
		"title":         "Membership Versioning",
	})
	incidentID := incident["incident_id"].(string)
	targetUserID := phase2test.SeedLocalUserFlags(t, harness.DB, "phase2-u207@example.test", "Phase2 U207", "Phase2U207Pass!", false, false, true)

	createResp := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-u-2-07-membership",
			"user_id":       targetUserID,
			"role":          "viewer",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	createdMembership := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)

	patchMembership := phase2test.DoJSON(
		t,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetUserID,
		map[string]any{
			"base_membership_version": createdMembership["membership_version"],
			"role":                    "reviewer",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	patchedMembership := httptestx.RequireSuccessEnvelope(t, patchMembership, http.StatusOK)["data"].(map[string]any)

	phase2test.RequireErrorContract(t, "membership_version_conflict", http.StatusConflict)
	stalePatch := phase2test.DoJSON(
		t,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetUserID,
		map[string]any{
			"base_membership_version": createdMembership["membership_version"],
			"role":                    "admin",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, stalePatch, http.StatusConflict, "membership_version_conflict")

	clientTxnPatch := phase2test.DoJSON(
		t,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetUserID,
		map[string]any{
			"client_txn_id":           "forbidden",
			"base_membership_version": patchedMembership["membership_version"],
			"role":                    "admin",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	patchError := httptestx.RequireErrorEnvelope(t, clientTxnPatch, http.StatusBadRequest, "invalid_mutation_payload")
	requireErrorDetails(t, patchError, "client_txn_id", "unknown_field")

	staleDelete := phase2test.DoJSON(
		t,
		http.MethodDelete,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetUserID,
		map[string]any{
			"base_membership_version": createdMembership["membership_version"],
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, staleDelete, http.StatusConflict, "membership_version_conflict")

	clientTxnDelete := phase2test.DoJSON(
		t,
		http.MethodDelete,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetUserID,
		map[string]any{
			"client_txn_id":           "forbidden",
			"base_membership_version": patchedMembership["membership_version"],
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	deleteError := httptestx.RequireErrorEnvelope(t, clientTxnDelete, http.StatusBadRequest, "invalid_mutation_payload")
	requireErrorDetails(t, deleteError, "client_txn_id", "unknown_field")

	phase2test.RequireErrorContract(t, "last_incident_admin", http.StatusConflict)
	lastAdmin := phase2test.DoJSON(
		t,
		http.MethodDelete,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+adminID,
		map[string]any{
			"base_membership_version": 1,
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, lastAdmin, http.StatusConflict, "last_incident_admin")
}

func TestSupportPhase2_ControlBoundaryInventoryDeploymentAdminWithoutMembershipDenied(t *testing.T) {
	for _, route := range phase2test.ControlBoundaryInventory() {
		route := route
		t.Run(route.Name, func(t *testing.T) {
			fixtureCtx := newPhase2RouteFixture(t, "phase2-control-denied-"+route.Name)
			deploymentEmail := phase2FixtureSlug("phase2-control-denied-" + route.Name)
			phase2test.SeedLocalUserFlags(
				t,
				fixtureCtx.harness.DB,
				deploymentEmail+"@example.test",
				"Deployment Only "+deploymentEmail,
				"DeploymentOnly1!",
				false,
				true,
				true,
			)
			deploymentSession, deploymentCSRF := phase2test.LoginLocalUser(
				t,
				fixtureCtx.harness.Server,
				deploymentEmail+"@example.test",
				"DeploymentOnly1!",
			)
			requireControlRouteOutcome(
				t,
				fixtureCtx.harness.Server.HTTP.URL,
				route,
				fixtureCtx.routeFixture(route.Name),
				deploymentSession,
				deploymentCSRF,
				controlStageNoMembership,
			)
		})
	}
}

func requireErrorDetails(t testing.TB, envelope map[string]any, wantField string, wantReasonCode string) {
	t.Helper()

	errorBody := envelope["error"].(map[string]any)
	details := errorBody["details"].(map[string]any)
	if details["field"] != wantField {
		t.Fatalf("unexpected error field: got %v want %q", details["field"], wantField)
	}
	if details["reason_code"] != wantReasonCode {
		t.Fatalf("unexpected error reason_code: got %v want %q", details["reason_code"], wantReasonCode)
	}
}
