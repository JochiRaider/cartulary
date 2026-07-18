package incidents_test

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	hostroutetest "github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/testsupport/routetest"
	extensionroutetest "github.com/JochiRaider/cartulary/internal/modules/extensions/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	indicatorroutetest "github.com/JochiRaider/cartulary/internal/modules/indicators/testsupport/routetest"
	recordroutetest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	timelineroutetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/routetest"
	workbookroutetest "github.com/JochiRaider/cartulary/internal/modules/workbook/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
	"github.com/JochiRaider/cartulary/internal/testutil/dbassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/routeinventory"
)

func TestIncidentCreateBootstrapsCreatorAndWorkbookPreferencesHTTPConformance(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "incident_membership-u-2-02")

	adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	createResp := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-u-2-02-create",
			"incident_key":  "IR-U202",
			"title":         "Bootstrap Incident",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	createBody := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)
	incidentID := createBody["data"].(map[string]any)["incident_id"].(string)

	sessionResp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/auth/session", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	sessionBody := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)
	memberships := sessionBody["data"].(map[string]any)["memberships"].([]any)
	if len(memberships) != 1 {
		t.Fatalf("expected one bootstrap membership summary, got %#v", memberships)
	}
	sessionMembership := memberships[0].(map[string]any)
	if sessionMembership["incident_id"] != incidentID || sessionMembership["role"] != "admin" {
		t.Fatalf("unexpected bootstrap session membership: %#v", sessionMembership)
	}

	defaultPrefsResp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/default", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	defaultPrefs := httptestx.RequireSuccessEnvelope(t, defaultPrefsResp, http.StatusOK)["data"].(map[string]any)
	if defaultPrefs["incident_id"] != incidentID || defaultPrefs["default_sheet_ref"] != nil || defaultPrefs["updated_by_user_id"] != adminID {
		t.Fatalf("unexpected default workbook preferences payload: %#v", defaultPrefs)
	}

	userPrefsResp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/me", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	userPrefs := httptestx.RequireSuccessEnvelope(t, userPrefsResp, http.StatusOK)["data"].(map[string]any)
	if userPrefs["incident_id"] != incidentID || userPrefs["user_id"] != adminID || userPrefs["home_sheet_ref"] != nil {
		t.Fatalf("unexpected user workbook preferences payload: %#v", userPrefs)
	}

	viewerID := flowtest.SeedLocalUserFlags(t, harness.DB, "incident_membership-u202-viewer@example.test", "IncidentMembership U202 Viewer", "IncidentMembershipU202Viewer1!", false, false, true)
	viewerSession, viewerCSRF := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, "incident_membership-u202-viewer@example.test", "IncidentMembershipU202Viewer1!", nil)
	reviewerID := flowtest.SeedLocalUserFlags(t, harness.DB, "incident_membership-u202-reviewer@example.test", "IncidentMembership U202 Reviewer", "IncidentMembershipU202Reviewer1!", false, false, true)
	reviewerSession, reviewerCSRF := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, "incident_membership-u202-reviewer@example.test", "IncidentMembershipU202Reviewer1!", nil)
	flowtest.SeedLocalUserFlags(t, harness.DB, "incident_membership-u202-nonmember@example.test", "IncidentMembership U202 Nonmember", "IncidentMembershipU202Nonmember1!", false, false, true)
	nonMemberSession, nonMemberCSRF := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, "incident_membership-u202-nonmember@example.test", "IncidentMembershipU202Nonmember1!", nil)
	scenariotest.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{
		"client_txn_id": "txn-u-2-02-viewer",
		"user_id":       viewerID,
		"role":          "viewer",
	})
	scenariotest.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{
		"client_txn_id": "txn-u-2-02-reviewer",
		"user_id":       reviewerID,
		"role":          "reviewer",
	})

	viewerPut := httptestx.DoJSON(
		t,
		http.MethodPut,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/me",
		map[string]any{
			"home_sheet_ref": map[string]any{
				"kind": "view_schema",
				"id":   timeline.TimelineViewSchemaID,
			},
		},
		httptestx.WithCookies(viewerSession, viewerCSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, viewerCSRF.Value),
	)
	viewerPrefs := httptestx.RequireSuccessEnvelope(t, viewerPut, http.StatusOK)["data"].(map[string]any)
	if viewerPrefs["incident_id"] != incidentID || viewerPrefs["user_id"] != viewerID {
		t.Fatalf("unexpected viewer workbook preferences payload: %#v", viewerPrefs)
	}
	requireWorkbookSheetRef(t, viewerPrefs["home_sheet_ref"], timeline.TimelineViewSchemaID)

	viewerGet := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/me", nil, httptestx.WithCookies(viewerSession))
	viewerGetPrefs := httptestx.RequireSuccessEnvelope(t, viewerGet, http.StatusOK)["data"].(map[string]any)
	requireWorkbookSheetRef(t, viewerGetPrefs["home_sheet_ref"], timeline.TimelineViewSchemaID)

	adminUserPrefsResp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/me", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	adminUserPrefs := httptestx.RequireSuccessEnvelope(t, adminUserPrefsResp, http.StatusOK)["data"].(map[string]any)
	if adminUserPrefs["user_id"] != adminID || adminUserPrefs["home_sheet_ref"] != nil {
		t.Fatalf("user workbook preferences PUT must only update the caller row: %#v", adminUserPrefs)
	}

	defaultBody := map[string]any{
		"default_sheet_ref": map[string]any{
			"kind": "view_schema",
			"id":   timeline.TimelineViewSchemaID,
		},
	}
	viewerDefault := httptestx.DoJSON(
		t,
		http.MethodPut,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/default",
		defaultBody,
		httptestx.WithCookies(viewerSession, viewerCSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, viewerCSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, viewerDefault, http.StatusForbidden, "authorization_denied")
	reviewerDefault := httptestx.DoJSON(
		t,
		http.MethodPut,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/default",
		defaultBody,
		httptestx.WithCookies(reviewerSession, reviewerCSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, reviewerCSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, reviewerDefault, http.StatusForbidden, "authorization_denied")

	adminDefault := httptestx.DoJSON(
		t,
		http.MethodPut,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/default",
		defaultBody,
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	adminDefaultPrefs := httptestx.RequireSuccessEnvelope(t, adminDefault, http.StatusOK)["data"].(map[string]any)
	if adminDefaultPrefs["incident_id"] != incidentID || adminDefaultPrefs["updated_by_user_id"] != adminID {
		t.Fatalf("unexpected admin default workbook preferences payload: %#v", adminDefaultPrefs)
	}
	requireWorkbookSheetRef(t, adminDefaultPrefs["default_sheet_ref"], timeline.TimelineViewSchemaID)

	defaultGet := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/default", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	defaultGetPrefs := httptestx.RequireSuccessEnvelope(t, defaultGet, http.StatusOK)["data"].(map[string]any)
	requireWorkbookSheetRef(t, defaultGetPrefs["default_sheet_ref"], timeline.TimelineViewSchemaID)

	missingCSRF := httptestx.DoJSON(
		t,
		http.MethodPut,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/me",
		map[string]any{
			"home_sheet_ref": map[string]any{
				"kind": "view_schema",
				"id":   "cartulary.view.hosts.v1",
			},
		},
		httptestx.WithCookies(viewerSession),
	)
	httptestx.RequireErrorEnvelope(t, missingCSRF, http.StatusForbidden, "csrf_verification_failed")
	viewerAfterCSRF := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/me", nil, httptestx.WithCookies(viewerSession))
	viewerAfterCSRFPrefs := httptestx.RequireSuccessEnvelope(t, viewerAfterCSRF, http.StatusOK)["data"].(map[string]any)
	requireWorkbookSheetRef(t, viewerAfterCSRFPrefs["home_sheet_ref"], timeline.TimelineViewSchemaID)

	nonMemberPut := httptestx.DoJSON(
		t,
		http.MethodPut,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/me",
		map[string]any{"home_sheet_ref": nil},
		httptestx.WithCookies(nonMemberSession, nonMemberCSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, nonMemberCSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, nonMemberPut, http.StatusNotFound, "incident_not_found")

	invalidDefault := httptestx.DoJSON(
		t,
		http.MethodPut,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/default",
		map[string]any{"default_sheet_ref": nil, "unexpected": true},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	invalidDefaultBody := httptestx.RequireErrorEnvelope(t, invalidDefault, http.StatusBadRequest, "invalid_mutation_payload")
	requireErrorDetails(t, invalidDefaultBody, "unexpected", "unknown_field")
}

func TestPublicRouteInventoryIncidentCoreEnvelopes(t *testing.T) {
	requirePublicRouteInventoryEnvelopes(t, routetest.PublicIncidentCore())
}

func TestPublicRouteInventoryMembershipAdminEnvelopes(t *testing.T) {
	requirePublicRouteInventoryEnvelopes(t, routetest.PublicMembershipAdmin())
}

func TestPublicRouteInventoryWorkbookPreferencesEnvelopes(t *testing.T) {
	requirePublicRouteInventoryEnvelopes(t, workbookroutetest.PublicPreferences())
}

func TestPublicRouteInventoryExtensionDiscoveryEnvelopes(t *testing.T) {
	requirePublicRouteInventoryEnvelopes(t, extensionroutetest.PublicDiscovery())
}

func requirePublicRouteInventoryEnvelopes(t *testing.T, routes []routeinventory.Entry) {
	t.Helper()

	for _, route := range routes {
		route := route
		t.Run(route.Name, func(t *testing.T) {
			fixtureCtx := newRouteFixture(t, "incident_membership-public-"+route.Name)
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

func TestIncidentCreateReturnsStableLocationHeaderHTTPConformance(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "incident_membership-u-2-03")

	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	createResp := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-u-2-03-create",
			"incident_key":  "IR-U203",
			"title":         "Location Incident",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	body := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)
	incidentID := body["data"].(map[string]any)["incident_id"].(string)
	if got := createResp.Header.Get("Location"); got != "/api/v1/incidents/"+incidentID {
		t.Fatalf("unexpected incident Location header: got %q", got)
	}
}

func TestIncidentCreateIdempotencyUsesActorAndNormalizedReplayHTTPConformance(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "incident_membership-u-2-04")

	firstActor, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	firstCreate := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-u-2-04",
			"incident_key":  "  IR-U204  ",
			"title":         "  Replay Incident  ",
		},
		httptestx.WithCookies(firstActor.SessionCookie, firstActor.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, firstActor.CSRFCookie.Value),
	)
	firstData := httptestx.RequireSuccessEnvelope(t, firstCreate, http.StatusCreated)["data"].(map[string]any)

	replay := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-u-2-04",
			"incident_key":  "IR-U204",
			"title":         "Replay Incident",
		},
		httptestx.WithCookies(firstActor.SessionCookie, firstActor.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, firstActor.CSRFCookie.Value),
	)
	replayData := httptestx.RequireSuccessEnvelope(t, replay, http.StatusOK)["data"].(map[string]any)
	if !reflect.DeepEqual(firstData, replayData) {
		t.Fatalf("expected replayed incident payload to match original result: first=%#v replay=%#v", firstData, replayData)
	}

	flowtest.SeedLocalUserFlags(t, harness.DB, "incident_membership-u204@example.test", "IncidentMembership U204", "IncidentMembershipU204Pass!", false, false, true)
	secondSession, secondCSRF := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, "incident_membership-u204@example.test", "IncidentMembershipU204Pass!", nil)
	secondCreate := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-u-2-04",
			"incident_key":  "IR-U204-SECOND",
			"title":         "Second Actor Incident",
		},
		httptestx.WithCookies(secondSession, secondCSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, secondCSRF.Value),
	)
	secondData := httptestx.RequireSuccessEnvelope(t, secondCreate, http.StatusCreated)["data"].(map[string]any)
	if secondData["incident_id"] == firstData["incident_id"] {
		t.Fatalf("expected actor-scoped idempotency to allow a distinct create, got %#v", secondData)
	}

	contracttest.RequireErrorContract(t, "client_txn_conflict", http.StatusConflict)
	divergent := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-u-2-04",
			"incident_key":  "IR-U204",
			"title":         "Different title",
		},
		httptestx.WithCookies(firstActor.SessionCookie, firstActor.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, firstActor.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, divergent, http.StatusConflict, "client_txn_conflict")
}

