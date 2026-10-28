package serverprocess

import (
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
)

func TestIncidentCreateListAndWorkbookPrefs_Process(t *testing.T) {
	server := startServerProcess(t, "incident_membership-e-2-01")

	adminLogin, _ := provisionBootstrapAdmin(t, server)
	created := createIncident(t, server, adminLogin.Session, adminLogin.CSRF, map[string]any{
		"client_txn_id": "txn-e-2-01-create",
		"incident_key":  "IR-E201",
		"title":         "E2E Incident",
	})
	incidentID := created["incident_id"].(string)

	listResp := doJSON(t, server, http.MethodGet, "/api/v1/incidents", nil, httptestx.WithCookies(adminLogin.Session))
	listBody := httptestx.RequireSuccessEnvelope(t, listResp, http.StatusOK)["data"].(map[string]any)
	incidents := listBody["incidents"].([]any)
	if len(incidents) != 1 {
		t.Fatalf("expected one listed incident, got %#v", incidents)
	}

	getResp := doJSON(t, server, http.MethodGet, "/api/v1/incidents/"+incidentID, nil, httptestx.WithCookies(adminLogin.Session))
	getBody := httptestx.RequireSuccessEnvelope(t, getResp, http.StatusOK)["data"].(map[string]any)
	if getBody["incident_id"] != incidentID || getBody["status"] != "active" {
		t.Fatalf("unexpected incident get payload: %#v", getBody)
	}

	defaultPrefs := doJSON(t, server, http.MethodGet, "/api/v1/incidents/"+incidentID+"/workbook-preferences/default", nil, httptestx.WithCookies(adminLogin.Session))
	defaultPrefsBody := httptestx.RequireSuccessEnvelope(t, defaultPrefs, http.StatusOK)["data"].(map[string]any)
	if defaultPrefsBody["default_sheet_ref"] != nil {
		t.Fatalf("unexpected default workbook preferences payload: %#v", defaultPrefsBody)
	}

	sessionResp := doJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, httptestx.WithCookies(adminLogin.Session))
	sessionBody := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	memberships := sessionBody["memberships"].([]any)
	if len(memberships) != 1 {
		t.Fatalf("expected one session membership summary, got %#v", memberships)
	}
}

