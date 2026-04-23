package entities_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/golden"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
	"github.com/JochiRaider/cartulary/internal/testutil/wstest"
)

func TestSupportPhase4Integration_SurfaceEnvelope(t *testing.T) {
	for _, route := range phase4test.RoutesForHarness(
		t,
		phase4test.Phase4RouteInventory(phase4test.RouteInventoryContext{}),
		phase4test.RouteHarnessSurfaceEnvelope,
	) {
		t.Run(string(route.Key), func(t *testing.T) {
			scenario := newPhase4SupportScenario(t, "surface-envelope", route)
			data := scenario.requireRouteSuccess(t, route, supportTxn("surface", route.Key), nil)
			assertRouteSuccessShape(t, route, scenario, data)
		})
	}
}

func TestSupportPhase4Integration_CSRFProtection(t *testing.T) {
	for _, route := range phase4test.RoutesForHarness(
		t,
		phase4test.Phase4RouteInventory(phase4test.RouteInventoryContext{}),
		phase4test.RouteHarnessCSRF,
	) {
		t.Run(string(route.Key), func(t *testing.T) {
			scenario := newPhase4SupportScenario(t, "csrf", route)
			resp := scenario.doRoute(t, route, supportTxn("csrf", route.Key), nil, false)
			phase4test.RequireErrorBody(t, resp, http.StatusForbidden, "csrf_verification_failed")
		})
	}
}

func TestSupportPhase4Integration_ReplayAndDivergentConflict(t *testing.T) {
	for _, route := range phase4test.RoutesForHarness(
		t,
		phase4test.Phase4RouteInventory(phase4test.RouteInventoryContext{}),
		phase4test.RouteHarnessReplayDivergent,
	) {
		t.Run(string(route.Key), func(t *testing.T) {
			scenario := newPhase4SupportScenario(t, "replay", route)
			clientTxnID := supportTxn("replay", route.Key)

			firstData := scenario.requireRouteSuccess(t, route, clientTxnID, nil)
			replayResp := scenario.doRoute(t, route, clientTxnID, nil, true)
			replayData := scenario.requireRouteSuccessStatus(t, route, replayResp, route.ReplayStatus)
			requireStableReplayPayload(t, route, firstData, replayData)

			divergentResp := scenario.doRoute(t, route, clientTxnID, route.BuildDivergentBody(scenario.routeCtx, clientTxnID), true)
			divergentBody := phase4test.RequireErrorBody(t, divergentResp, route.DivergentStatus, route.DivergentCode)
			httptestx.RequireDivergentReplayRejected(
				t,
				divergentResp.StatusCode,
				divergentBody["error"].(map[string]any)["code"].(string),
				route.DivergentCode,
			)
		})
	}
}

func TestSupportPhase4Integration_AuthorizationReDerivation(t *testing.T) {
	for _, route := range phase4test.RoutesForHarness(
		t,
		phase4test.Phase4RouteInventory(phase4test.RouteInventoryContext{}),
		phase4test.RouteHarnessAuthorization,
	) {
		t.Run(string(route.Key), func(t *testing.T) {
			scenario := newPhase4SupportScenario(t, "authorization", route)
			scenario.applyAuthorizationChange(t, route)

			resp := scenario.doRoute(t, route, supportTxn("authorization", route.Key), nil, true)
			body := phase4test.RequireErrorBody(t, resp, route.AuthorizationStatus, route.AuthorizationCode)
			httptestx.RequireAuthorizationReDerived(
				t,
				httptestx.AuthorizationOutcome{Status: route.SuccessStatus},
				httptestx.AuthorizationOutcome{Status: resp.StatusCode, Code: body["error"].(map[string]any)["code"].(string)},
			)
		})
	}
}

