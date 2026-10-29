package incidents_test

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/dbassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestIncidentCreationBootstrapsCreatorAndWorkbookPreferencesHTTPConformance(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "incident_membership-u-2-02", TestRouteMode: httptestx.TestRouteModeDisabled})

	adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	createResp := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-u-2-02-create",
			"incident_key":  "IR-U202",
			"title":         "Bootstrap Incident",
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	createBody := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)
	incidentID := createBody["data"].(map[string]any)["incident_id"].(string)

	sessionResp := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/auth/session", nil, httptestx.WithCookies(adminLogin.Session))
	sessionBody := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)
	memberships := sessionBody["data"].(map[string]any)["memberships"].([]any)
	if len(memberships) != 1 {
		t.Fatalf("expected one bootstrap membership summary, got %#v", memberships)
	}
	sessionMembership := memberships[0].(map[string]any)
	if sessionMembership["incident_id"] != incidentID || sessionMembership["role"] != "admin" {
		t.Fatalf("unexpected bootstrap session membership: %#v", sessionMembership)
	}

	defaultPrefsResp := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/default", nil, httptestx.WithCookies(adminLogin.Session))
	defaultPrefs := httptestx.RequireSuccessEnvelope(t, defaultPrefsResp, http.StatusOK)["data"].(map[string]any)
	if defaultPrefs["incident_id"] != incidentID || defaultPrefs["default_sheet_ref"] != nil || defaultPrefs["updated_by_user_id"] != adminID.String() {
		t.Fatalf("unexpected default workbook preferences payload: %#v", defaultPrefs)
	}

	userPrefsResp := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/me", nil, httptestx.WithCookies(adminLogin.Session))
	userPrefs := httptestx.RequireSuccessEnvelope(t, userPrefsResp, http.StatusOK)["data"].(map[string]any)
	if userPrefs["incident_id"] != incidentID || userPrefs["user_id"] != adminID.String() || userPrefs["home_sheet_ref"] != nil {
		t.Fatalf("unexpected user workbook preferences payload: %#v", userPrefs)
	}

	viewerID := flowtest.SeedLocalUserFlags(t, harness.DB, "incident_membership-u202-viewer@example.test", "IncidentMembership U202 Viewer", "IncidentMembershipU202Viewer1!", false, false, true)
	viewerLogin := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, "incident_membership-u202-viewer@example.test", "IncidentMembershipU202Viewer1!", nil)
	reviewerID := flowtest.SeedLocalUserFlags(t, harness.DB, "incident_membership-u202-reviewer@example.test", "IncidentMembership U202 Reviewer", "IncidentMembershipU202Reviewer1!", false, false, true)
	reviewerLogin := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, "incident_membership-u202-reviewer@example.test", "IncidentMembershipU202Reviewer1!", nil)
	flowtest.SeedLocalUserFlags(t, harness.DB, "incident_membership-u202-nonmember@example.test", "IncidentMembership U202 Nonmember", "IncidentMembershipU202Nonmember1!", false, false, true)
	nonMemberLogin := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, "incident_membership-u202-nonmember@example.test", "IncidentMembershipU202Nonmember1!", nil)
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

	t.Run("registered mutation failure precedence", func(t *testing.T) {
		beforeReceipts := dbassert.CountSQL(t, harness.DB, `SELECT count(*) FROM route_idempotency`)
		beforeMemberships := dbassert.CountSQL(t, harness.DB, `SELECT count(*) FROM incident_memberships`)
		root := "/api/v1/incidents/" + incidentID
		operations := []struct{ name, method, path, bodyCode string }{
			{"metadata", "PATCH", root, "invalid_incident_patch"},
			{"close", "POST", root + "/close", "invalid_incident_lifecycle_request"},
			{"reopen", "POST", root + "/reopen", "invalid_incident_lifecycle_request"},
			{"create membership", "POST", root + "/memberships", "invalid_mutation_payload"},
			{"patch membership", "PATCH", root + "/memberships/" + viewerID, "invalid_mutation_payload"},
			{"delete membership", "DELETE", root + "/memberships/" + viewerID, "invalid_mutation_payload"},
		}
		for _, operation := range operations {
			t.Run(operation.name, func(t *testing.T) {
				cases := []struct {
					name    string
					session *http.Cookie
					csrf    string
					status  int
					code    string
				}{
					{"unauthenticated", nil, "", 401, "session_required"},
					{"missing csrf", adminLogin.Session, "", 403, "csrf_verification_failed"},
					{"hidden before body", nonMemberLogin.Session, nonMemberLogin.CSRF.Value, 404, "incident_not_found"},
					{"role before body", viewerLogin.Session, viewerLogin.CSRF.Value, 403, "authorization_denied"},
					{"authorized body", adminLogin.Session, adminLogin.CSRF.Value, 400, operation.bodyCode},
				}
				for _, test := range cases {
					t.Run(test.name, func(t *testing.T) {
						var options []func(*http.Request)
						if test.session != nil {
							options = append(options, httptestx.WithCookies(test.session))
						}
						if test.csrf != "" {
							options = append(options, httptestx.WithCookies(&http.Cookie{Name: authn.CSRFCookieName, Value: test.csrf}), httptestx.WithHeader(authn.CSRFHeaderName, test.csrf))
						}
						response := httptestx.DoJSON(t, http.DefaultClient, operation.method, harness.Server.HTTP.URL+operation.path, map[string]any{"unexpected": true}, options...)
						httptestx.RequireErrorEnvelope(t, response, test.status, test.code)
					})
				}
			})
		}
		for _, suffix := range []string{"/memberships?limit=bad", "/membership-audit-events?limit=bad"} {
			response := httptestx.DoJSON(t, http.DefaultClient, "GET", harness.Server.HTTP.URL+root+suffix, nil, httptestx.WithCookies(nonMemberLogin.Session))
			httptestx.RequireErrorEnvelope(t, response, 404, "incident_not_found")
		}
		response := httptestx.DoJSON(t, http.DefaultClient, "GET", harness.Server.HTTP.URL+root+"/membership-audit-events?limit=bad", nil, httptestx.WithCookies(viewerLogin.Session))
		httptestx.RequireErrorEnvelope(t, response, 403, "authorization_denied")
		if got := dbassert.CountSQL(t, harness.DB, `SELECT count(*) FROM route_idempotency`); got != beforeReceipts {
			t.Fatalf("denials changed receipts: %d -> %d", beforeReceipts, got)
		}
		if got := dbassert.CountSQL(t, harness.DB, `SELECT count(*) FROM incident_memberships`); got != beforeMemberships {
			t.Fatalf("denials changed memberships: %d -> %d", beforeMemberships, got)
		}
	})

	viewerPut := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPut,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/me",
		map[string]any{
			"home_sheet_ref": map[string]any{
				"kind": "view_schema",
				"id":   timeline.TimelineViewSchemaID,
			},
		},
		httptestx.WithCookies(viewerLogin.Session, viewerLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, viewerLogin.CSRF.Value),
	)
	viewerPrefs := httptestx.RequireSuccessEnvelope(t, viewerPut, http.StatusOK)["data"].(map[string]any)
	if viewerPrefs["incident_id"] != incidentID || viewerPrefs["user_id"] != viewerID {
		t.Fatalf("unexpected viewer workbook preferences payload: %#v", viewerPrefs)
	}
	requireWorkbookSheetRef(t, viewerPrefs["home_sheet_ref"], timeline.TimelineViewSchemaID)

	viewerGet := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/me", nil, httptestx.WithCookies(viewerLogin.Session))
	viewerGetPrefs := httptestx.RequireSuccessEnvelope(t, viewerGet, http.StatusOK)["data"].(map[string]any)
	requireWorkbookSheetRef(t, viewerGetPrefs["home_sheet_ref"], timeline.TimelineViewSchemaID)

	adminUserPrefsResp := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/me", nil, httptestx.WithCookies(adminLogin.Session))
	adminUserPrefs := httptestx.RequireSuccessEnvelope(t, adminUserPrefsResp, http.StatusOK)["data"].(map[string]any)
	if adminUserPrefs["user_id"] != adminID.String() || adminUserPrefs["home_sheet_ref"] != nil {
		t.Fatalf("user workbook preferences PUT must only update the caller row: %#v", adminUserPrefs)
	}

	defaultBody := map[string]any{
		"default_sheet_ref": map[string]any{
			"kind": "view_schema",
			"id":   timeline.TimelineViewSchemaID,
		},
	}
	viewerDefault := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPut,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/default",
		defaultBody,
		httptestx.WithCookies(viewerLogin.Session, viewerLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, viewerLogin.CSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, viewerDefault, http.StatusForbidden, "authorization_denied")
	reviewerDefault := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPut,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/default",
		defaultBody,
		httptestx.WithCookies(reviewerLogin.Session, reviewerLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, reviewerLogin.CSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, reviewerDefault, http.StatusForbidden, "authorization_denied")

	adminDefault := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPut,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/default",
		defaultBody,
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	adminDefaultPrefs := httptestx.RequireSuccessEnvelope(t, adminDefault, http.StatusOK)["data"].(map[string]any)
	if adminDefaultPrefs["incident_id"] != incidentID || adminDefaultPrefs["updated_by_user_id"] != adminID.String() {
		t.Fatalf("unexpected admin default workbook preferences payload: %#v", adminDefaultPrefs)
	}
	requireWorkbookSheetRef(t, adminDefaultPrefs["default_sheet_ref"], timeline.TimelineViewSchemaID)

	defaultGet := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/default", nil, httptestx.WithCookies(adminLogin.Session))
	defaultGetPrefs := httptestx.RequireSuccessEnvelope(t, defaultGet, http.StatusOK)["data"].(map[string]any)
	requireWorkbookSheetRef(t, defaultGetPrefs["default_sheet_ref"], timeline.TimelineViewSchemaID)

	missingCSRF := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPut,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/me",
		map[string]any{
			"home_sheet_ref": map[string]any{
				"kind": "view_schema",
				"id":   "cartulary.view.hosts.v1",
			},
		},
		httptestx.WithCookies(viewerLogin.Session),
	)
	httptestx.RequireErrorEnvelope(t, missingCSRF, http.StatusForbidden, "csrf_verification_failed")
	viewerAfterCSRF := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/me", nil, httptestx.WithCookies(viewerLogin.Session))
	viewerAfterCSRFPrefs := httptestx.RequireSuccessEnvelope(t, viewerAfterCSRF, http.StatusOK)["data"].(map[string]any)
	requireWorkbookSheetRef(t, viewerAfterCSRFPrefs["home_sheet_ref"], timeline.TimelineViewSchemaID)

	nonMemberPut := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPut,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/me",
		map[string]any{"home_sheet_ref": nil},
		httptestx.WithCookies(nonMemberLogin.Session, nonMemberLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, nonMemberLogin.CSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, nonMemberPut, http.StatusNotFound, "incident_not_found")

	invalidDefault := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPut,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/default",
		map[string]any{"default_sheet_ref": nil, "unexpected": true},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	invalidDefaultBody := httptestx.RequireErrorEnvelope(t, invalidDefault, http.StatusBadRequest, "invalid_mutation_payload")
	requireErrorDetails(t, invalidDefaultBody, "unexpected", "unknown_field")
}

