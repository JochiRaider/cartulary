package scenariotest

import (
	"database/sql"
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/asserttest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/auditassert"
	"github.com/JochiRaider/cartulary/internal/testutil/contractassert"
)

type RouteConformanceCase struct {
	Route                  RouteInventoryEntry
	Context                RouteInventoryContext
	ClientTxnID            string
	Login                  appsupport.LoginResult
	ActorUserID            string
	ExpectedMutationSource string
}

func RequireRouteReplayHistoryConformance(t testing.TB, db *sql.DB, serverURL string, c RouteConformanceCase) map[string]any {
	t.Helper()
	if c.Route.ReplayCapability == RouteReplayNotApplicable {
		t.Fatalf("route %s does not declare replay conformance", c.Route.Key)
	}
	if c.Route.BuildDivergentBody == nil {
		t.Fatalf("route %s does not declare a divergent replay payload", c.Route.Key)
	}
	if c.ClientTxnID == "" {
		t.Fatalf("route %s conformance case missing client transaction id", c.Route.Key)
	}

	firstResp := DoRoute(t, serverURL, c.Route, c.Context, c.ClientTxnID, c.Login, c.Route.BuildBody(c.Context, c.ClientTxnID))
	firstData := appsupport.RequireSuccessData(t, firstResp, c.Route.SuccessStatus)
	if c.ExpectedMutationSource != "" {
		changeSetID, ok := firstData["change_set_id"].(string)
		if !ok || changeSetID == "" {
			t.Fatalf("route %s expected attributed mutation response with change_set_id, got %#v", c.Route.Key, firstData)
		}
		RequireChangeSetAttribution(t, db, changeSetID, c.ActorUserID, c.ExpectedMutationSource, c.ClientTxnID)
	}

	affectedRecordID := ""
	if c.Route.AffectedRecordID != nil {
		affectedRecordID = c.Route.AffectedRecordID(c.Context, firstData)
	}
	stableBefore := SnapshotReplayCounts(t, db, c.Context.IncidentID, affectedRecordID)
	replayResp := DoRoute(t, serverURL, c.Route, c.Context, c.ClientTxnID, c.Login, c.Route.BuildBody(c.Context, c.ClientTxnID))
	replayData := appsupport.RequireSuccessData(t, replayResp, c.Route.ReplayStatus)
	if firstChangeSet, ok := firstData["change_set_id"].(string); ok && firstChangeSet != "" {
		if replayData["change_set_id"] != firstChangeSet {
			t.Fatalf("route %s replay returned different change_set_id: first=%#v replay=%#v", c.Route.Key, firstData, replayData)
		}
	}
	stableAfter := SnapshotReplayCounts(t, db, c.Context.IncidentID, affectedRecordID)
	contractassert.RequireReplayScaffold(t, contractassert.ReplayExpectation{
		FirstStatus:     c.Route.SuccessStatus,
		ReplayStatus:    c.Route.ReplayStatus,
		DivergentStatus: c.Route.DivergentStatus,
		DivergentCode:   c.Route.DivergentCode,
		StableBefore:    stableBefore,
		StableAfter:     stableAfter,
	})

	divergentResp := DoRoute(t, serverURL, c.Route, c.Context, c.ClientTxnID, c.Login, c.Route.BuildDivergentBody(c.Context, c.ClientTxnID))
	divergentBody := appsupport.RequireErrorBody(
		t,
		divergentResp,
		c.Route.DivergentStatus,
		c.Route.DivergentCode,
	)
	contractassert.RequireDivergentReplayRejected(
		t,
		divergentResp.StatusCode,
		divergentBody["error"].(map[string]any)["code"].(string),
		c.Route.DivergentCode,
	)

	return firstData
}

func DoRoute(t testing.TB, serverURL string, route RouteInventoryEntry, ctx RouteInventoryContext, clientTxnID string, login appsupport.LoginResult, body any) *http.Response {
	t.Helper()
	options := []func(*http.Request){
		appsupport.WithCookies(login.SessionCookie, login.CSRFCookie),
	}
	if route.RequiresCSRF {
		options = append(
			options,
			appsupport.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
		)
	}
	return appsupport.DoJSON(
		t,
		route.Method,
		serverURL+route.BuildPath(ctx),
		body,
		options...,
	)
}

func SnapshotReplayCounts(t testing.TB, db *sql.DB, incidentID string, recordID string) contractassert.ReplayCounts {
	t.Helper()

	counts := contractassert.ReplayCounts{
		ChangeSets: appsupport.QueryCount(t, db, `SELECT COUNT(*) FROM change_sets WHERE incident_id::text = $1`, incidentID),
		MutationRows: appsupport.QueryCount(t, db, `
SELECT COUNT(*)
  FROM change_set_mutations m
  JOIN change_sets c ON c.change_set_id = m.change_set_id
 WHERE c.incident_id::text = $1
`, incidentID),
	}
	if recordID != "" {
		counts.Revisions = appsupport.QueryCount(t, db, `SELECT COUNT(*) FROM record_revisions WHERE record_id::text = $1`, recordID)
	}
	return counts
}

func RequireChangeSetAttribution(t testing.TB, db *sql.DB, changeSetID string, actorUserID string, source string, clientTxnID string) {
	t.Helper()
	changeSet := asserttest.LookupChangeSet(t, asserttest.SQLDatabase(db), changeSetID)
	auditassert.RequireMutationAttribution(t, auditassert.MutationAttribution{
		ActorUserID: changeSet.ActorUserID,
		Source:      changeSet.Source,
		ClientTxnID: changeSet.ClientTxnID,
		RequestID:   changeSet.RequestID,
		CreatedAt:   changeSet.CreatedAt,
	}, actorUserID, source, clientTxnID)
}

func MustRoute(t testing.TB, key RouteKey, ctx RouteInventoryContext) RouteInventoryEntry {
	t.Helper()
	for _, route := range WorkbookRouteInventory(ctx) {
		if route.Key == key {
			return route
		}
	}
	t.Fatalf("workbook route inventory missing %s", key)
	return RouteInventoryEntry{}
}