func TestSupportPhase4Integration_DefaultQueryMetaAndFieldKeyConformance(t *testing.T) {
	for _, route := range phase4test.RoutesForHarness(
		t,
		phase4test.Phase4RouteInventory(phase4test.RouteInventoryContext{}),
		phase4test.RouteHarnessQueryFieldMatrix,
	) {
		t.Run(string(route.Key), func(t *testing.T) {
			scenario := newPhase4SupportScenario(t, "query-matrix", route)
			data := scenario.requireRouteSuccess(t, route, supportTxn("query", route.Key), nil)
			recordID := requireAffectedRecordID(t, route, scenario, data)

			envelope, row := scenario.queryAffectedRow(t, route, recordID)
			httptestx.RequireDefaultQueryMeta(t, envelope, route.ExpectedViewSchemaID)
			httptestx.RequireFieldKeyConformance(
				t,
				phase4test.SortedRowFieldKeys(t, row),
				phase4test.AllowedFieldKeys(t, string(route.Key), route.ExpectedViewSchemaID),
			)
		})
	}
}

func TestSupportPhase4Integration_ProjectionAndWebsocketConsequences(t *testing.T) {
	for _, route := range phase4test.RoutesForHarness(
		t,
		phase4test.Phase4RouteInventory(phase4test.RouteInventoryContext{}),
		phase4test.RouteHarnessEffects,
	) {
		t.Run(string(route.Key), func(t *testing.T) {
			scenario := newPhase4SupportScenario(t, "effects", route)
			var wsClient *wstest.Client
			if route.WebSocketExpectation == phase4test.RouteWebSocketRecordChanged {
				wsClient = phase4test.ConnectViewSocket(
					t,
					scenario.harness.Server,
					scenario.IncidentID.String(),
					route.WebSocketViewSchemaID,
					scenario.actorLogin.SessionCookie.Value,
				)
				defer wsClient.Close(1000, "test_complete")
			}

			data := scenario.requireRouteSuccess(t, route, supportTxn("effects", route.Key), nil)

			if wsClient != nil {
				socketChange := phase4test.RequireRecordChanged(
					t,
					wsClient,
					route.BuildWebSocketRecordID(scenario.routeCtx),
					route.WebSocketRowVersion,
				)
				if changeSetID, ok := data["change_set_id"].(string); ok && changeSetID != "" && socketChange.ChangeSetID != changeSetID {
					t.Fatalf("expected websocket change_set_id to match route response for %s: payload=%#v response=%#v", route.Key, socketChange, data)
				}
				phase4test.ExpectNoSocketMessage(t, wsClient)
			}

			recordID := requireAffectedRecordID(t, route, scenario, data)
			_, rowBefore := scenario.queryAffectedRow(t, route, recordID)
			scenario.rebuildProjection(t, route)
			_, rowAfter := scenario.queryAffectedRow(t, route, recordID)
			httptestx.RequireProjectionDeterminism(t, rowBefore["cells"], rowAfter["cells"])
		})
	}
}

type phase4SupportScenario struct {
	harness         *phase4test.ServerHarness
	bootstrapUserID uuid.UUID
	actorLogin      phase4test.LoginResult
	actorUserID     uuid.UUID
	IncidentID      uuid.UUID
	routeCtx        phase4test.RouteInventoryContext
}