func TestIncidentCreateReturnsStableLocationHeaderHTTPConformance(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "incident_membership-u-2-03", TestRouteMode: httptestx.TestRouteModeDisabled})

	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	createResp := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-u-2-03-create",
			"incident_key":  "IR-U203",
			"title":         "Location Incident",
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	body := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)
	incidentID := body["data"].(map[string]any)["incident_id"].(string)
	if got := createResp.Header.Get("Location"); got != "/api/v1/incidents/"+incidentID {
		t.Fatalf("unexpected incident Location header: got %q", got)
	}
}

func TestIncidentCreateIdempotencyUsesActorAndNormalizedReplayHTTPConformance(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "incident_membership-u-2-04", TestRouteMode: httptestx.TestRouteModeDisabled})

	firstActor, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	firstCreate := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-u-2-04",
			"incident_key":  "  IR-U204  ",
			"title":         "  Replay Incident  ",
		},
		httptestx.WithCookies(firstActor.Session, firstActor.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, firstActor.CSRF.Value),
	)
	firstData := httptestx.RequireSuccessEnvelope(t, firstCreate, http.StatusCreated)["data"].(map[string]any)
	expectedLocation := "/api/v1/incidents/" + firstData["incident_id"].(string)
	if got := firstCreate.Header.Get("Location"); got != expectedLocation {
		t.Fatalf("unexpected incident create Location header: got %q, want %q", got, expectedLocation)
	}

	replay := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-u-2-04",
			"incident_key":  "IR-U204",
			"title":         "Replay Incident",
		},
		httptestx.WithCookies(firstActor.Session, firstActor.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, firstActor.CSRF.Value),
	)
	replayData := httptestx.RequireSuccessEnvelope(t, replay, http.StatusOK)["data"].(map[string]any)
	if got := replay.Header.Get("Location"); got != expectedLocation {
		t.Fatalf("unexpected incident replay Location header: got %q, want %q", got, expectedLocation)
	}
	if !reflect.DeepEqual(firstData, replayData) {
		t.Fatalf("expected replayed incident payload to match original result: first=%#v replay=%#v", firstData, replayData)
	}

	flowtest.SeedLocalUserFlags(t, harness.DB, "incident_membership-u204@example.test", "IncidentMembership U204", "IncidentMembershipU204Pass!", false, false, true)
	secondLogin := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, "incident_membership-u204@example.test", "IncidentMembershipU204Pass!", nil)
	secondCreate := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-u-2-04",
			"incident_key":  "IR-U204-SECOND",
			"title":         "Second Actor Incident",
		},
		httptestx.WithCookies(secondLogin.Session, secondLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, secondLogin.CSRF.Value),
	)
	secondData := httptestx.RequireSuccessEnvelope(t, secondCreate, http.StatusCreated)["data"].(map[string]any)
	if secondData["incident_id"] == firstData["incident_id"] {
		t.Fatalf("expected actor-scoped idempotency to allow a distinct create, got %#v", secondData)
	}

	contracttest.RequireErrorContract(t, "client_txn_conflict", http.StatusConflict)
	divergent := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-u-2-04",
			"incident_key":  "IR-U204",
			"title":         "Different title",
		},
		httptestx.WithCookies(firstActor.Session, firstActor.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, firstActor.CSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, divergent, http.StatusConflict, "client_txn_conflict")
}

