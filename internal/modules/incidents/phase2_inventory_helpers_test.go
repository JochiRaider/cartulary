package incidents_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/coder/websocket"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/phase2test"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
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

type phase2RouteFixture struct {
	harness     *phase2test.ServerHarness
	adminLogin  phase2test.LoginResult
	adminID     string
	candidateID string
	memberID    string
	fixture     phase2test.RouteInventoryFixture
}

func newPhase2RouteFixture(t testing.TB, prefix string) *phase2RouteFixture {
	t.Helper()

	slug := phase2FixtureSlug(prefix)
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, slug)
	adminLogin, adminID := phase2test.ProvisionBootstrapAdmin(t, harness.Server)

	candidateID := phase2test.SeedLocalUserFlags(
		t,
		harness.DB,
		slug+"-candidate@example.test",
		"Phase2 Candidate "+slug,
		"Phase2Candidate1!",
		false,
		false,
		true,
	)
	memberID := phase2test.SeedLocalUserFlags(
		t,
		harness.DB,
		slug+"-member@example.test",
		"Phase2 Member "+slug,
		"Phase2Member1!",
		false,
		false,
		true,
	)

	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-" + slug + "-incident",
		"incident_key":  "IR-" + strings.ToUpper(slug),
		"title":         "Phase2 " + slug,
	})
	incidentID := incident["incident_id"].(string)
	member := phase2test.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{
		"client_txn_id": "txn-" + slug + "-member-bootstrap",
		"user_id":       memberID,
		"role":          "viewer",
	})
	primaryRow := phase2test.CreateTimelineRow(t, harness.Server, adminLogin, incidentID, map[string]any{
		"client_txn_id":                   "txn-" + slug + "-primary-row",
		"timeline.activity_synopsis_text": "Primary row " + slug,
	})
	replacementRow := phase2test.CreateTimelineRow(t, harness.Server, adminLogin, incidentID, map[string]any{
		"client_txn_id":                   "txn-" + slug + "-replacement-row",
		"timeline.activity_synopsis_text": "Replacement row " + slug,
	})

	return &phase2RouteFixture{
		harness:     harness,
		adminLogin:  adminLogin,
		adminID:     adminID,
		candidateID: candidateID,
		memberID:    memberID,
		fixture: phase2test.RouteInventoryFixture{
			IncidentID:            incidentID,
			AdminUserID:           adminID,
			CandidateUserID:       candidateID,
			MemberUserID:          memberID,
			PrimaryRecordID:       primaryRow["row"].(map[string]any)["record_id"].(string),
			ReplacementRecordID:   replacementRow["row"].(map[string]any)["record_id"].(string),
			BaseIncidentVersion:   phase2ValueAsInt64(t, incident["incident_version"]),
			BaseRecordVersion:     phase2ValueAsInt64(t, primaryRow["row"].(map[string]any)["row_version"]),
			BaseMembershipVersion: phase2ValueAsInt64(t, member["membership_version"]),
		},
	}
}

func (f *phase2RouteFixture) routeFixture(clientTxnSuffix string) phase2test.RouteInventoryFixture {
	fixture := f.fixture
	fixture.ClientTxnSuffix = phase2FixtureSlug(clientTxnSuffix)
	return fixture
}

func (f *phase2RouteFixture) resetRecordTargets(t testing.TB, suffix string) {
	t.Helper()

	slug := phase2FixtureSlug(suffix)
	primaryRow := phase2test.CreateTimelineRow(t, f.harness.Server, f.adminLogin, f.fixture.IncidentID, map[string]any{
		"client_txn_id":                   "txn-" + slug + "-primary-row",
		"timeline.activity_synopsis_text": "Primary row " + slug,
	})
	replacementRow := phase2test.CreateTimelineRow(t, f.harness.Server, f.adminLogin, f.fixture.IncidentID, map[string]any{
		"client_txn_id":                   "txn-" + slug + "-replacement-row",
		"timeline.activity_synopsis_text": "Replacement row " + slug,
	})
	f.fixture.PrimaryRecordID = primaryRow["row"].(map[string]any)["record_id"].(string)
	f.fixture.ReplacementRecordID = replacementRow["row"].(map[string]any)["record_id"].(string)
	f.fixture.BaseRecordVersion = phase2ValueAsInt64(t, primaryRow["row"].(map[string]any)["row_version"])
}

