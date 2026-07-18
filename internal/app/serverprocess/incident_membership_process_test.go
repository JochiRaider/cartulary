package serverprocess

import (
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
)

func TestIncidentCreateListAndWorkbookPrefs_Process(t *testing.T) {
	server := startServerProcess(t, "phase2-e-2-01")

	adminLogin, _ := ProvisionBootstrapAdmin(t, server)
	created := CreateIncident(t, server, adminLogin.sessionCookie, adminLogin.csrfCookie, map[string]any{
		"client_txn_id": "txn-e-2-01-create",
		"incident_key":  "IR-E201",
		"title":         "E2E Incident",
	})
	incidentID := created["incident_id"].(string)

	listResp := DoJSON(t, server, http.MethodGet, "/api/v1/incidents", nil, withCookies(adminLogin.sessionCookie))
	listBody := httptestx.RequireSuccessEnvelope(t, listResp, http.StatusOK)["data"].(map[string]any)
	incidents := listBody["incidents"].([]any)
	if len(incidents) != 1 {
		t.Fatalf("expected one listed incident, got %#v", incidents)
	}

	getResp := DoJSON(t, server, http.MethodGet, "/api/v1/incidents/"+incidentID, nil, withCookies(adminLogin.sessionCookie))
	getBody := httptestx.RequireSuccessEnvelope(t, getResp, http.StatusOK)["data"].(map[string]any)
	if getBody["incident_id"] != incidentID || getBody["status"] != "active" {
		t.Fatalf("unexpected incident get payload: %#v", getBody)
	}

	defaultPrefs := DoJSON(t, server, http.MethodGet, "/api/v1/incidents/"+incidentID+"/workbook-preferences/default", nil, withCookies(adminLogin.sessionCookie))
	defaultPrefsBody := httptestx.RequireSuccessEnvelope(t, defaultPrefs, http.StatusOK)["data"].(map[string]any)
	if defaultPrefsBody["default_sheet_ref"] != nil {
		t.Fatalf("unexpected default workbook preferences payload: %#v", defaultPrefsBody)
	}

	sessionResp := DoJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, withCookies(adminLogin.sessionCookie))
	sessionBody := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	memberships := sessionBody["memberships"].([]any)
	if len(memberships) != 1 {
		t.Fatalf("expected one session membership summary, got %#v", memberships)
	}
}