func newPhase4SupportScenario(t *testing.T, label string, route phase4test.RouteInventoryEntry) *phase4SupportScenario {
	t.Helper()

	harness := phase4test.StartServer(t, "phase4-support-"+label+"-"+string(route.Key))
	bootstrapLogin, bootstrapUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, bootstrapLogin, map[string]any{
		"client_txn_id": supportTxn("incident", route.Key),
		"incident_key":  "IR-P4-" + strings.ToUpper(strings.ReplaceAll(string(route.Key), "_", "-")),
		"title":         "Phase 4 support matrix " + string(route.Key),
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	const actorPassword = "SupportAdminPass1!"
	actorRecord := phase4test.SeedLocalUserFlags(
		t,
		harness.DB,
		"phase4-support-"+string(route.Key)+"@example.test",
		"Phase 4 Support Admin",
		actorPassword,
		false,
		false,
		true,
	)
	phase4test.SeedIncidentMembership(t, harness.DB, incidentID, actorRecord.ID, actorRecord.DisplayName, "admin", bootstrapUserID)
	actorLogin := loginLocalSupportUser(t, harness, actorRecord.Email, actorPassword)

	scenario := &phase4SupportScenario{
		harness:         harness,
		bootstrapUserID: bootstrapUserID,
		actorLogin:      actorLogin,
		actorUserID:     actorRecord.ID,
		IncidentID:      incidentID,
		routeCtx: phase4test.RouteInventoryContext{
			IncidentID:            incidentID.String(),
			TimelineRecordID:      golden.Phase4TimelineRecordID.String(),
			MentionID:             golden.Phase4HostMentionID.String(),
			MergeSurvivorRecordID: golden.Phase4CanonicalHostRecordID.String(),
			MergeLoserRecordID:    golden.Phase4DuplicateHostRecordID.String(),
			HostRecordID:          golden.Phase4CanonicalHostRecordID.String(),
			IdentityRecordID:      golden.Phase4CanonicalIdentityID.String(),
		},
	}

	scenario.seedBaseData(t, route)
	return scenario
}

func (s *phase4SupportScenario) seedBaseData(t *testing.T, route phase4test.RouteInventoryEntry) {
	t.Helper()

	phase4test.SeedTimelineRecord(t, s.harness.DB, s.IncidentID, s.actorUserID, golden.Phase4TimelineRecordID)
	phase4test.SeedHostRecord(t, s.harness.DB, s.IncidentID, s.actorUserID, golden.Phase4CanonicalHostRecordID, "WS-023", "WS-023", "", "")
	phase4test.SeedHostRecord(t, s.harness.DB, s.IncidentID, s.actorUserID, golden.Phase4DuplicateHostRecordID, "WS-024", "WS-024", "ws-024.corp.example.test", "")
	phase4test.SeedIdentityRecord(t, s.harness.DB, s.IncidentID, s.actorUserID, golden.Phase4CanonicalIdentityID, "Alex Analyst", "alex.analyst@example.test", "alex.analyst@example.test", "ALEXA")
	phase4test.SeedIdentityRecord(t, s.harness.DB, s.IncidentID, s.actorUserID, golden.Phase4DuplicateIdentityID, "Legacy Analyst", "legacy.analyst@example.test", "legacy.analyst@example.test", "LEGACYA")

	switch route.Key {
	case phase4test.RouteMentionResolve:
		phase4test.SeedMention(
			t,
			s.harness.DB,
			s.actorUserID,
			golden.Phase4HostMentionID,
			golden.Phase4TimelineRecordID,
			golden.Phase4FieldTimelineHostRefs,
			"host",
			"WS-023",
			"unresolved",
			nil,
			nil,
		)
	case phase4test.RouteExplicitMerge:
		phase4test.SeedResolvedMention(
			t,
			s.harness.DB,
			s.actorUserID,
			golden.Phase4HostMentionID,
			golden.Phase4TimelineRecordID,
			golden.Phase4DuplicateHostRecordID,
			golden.Phase4FieldTimelineHostRefs,
			"host",
			"WS-024",
		)
		phase4test.SeedRecordLink(
			t,
			s.harness.DB,
			s.IncidentID,
			s.actorUserID,
			golden.Phase4DuplicateLinkID,
			golden.Phase4TimelineRecordID,
			golden.Phase4DuplicateHostRecordID,
			"observed_on_host",
			"manual",
			nil,
		)
		phase4test.SeedRecordTag(t, s.harness.DB, s.IncidentID, s.actorUserID, golden.Phase4TagIDSurvivor, golden.Phase4CanonicalHostRecordID, "critical-host")
		phase4test.SeedRecordTag(t, s.harness.DB, s.IncidentID, s.actorUserID, golden.Phase4TagIDLoser, golden.Phase4DuplicateHostRecordID, "critical-host")
		phase4test.SeedAssessment(t, s.harness.DB, s.IncidentID, s.actorUserID, golden.Phase4AssessmentHostID, golden.Phase4DuplicateHostRecordID, "host", "confirmed")
	}

	s.rebuildBaseProjections(t)

	if route.Key == phase4test.RouteIndicatorsQuery {
		data := s.requireRouteSuccess(
			t,
			findRouteByKey(t, phase4test.RouteIndicatorsCreate),
			supportTxn("seed-indicator", route.Key),
			nil,
		)
		s.routeCtx.IndicatorRecordID = requireRowRecordID(t, data)
	}
}

func (s *phase4SupportScenario) rebuildBaseProjections(t *testing.T) {
	t.Helper()

	store := projections.NewStore(s.harness.Server.Runtime.Postgres)
	if err := store.RebuildIncidentTimeline(context.Background(), s.IncidentID); err != nil {
		t.Fatalf("rebuild timeline projections: %v", err)
	}
	if err := store.RebuildIncidentHosts(context.Background(), s.IncidentID); err != nil {
		t.Fatalf("rebuild host projections: %v", err)
	}
	if err := store.RebuildIncidentIdentities(context.Background(), s.IncidentID); err != nil {
		t.Fatalf("rebuild identity projections: %v", err)
	}
}

func (s *phase4SupportScenario) requireRouteSuccess(t *testing.T, route phase4test.RouteInventoryEntry, clientTxnID string, body any) map[string]any {
	t.Helper()

	resp := s.doRoute(t, route, clientTxnID, body, true)
	return s.requireRouteSuccessStatus(t, route, resp, route.SuccessStatus)
}

func (s *phase4SupportScenario) requireRouteSuccessStatus(t *testing.T, route phase4test.RouteInventoryEntry, resp *http.Response, wantStatus int) map[string]any {
	t.Helper()

	data := phase4test.RequireSuccessData(t, resp, wantStatus)
	assertRouteSuccessShape(t, route, s, data)
	return data
}

func (s *phase4SupportScenario) doRoute(
	t *testing.T,
	route phase4test.RouteInventoryEntry,
	clientTxnID string,
	body any,
	includeCSRFFHeader bool,
) *http.Response {
	t.Helper()

	requestBody := body
	if requestBody == nil {
		requestBody = route.BuildBody(s.routeCtx, clientTxnID)
	}

	options := []func(*http.Request){}
	if route.RequiresCSRF {
		options = append(options, phase4test.WithCookies(s.actorLogin.SessionCookie, s.actorLogin.CSRFCookie))
		if includeCSRFFHeader {
			options = append(options, phase4test.WithHeader(authn.CSRFHeaderName, s.actorLogin.CSRFCookie.Value))
		}
	} else {
		options = append(options, phase4test.WithCookies(s.actorLogin.SessionCookie))
	}

	return phase4test.DoJSON(
		t,
		route.Method,
		s.harness.Server.HTTP.URL+route.BuildPath(s.routeCtx),
		requestBody,
		options...,
	)
}

func (s *phase4SupportScenario) applyAuthorizationChange(t *testing.T, route phase4test.RouteInventoryEntry) {
	t.Helper()

	switch route.AuthorizationChange {
	case phase4test.RouteAuthorizationDemoteViewer:
		if _, err := s.harness.DB.ExecContext(
			context.Background(),
			`
UPDATE incident_memberships
   SET role = 'viewer',
       updated_at = now(),
       updated_by_user_id = $3
 WHERE incident_id = $1
   AND user_id = $2
`,
			s.IncidentID,
			s.actorUserID,
			s.bootstrapUserID,
		); err != nil {
			t.Fatalf("demote support actor membership: %v", err)
		}
	case phase4test.RouteAuthorizationRemoveMember:
		if _, err := s.harness.DB.ExecContext(
			context.Background(),
			`DELETE FROM incident_memberships WHERE incident_id = $1 AND user_id = $2`,
			s.IncidentID,
			s.actorUserID,
		); err != nil {
			t.Fatalf("remove support actor membership: %v", err)
		}
	case phase4test.RouteAuthorizationNotApplicable:
		t.Fatalf("route %s does not declare authorization change", route.Key)
	default:
		t.Fatalf("unsupported authorization change %s", route.AuthorizationChange)
	}
}

func (s *phase4SupportScenario) queryAffectedRow(t *testing.T, route phase4test.RouteInventoryEntry, recordID string) (map[string]any, map[string]any) {
	t.Helper()

	envelope := phase4test.QueryViewEnvelope(
		t,
		s.harness.Server.HTTP.URL,
		s.IncidentID.String(),
		route.ExpectedViewSchemaID,
		s.actorLogin,
	)
	rows := rowsFromQueryData(t, envelope["data"].(map[string]any))
	return envelope, phase4test.FindRow(t, rows, recordID)
}

func (s *phase4SupportScenario) rebuildProjection(t *testing.T, route phase4test.RouteInventoryEntry) {
	t.Helper()

	store := projections.NewStore(s.harness.Server.Runtime.Postgres)
	var err error
	switch route.ProjectionTarget {
	case phase4test.RouteProjectionHosts:
		err = store.RebuildIncidentHosts(context.Background(), s.IncidentID)
	case phase4test.RouteProjectionIdentities:
		err = store.RebuildIncidentIdentities(context.Background(), s.IncidentID)
	case phase4test.RouteProjectionIndicators:
		err = store.RebuildIncidentIndicators(context.Background(), s.IncidentID)
	case phase4test.RouteProjectionTimeline:
		err = store.RebuildIncidentTimeline(context.Background(), s.IncidentID)
	case phase4test.RouteProjectionNotApplicable:
		return
	default:
		t.Fatalf("unsupported projection target %s", route.ProjectionTarget)
	}
	if err != nil {
		t.Fatalf("rebuild %s projection: %v", route.ProjectionTarget, err)
	}
}

func assertRouteSuccessShape(t testing.TB, route phase4test.RouteInventoryEntry, scenario *phase4SupportScenario, data map[string]any) {
	t.Helper()

	switch route.SuccessShape {
	case phase4test.RouteSuccessShapeMentionResolution:
		if data["incident_id"] != scenario.IncidentID.String() {
			t.Fatalf("unexpected incident_id for %s: %#v", route.Key, data)
		}
		if _, ok := data["source_record"].(map[string]any); !ok {
			t.Fatalf("expected mention resolution source_record for %s, got %#v", route.Key, data)
		}
		requireNonEmptyString(t, data, "change_set_id")
	case phase4test.RouteSuccessShapeMerge:
		if data["survivor_record_id"] != scenario.routeCtx.MergeSurvivorRecordID {
			t.Fatalf("unexpected survivor_record_id for %s: %#v", route.Key, data)
		}
		if data["loser_record_id"] != scenario.routeCtx.MergeLoserRecordID {
			t.Fatalf("unexpected loser_record_id for %s: %#v", route.Key, data)
		}
		if _, ok := data["merge_summary"].(map[string]any); !ok {
			t.Fatalf("expected merge_summary for %s, got %#v", route.Key, data)
		}
		requireNonEmptyString(t, data, "change_set_id")
	case phase4test.RouteSuccessShapeMutationRow:
		requireNonEmptyString(t, data, "change_set_id")
		if requireRowRecordID(t, data) == "" {
			t.Fatalf("expected row record_id for %s, got %#v", route.Key, data)
		}
	case phase4test.RouteSuccessShapeQueryRows:
		rows := rowsFromQueryData(t, data)
		if len(rows) == 0 {
			t.Fatalf("expected non-empty query rows for %s, got %#v", route.Key, data)
		}
	default:
		t.Fatalf("unsupported success shape %s", route.SuccessShape)
	}
}

func requireStableReplayPayload(t testing.TB, route phase4test.RouteInventoryEntry, firstData map[string]any, replayData map[string]any) {
	t.Helper()

	firstChangeSet, firstHasChangeSet := firstData["change_set_id"].(string)
	replayChangeSet, replayHasChangeSet := replayData["change_set_id"].(string)
	if firstHasChangeSet && replayHasChangeSet && firstChangeSet != replayChangeSet {
		t.Fatalf("expected replay to preserve change_set_id for %s: first=%#v replay=%#v", route.Key, firstData, replayData)
	}
	if route.SuccessShape == phase4test.RouteSuccessShapeMutationRow {
		if requireRowRecordID(t, firstData) != requireRowRecordID(t, replayData) {
			t.Fatalf("expected replay to preserve row record_id for %s: first=%#v replay=%#v", route.Key, firstData, replayData)
		}
	}
}

func requireAffectedRecordID(t testing.TB, route phase4test.RouteInventoryEntry, scenario *phase4SupportScenario, data map[string]any) string {
	t.Helper()

	recordID := route.AffectedRecordID(scenario.routeCtx, data)
	if recordID == "" {
		t.Fatalf("route %s returned no affected record id for %#v", route.Key, data)
	}
	return recordID
}

func rowsFromQueryData(t testing.TB, data map[string]any) []map[string]any {
	t.Helper()

	rawRows, ok := data["rows"].([]any)
	if !ok {
		t.Fatalf("expected query rows array, got %#v", data)
	}
	rows := make([]map[string]any, 0, len(rawRows))
	for _, rawRow := range rawRows {
		row, ok := rawRow.(map[string]any)
		if !ok {
			t.Fatalf("expected query row object, got %#v", rawRow)
		}
		rows = append(rows, row)
	}
	return rows
}

func requireRowRecordID(t testing.TB, data map[string]any) string {
	t.Helper()

	row, ok := data["row"].(map[string]any)
	if !ok {
		t.Fatalf("expected row payload, got %#v", data)
	}
	recordID, ok := row["record_id"].(string)
	if !ok || recordID == "" {
		t.Fatalf("expected non-empty row.record_id, got %#v", row)
	}
	return recordID
}

func requireNonEmptyString(t testing.TB, payload map[string]any, key string) string {
	t.Helper()

	value, ok := payload[key].(string)
	if !ok || value == "" {
		t.Fatalf("expected non-empty %s in %#v", key, payload)
	}
	return value
}

func loginLocalSupportUser(t testing.TB, harness *phase4test.ServerHarness, username string, password string) phase4test.LoginResult {
	t.Helper()

	resp := phase4test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/auth/login",
		map[string]any{
			"username": username,
			"password": password,
		},
	)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("support login failed: status=%d body=%#v", resp.StatusCode, httptestx.ReadJSONBody(t, resp))
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
		t.Fatalf("expected support login to set session and csrf cookies, got %#v", resp.Cookies())
	}
	return phase4test.LoginResult{
		SessionCookie: sessionCookie,
		CSRFCookie:    csrfCookie,
	}
}

func findRouteByKey(t testing.TB, key phase4test.RouteKey) phase4test.RouteInventoryEntry {
	t.Helper()

	for _, route := range phase4test.Phase4RouteInventory(phase4test.RouteInventoryContext{}) {
		if route.Key == key {
			return route
		}
	}
	t.Fatalf("missing phase4 route inventory entry %s", key)
	return phase4test.RouteInventoryEntry{}
}

func supportTxn(label string, routeKey phase4test.RouteKey) string {
	return "txn-phase4-support-" + label + "-" + string(routeKey)
}