func TestMembershipCreateRequiresOneSelectorClosedRolesAndNoInvitationFieldsHTTPConformance(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "incident_membership-u-2-06")

	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-u-2-06-incident",
		"incident_key":  "IR-U206",
		"title":         "Membership Targets",
	})
	incidentID := incident["incident_id"].(string)

	missingSelector := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-u-2-06-missing-target",
			"role":          "viewer",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	missingSelectorBody := httptestx.RequireErrorEnvelope(t, missingSelector, http.StatusBadRequest, "invalid_mutation_payload")
	requireErrorDetails(t, missingSelectorBody, "user_id", "exactly_one_target_selector")

	dualSelector := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-u-2-06-dual-target",
			"user_id":       "00000000-0000-0000-0000-000000000602",
			"email":         "dual@example.test",
			"role":          "viewer",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	dualSelectorBody := httptestx.RequireErrorEnvelope(t, dualSelector, http.StatusBadRequest, "invalid_mutation_payload")
	requireErrorDetails(t, dualSelectorBody, "user_id", "exactly_one_target_selector")

	invalidRole := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-u-2-06-invalid-role",
			"email":         "solo@example.test",
			"role":          "owner",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	invalidRoleBody := httptestx.RequireErrorEnvelope(t, invalidRole, http.StatusBadRequest, "invalid_mutation_payload")
	requireErrorDetails(t, invalidRoleBody, "role", "invalid_role")

	unknownInvitation := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id":    "txn-u-2-06-no-invite",
			"email":            "solo@example.test",
			"role":             "viewer",
			"invitation_email": "new@example.test",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	unknownInvitationBody := httptestx.RequireErrorEnvelope(t, unknownInvitation, http.StatusBadRequest, "invalid_mutation_payload")
	requireErrorDetails(t, unknownInvitationBody, "invitation_email", "unknown_field")

	contracttest.RequireErrorContract(t, "user_not_found", http.StatusNotFound)
	missingUser := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-u-2-06-missing-user",
			"email":         "missing@example.test",
			"role":          "viewer",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, missingUser, http.StatusNotFound, "user_not_found")
	if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM users WHERE email = $1`, "missing@example.test"); got != 0 {
		t.Fatalf("membership create must not auto-create a missing user, got %d rows", got)
	}

	flowtest.SeedLocalUserFlags(t, harness.DB, "inactive@example.test", "Inactive User", "InactiveUser1!", false, false, false)
	contracttest.RequireErrorContract(t, "user_inactive", http.StatusConflict)
	inactiveUser := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-u-2-06-inactive-user",
			"email":         "inactive@example.test",
			"role":          "viewer",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, inactiveUser, http.StatusConflict, "user_inactive")
}

func TestMembershipPatchAndDeleteEnforceBaseVersionAndLastAdminGuardHTTPConformance(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "incident_membership-u-2-07")

	adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-u-2-07-incident",
		"incident_key":  "IR-U207",
		"title":         "Membership Versioning",
	})
	incidentID := incident["incident_id"].(string)
	targetUserID := flowtest.SeedLocalUserFlags(t, harness.DB, "incident_membership-u207@example.test", "IncidentMembership U207", "IncidentMembershipU207Pass!", false, false, true)

	createResp := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-u-2-07-membership",
			"user_id":       targetUserID,
			"role":          "viewer",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	createdMembership := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)

	patchMembership := httptestx.DoJSON(
		t,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetUserID,
		map[string]any{
			"base_membership_version": createdMembership["membership_version"],
			"role":                    "reviewer",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	patchedMembership := httptestx.RequireSuccessEnvelope(t, patchMembership, http.StatusOK)["data"].(map[string]any)

	contracttest.RequireErrorContract(t, "membership_version_conflict", http.StatusConflict)
	stalePatch := httptestx.DoJSON(
		t,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetUserID,
		map[string]any{
			"base_membership_version": createdMembership["membership_version"],
			"role":                    "admin",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, stalePatch, http.StatusConflict, "membership_version_conflict")

	clientTxnPatch := httptestx.DoJSON(
		t,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetUserID,
		map[string]any{
			"client_txn_id":           "forbidden",
			"base_membership_version": patchedMembership["membership_version"],
			"role":                    "admin",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	patchError := httptestx.RequireErrorEnvelope(t, clientTxnPatch, http.StatusBadRequest, "invalid_mutation_payload")
	requireErrorDetails(t, patchError, "client_txn_id", "unknown_field")

	staleDelete := httptestx.DoJSON(
		t,
		http.MethodDelete,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetUserID,
		map[string]any{
			"base_membership_version": createdMembership["membership_version"],
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, staleDelete, http.StatusConflict, "membership_version_conflict")

	clientTxnDelete := httptestx.DoJSON(
		t,
		http.MethodDelete,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetUserID,
		map[string]any{
			"client_txn_id":           "forbidden",
			"base_membership_version": patchedMembership["membership_version"],
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	deleteError := httptestx.RequireErrorEnvelope(t, clientTxnDelete, http.StatusBadRequest, "invalid_mutation_payload")
	requireErrorDetails(t, deleteError, "client_txn_id", "unknown_field")

	contracttest.RequireErrorContract(t, "last_incident_admin", http.StatusConflict)
	lastAdmin := httptestx.DoJSON(
		t,
		http.MethodDelete,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+adminID,
		map[string]any{
			"base_membership_version": 1,
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, lastAdmin, http.StatusConflict, "last_incident_admin")
}

func TestControlBoundaryIncidentCoreDeploymentAdminWithoutMembershipDenied(t *testing.T) {
	requireControlBoundaryInventoryDeploymentAdminWithoutMembershipDenied(t, routetest.ControlIncidentCore())
}

func TestControlBoundaryMembershipAdminDeploymentAdminWithoutMembershipDenied(t *testing.T) {
	requireControlBoundaryInventoryDeploymentAdminWithoutMembershipDenied(t, routetest.ControlMembershipAdmin())
}

func TestControlBoundaryWorkbookPreferencesDeploymentAdminWithoutMembershipDenied(t *testing.T) {
	requireControlBoundaryInventoryDeploymentAdminWithoutMembershipDenied(t, workbookroutetest.ControlPreferences())
}

func TestControlBoundaryWorkbookQueriesDeploymentAdminWithoutMembershipDenied(t *testing.T) {
	routes := append(timelineroutetest.ControlQuery(), hostroutetest.ControlQueries()...)
	routes = append(routes, indicatorroutetest.ControlQuery()...)
	requireControlBoundaryInventoryDeploymentAdminWithoutMembershipDenied(t, routes)
}

func TestControlBoundaryTimelineRecordAndLiveDeploymentAdminWithoutMembershipDenied(t *testing.T) {
	routes := append(timelineroutetest.ControlCreateAndLive(), recordroutetest.ControlMutations()...)
	requireControlBoundaryInventoryDeploymentAdminWithoutMembershipDenied(t, routes)
}

func requireControlBoundaryInventoryDeploymentAdminWithoutMembershipDenied(t *testing.T, routes []routeinventory.Entry) {
	t.Helper()

	for _, route := range routes {
		route := route
		t.Run(route.Name, func(t *testing.T) {
			fixtureCtx := newRouteFixture(t, "incident_membership-control-denied-"+route.Name)
			deploymentEmail := FixtureSlug("incident_membership-control-denied-" + route.Name)
			flowtest.SeedLocalUserFlags(
				t,
				fixtureCtx.harness.DB,
				deploymentEmail+"@example.test",
				"Deployment Only "+deploymentEmail,
				"DeploymentOnly1!",
				false,
				true,
				true,
			)
			deploymentSession, deploymentCSRF := flowtest.LoginLocalUser(
				t,
				fixtureCtx.harness.Server.HTTP.URL,

				deploymentEmail+"@example.test",
				"DeploymentOnly1!", nil)

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

func requireWorkbookSheetRef(t testing.TB, value any, wantID string) {
	t.Helper()

	ref, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected sheet_ref object, got %#v", value)
	}
	if ref["kind"] != "view_schema" || ref["id"] != wantID {
		t.Fatalf("unexpected sheet_ref: got %#v want view_schema %q", ref, wantID)
	}
}
