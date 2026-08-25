package incidents_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/coder/websocket"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	timelineroutetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/incidentwstest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/routeinventory"
)

type controlProgressionStage string

const (
	controlStageNoMembership controlProgressionStage = "no-membership"
	controlStageViewer       controlProgressionStage = "viewer"
	controlStageReviewer     controlProgressionStage = "reviewer"
	controlStageAdmin        controlProgressionStage = "admin"
	controlStageRemoved      controlProgressionStage = "removed"
)

type controlRouteExpectation struct {
	success bool
	status  int
	code    string
}

type RouteFixture struct {
	harness     *appsupport.ServerHarness
	adminLogin  flowtest.LoginResult
	adminID     string
	candidateID string
	memberID    string
	fixture     routeinventory.Fixture
}

func newRouteFixture(t testing.TB, prefix string) *RouteFixture {
	t.Helper()

	slug := FixtureSlug(prefix)
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartDefaultServer(t, slug)
	adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)

	candidateID := flowtest.SeedLocalUserFlags(
		t,
		harness.DB,
		slug+"-candidate@example.test",
		"IncidentMembership Candidate "+slug,
		"IncidentMembershipCandidate1!",
		false,
		false,
		true,
	)
	memberID := flowtest.SeedLocalUserFlags(
		t,
		harness.DB,
		slug+"-member@example.test",
		"IncidentMembership Member "+slug,
		"IncidentMembershipMember1!",
		false,
		false,
		true,
	)

	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-" + slug + "-incident",
		"incident_key":  "IR-" + strings.ToUpper(slug),
		"title":         "IncidentMembership " + slug,
	})
	incidentID := incident["incident_id"].(string)
	member := scenariotest.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{
		"client_txn_id": "txn-" + slug + "-member-bootstrap",
		"user_id":       memberID,
		"role":          "viewer",
	})
	primaryRow := timelineroutetest.CreateRow(t, harness.Server, adminLogin, incidentID, map[string]any{
		"client_txn_id":                   "txn-" + slug + "-primary-row",
		"timeline.activity_synopsis_text": "Primary row " + slug,
	})
	replacementRow := timelineroutetest.CreateRow(t, harness.Server, adminLogin, incidentID, map[string]any{
		"client_txn_id":                   "txn-" + slug + "-replacement-row",
		"timeline.activity_synopsis_text": "Replacement row " + slug,
	})

	return &RouteFixture{
		harness:     harness,
		adminLogin:  adminLogin,
		adminID:     adminID,
		candidateID: candidateID,
		memberID:    memberID,
		fixture: routeinventory.Fixture{
			IncidentID:            incidentID,
			AdminUserID:           adminID,
			CandidateUserID:       candidateID,
			MemberUserID:          memberID,
			PrimaryRecordID:       primaryRow["row"].(map[string]any)["record_id"].(string),
			ReplacementRecordID:   replacementRow["row"].(map[string]any)["record_id"].(string),
			BaseIncidentVersion:   ValueAsInt64(t, incident["incident_version"]),
			BaseRecordVersion:     ValueAsInt64(t, primaryRow["row"].(map[string]any)["row_version"]),
			BaseMembershipVersion: ValueAsInt64(t, member["membership_version"]),
		},
	}
}

func (f *RouteFixture) routeFixture(clientTxnSuffix string) routeinventory.Fixture {
	fixture := f.fixture
	fixture.ClientTxnSuffix = FixtureSlug(clientTxnSuffix)
	return fixture
}

func (f *RouteFixture) resetRecordTargets(t testing.TB, suffix string) {
	t.Helper()

	slug := FixtureSlug(suffix)
	primaryRow := timelineroutetest.CreateRow(t, f.harness.Server, f.adminLogin, f.fixture.IncidentID, map[string]any{
		"client_txn_id":                   "txn-" + slug + "-primary-row",
		"timeline.activity_synopsis_text": "Primary row " + slug,
	})
	replacementRow := timelineroutetest.CreateRow(t, f.harness.Server, f.adminLogin, f.fixture.IncidentID, map[string]any{
		"client_txn_id":                   "txn-" + slug + "-replacement-row",
		"timeline.activity_synopsis_text": "Replacement row " + slug,
	})
	f.fixture.PrimaryRecordID = primaryRow["row"].(map[string]any)["record_id"].(string)
	f.fixture.ReplacementRecordID = replacementRow["row"].(map[string]any)["record_id"].(string)
	f.fixture.BaseRecordVersion = ValueAsInt64(t, primaryRow["row"].(map[string]any)["row_version"])
}

func executeRouteRequest(
	t testing.TB,
	serverURL string,
	route routeinventory.Entry,
	fixture routeinventory.Fixture,
	sessionCookie *http.Cookie,
	csrfCookie *http.Cookie,
) *http.Response {
	t.Helper()

	options := []func(*http.Request){httptestx.WithCookies(sessionCookie)}
	if route.RequiresCSRF {
		options = append(options, httptestx.WithCookies(csrfCookie))
		options = append(options, httptestx.WithHeader(authn.CSRFHeaderName, csrfCookie.Value))
	}

	var body any
	if route.Body != nil {
		body = route.Body(fixture)
	}
	return httptestx.DoJSON(
		t,
		route.Method,
		serverURL+routeinventory.BuildPath(route.Template, fixture),
		body,
		options...,
	)
}