func executeRouteRequest(
	t testing.TB,
	serverURL string,
	route phase2test.RouteInventoryEntry,
	fixture phase2test.RouteInventoryFixture,
	sessionCookie *http.Cookie,
	csrfCookie *http.Cookie,
) *http.Response {
	t.Helper()

	options := []func(*http.Request){phase2test.WithCookies(sessionCookie)}
	if route.RequiresCSRF {
		options = append(options, phase2test.WithCookies(csrfCookie))
		options = append(options, phase2test.WithHeader(authn.CSRFHeaderName, csrfCookie.Value))
	}

	var body any
	if route.Body != nil {
		body = route.Body(fixture)
	}
	return phase2test.DoJSON(
		t,
		route.Method,
		serverURL+phase2test.BuildRoutePath(route.Template, fixture),
		body,
		options...,
	)
}

func requireRouteSuccess(t testing.TB, resp *http.Response, route phase2test.RouteInventoryEntry) map[string]any {
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
	route phase2test.RouteInventoryEntry,
	fixture phase2test.RouteInventoryFixture,
	sessionCookie *http.Cookie,
	csrfCookie *http.Cookie,
	stage controlProgressionStage,
) map[string]any {
	t.Helper()

	expectation := controlExpectation(route, stage)
	if route.Transport == phase2test.RouteTransportWebSocket {
		if expectation.success {
			client := phase2test.ConnectTimelineSocket(t, serverURL, fixture.IncidentID, sessionCookie.Value)
			client.Close(websocket.StatusNormalClosure, "phase2_control_boundary_cleanup")
			return nil
		}
		phase2test.RequireTimelineSocketRejected(
			t,
			serverURL,
			fixture.IncidentID,
			sessionCookie.Value,
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

func controlExpectation(route phase2test.RouteInventoryEntry, stage controlProgressionStage) controlRouteExpectation {
	switch stage {
	case controlStageNoMembership, controlStageRemoved:
		return controlRouteExpectation{
			status: http.StatusNotFound,
			code:   "incident_not_found",
		}
	case controlStageViewer:
		if route.AllowedRole == phase2test.ControlRoleMembershipRequired {
			return controlRouteExpectation{success: true}
		}
		return controlRouteExpectation{
			status: http.StatusForbidden,
			code:   "authorization_denied",
		}
	case controlStageReviewer:
		if route.AllowedRole != phase2test.ControlRoleAdminOnly {
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
	fixtureCtx *phase2RouteFixture,
	route phase2test.RouteInventoryEntry,
	data map[string]any,
	stage controlProgressionStage,
) {
	t.Helper()

	if data == nil {
		return
	}

	switch route.Name {
	case "incident patch":
		fixtureCtx.fixture.BaseIncidentVersion = phase2ValueAsInt64(t, data["incident_version"])
	case "record patch":
		row, ok := data["row"].(map[string]any)
		if !ok {
			t.Fatalf("unexpected record patch payload: %#v", data)
		}
		fixtureCtx.fixture.BaseRecordVersion = phase2ValueAsInt64(t, row["row_version"])
	case "mark reviewed", "supersede":
		if stage == controlStageReviewer {
			fixtureCtx.resetRecordTargets(t, route.Name+"-admin-reset")
		}
	}
}

func ensureUserWorkbookPreferences(
	t testing.TB,
	fixtureCtx *phase2RouteFixture,
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

func phase2FixtureSlug(value string) string {
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
		return "phase2"
	}
	return slug
}

func phase2ValueAsInt64(t testing.TB, value any) int64 {
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