func TestMembershipCreateRequiresOneSelectorClosedRolesAndNoInvitationFieldsHTTPConformance(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "incident_membership-u-2-06", TestRouteMode: httptestx.TestRouteModeDisabled})

	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-u-2-06-incident",
		"incident_key":  "IR-U206",
		"title":         "Membership Targets",
	})
	incidentID := incident["incident_id"].(string)

	missingSelector := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-u-2-06-missing-target",
			"role":          "viewer",
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	missingSelectorBody := httptestx.RequireErrorEnvelope(t, missingSelector, http.StatusBadRequest, "invalid_mutation_payload")
	requireErrorDetails(t, missingSelectorBody, "user_id", "exactly_one_target_selector")

	dualSelector := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-u-2-06-dual-target",
			"user_id":       "00000000-0000-0000-0000-000000000602",
			"email":         "dual@example.test",
			"role":          "viewer",
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	dualSelectorBody := httptestx.RequireErrorEnvelope(t, dualSelector, http.StatusBadRequest, "invalid_mutation_payload")
	requireErrorDetails(t, dualSelectorBody, "user_id", "exactly_one_target_selector")

	invalidRole := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-u-2-06-invalid-role",
			"email":         "solo@example.test",
			"role":          "owner",
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	invalidRoleBody := httptestx.RequireErrorEnvelope(t, invalidRole, http.StatusBadRequest, "invalid_mutation_payload")
	requireErrorDetails(t, invalidRoleBody, "role", "invalid_role")

	unknownInvitation := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id":    "txn-u-2-06-no-invite",
			"email":            "solo@example.test",
			"role":             "viewer",
			"invitation_email": "new@example.test",
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	unknownInvitationBody := httptestx.RequireErrorEnvelope(t, unknownInvitation, http.StatusBadRequest, "invalid_mutation_payload")
	requireErrorDetails(t, unknownInvitationBody, "invitation_email", "unknown_field")

	contracttest.RequireErrorContract(t, "user_not_found", http.StatusNotFound)
	missingUser := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-u-2-06-missing-user",
			"email":         "missing@example.test",
			"role":          "viewer",
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, missingUser, http.StatusNotFound, "user_not_found")
	if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM users WHERE email = $1`, "missing@example.test"); got != 0 {
		t.Fatalf("membership create must not auto-create a missing user, got %d rows", got)
	}

	flowtest.SeedActor(t, harness.DB, flowtest.ActorAttributes{Email: "inactive@example.test", DisplayName: "Inactive User", Disabled: true})
	contracttest.RequireErrorContract(t, "user_inactive", http.StatusConflict)
	inactiveUser := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-u-2-06-inactive-user",
			"email":         "inactive@example.test",
			"role":          "viewer",
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, inactiveUser, http.StatusConflict, "user_inactive")
}

func TestMembershipPatchAndDeleteEnforceBaseVersionAndLastAdminGuardHTTPConformance(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "incident_membership-u-2-07", TestRouteMode: httptestx.TestRouteModeDisabled})

	adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-u-2-07-incident",
		"incident_key":  "IR-U207",
		"title":         "Membership Versioning",
	})
	incidentID := incident["incident_id"].(string)
	targetUserID := flowtest.SeedActor(t, harness.DB, flowtest.ActorAttributes{Email: "incident_membership-u207@example.test", DisplayName: "IncidentMembership U207"}).ID.String()

	// Membership administration remains available after incident closure.
	closeResp := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/close",
		map[string]any{"client_txn_id": "txn-u-2-07-close", "base_incident_version": incident["incident_version"], "reason": "Membership review after closure"},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value))
	httptestx.RequireSuccessEnvelope(t, closeResp, http.StatusOK)

	createResp := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-u-2-07-membership",
			"user_id":       targetUserID,
			"role":          "viewer",
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	createdMembership := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)

	patchMembership := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetUserID,
		map[string]any{
			"base_membership_version": createdMembership["membership_version"],
			"role":                    "reviewer",
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	patchedMembership := httptestx.RequireSuccessEnvelope(t, patchMembership, http.StatusOK)["data"].(map[string]any)

	contracttest.RequireErrorContract(t, "membership_version_conflict", http.StatusConflict)
	stalePatch := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetUserID,
		map[string]any{
			"base_membership_version": createdMembership["membership_version"],
			"role":                    "admin",
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, stalePatch, http.StatusConflict, "membership_version_conflict")

	clientTxnPatch := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetUserID,
		map[string]any{
			"client_txn_id":           "forbidden",
			"base_membership_version": patchedMembership["membership_version"],
			"role":                    "admin",
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	patchError := httptestx.RequireErrorEnvelope(t, clientTxnPatch, http.StatusBadRequest, "invalid_mutation_payload")
	requireErrorDetails(t, patchError, "client_txn_id", "unknown_field")

	staleDelete := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodDelete,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetUserID,
		map[string]any{
			"base_membership_version": createdMembership["membership_version"],
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, staleDelete, http.StatusConflict, "membership_version_conflict")

	clientTxnDelete := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodDelete,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetUserID,
		map[string]any{
			"client_txn_id":           "forbidden",
			"base_membership_version": patchedMembership["membership_version"],
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	deleteError := httptestx.RequireErrorEnvelope(t, clientTxnDelete, http.StatusBadRequest, "invalid_mutation_payload")
	requireErrorDetails(t, deleteError, "client_txn_id", "unknown_field")

	removed := httptestx.DoJSON(t, http.DefaultClient, http.MethodDelete, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetUserID,
		map[string]any{"base_membership_version": patchedMembership["membership_version"]},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value))
	if removed.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d, want 204", removed.StatusCode)
	}
	_ = removed.Body.Close()
	missing := httptestx.DoJSON(t, http.DefaultClient, http.MethodDelete, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetUserID,
		map[string]any{"base_membership_version": patchedMembership["membership_version"]},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value))
	httptestx.RequireErrorEnvelope(t, missing, http.StatusNotFound, "membership_not_found")

	contracttest.RequireErrorContract(t, "last_incident_admin", http.StatusConflict)
	lastAdmin := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodDelete,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+adminID.String(),
		map[string]any{
			"base_membership_version": 1,
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, lastAdmin, http.StatusConflict, "last_incident_admin")
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