func requireRouteSuccess(t testing.TB, resp *http.Response, route routeinventory.Entry) map[string]any {
	t.Helper()

	if route.SuccessEnvelope {
		return httptestx.RequireSuccessEnvelope(t, resp, route.SuccessStatus)["data"].(map[string]any)
	}
	httptestx.RequireStatus(t, resp, route.SuccessStatus)
	return nil
}

func requireControlRouteOutcome(
	t testing.TB,
	serverURL string,
	route routeinventory.Entry,
	fixture routeinventory.Fixture,
	sessionCookie *http.Cookie,
	csrfCookie *http.Cookie,
	stage controlProgressionStage,
) map[string]any {
	t.Helper()

	expectation := controlExpectation(route, stage)
	if route.Transport == routeinventory.TransportWebSocket {
		if expectation.success {
			client := incidentwstest.ConnectAndHello(t, serverURL, fixture.IncidentID, incidentwstest.ConnectOptions{SessionToken: sessionCookie.Value})
			client.Close(websocket.StatusNormalClosure, "incident_membership_control_boundary_cleanup")
			return nil
		}
		incidentwstest.RequireDialErrorEnvelope(
			t,
			serverURL,
			fixture.IncidentID,
			incidentwstest.ConnectOptions{SessionToken: sessionCookie.Value},
			expectation.status,
			expectation.code,
		)
		return nil
	}

	resp := executeRouteRequest(t, serverURL, route, fixture, sessionCookie, csrfCookie)
	if expectation.success {
		return requireRouteSuccess(t, resp, route)
	}
	httptestx.RequireErrorEnvelope(t, resp, expectation.status, expectation.code)
	return nil
}

func controlExpectation(route routeinventory.Entry, stage controlProgressionStage) controlRouteExpectation {
	switch stage {
	case controlStageNoMembership, controlStageRemoved:
		return controlRouteExpectation{
			status: http.StatusNotFound,
			code:   "incident_not_found",
		}
	case controlStageViewer:
		if route.AllowedRole == routeinventory.ControlRoleMembershipRequired {
			return controlRouteExpectation{success: true}
		}
		return controlRouteExpectation{
			status: http.StatusForbidden,
			code:   "authorization_denied",
		}
	case controlStageReviewer:
		if route.AllowedRole != routeinventory.ControlRoleAdminOnly {
			return controlRouteExpectation{success: true}
		}
		return controlRouteExpectation{
			status: http.StatusForbidden,
			code:   "authorization_denied",
		}
	case controlStageAdmin:
		return controlRouteExpectation{success: true}
	default:
		return controlRouteExpectation{
			status: http.StatusInternalServerError,
			code:   "unexpected_control_stage",
		}
	}
}

func updateRouteFixtureAfterSuccess(
	t testing.TB,
	fixtureCtx *RouteFixture,
	route routeinventory.Entry,
	data map[string]any,
	stage controlProgressionStage,
) {
	t.Helper()

	if data == nil {
		return
	}

	switch route.Name {
	case "incident patch":
		fixtureCtx.fixture.BaseIncidentVersion = ValueAsInt64(t, data["incident_version"])
	case "record patch":
		row, ok := data["row"].(map[string]any)
		if !ok {
			t.Fatalf("unexpected record patch payload: %#v", data)
		}
		fixtureCtx.fixture.BaseRecordVersion = ValueAsInt64(t, row["row_version"])
	case "mark reviewed", "supersede":
		if stage == controlStageReviewer {
			fixtureCtx.resetRecordTargets(t, route.Name+"-admin-reset")
		}
	}
}

func ensureUserWorkbookPreferences(
	t testing.TB,
	fixtureCtx *RouteFixture,
	userID string,
) {
	t.Helper()

	if _, err := fixtureCtx.harness.DB.ExecContext(
		context.Background(),
		`
INSERT INTO user_workbook_preferences (incident_id, user_id, home_sheet_ref, created_at, updated_at)
VALUES ($1::uuid, $2::uuid, NULL, NOW(), NOW())
ON CONFLICT (incident_id, user_id) DO NOTHING
`,
		fixtureCtx.fixture.IncidentID,
		userID,
	); err != nil {
		t.Fatalf("ensure user workbook preferences: %v", err)
	}
}

func FixtureSlug(value string) string {
	replacer := strings.NewReplacer(
		" ", "-",
		"/", "-",
		"_", "-",
		".", "-",
		"(", "",
		")", "",
	)
	slug := strings.ToLower(replacer.Replace(value))
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return "incident_membership"
	}
	return slug
}

func ValueAsInt64(t testing.TB, value any) int64 {
	t.Helper()

	switch typed := value.(type) {
	case float64:
		return int64(typed)
	case int64:
		return typed
	case int:
		return int64(typed)
	default:
		t.Fatalf("unexpected numeric payload type: %T", value)
		return 0
	}
}
