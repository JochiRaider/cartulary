package main

import (
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
)

func TestPhase2_IncidentCreateListAndWorkbookPrefs_E_2_SMOKE_01_ProcessSmoke(t *testing.T) {
	server := startPhase1ServerProcess(t, "phase2-e-2-01")

	adminLogin, _ := phase1ProvisionBootstrapAdmin(t, server)
	created := phase2CreateIncident(t, server, adminLogin.sessionCookie, adminLogin.csrfCookie, map[string]any{
		"client_txn_id": "txn-e-2-01-create",
		"incident_key":  "IR-E201",
		"title":         "E2E Incident",
	})
	incidentID := created["incident_id"].(string)

	listResp := phase1DoJSON(t, server, http.MethodGet, "/api/v1/incidents", nil, withCookies(adminLogin.sessionCookie))
	listBody := httptestx.RequireSuccessEnvelope(t, listResp, http.StatusOK)["data"].(map[string]any)
	incidents := listBody["incidents"].([]any)
	if len(incidents) != 1 {
		t.Fatalf("expected one listed incident, got %#v", incidents)
	}

	getResp := phase1DoJSON(t, server, http.MethodGet, "/api/v1/incidents/"+incidentID, nil, withCookies(adminLogin.sessionCookie))
	getBody := httptestx.RequireSuccessEnvelope(t, getResp, http.StatusOK)["data"].(map[string]any)
	if getBody["incident_id"] != incidentID || getBody["status"] != "active" {
		t.Fatalf("unexpected incident get payload: %#v", getBody)
	}

	defaultPrefs := phase1DoJSON(t, server, http.MethodGet, "/api/v1/incidents/"+incidentID+"/workbook-preferences/default", nil, withCookies(adminLogin.sessionCookie))
	defaultPrefsBody := httptestx.RequireSuccessEnvelope(t, defaultPrefs, http.StatusOK)["data"].(map[string]any)
	if defaultPrefsBody["default_sheet_ref"] != nil {
		t.Fatalf("unexpected default workbook preferences payload: %#v", defaultPrefsBody)
	}

	sessionResp := phase1DoJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, withCookies(adminLogin.sessionCookie))
	sessionBody := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	memberships := sessionBody["memberships"].([]any)
	if len(memberships) != 1 {
		t.Fatalf("expected one session membership summary, got %#v", memberships)
	}
}

func TestPhase2_IncidentValidationAndPatch_E_2_SMOKE_01_ProcessSmoke(t *testing.T) {
	server := startPhase1ServerProcess(t, "phase2-e-2-02")

	adminLogin, _ := phase1ProvisionBootstrapAdmin(t, server)

	invalidCreate := phase1DoJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/incidents",
		map[string]any{
			"client_txn_id":       "txn-e-2-02-invalid",
			"incident_key":        "IR-E202",
			"title":               "Invalid",
			"initial_memberships": []any{},
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, invalidCreate, http.StatusBadRequest, "invalid_incident_create")

	created := phase2CreateIncident(t, server, adminLogin.sessionCookie, adminLogin.csrfCookie, map[string]any{
		"client_txn_id": "txn-e-2-02-create",
		"incident_key":  "IR-E202",
		"title":         "Patchable",
	})
	incidentID := created["incident_id"].(string)

	invalidPatch := phase1DoJSON(
		t,
		server,
		http.MethodPatch,
		"/api/v1/incidents/"+incidentID,
		map[string]any{
			"base_incident_version": 1,
			"title":                 "forbidden",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, invalidPatch, http.StatusBadRequest, "invalid_incident_patch")

	patchResp := phase1DoJSON(
		t,
		server,
		http.MethodPatch,
		"/api/v1/incidents/"+incidentID,
		map[string]any{
			"base_incident_version":     1,
			"tlp":                       "amber",
			"primary_external_case_ref": "CASE-E202",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	patchBody := httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)["data"].(map[string]any)
	if patchBody["incident_version"] != float64(2) || patchBody["tlp"] != "amber" {
		t.Fatalf("unexpected incident patch payload: %#v", patchBody)
	}
}

func TestPhase2_MembershipAdminFlow_E_2_SMOKE_01_ProcessSmoke(t *testing.T) {
	server := startPhase1ServerProcess(t, "phase2-e-2-03")

	adminLogin, _ := phase1ProvisionBootstrapAdmin(t, server)
	createdUser := phase1CreateUser(t, server, adminLogin.sessionCookie, adminLogin.csrfCookie, map[string]any{
		"client_txn_id":    "txn-e-2-03-user",
		"auth_kind":        "local",
		"email":            "phase2-e-2-03@example.test",
		"display_name":     "Phase2 E203",
		"initial_password": "Phase2E203Pass!",
		"mfa_required":     false,
	})
	userID := createdUser["user_id"].(string)
	userLogin := phase1LoginLocalUserWithSecondFactor(t, server, "phase2-e-2-03@example.test", "Phase2E203Pass!", "")

	incident := phase2CreateIncident(t, server, adminLogin.sessionCookie, adminLogin.csrfCookie, map[string]any{
		"client_txn_id": "txn-e-2-03-incident",
		"incident_key":  "IR-E203",
		"title":         "Membership Flow",
	})
	incidentID := incident["incident_id"].(string)

	createMembership := phase1DoJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-e-2-03-membership",
			"user_id":       userID,
			"role":          "viewer",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, createMembership, http.StatusCreated)

	listMemberships := phase1DoJSON(t, server, http.MethodGet, "/api/v1/incidents/"+incidentID+"/memberships", nil, withCookies(adminLogin.sessionCookie))
	listBody := httptestx.RequireSuccessEnvelope(t, listMemberships, http.StatusOK)["data"].(map[string]any)
	if len(listBody["memberships"].([]any)) != 2 {
		t.Fatalf("unexpected incident membership list: %#v", listBody)
	}

	nonAdminCreate := phase1DoJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-e-2-03-non-admin",
			"email":         "nobody@example.test",
			"role":          "viewer",
		},
		withCookies(userLogin.sessionCookie, userLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, userLogin.csrfCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, nonAdminCreate, http.StatusForbidden, "authorization_denied")
}

