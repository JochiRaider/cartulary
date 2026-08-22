package entities_test

import (
	"database/sql"
	"net/http"
	"slices"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/testsupport/incidentwstest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/contractassert"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	workbookscenariotest "github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

func TestSurfaceEnvelope(t *testing.T) {
	suite := newSupportSuite(t, "surface-envelope")
	for _, route := range workbookscenariotest.RoutesForHarness(
		t,
		workbookscenariotest.WorkbookRouteInventory(workbookscenariotest.RouteInventoryContext{}),
		workbookscenariotest.RouteHarnessSurfaceEnvelope,
	) {
		t.Run(string(route.Key), func(t *testing.T) {
			scenario := suite.newScenario(t, route)
			data := scenario.requireRouteSuccess(t, route, supportTxn("surface", route.Key), nil)
			assertRouteSuccessShape(t, route, scenario, data)
		})
	}
}

func TestCSRFProtection(t *testing.T) {
	suite := newSupportSuite(t, "csrf")
	for _, route := range workbookscenariotest.RoutesForHarness(
		t,
		workbookscenariotest.WorkbookRouteInventory(workbookscenariotest.RouteInventoryContext{}),
		workbookscenariotest.RouteHarnessCSRF,
	) {
		t.Run(string(route.Key), func(t *testing.T) {
			scenario := suite.newScenario(t, route)
			resp := scenario.doRoute(t, route, supportTxn("csrf", route.Key), nil, false)
			appsupport.RequireErrorBody(t, resp, http.StatusForbidden, "csrf_verification_failed")
		})
	}
}

func TestReplayAndDivergentConflict(t *testing.T) {
	suite := newSupportSuite(t, "replay")
	for _, route := range workbookscenariotest.RoutesForHarness(
		t,
		workbookscenariotest.WorkbookRouteInventory(workbookscenariotest.RouteInventoryContext{}),
		workbookscenariotest.RouteHarnessReplayDivergent,
	) {
		t.Run(string(route.Key), func(t *testing.T) {
			scenario := suite.newScenario(t, route)
			clientTxnID := supportTxn("replay", route.Key)

			firstData := scenario.requireRouteSuccess(t, route, clientTxnID, nil)
			replayResp := scenario.doRoute(t, route, clientTxnID, nil, true)
			replayData := scenario.requireRouteSuccessStatus(t, route, replayResp, route.ReplayStatus)
			requireStableReplayPayload(t, route, firstData, replayData)

			divergentResp := scenario.doRoute(t, route, clientTxnID, route.BuildDivergentBody(scenario.routeCtx, clientTxnID), true)
			divergentBody := appsupport.RequireErrorBody(t, divergentResp, route.DivergentStatus, route.DivergentCode)
			contractassert.RequireDivergentReplayRejected(
				t,
				divergentResp.StatusCode,
				divergentBody["error"].(map[string]any)["code"].(string),
				route.DivergentCode,
			)
		})
	}
}

func TestAuthorizationReDerivation(t *testing.T) {
	suite := newSupportSuite(t, "authorization")
	for _, route := range workbookscenariotest.RoutesForHarness(
		t,
		workbookscenariotest.WorkbookRouteInventory(workbookscenariotest.RouteInventoryContext{}),
		workbookscenariotest.RouteHarnessAuthorization,
	) {
		t.Run(string(route.Key), func(t *testing.T) {
			scenario := suite.newScenario(t, route)
			scenario.applyAuthorizationChange(t, route)

			resp := scenario.doRoute(t, route, supportTxn("authorization", route.Key), nil, true)
			body := appsupport.RequireErrorBody(t, resp, route.AuthorizationStatus, route.AuthorizationCode)
			contractassert.RequireAuthorizationReDerived(
				t,
				contractassert.AuthorizationOutcome{Status: route.SuccessStatus},
				contractassert.AuthorizationOutcome{Status: resp.StatusCode, Code: body["error"].(map[string]any)["code"].(string)},
			)
		})
	}
}

func TestDefaultQueryMetaAndFieldKeyConformance(t *testing.T) {
	suite := newSupportSuite(t, "query-matrix")
	for _, route := range workbookscenariotest.RoutesForHarness(
		t,
		workbookscenariotest.WorkbookRouteInventory(workbookscenariotest.RouteInventoryContext{}),
		workbookscenariotest.RouteHarnessQueryFieldMatrix,
	) {
		t.Run(string(route.Key), func(t *testing.T) {
			scenario := suite.newScenario(t, route)
			data := scenario.requireRouteSuccess(t, route, supportTxn("query", route.Key), nil)
			recordID := requireAffectedRecordID(t, route, scenario, data)

			envelope, row := scenario.queryAffectedRow(t, route, recordID)
			contractassert.RequireDefaultQueryMeta(t, envelope, route.ExpectedViewSchemaID)
			contractassert.RequireFieldKeyConformance(
				t,
				workbookscenariotest.SortedRowFieldKeys(t, row),
				workbookscenariotest.AllowedFieldKeys(t, string(route.Key), route.ExpectedViewSchemaID),
			)
		})
	}
}

func TestProjectionAndWebsocketConsequences(t *testing.T) {
	suite := newSupportSuite(t, "effects")
	for _, route := range workbookscenariotest.RoutesForHarness(
		t,
		workbookscenariotest.WorkbookRouteInventory(workbookscenariotest.RouteInventoryContext{}),
		workbookscenariotest.RouteHarnessEffects,
	) {
		t.Run(string(route.Key), func(t *testing.T) {
			scenario := suite.newScenario(t, route)
			var wsClient *incidentwstest.Client
			if route.WebSocketExpectation == workbookscenariotest.RouteWebSocketRecordChanged {
				wsClient = incidentwstest.ConnectViewSocket(
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
				socketChange := incidentwstest.RequireRecordChanged(
					t,
					wsClient,
					route.BuildWebSocketRecordID(scenario.routeCtx),
					route.WebSocketRowVersion,
				)
				requireRouteSocketChange(t, route.Key, data, socketChange, route.WebSocketViewSchemaID, nil)
				for _, expectation := range route.AdditionalWebSocketChanges {
					additionalChange := incidentwstest.RequireRecordChanged(
						t,
						wsClient,
						expectation.BuildRecordID(scenario.routeCtx),
						expectation.RowVersion,
					)
					requireRouteSocketChange(t, route.Key, data, additionalChange, expectation.ViewSchemaID, expectation.ChangedKeys)
				}
				incidentwstest.ExpectNoSocketMessage(t, wsClient)
			}

			recordID := requireAffectedRecordID(t, route, scenario, data)
			_, rowBefore := scenario.queryAffectedRow(t, route, recordID)
			scenario.rebuildProjection(t, route)
			_, rowAfter := scenario.queryAffectedRow(t, route, recordID)
			contractassert.RequireProjectionDeterminism(t, rowBefore["cells"], rowAfter["cells"])
		})
	}
}

func requireRouteSocketChange(t testing.TB, routeKey workbookscenariotest.RouteKey, responseData map[string]any, socketChange incidentwstest.RecordChangeSocketPayload, viewSchemaID string, changedKeys []string) {
	t.Helper()
	if changeSetID, ok := responseData["change_set_id"].(string); ok && changeSetID != "" && socketChange.ChangeSetID != changeSetID {
		t.Fatalf("expected websocket change_set_id to match route response for %s: payload=%#v response=%#v", routeKey, socketChange, responseData)
	}
	if viewSchemaID != "" && !socketChangeIncludesView(socketChange, viewSchemaID) {
		t.Fatalf("expected websocket affected view %s for %s, got %#v", viewSchemaID, routeKey, socketChange)
	}
	for _, key := range changedKeys {
		if !slices.Contains(socketChange.ChangedFieldKeys, key) {
			t.Fatalf("expected websocket changed key %s for %s, got %#v", key, routeKey, socketChange)
		}
	}
}

func socketChangeIncludesView(socketChange incidentwstest.RecordChangeSocketPayload, viewSchemaID string) bool {
	for _, view := range socketChange.AffectedViews {
		if view.ViewSchemaID == viewSchemaID {
			return true
		}
	}
	return false
}

func TestRecordEnvelopeHeadSchema(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "entity_linking-records-head")

	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open head schema database: %v", err)
	}
	defer db.Close()

	requireColumns(t, db, "records", "record_id", "incident_id", "record_type", "created_by_user_id", "created_at", "updated_by_user_id", "updated_at", "row_version", "deleted_at", "deleted_by_user_id")
	requireNoColumns(t, db, "record_id", "users", "user_sessions", "bootstrap_tokens", "pending_totp_enrollments", "incident_memberships", "deployment_bootstrap_state")
	requireGenericRecordFKsDoNotTargetTimeline(t, db)
}