func TestIncidentValidationAndPatch_Process(t *testing.T) {
	server := startServerProcess(t, "incident_membership-e-2-02")

	adminLogin, _ := provisionBootstrapAdmin(t, server)

	invalidCreate := doJSON(
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
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, invalidCreate, http.StatusBadRequest, "invalid_incident_create")

	created := createIncident(t, server, adminLogin.Session, adminLogin.CSRF, map[string]any{
		"client_txn_id": "txn-e-2-02-create",
		"incident_key":  "IR-E202",
		"title":         "Patchable",
	})
	incidentID := created["incident_id"].(string)

	invalidPatch := doJSON(
		t,
		server,
		http.MethodPatch,
		"/api/v1/incidents/"+incidentID,
		map[string]any{
			"base_incident_version": 1,
			"title":                 "forbidden",
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, invalidPatch, http.StatusBadRequest, "invalid_incident_patch")

	patchResp := doJSON(
		t,
		server,
		http.MethodPatch,
		"/api/v1/incidents/"+incidentID,
		map[string]any{
			"base_incident_version":     1,
			"tlp":                       "TLP:AMBER",
			"primary_external_case_ref": "CASE-E202",
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	patchBody := httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)["data"].(map[string]any)
	if patchBody["incident_version"] != float64(2) || patchBody["tlp"] != "TLP:AMBER" {
		t.Fatalf("unexpected incident patch payload: %#v", patchBody)
	}
}

func TestMembershipAdminFlow_Process(t *testing.T) {
	server := startServerProcess(t, "incident_membership-e-2-03")

	adminLogin, _ := provisionBootstrapAdmin(t, server)
	createdUser := createUser(t, server, adminLogin.Session, adminLogin.CSRF, map[string]any{
		"client_txn_id":    "txn-e-2-03-user",
		"auth_kind":        "local",
		"email":            "incident_membership-e-2-03@example.test",
		"display_name":     "IncidentMembership E203",
		"initial_password": "IncidentMembershipE203Pass!",
		"mfa_required":     false,
	})
	userID := createdUser["user_id"].(string)
	userLogin := flowtest.LoginLocalUser(t, server.Client, server.BaseURL, "incident_membership-e-2-03@example.test", "IncidentMembershipE203Pass!", nil)

	incident := createIncident(t, server, adminLogin.Session, adminLogin.CSRF, map[string]any{
		"client_txn_id": "txn-e-2-03-incident",
		"incident_key":  "IR-E203",
		"title":         "Membership Flow",
	})
	incidentID := incident["incident_id"].(string)

	createMembership := doJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-e-2-03-membership",
			"user_id":       userID,
			"role":          "viewer",
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	httptestx.RequireSuccessEnvelope(t, createMembership, http.StatusCreated)

	listMemberships := doJSON(t, server, http.MethodGet, "/api/v1/incidents/"+incidentID+"/memberships", nil, httptestx.WithCookies(adminLogin.Session))
	listBody := httptestx.RequireSuccessEnvelope(t, listMemberships, http.StatusOK)["data"].(map[string]any)
	if len(listBody["memberships"].([]any)) != 2 {
		t.Fatalf("unexpected incident membership list: %#v", listBody)
	}

	nonAdminCreate := doJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-e-2-03-non-admin",
			"email":         "nobody@example.test",
			"role":          "viewer",
		},
		httptestx.WithCookies(userLogin.Session, userLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, userLogin.CSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, nonAdminCreate, http.StatusForbidden, "authorization_denied")
}

func TestMembershipPatchDeleteAndLastAdmin_Process(t *testing.T) {
	server := startServerProcess(t, "incident_membership-e-2-04")

	adminLogin, _ := provisionBootstrapAdmin(t, server)
	adminSession := adminLogin.Session
	adminCSRF := adminLogin.CSRF
	sessionResp := doJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, httptestx.WithCookies(adminSession))
	sessionBody := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	adminUserID := sessionBody["user_id"].(string)
	createdUser := createUser(t, server, adminSession, adminCSRF, map[string]any{
		"client_txn_id":    "txn-e-2-04-user",
		"auth_kind":        "local",
		"email":            "incident_membership-e-2-04@example.test",
		"display_name":     "IncidentMembership E204",
		"initial_password": "IncidentMembershipE204Pass!",
	})
	userID := createdUser["user_id"].(string)
	incident := createIncident(t, server, adminSession, adminCSRF, map[string]any{
		"client_txn_id": "txn-e-2-04-incident",
		"incident_key":  "IR-E204",
		"title":         "Membership Mutations",
	})
	incidentID := incident["incident_id"].(string)

	createMembership := doJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-e-2-04-membership",
			"user_id":       userID,
			"role":          "viewer",
		},
		httptestx.WithCookies(adminSession, adminCSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	createMembershipBody := httptestx.RequireSuccessEnvelope(t, createMembership, http.StatusCreated)["data"].(map[string]any)

	patchMembership := doJSON(
		t,
		server,
		http.MethodPatch,
		"/api/v1/incidents/"+incidentID+"/memberships/"+userID,
		map[string]any{
			"base_membership_version": createMembershipBody["membership_version"],
			"role":                    "reviewer",
		},
		httptestx.WithCookies(adminSession, adminCSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	patchMembershipBody := httptestx.RequireSuccessEnvelope(t, patchMembership, http.StatusOK)["data"].(map[string]any)
	if patchMembershipBody["role"] != "reviewer" || patchMembershipBody["membership_version"] != float64(2) {
		t.Fatalf("unexpected membership patch payload: %#v", patchMembershipBody)
	}

	deleteMembership := doJSON(
		t,
		server,
		http.MethodDelete,
		"/api/v1/incidents/"+incidentID+"/memberships/"+userID,
		map[string]any{
			"base_membership_version": patchMembershipBody["membership_version"],
		},
		httptestx.WithCookies(adminSession, adminCSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	httptestx.RequireNoContent(t, deleteMembership)

	lastAdminGuard := doJSON(
		t,
		server,
		http.MethodDelete,
		"/api/v1/incidents/"+incidentID+"/memberships/"+adminUserID,
		map[string]any{
			"base_membership_version": 1,
		},
		httptestx.WithCookies(adminSession, adminCSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, lastAdminGuard, http.StatusConflict, "last_incident_admin")
}

func TestExtensionDiscoveryAndReservedRoutes_Process(t *testing.T) {
	server := startServerProcess(t, "incident_membership-e-2-05")

	adminLogin, _ := provisionBootstrapAdmin(t, server)
	createdUser := createUser(t, server, adminLogin.Session, adminLogin.CSRF, map[string]any{
		"client_txn_id":    "txn-e-2-05-user",
		"auth_kind":        "local",
		"email":            "incident_membership-e-2-05@example.test",
		"display_name":     "IncidentMembership E205",
		"initial_password": "IncidentMembershipE205Pass!",
		"mfa_required":     false,
	})
	userID := createdUser["user_id"].(string)
	userLogin := flowtest.LoginLocalUser(t, server.Client, server.BaseURL, "incident_membership-e-2-05@example.test", "IncidentMembershipE205Pass!", nil)

	extensions := doJSON(t, server, http.MethodGet, "/api/v1/extensions", nil, httptestx.WithCookies(userLogin.Session))
	extensionsBody := httptestx.RequireSuccessEnvelope(t, extensions, http.StatusOK)["data"].(map[string]any)
	if len(extensionsBody["extensions"].([]any)) != 6 {
		t.Fatalf("unexpected extensions payload: %#v", extensionsBody)
	}
	requireSmokeExtensionClaim(t, extensionsBody, "network_flow_activity", false)

	rootReserved := doJSON(t, server, http.MethodGet, "/api/v1/auth/providers", nil, httptestx.WithCookies(userLogin.Session))
	rootReservedBody := httptestx.RequireErrorEnvelope(t, rootReserved, http.StatusNotFound, "extension_profile_not_claimed")
	rootDetails := rootReservedBody["error"].(map[string]any)["details"].(map[string]any)
	if rootDetails["profile_id"] != "enterprise_authentication" {
		t.Fatalf("unexpected enterprise-authentication reserved dispatch details: %#v", rootDetails)
	}

	nestedReserved := doJSON(t, server, http.MethodGet, "/api/v1/users/"+userID+"/auth-bindings", nil, httptestx.WithCookies(userLogin.Session))
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
	server := startServerProcess(t, "incident_membership-e-2-06")

	adminLogin, _ := provisionBootstrapAdmin(t, server)
	incident := createIncident(t, server, adminLogin.Session, adminLogin.CSRF, map[string]any{
		"client_txn_id": "txn-e-2-06-incident",
		"incident_key":  "IR-E206",
		"title":         "Boundary Incident",
	})
	incidentID := incident["incident_id"].(string)

	deploymentOnly := createUser(t, server, adminLogin.Session, adminLogin.CSRF, map[string]any{
		"client_txn_id":       "txn-e-2-06-user",
		"auth_kind":           "local",
		"email":               "incident_membership-e-2-06@example.test",
		"display_name":        "IncidentMembership E206",
		"initial_password":    "IncidentMembershipE206Pass!",
		"mfa_required":        false,
		"is_deployment_admin": true,
	})
	_ = deploymentOnly
	deploymentLogin := flowtest.LoginLocalUser(t, server.Client, server.BaseURL, "incident_membership-e-2-06@example.test", "IncidentMembershipE206Pass!", nil)

	getIncident := doJSON(t, server, http.MethodGet, "/api/v1/incidents/"+incidentID, nil, httptestx.WithCookies(deploymentLogin.Session))
	httptestx.RequireErrorEnvelope(t, getIncident, http.StatusNotFound, "incident_not_found")

	listIncidentMemberships := doJSON(t, server, http.MethodGet, "/api/v1/incidents/"+incidentID+"/memberships", nil, httptestx.WithCookies(deploymentLogin.Session))
	httptestx.RequireErrorEnvelope(t, listIncidentMemberships, http.StatusNotFound, "incident_not_found")
}

func createIncident(t testing.TB, server *processtest.Server, adminSession *http.Cookie, adminCSRF *http.Cookie, body map[string]any) map[string]any {
	t.Helper()

	resp := doJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/incidents",
		body,
		httptestx.WithCookies(adminSession, adminCSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}