func TestPhase2_MembershipPatchDeleteAndLastAdmin_E_2_SMOKE_01_ProcessSmoke(t *testing.T) {
	server := startPhase1ServerProcess(t, "phase2-e-2-04")

	adminLogin, _ := phase1ProvisionBootstrapAdmin(t, server)
	adminSession := adminLogin.sessionCookie
	adminCSRF := adminLogin.csrfCookie
	sessionResp := phase1DoJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, withCookies(adminSession))
	sessionBody := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	adminUserID := sessionBody["user_id"].(string)
	createdUser := phase1CreateUser(t, server, adminSession, adminCSRF, map[string]any{
		"client_txn_id":    "txn-e-2-04-user",
		"auth_kind":        "local",
		"email":            "phase2-e-2-04@example.test",
		"display_name":     "Phase2 E204",
		"initial_password": "Phase2E204Pass!",
	})
	userID := createdUser["user_id"].(string)
	incident := phase2CreateIncident(t, server, adminSession, adminCSRF, map[string]any{
		"client_txn_id": "txn-e-2-04-incident",
		"incident_key":  "IR-E204",
		"title":         "Membership Mutations",
	})
	incidentID := incident["incident_id"].(string)

	createMembership := phase1DoJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-e-2-04-membership",
			"user_id":       userID,
			"role":          "viewer",
		},
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	createMembershipBody := httptestx.RequireSuccessEnvelope(t, createMembership, http.StatusCreated)["data"].(map[string]any)

	patchMembership := phase1DoJSON(
		t,
		server,
		http.MethodPatch,
		"/api/v1/incidents/"+incidentID+"/memberships/"+userID,
		map[string]any{
			"base_membership_version": createMembershipBody["membership_version"],
			"role":                    "reviewer",
		},
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	patchMembershipBody := httptestx.RequireSuccessEnvelope(t, patchMembership, http.StatusOK)["data"].(map[string]any)
	if patchMembershipBody["role"] != "reviewer" || patchMembershipBody["membership_version"] != float64(2) {
		t.Fatalf("unexpected membership patch payload: %#v", patchMembershipBody)
	}

	deleteMembership := phase1DoJSON(
		t,
		server,
		http.MethodDelete,
		"/api/v1/incidents/"+incidentID+"/memberships/"+userID,
		map[string]any{
			"base_membership_version": patchMembershipBody["membership_version"],
		},
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	httptestx.RequireStatus(t, deleteMembership, http.StatusNoContent)

	lastAdminGuard := phase1DoJSON(
		t,
		server,
		http.MethodDelete,
		"/api/v1/incidents/"+incidentID+"/memberships/"+adminUserID,
		map[string]any{
			"base_membership_version": 1,
		},
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, lastAdminGuard, http.StatusConflict, "last_incident_admin")
}

func TestPhase2_ExtensionDiscoveryAndReservedRoutes_E_2_SMOKE_01_ProcessSmoke(t *testing.T) {
	server := startPhase1ServerProcess(t, "phase2-e-2-05")

	adminLogin, _ := phase1ProvisionBootstrapAdmin(t, server)
	createdUser := phase1CreateUser(t, server, adminLogin.sessionCookie, adminLogin.csrfCookie, map[string]any{
		"client_txn_id":    "txn-e-2-05-user",
		"auth_kind":        "local",
		"email":            "phase2-e-2-05@example.test",
		"display_name":     "Phase2 E205",
		"initial_password": "Phase2E205Pass!",
		"mfa_required":     false,
	})
	userID := createdUser["user_id"].(string)
	userLogin := phase1LoginLocalUserWithSecondFactor(t, server, "phase2-e-2-05@example.test", "Phase2E205Pass!", "")

	extensions := phase1DoJSON(t, server, http.MethodGet, "/api/v1/extensions", nil, withCookies(userLogin.sessionCookie))
	extensionsBody := httptestx.RequireSuccessEnvelope(t, extensions, http.StatusOK)["data"].(map[string]any)
	if len(extensionsBody["extensions"].([]any)) != 5 {
		t.Fatalf("unexpected extensions payload: %#v", extensionsBody)
	}

	rootReserved := phase1DoJSON(t, server, http.MethodGet, "/api/v1/reference-packs", nil, withCookies(userLogin.sessionCookie))
	rootReservedBody := httptestx.RequireErrorEnvelope(t, rootReserved, http.StatusNotFound, "extension_profile_not_claimed")
	rootDetails := rootReservedBody["error"].(map[string]any)["details"].(map[string]any)
	if rootDetails["profile_id"] != "reference_pack" {
		t.Fatalf("unexpected reference-pack reserved dispatch details: %#v", rootDetails)
	}

	nestedReserved := phase1DoJSON(t, server, http.MethodGet, "/api/v1/users/"+userID+"/auth-bindings", nil, withCookies(userLogin.sessionCookie))
	httptestx.RequireErrorEnvelope(t, nestedReserved, http.StatusNotFound, "extension_profile_not_claimed")
}

func TestPhase2_DeploymentAdminBoundary_E_2_SMOKE_01_ProcessSmoke(t *testing.T) {
	server := startPhase1ServerProcess(t, "phase2-e-2-06")

	adminLogin, _ := phase1ProvisionBootstrapAdmin(t, server)
	incident := phase2CreateIncident(t, server, adminLogin.sessionCookie, adminLogin.csrfCookie, map[string]any{
		"client_txn_id": "txn-e-2-06-incident",
		"incident_key":  "IR-E206",
		"title":         "Boundary Incident",
	})
	incidentID := incident["incident_id"].(string)

	deploymentOnly := phase1CreateUser(t, server, adminLogin.sessionCookie, adminLogin.csrfCookie, map[string]any{
		"client_txn_id":       "txn-e-2-06-user",
		"auth_kind":           "local",
		"email":               "phase2-e-2-06@example.test",
		"display_name":        "Phase2 E206",
		"initial_password":    "Phase2E206Pass!",
		"mfa_required":        false,
		"is_deployment_admin": true,
	})
	_ = deploymentOnly
	deploymentLogin := phase1LoginLocalUserWithSecondFactor(t, server, "phase2-e-2-06@example.test", "Phase2E206Pass!", "")

	getIncident := phase1DoJSON(t, server, http.MethodGet, "/api/v1/incidents/"+incidentID, nil, withCookies(deploymentLogin.sessionCookie))
	httptestx.RequireErrorEnvelope(t, getIncident, http.StatusNotFound, "incident_not_found")

	listIncidentMemberships := phase1DoJSON(t, server, http.MethodGet, "/api/v1/incidents/"+incidentID+"/memberships", nil, withCookies(deploymentLogin.sessionCookie))
	httptestx.RequireErrorEnvelope(t, listIncidentMemberships, http.StatusNotFound, "incident_not_found")
}

func phase2CreateIncident(t testing.TB, server *processtest.Server, adminSession *http.Cookie, adminCSRF *http.Cookie, body map[string]any) map[string]any {
	t.Helper()

	resp := phase1DoJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/incidents",
		body,
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}