func TestIncidentValidationAndPatch_Process(t *testing.T) {
	server := startServerProcess(t, "phase2-e-2-02")

	adminLogin, _ := ProvisionBootstrapAdmin(t, server)

	invalidCreate := DoJSON(
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

	created := CreateIncident(t, server, adminLogin.sessionCookie, adminLogin.csrfCookie, map[string]any{
		"client_txn_id": "txn-e-2-02-create",
		"incident_key":  "IR-E202",
		"title":         "Patchable",
	})
	incidentID := created["incident_id"].(string)

	invalidPatch := DoJSON(
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

	patchResp := DoJSON(
		t,
		server,
		http.MethodPatch,
		"/api/v1/incidents/"+incidentID,
		map[string]any{
			"base_incident_version":     1,
			"tlp":                       "TLP:AMBER",
			"primary_external_case_ref": "CASE-E202",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	patchBody := httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)["data"].(map[string]any)
	if patchBody["incident_version"] != float64(2) || patchBody["tlp"] != "TLP:AMBER" {
		t.Fatalf("unexpected incident patch payload: %#v", patchBody)
	}
}

func TestMembershipAdminFlow_Process(t *testing.T) {
	server := startServerProcess(t, "phase2-e-2-03")

	adminLogin, _ := ProvisionBootstrapAdmin(t, server)
	createdUser := CreateUser(t, server, adminLogin.sessionCookie, adminLogin.csrfCookie, map[string]any{
		"client_txn_id":    "txn-e-2-03-user",
		"auth_kind":        "local",
		"email":            "phase2-e-2-03@example.test",
		"display_name":     "Phase2 E203",
		"initial_password": "Phase2E203Pass!",
		"mfa_required":     false,
	})
	userID := createdUser["user_id"].(string)
	userLogin := LoginLocalUserWithSecondFactor(t, server, "phase2-e-2-03@example.test", "Phase2E203Pass!", "")

	incident := CreateIncident(t, server, adminLogin.sessionCookie, adminLogin.csrfCookie, map[string]any{
		"client_txn_id": "txn-e-2-03-incident",
		"incident_key":  "IR-E203",
		"title":         "Membership Flow",
	})
	incidentID := incident["incident_id"].(string)

	createMembership := DoJSON(
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

	listMemberships := DoJSON(t, server, http.MethodGet, "/api/v1/incidents/"+incidentID+"/memberships", nil, withCookies(adminLogin.sessionCookie))
	listBody := httptestx.RequireSuccessEnvelope(t, listMemberships, http.StatusOK)["data"].(map[string]any)
	if len(listBody["memberships"].([]any)) != 2 {
		t.Fatalf("unexpected incident membership list: %#v", listBody)
	}

	nonAdminCreate := DoJSON(
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

func TestMembershipPatchDeleteAndLastAdmin_Process(t *testing.T) {
	server := startServerProcess(t, "phase2-e-2-04")

	adminLogin, _ := ProvisionBootstrapAdmin(t, server)
	adminSession := adminLogin.sessionCookie
	adminCSRF := adminLogin.csrfCookie
	sessionResp := DoJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, withCookies(adminSession))
	sessionBody := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	adminUserID := sessionBody["user_id"].(string)
	createdUser := CreateUser(t, server, adminSession, adminCSRF, map[string]any{
		"client_txn_id":    "txn-e-2-04-user",
		"auth_kind":        "local",
		"email":            "phase2-e-2-04@example.test",
		"display_name":     "Phase2 E204",
		"initial_password": "Phase2E204Pass!",
	})
	userID := createdUser["user_id"].(string)
	incident := CreateIncident(t, server, adminSession, adminCSRF, map[string]any{
		"client_txn_id": "txn-e-2-04-incident",
		"incident_key":  "IR-E204",
		"title":         "Membership Mutations",
	})
	incidentID := incident["incident_id"].(string)

	createMembership := DoJSON(
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

	patchMembership := DoJSON(
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

	deleteMembership := DoJSON(
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

	lastAdminGuard := DoJSON(
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

func TestExtensionDiscoveryAndReservedRoutes_Process(t *testing.T) {
	server := startServerProcess(t, "phase2-e-2-05")

	adminLogin, _ := ProvisionBootstrapAdmin(t, server)
	createdUser := CreateUser(t, server, adminLogin.sessionCookie, adminLogin.csrfCookie, map[string]any{
		"client_txn_id":    "txn-e-2-05-user",
		"auth_kind":        "local",
		"email":            "phase2-e-2-05@example.test",
		"display_name":     "Phase2 E205",
		"initial_password": "Phase2E205Pass!",
		"mfa_required":     false,
	})
	userID := createdUser["user_id"].(string)
	userLogin := LoginLocalUserWithSecondFactor(t, server, "phase2-e-2-05@example.test", "Phase2E205Pass!", "")

	extensions := DoJSON(t, server, http.MethodGet, "/api/v1/extensions", nil, withCookies(userLogin.sessionCookie))
	extensionsBody := httptestx.RequireSuccessEnvelope(t, extensions, http.StatusOK)["data"].(map[string]any)
	if len(extensionsBody["extensions"].([]any)) != 6 {
		t.Fatalf("unexpected extensions payload: %#v", extensionsBody)
	}
	requireSmokeExtensionClaim(t, extensionsBody, "network_flow_activity", false)

	rootReserved := DoJSON(t, server, http.MethodGet, "/api/v1/auth/providers", nil, withCookies(userLogin.sessionCookie))
	rootReservedBody := httptestx.RequireErrorEnvelope(t, rootReserved, http.StatusNotFound, "extension_profile_not_claimed")
	rootDetails := rootReservedBody["error"].(map[string]any)["details"].(map[string]any)
	if rootDetails["profile_id"] != "enterprise_authentication" {
		t.Fatalf("unexpected enterprise-authentication reserved dispatch details: %#v", rootDetails)
	}

	nestedReserved := DoJSON(t, server, http.MethodGet, "/api/v1/users/"+userID+"/auth-bindings", nil, withCookies(userLogin.sessionCookie))
	httptestx.RequireErrorEnvelope(t, nestedReserved, http.StatusNotFound, "extension_profile_not_claimed")
}

func requireSmokeExtensionClaim(t testing.TB, body map[string]any, profileID string, claimed bool) {
	t.Helper()
	for _, item := range body["extensions"].([]any) {
		extension := item.(map[string]any)
		if extension["profile_id"] == profileID {
			if extension["claimed"] != claimed {
				t.Fatalf("extension %s claimed=%#v want %v in %#v", profileID, extension["claimed"], claimed, body)
			}
			return
		}
	}
	t.Fatalf("extension %s missing from %#v", profileID, body)
}

func TestDeploymentAdminBoundary_Process(t *testing.T) {
	server := startServerProcess(t, "phase2-e-2-06")

	adminLogin, _ := ProvisionBootstrapAdmin(t, server)
	incident := CreateIncident(t, server, adminLogin.sessionCookie, adminLogin.csrfCookie, map[string]any{
		"client_txn_id": "txn-e-2-06-incident",
		"incident_key":  "IR-E206",
		"title":         "Boundary Incident",
	})
	incidentID := incident["incident_id"].(string)

	deploymentOnly := CreateUser(t, server, adminLogin.sessionCookie, adminLogin.csrfCookie, map[string]any{
		"client_txn_id":       "txn-e-2-06-user",
		"auth_kind":           "local",
		"email":               "phase2-e-2-06@example.test",
		"display_name":        "Phase2 E206",
		"initial_password":    "Phase2E206Pass!",
		"mfa_required":        false,
		"is_deployment_admin": true,
	})
	_ = deploymentOnly
	deploymentLogin := LoginLocalUserWithSecondFactor(t, server, "phase2-e-2-06@example.test", "Phase2E206Pass!", "")

	getIncident := DoJSON(t, server, http.MethodGet, "/api/v1/incidents/"+incidentID, nil, withCookies(deploymentLogin.sessionCookie))
	httptestx.RequireErrorEnvelope(t, getIncident, http.StatusNotFound, "incident_not_found")

	listIncidentMemberships := DoJSON(t, server, http.MethodGet, "/api/v1/incidents/"+incidentID+"/memberships", nil, withCookies(deploymentLogin.sessionCookie))
	httptestx.RequireErrorEnvelope(t, listIncidentMemberships, http.StatusNotFound, "incident_not_found")
}

func CreateIncident(t testing.TB, server *processtest.Server, adminSession *http.Cookie, adminCSRF *http.Cookie, body map[string]any) map[string]any {
	t.Helper()

	resp := DoJSON(
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
